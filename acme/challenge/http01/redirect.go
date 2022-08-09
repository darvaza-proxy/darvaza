package http01

import (
	"fmt"
	"net/http"
)

type HttpsRedirectHandler struct{}

func (h *HttpsRedirectHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Scheme != "https" {
		url := *req.URL
		url.Scheme = "https"

		loc := url.String()

		rw.Header().Add("Location", loc)
		rw.WriteHeader(http.StatusPermanentRedirect)
		fmt.Fprintf(rw, "Redirected to %s", loc)
	} else {
		http.NotFound(rw, req)
	}
}

func NewHttpsRedirectHandler() http.Handler {
	return &HttpsRedirectHandler{}
}
