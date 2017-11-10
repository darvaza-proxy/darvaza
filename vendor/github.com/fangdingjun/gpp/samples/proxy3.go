package main

import (
	. "fmt"
	"github.com/fangdingjun/gpp"
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

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/welcome", welcome)
	http.Handle("/", http.FileServer(http.Dir(".")))

	log.Print("Listen on: ", Sprintf("0.0.0.0:%d", port))
	err := http.ListenAndServe(Sprintf(":%d", port), &gpp.Handler{})
	if err != nil {
		log.Fatal(err)
	}
}
