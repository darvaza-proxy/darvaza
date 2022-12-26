// Package http01 provides logic regarding ACME-HTTP-01 protocol
package http01

import (
	"net"
	"net/http"
	"strings"

	"github.com/darvaza-proxy/darvaza/acme"
)

var (
	_ http.Handler = (*ChallengeHandler)(nil)
)

// ChallengeHandler handles /.well-known/acme-challenge requests
// against a given HTTP-01 Challenge Resolver
type ChallengeHandler struct {
	Resolver acme.HTTP01Resolver
	next     http.Handler
}

func (h *ChallengeHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	host := req.URL.Hostname()
	path := req.URL.Path

	if h.Resolver != nil && net.ParseIP(host) == nil {
		// only process named hosts

		h.Resolver.AnnounceHost(host)

		token := strings.TrimPrefix(path, "/.well-known/acme-challenge")
		if token == path {
			// invalid prefix
			goto next
		} else if l := len(token); l == 0 {
			// no token
			http.NotFound(rw, req)
		} else if token[0] != '/' {
			// invalid prefix
			goto next
		} else if c := h.Resolver.LookupChallenge(host, token[1:]); c == nil {
			// host,token pair not recognised
			http.NotFound(rw, req)
		} else {
			// host,token pair recognised, proceed
			c.ServeHTTP(rw, req)
		}

		return
	}

next:
	if h.next == nil {
		h.next = NewHTTPSRedirectHandler()
	}
	h.next.ServeHTTP(rw, req)
}

// NewChallengeHandler creates a handler for the provided HTTP-01 challenge resolver
func NewChallengeHandler(resolver acme.HTTP01Resolver) *ChallengeHandler {
	return &ChallengeHandler{
		Resolver: resolver,
	}
}

// NewChallengeMiddleware creates middleware using a provided HTTP-01 challenge resolver
func NewChallengeMiddleware(resolver acme.HTTP01Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &ChallengeHandler{
			Resolver: resolver,
			next:     next,
		}
	}
}
