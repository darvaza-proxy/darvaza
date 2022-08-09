package acme

import (
	"net/http"
)

type Http01Challenge interface {
	http.Handler
}

type Http01Resolver interface {
	AnnounceHost(hostname string)
	LookupChallenge(hostname, key string) Http01Challenge
}
