//go:build !go1.20

package httpserver

import "net/http"

// TODO: add mechanism to disable HTTP/3 support instead
var (
	_ int = "unfortunately your version of Go can't be supported"
)

// QuicHeaderMiddleware does nothing when building with go 1.19 or older
func (*Server) QuicHeadersMiddleware(next http.Handler) http.Handler {
	return next
}
