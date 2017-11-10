/*Package gpp is a http proxy handler, it can act as a proxy server and a http server.

Use it as a normal http.Handler. it determines the proxy request and the local request automatically.

It handle the proxy request itself, and route the local request to http.DefaultServerMux.

you can set its Handler options to yourself handler.

you can set EnableProxy to false to disable proxy function.

Example

a proxy example
    package main

    import (
        . "fmt"
        "github.com/fangdingjun/gpp"
        "log"
        "net/http"
    )

    func main() {
        port := 8080

        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(200)
            w.Write([]byte("<h1>welcome!</h1>"))
        })

        log.Print("Listen on: ", Sprintf("0.0.0.0:%d", port))
        err := http.ListenAndServe(Sprintf(":%d", port), &gpp.Handler{EnableProxy:true})
        if err != nil {
            log.Fatal(err)
        }
    }
Run above example and use curl to test it.

Run the follow command you will see a welcome message
    $ curl http://127.0.0.1:8080/
Run the follow command to test proxy function
    $ curl --proxy 127.0.0.1:8080 http://httpbin.org/ip */
package gpp

import (
	"fmt"
	"github.com/fangdingjun/gpp/util"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

/*
Handler is proxy handler, you can use this as a http.Handler.
*/
type Handler struct {
	// the handler to process local path request
	Handler http.Handler

	// enable proxy or not
	EnableProxy bool

	// when enable http2 support
	// enable proxy on http/1.1 or not
	EnableProxyHTTP11 bool

	// the local domain name, only required when http2 enabled
	LocalDomains []string

	// the RoundTripper for http proxy
	Transport http.RoundTripper

	// log instance
	Logger *log.Logger

	// proxy require auth
	ProxyAuth bool

	/*
	   if ProxyAuth is true, ProxyAuthFunc used to check the user authorization,
	   return true if success, false if failed,

	   when failed, the response must be replyed to the client before the
	   function ProxyAuthFunc return
	*/
	ProxyAuthFunc func(w http.ResponseWriter, r *http.Request) bool
}

/*
Log a shortcut for log.Printf, if h.Logger is nil this does nothing.
*/
func (h *Handler) Log(format string, args ...interface{}) {
	if h.Logger != nil {
		h.Logger.Printf(format, args...)
	}
}

/*
Impelemnt the http.Handler inferface.

It determimes the proxy request and the local page request automitically.

If the h.EnableProxy is false, all proxy requests will be denied.

If the h.Handler is nil, the local page request will be routed to http.DefaultServerMux.

If the h.Handler is not nil, will use h.Handler to handle the request.
*/
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if b := h.isLocalRequest(r); b {
		if h.Handler != nil {
			/* invoke handler */
			h.Handler.ServeHTTP(w, r)
			return
		}

		/* invoke default handler */
		http.DefaultServeMux.ServeHTTP(w, r)
		return
	}

	if !h.EnableProxy {
		/* proxy not enabled */
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<h1>Not Found</h1>"))
		return
	}

	if r.ProtoMajor == 1 && !h.EnableProxyHTTP11 {
		/* proxy on http/1.1 not enabled */
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("<h1>Not Found</h1>"))
		return
	}

	/* proxy */

	if h.ProxyAuth {
		if h.ProxyAuthFunc == nil {
			panic("ProxyAuth is true but ProxyAuthFunc is nil")
		}

		if !h.ProxyAuthFunc(w, r) {
			return
		}
	}

	if r.Method == "CONNECT" {
		h.HandleConnect(w, r)
		return
	}

	h.HandleHTTP(w, r)
}

type flushWriter struct {
	w io.Writer
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if f, ok := fw.w.(http.Flusher); ok {
		f.Flush()
	}
	return
}

/*
HandleConnect handle the CONNECT request
*/
func (h *Handler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	srv := r.RequestURI
	if r.ProtoMajor == 2 {
		/* http/2.0 */
		srv = r.URL.Host
	}

	if _, _, err := net.SplitHostPort(srv); err != nil {
		/* no port specialed, set port to 443 */
		srv = net.JoinHostPort(srv, "443")
	}

	serverConn, err := util.Dial("tcp", srv)
	if err != nil {
		h.Log("dial to server: %s\n", err.Error())

		w.WriteHeader(http.StatusServiceUnavailable)

		w.Write([]byte(err.Error()))

		return
	}

	defer serverConn.Close()

	if r.ProtoMajor == 1 {
		/* HTTP/1.1 */
		hj, _ := w.(http.Hijacker)
		c, _, _ := hj.Hijack()

		defer c.Close()

		fmt.Fprintf(c, "HTTP/1.1 200 connection established\r\n\r\n")

		done := make(chan int)

		go forward(c, serverConn, done)
		go forward(serverConn, c, done)

		<-done

		return
	}

	/* HTTP/2.0 */

	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	done := make(chan int)

	go forward(serverConn, r.Body, done)
	go forward(flushWriter{w}, serverConn, done)

	<-done
}

func forward(dst io.Writer, src io.Reader, done chan int) {
	io.Copy(dst, src)
	select {
	case done <- 1:
	default:
	}
}

/*
HandleHTTP handle the other http proxy request, like GET, POST, HEAD.

If h.Transport is nil, will use http.DefaultTransport to process the request.

*/
func (h *Handler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error

	/* delete proxy-connection header */
	r.Header.Del("proxy-connection")

	/* set URL.Scheme, URL.Host for http/2.0 */
	if r.ProtoMajor == 2 {
		r.URL.Scheme = "http"
		r.URL.Host = r.Host
		r.RequestURI = r.URL.String()

		if r.Method != "POST" && r.Method != "PUT" {
			r.ContentLength = 0
			r.Body = nil
		}
	}

	if h.Transport != nil {
		/* invoke user defined transport */
		resp, err = h.Transport.RoundTrip(r)
	} else {
		/* invoke default transport */
		resp, err = http.DefaultTransport.RoundTrip(r)
	}

	if err != nil {
		h.Log("proxy err: %s\n", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	hdr := w.Header()
	for k, v := range resp.Header {
		//h.Log("header: %s = %s\n", k, v)
		if strings.ToLower(k) != "connection" {
			for _, v1 := range v {
				hdr.Add(k, v1)
			}
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)

	resp.Body.Close()
}

func (h *Handler) isLocalRequest(r *http.Request) bool {
	// connect is always a proxy request
	if r.Method == "CONNECT" {
		return false
	}

	// not enable proxy
	// all request trust as local request
	if !h.EnableProxy {
		return true
	}

	/* http/1.x */
	if r.ProtoMajor == 1 {
		if r.RequestURI[0] == '/' {
			return true
		}
		return false
	}

	/* http/2.x */
	if r.ProtoMajor == 2 {

		// LocalDomain not set
		// trust all as local request
		if len(h.LocalDomains) == 0 {
			return true
		}

		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}

		for _, d := range h.LocalDomains {
			if strings.HasSuffix(host, d) {
				return true
			}
		}

		return false
	}

	return true
}
