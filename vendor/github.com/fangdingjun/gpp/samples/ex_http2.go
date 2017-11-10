/*
a example to support http2
*/
package main

import (
	. "fmt"
	"github.com/fangdingjun/gpp"
	"github.com/fangdingjun/http2"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>hello</h1>"))
}

func welcome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>welcome</h1>"))
}

func main() {
	port := 8080

	var srv http.Server

	router := mux.NewRouter()

	router.HandleFunc("/hello", hello)
	router.HandleFunc("/welcome", welcome)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(".")))

	srv.Addr = Sprintf(":%d", port)

	/* set the http handler */
	srv.Handler = &gpp.Handler{EnableProxy: true, Handler: router}

	/* initial http2 support */
	http2.ConfigureServer(&srv, nil)

	log.Print("Listen on: ", Sprintf("https://0.0.0.0:%d", port))
	srv.ListenAndServeTLS("server.crt", "server.key")
	if err != nil {
		log.Fatal(err)
	}
}
