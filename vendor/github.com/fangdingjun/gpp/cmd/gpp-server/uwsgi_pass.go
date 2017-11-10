package main

import (
	//"fmt"
	uwsgi "github.com/fangdingjun/go-uwsgi"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// Uwsgi is a struct for uwsgi
type Uwsgi struct {
	Passenger *uwsgi.Passenger
	URLPrefix string
}

// NewUwsgi create a new Uwsgi
func NewUwsgi(network, addr, urlPrefix string) *Uwsgi {
	u := strings.TrimRight(urlPrefix, "/")
	return &Uwsgi{&uwsgi.Passenger{network, addr}, u}
}

// ServeHTTP implements http.Handler interface
func (u *Uwsgi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u.UwsgiPass(w, r)
}

// UwsgiPass pass the request to uwsgi interface
func (u *Uwsgi) UwsgiPass(w http.ResponseWriter, r *http.Request) {
	params := buildParams(r, u.URLPrefix)
	u.Passenger.UwsgiPass(w, r, params)
}

func buildParams(req *http.Request, urlPrefix string) map[string][]string {
	var err error

	header := make(map[string][]string)

	if urlPrefix != "" {
		header["SCRIPT_NAME"] = []string{urlPrefix}
		p := strings.Replace(req.URL.Path, urlPrefix, "", 1)
		header["PATH_INFO"] = []string{p}
	} else {
		header["PATH_INFO"] = []string{req.URL.Path}
	}

	//fmt.Printf("url: %s, scheme: %s\n", req.URL.String(), req.URL.Scheme)

	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	header["REQUEST_SCHEME"] = []string{scheme}

	header["HTTPS"] = []string{"off"}

	/* https */
	if scheme == "https" {
		header["HTTPS"] = []string{"on"}
	}

	/* speicial port */
	host, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		host = req.Host
		if scheme == "http" {
			port = "80"
		} else {
			port = "443"
		}
	}
	header["SERVER_NAME"] = []string{host}
	header["SERVER_PORT"] = []string{port}

	host, port, err = net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
		port = "80"
	}
	header["REMOTE_PORT"] = []string{port}
	header["REMOTE_ADDR"] = []string{host}

	header["REQUEST_METHOD"] = []string{req.Method}
	header["REQUEST_URI"] = []string{req.RequestURI}
	header["CONTENT_LENGTH"] = []string{strconv.Itoa(int(req.ContentLength))}
	header["SERVER_PROTOCOL"] = []string{req.Proto}
	header["QUERY_STRING"] = []string{req.URL.RawQuery}

	if ctype := req.Header.Get("Content-Type"); ctype != "" {
		header["CONTENT_TYPE"] = []string{ctype}
	}

	for k, v := range req.Header {
		k = "HTTP_" + strings.ToUpper(strings.Replace(k, "-", "_", -1))
		if _, ok := header[k]; ok == false {
			header[k] = v
		}
	}
	return header
}
