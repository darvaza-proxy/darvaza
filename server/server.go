package darvaza

import (
	"net"

	"github.com/darvaza-proxy/darvaza/shared"
)

// Runner is an interface which is implemented by all proxies
type Runner interface {
	darvaza.Worker

	TLSHandler(func(net.Conn))
}

func NewServer() *darvaza.WorkGroup {
	return darvaza.NewWorkGroup()
}
