package main

import (
	. "fmt"
	"github.com/fangdingjun/gpp"
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

	router := mux.NewRouter()

	router.HandleFunc("/hello", hello)
	router.HandleFunc("/welcome", welcome)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(".")))

	http.Handle("/", router)

	log.Print("Listen on: ", Sprintf("0.0.0.0:%d", port))
	err := http.ListenAndServe(Sprintf(":%d", port), &gpp.Handler{})
	if err != nil {
		log.Fatal(err)
	}
}
