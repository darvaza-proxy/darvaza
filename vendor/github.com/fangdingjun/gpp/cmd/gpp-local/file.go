package main

import (
	"net/http"
)

func initRouters() {
	http.Handle("/", http.FileServer(http.Dir(docroot)))
}
