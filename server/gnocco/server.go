// Package gnocco provides a cached DNS resolver
package gnocco

import (
	"net"
	"strconv"

	"darvaza.org/slog"
	"github.com/miekg/dns"
)

// Resolver implements a cached DNS resolver
type Resolver struct {
	Host       string
	Port       int
	MaxJobs    int
	MaxQueries int
	handler    *gnoccoHandler
	cf         *Gnocco
}

// Addr returns the server's address
func (s *Resolver) Addr() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
}

// Logger returns the slog.Logger associated with the server's config
func (s *Resolver) Logger() slog.Logger {
	return s.cf.logger
}

func (s *Resolver) newHandler() *gnoccoHandler {
	return s.cf.newHandler(s.MaxJobs)
}

// Run runs the Server according to the Gnocco config
func (s *Resolver) Run() {
	s.handler = s.newHandler()

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", s.handler.do)
	tcpResolver := &dns.Server{Addr: s.Addr(),
		Net:     "tcp",
		Handler: tcpHandler}
	go s.start(tcpResolver)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", s.handler.do)

	udpResolver := &dns.Server{Addr: s.Addr(),
		Net:     "udp",
		Handler: udpHandler,
		UDPSize: 65535}

	go s.start(udpResolver)
}

func (s *Resolver) start(ds *dns.Server) {
	s.Logger().Info().Printf("Start %s listener on %s", ds.Net, ds.Addr)
	err := ds.ListenAndServe()
	if err != nil {
		s.Logger().Fatal().Printf("Start %s listener on %s failed:%s", ds.Net, ds.Addr, err.Error())
	}
}

// ShutDown pretends to shutdown the server
func (s *Resolver) ShutDown() {
	s.Logger().Info().Print("Shutdown called.")
}

// NewResolver returns a pointer to a gnocco.Resolver from a gnocco.Gnocco pointer
func NewResolver(cf *Gnocco) *Resolver {
	return &Resolver{
		Host:       cf.Listen.Host,
		Port:       cf.Listen.Port,
		MaxJobs:    cf.MaxJobs,
		MaxQueries: cf.MaxQueries,
		cf:         cf,
	}
}
