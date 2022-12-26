// Package server provides logic for applications facing browsers
package server

import (
	"net"

	"github.com/darvaza-proxy/darvaza/shared"
)

// Runner is an interface which is implemented by all proxies
type Runner interface {
	shared.Worker

	TLSHandler(func(net.Conn))
}

// NewServer creates a new WorkGroup
func NewServer() *shared.WorkGroup {
	return shared.NewWorkGroup()
}
