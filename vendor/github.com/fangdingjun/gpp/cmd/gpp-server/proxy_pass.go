package main

import (
	"net"
	"net/http"
	//"bufio"
	"io"
	"log"
	//"strings"
)

type proxy struct {
	transport *http.Transport
	addr      string
	//prefix string
}

func newProxy(addr string) *proxy {
	p := &proxy{addr: addr}
	p.transport = &http.Transport{
		Dial: p.dial,
	}
	/*
	   p.prefix = prefix
	   if p.prefix[len(prefix)-1] == '/'{
	       p.prefix = strings.TrimRight(p.prefix, "/")
	   }
	*/
	return p
}

func (p *proxy) dial(network string, addr string) (conn net.Conn, err error) {
	return net.Dial("tcp", p.addr)
}

func (p *proxy) proxyPass(w http.ResponseWriter, r *http.Request) {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	r.Header.Add("X-Forwarded-For", host)
	//r.RequestURI = ""
	r.URL.Scheme = "http"
	r.URL.Host = r.Host
	//r.URL.Path = strings.TrimLeft(r.URL.Path, p.prefix)
	resp, err := p.transport.RoundTrip(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("<h1>502 Bad Gateway</h1>"))
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
