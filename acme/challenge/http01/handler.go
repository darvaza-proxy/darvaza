// Package http01 provides logic regarding ACME-HTTP-01 protocol
package http01

import (
	"net/http"
	"strings"

	"darvaza.org/darvaza/acme"
	"darvaza.org/middleware"
)

const (
	// WellKnownPath is the directory handled to resolve this challenge
	WellKnownPath = "/.well-known/acme-challenge"
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
	if h.Resolver != nil {
		if c := h.resolveHandler(req.URL.Hostname(), req.URL.Path); c != nil {
			c.ServeHTTP(rw, req)
			return
		}
	}

	if h.next == nil {
		h.next = middleware.NewHTTPSRedirectHandler(0)
	}
	h.next.ServeHTTP(rw, req)
}

func (h *ChallengeHandler) resolveHandler(host, path string) http.Handler {
	var c http.Handler

	token, ok := TokenFromPath(path)
	if ok {
		h.Resolver.AnnounceHost(host)

		if token != "" {
			c = h.Resolver.LookupChallenge(host, token)
		}

		if c == nil {
			c = http.NotFoundHandler()
		}
	}

	return c
}

// TokenFromPath returns the path within the well-known
// directory and an indicator if the path pointed
// to the well-known directory or not
func TokenFromPath(path string) (string, bool) {
	token := strings.TrimPrefix(path, WellKnownPath)
	if token == path {
		return "", false
	} else if token == "" {
		return "", true
	} else if token[0] != '/' {
		return "", false
	} else {
		return token[1:], true
	}
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
