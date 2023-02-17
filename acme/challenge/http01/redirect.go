package http01

import (
	"fmt"
	"net/http"
)

// HTTPSRedirectHandler provides an automatic redirect to HTTPS
type HTTPSRedirectHandler struct{}

func (*HTTPSRedirectHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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

// NewHTTPSRedirectHandler creates a new automatic redirect to HTTPS handler
func NewHTTPSRedirectHandler() http.Handler {
	return &HTTPSRedirectHandler{}
}
