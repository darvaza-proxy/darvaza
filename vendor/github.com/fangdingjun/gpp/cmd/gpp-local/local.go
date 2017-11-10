/*
gpp-local is application to convert the https proxy(http proxy over TLS) to a normal http proxy.

Usage

generate a configure file and edit it
    $ gpp-local -dumpflags > client.ini

run it
    $ gpp-local -config client.ini
*/
package main

import (
	"crypto/tls"
	//"errors"
	"flag"
	"fmt"
	//"github.com/fangdingjun/net/http2"
	"github.com/vharitonsky/iniflags"
	"golang.org/x/net/http2"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	//"net/url"
	//"strings"
	//"github.com/fangdingjun/gpp/util"
	"github.com/fangdingjun/handlers"
	_ "net/http/pprof"
	"os"
	"time"
)

var serverName string
var port = 0

type myhandler struct {
	proxy *proxy
}

type proxy struct {
	handler http.RoundTripper
	index   int
}

func (p *proxy) do(r *http.Request) (*http.Response, error) {
	return p.handler.RoundTrip(r)
}

type conn struct {
	net.Conn
	timeout time.Duration
}

func (c *conn) Read(buf []byte) (n int, err error) {
	err = c.Conn.SetReadDeadline(time.Now().Add(c.timeout))
	if err != nil {
		return
	}
	return c.Conn.Read(buf)
}

func (c *conn) Write(buf []byte) (n int, err error) {
	err = c.Conn.SetWriteDeadline(time.Now().Add(c.timeout))
	if err != nil {
		return
	}
	return c.Conn.Write(buf)
}

func (p *proxy) dialTLS(network, addr string, cfg *tls.Config) (net.Conn, error) {
	name := addr
	if serverName != "" {
		name = serverName
	}
	addr = hosts[0]
	c, err := net.DialTimeout(network, addr, 3*time.Second)
	if err != nil {
		return nil, err
	}

	cfg.ServerName = name
	cfg.InsecureSkipVerify = false

	cc := tls.Client(&conn{c, 80 * time.Second}, cfg)

	err = cc.Handshake()
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return cc, nil
}

func (mhd *myhandler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	r.Header.Del("proxy-connection")

	pr, pw := io.Pipe()

	defer pr.Close()

	r.Body = ioutil.NopCloser(pr)
	r.URL.Scheme = "https"
	r.URL.Host = hosts[0]
	r.ContentLength = -1
	s, err := mhd.proxy.do(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	defer s.Body.Close()

	w.WriteHeader(s.StatusCode)

	c, _, err := hj.Hijack()
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	defer c.Close()

	done := make(chan int)

	go forward(c, s.Body, done)
	go forward(pw, c, done)

	<-done

}

func forward(dst io.Writer, src io.Reader, done chan int) {
	io.Copy(dst, src)
	select {
	case done <- 1:
	default:
	}
}

func (mhd *myhandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("proxy-connection")
	//r.URL.Scheme = "https"
	r.URL.Host = hosts[0]
	resp, err := mhd.proxy.do(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	header := w.Header()
	for k, v := range resp.Header {
		for _, v1 := range v {
			header.Add(k, v1)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func (mhd *myhandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		mhd.HandleConnect(w, r)
		return
	}

	// local request
	if r.RequestURI[0] == '/' {
		http.DefaultServeMux.ServeHTTP(w, r)
		return
	}

	mhd.HandleHTTP(w, r)
}

type myargs []string

func (m *myargs) Set(s string) error {
	*m = append(*m, s)
	return nil
}

func (m *myargs) String() string {
	return ""
}

var hosts myargs
var docroot string

func main() {

	flag.IntVar(&port, "port", 8080, "the port listen to")
	flag.StringVar(&serverName, "server_name", "", "the server name")
	flag.Var(&hosts, "server", "the server connect to")
	flag.StringVar(&docroot, "docroot", ".", "the local http www root")

	iniflags.Parse()

	initRouters()

	if len(hosts) == 0 {
		log.Fatal("you must special a server")
	}
	http2.VerboseLogs = false
	log.Printf("Listening on :%d", port)
	p := &proxy{}
	p.handler = &http2.Transport{
		DialTLS:   p.dialTLS,
		AllowHTTP: true,
	}
	hdr := &myhandler{proxy: p}
	err := http.ListenAndServe(fmt.Sprintf(":%d", port),
		handlers.CombinedLoggingHandler(os.Stdout, hdr))
	if err != nil {
		log.Fatal(err)
	}
}
