package http01

import (
	"net"
	"net/http"
	"strings"

	"github.com/darvaza-proxy/darvaza/acme"
)

var (
	_ http.Handler = (*Http01ChallengeHandler)(nil)
)

type Http01ChallengeHandler struct {
	Resolver acme.Http01Resolver
	next     http.Handler
}

func (h *Http01ChallengeHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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
		h.next = NewHttpsRedirectHandler()
	}
	h.next.ServeHTTP(rw, req)
}

func NewHtt01ChallengeHandler(resolver acme.Http01Resolver) *Http01ChallengeHandler {
	return &Http01ChallengeHandler{
		Resolver: resolver,
	}
}

func NewHttp01ChallengeMiddleware(resolver acme.Http01Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &Http01ChallengeHandler{
			Resolver: resolver,
			next:     next,
		}
	}
}
