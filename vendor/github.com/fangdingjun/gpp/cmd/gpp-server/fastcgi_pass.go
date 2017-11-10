package main

import (
	"github.com/yookoala/gofast"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FastCGI is a fastcgi client connection
type FastCGI struct {
	Network   string
	Addr      string
	DocRoot   string
	URLPrefix string
	//client  gofast.Client
}

// NewFastCGI creates a new FastCGI struct
func NewFastCGI(network, addr, docroot, urlPrefix string) (*FastCGI, error) {
	u := strings.TrimRight(urlPrefix, "/")
	return &FastCGI{network, addr, docroot, u}, nil
}

var fcgiPathInfo = regexp.MustCompile(`^(.*?\.php)(.*)$`)

// ServeHTTP implements http.Handler interface
func (f FastCGI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.FastCGIPass(w, r)
}

// FastCGIPass pass the request to fastcgi socket
func (f FastCGI) FastCGIPass(w http.ResponseWriter, r *http.Request) {
	var scriptName, pathInfo, scriptFileName string

	conn, err := net.Dial(f.Network, f.Addr)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	defer conn.Close()

	client := gofast.NewClient(conn, 20)

	urlPath := r.URL.Path
	if f.URLPrefix != "" {
		urlPath = strings.Replace(r.URL.Path, f.URLPrefix, "", 1)
	}

	p := fcgiPathInfo.FindStringSubmatch(urlPath)

	if len(p) < 2 {
		if strings.HasSuffix(r.URL.Path, "/") {
			// redirect to index.php
			scriptName = ""
			pathInfo = ""
			scriptFileName = filepath.Join(f.DocRoot, urlPath, "index.php")
		} else {
			// serve static file in php directory
			fn := filepath.Join(f.DocRoot, urlPath)
			http.ServeFile(w, r, fn)
			return
		}
	} else {
		scriptName = p[1]
		pathInfo = p[2]
		scriptFileName = filepath.Join(f.DocRoot, scriptName)
	}

	req := client.NewRequest(r)

	req.Params["PATH_INFO"] = pathInfo
	req.Params["SCRIPT_FILENAME"] = scriptFileName

	https := "off"
	scheme := "http"
	if r.TLS != nil {
		https = "on"
		scheme = "https"
	}

	req.Params["REQUEST_SCHEME"] = scheme
	req.Params["HTTPS"] = https

	host, port, _ := net.SplitHostPort(r.RemoteAddr)
	req.Params["REMOTE_ADDR"] = host
	req.Params["REMOTE_PORT"] = port

	host, port, err = net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
		if scheme == "http" {
			port = "80"
		} else {
			port = "443"
		}
	}
	req.Params["SERVER_NAME"] = host
	req.Params["SERVER_PORT"] = port

	req.Params["SERVER_PROTOCOL"] = r.Proto

	for k, v := range r.Header {
		k = "HTTP_" + strings.ToUpper(strings.Replace(k, "-", "_", -1))
		if _, ok := req.Params[k]; ok == false {
			req.Params[k] = strings.Join(v, ";")
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	err = resp.WriteTo(w, os.Stderr)
	if err != nil {
		log.Println(err)
	}

	resp.Close()
}
