package main

import (
	"encoding/base64"
	"log"
	"net"
	"net/http"
	"strings"
)

func getPass(r *http.Request) (user, pass string) {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
		if r.TLS != nil {
			port = "443"
		} else {
			port = "80"
		}
	}

	for _, h := range cfg.Host {
		host1, port1, _ := net.SplitHostPort(h.Host)
		if host1 == host && port == port1 {
			// host, port match
			return h.ProxyUser, h.ProxyPasswd
		} else if host1 == "" && port1 == port {
			// port match and host is wildcard
			return h.ProxyUser, h.ProxyPasswd
		} else if host1 == "0.0.0.0" && port == port1 {
			// port match and host is wildcard
			return h.ProxyUser, h.ProxyPasswd
		} else if host1 == "::" && port == port1 {
			// port match and host is wildcard
			return h.ProxyUser, h.ProxyPasswd
		}
	}
	return "", ""
}

func proxyAuthFunc(w http.ResponseWriter, r *http.Request) bool {
	user, pass := getBasicUserpass(r)
	if user == "" && pass == "" {
		authFailed(w)
		return false
	}
	u1, p1 := getPass(r)
	if user == u1 && pass == p1 {
		return true
	}

	authFailed(w)
	return false
}

func authFailed(w http.ResponseWriter) {
	w.Header().Add("Proxy-Authenticate", "Basic realm=\"xxxx.com\"")
	w.WriteHeader(407)
	w.Write([]byte("<h1>unauthenticate</h1>"))
}

func getBasicUserpass(r *http.Request) (string, string) {
	proxyHeader := r.Header.Get("Proxy-Authorization")
	if proxyHeader == "" {
		return "", ""
	}

	r.Header.Del("Proxy-Authorization")

	ss := strings.Split(proxyHeader, " ")
	if len(ss) != 2 {
		return "", ""
	}

	if strings.ToLower(ss[0]) != "basic" {
		return "", ""
	}

	data, err := base64.StdEncoding.DecodeString(ss[1])
	if err != nil {
		log.Printf("%s\n", err.Error())
		return "", ""
	}

	uu := strings.SplitN(string(data), ":", 2)

	return uu[0], uu[1]
}
