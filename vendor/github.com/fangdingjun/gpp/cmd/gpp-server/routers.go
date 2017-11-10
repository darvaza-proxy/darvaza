package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"os"
	"regexp"
	//"path/filepath"
	//"strings"
)

func initRouters() {
	router := mux.NewRouter()

	for _, r := range cfg.Routes {
		switch r.URLType {
		case "file":
			registerFileHandler(r, router)
		case "dir":
			registerDirHandler(r, router)
		case "uwsgi":
			registerUwsgiHandler(r, router)
		case "fastcgi":
			registerFastCGIHandler(r, router)
		default:
			fmt.Printf("invalid type: %s\n", r.URLType)
		}
	}

	http.Handle("/", router)
}

func registerFileHandler(r URLRoute, router *mux.Router) {
	router.HandleFunc(r.URLPrefix, func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, r.Path)
	})
}

func registerDirHandler(r URLRoute, router *mux.Router) {
	router.PathPrefix(r.URLPrefix).Handler(http.FileServer(http.Dir(r.DocRoot)))
}

func registerUwsgiHandler(r URLRoute, router *mux.Router) {
	_u1, err := url.Parse(r.Path)
	if err != nil {
		fmt.Printf("invalid path: %s\n", r.Path)
		os.Exit(-1)
	}
	_p := _u1.Path
	switch _u1.Scheme {
	case "unix":
	case "tcp":
		_p = _u1.Host
	default:
		fmt.Printf("invalid scheme: %s, only support unix, tcp", _u1.Scheme)
		os.Exit(-1)
	}
	if r.UseRegex {
		m1 := myURLMatch{regexp.MustCompile(r.URLPrefix)}
		_u := NewUwsgi(_u1.Scheme, _p, "")
		router.MatcherFunc(m1.match).Handler(_u)
	} else {
		_u := NewUwsgi(_u1.Scheme, _p, r.URLPrefix)
		router.PathPrefix(r.URLPrefix).Handler(_u)
	}
}

func registerFastCGIHandler(r URLRoute, router *mux.Router) {
	_u1, err := url.Parse(r.Path)
	if err != nil {
		fmt.Printf("invalid path: %s\n", r.Path)
		os.Exit(-1)
	}
	_p := _u1.Path
	switch _u1.Scheme {
	case "unix":
	case "tcp":
		_p = _u1.Host
	default:
		fmt.Printf("invalid scheme: %s, only support unix, tcp", _u1.Scheme)
		os.Exit(-1)
	}
	if r.UseRegex {
		m1 := myURLMatch{regexp.MustCompile(r.URLPrefix)}
		_u, _ := NewFastCGI(_u1.Scheme, _p, r.DocRoot, "")
		router.MatcherFunc(m1.match).Handler(_u)
	} else {
		_u, _ := NewFastCGI(_u1.Scheme, _p, r.DocRoot, r.URLPrefix)
		router.PathPrefix(r.URLPrefix).Handler(_u)
	}
}

type myURLMatch struct {
	re *regexp.Regexp
}

func (m myURLMatch) match(r *http.Request, route *mux.RouteMatch) bool {
	ret := m.re.MatchString(r.URL.Path)
	return ret
}
