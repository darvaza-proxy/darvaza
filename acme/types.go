package acme

import (
	"net/http"
)

// HTTP01Challenge represents a ACME-HTTP-01 challenge handler
type HTTP01Challenge interface {
	http.Handler
}

// HTTP01Resolver represents the interface using to handle
// a ACME-HTTP-01 challenge
type HTTP01Resolver interface {
	AnnounceHost(hostname string)
	LookupChallenge(hostname, key string) HTTP01Challenge
}
