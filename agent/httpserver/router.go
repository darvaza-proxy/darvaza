package httpserver

import (
	"net/http"

	"github.com/darvaza-proxy/darvaza/acme/challenge/http01"
)

func (srv *Server) mightInitMux() {
	if srv.mux == nil {
		srv.mux = http.NewServeMux()

		if h := srv.cfg.Handler; h != nil {
			// Application Handler from Config.
			// Make sure you pass `nil` to Serve()
			srv.mux.Handle("/", h)
		}

		// ACME-HTTP-01
		if r := srv.cfg.AcmeHTTP01; r != nil {
			h := http01.NewChallengeHandler(r)
			srv.mux.Handle(http01.WellKnownPath, h)
		}
	}
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (srv *Server) Handle(pattern string, handler http.Handler) {
	srv.mightInitMux()

	srv.mux.Handle(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (srv *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	srv.mux.Handle(pattern, http.HandlerFunc(handler))
}

func (srv *Server) newInsecureHandler() http.Handler {
	h := http01.NewHTTPSRedirectHandler()
	// ACME-HTTP-01
	if r := srv.cfg.AcmeHTTP01; r != nil {
		m := http01.NewChallengeMiddleware(r)
		h = m(h)
	}
	return h
}
