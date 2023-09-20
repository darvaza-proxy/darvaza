//go:build !go1.20

package httpserver

// TODO: add mechanism to disable HTTP/3 support instead
var (
	_ int = "unfortunately your version of Go can't be supported"
)
