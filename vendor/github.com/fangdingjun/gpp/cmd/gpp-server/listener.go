package main

import (
	"fmt"
	"github.com/fangdingjun/gpp"
	"github.com/fangdingjun/handlers"
	"log"
	"net/http"
	"os"
)

func initListeners() {
	var out = os.Stdout

	if cfg.LogFile != "" {
		out1, err := os.Create(cfg.LogFile)
		if err != nil {
			log.Print(err)
		} else {
			out = out1
		}
	}
	log.SetOutput(out)
	logger := log.New(out, "", log.LstdFlags)

	for _, l := range cfg.Host {
		hdr1 := &gpp.Handler{
			//Handler:           Router,
			EnableProxy:       l.EnableProxy,
			EnableProxyHTTP11: true,
			LocalDomains:      cfg.LocalDomains,
			Logger:            logger,
			ProxyAuth:         l.ProxyAuth,
			ProxyAuthFunc:     proxyAuthFunc,
		}

		hdr := handlers.CombinedLoggingHandler(out, hdr1)
		if l.Cert != "" && l.Key != "" {
			go func(l ListenEntry) {
				fmt.Printf("Listen on https %s\n", l.Host)
				err := http.ListenAndServeTLS(l.Host, l.Cert, l.Key, hdr)
				if err != nil {
					fmt.Printf("listen failed on %s: %s\n", l.Host, err)
					os.Exit(-1)
				}
			}(l)
		} else {
			go func(l ListenEntry) {
				fmt.Printf("Listen on http %s\n", l.Host)
				err := http.ListenAndServe(l.Host, hdr)
				if err != nil {
					fmt.Printf("listen failed on %s: %s\n", l.Host, err)
					os.Exit(-1)
				}
			}(l)
		}

	}
}
