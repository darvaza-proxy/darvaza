// Package gnocco provides a cached DNS resolver
package gnocco

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/darvaza-proxy/slog"
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

// DumpCache writes the cache to disc
func (s *Resolver) DumpCache() {
	cache := &s.cf.Cache

	s.Logger().Info().Printf("Dumping cache at %v", time.Now())
	files := []string{cache.CachePath + "/pcache", cache.CachePath + "/ncache"}

	for _, file := range files {
		_ = os.MkdirAll(cache.CachePath, 0755)
		if fl, err := os.Create(file); err == nil {

			defer fl.Close()
			if strings.HasSuffix(file, "pcache") {
				s.handler.Cache.dump(fl, true)
			} else {
				s.handler.Cache.dump(fl, false)
			}

		} else {
			s.Logger().Error().Print(err)
		}
	}

}

func (s *Resolver) newHandler() *gnoccoHandler {
	return s.cf.newHandler(s.MaxJobs)
}

// Run runs the Server according to the Gnocco config
func (s *Resolver) Run() {
	cache := &s.cf.Cache

	s.handler = s.newHandler()
	if _, err := os.Stat(cache.CachePath + "/pcache"); err == nil {
		var file, err = os.OpenFile(cache.CachePath+"/pcache", os.O_RDWR, 0644)
		if err == nil {
			s.handler.Cache.load(file, true)
		}
	}
	if _, err := os.Stat(cache.CachePath + "/ncache"); err == nil {
		var file, err = os.OpenFile(cache.CachePath+"/ncache", os.O_RDWR, 0644)
		if err == nil {
			s.handler.Cache.load(file, false)
		}

	}

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", s.handler.doTCP)
	tcpResolver := &dns.Server{Addr: s.Addr(),
		Net:     "tcp",
		Handler: tcpHandler}
	go s.start(tcpResolver)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", s.handler.doUDP)

	udpResolver := &dns.Server{Addr: s.Addr(),
		Net:     "udp",
		Handler: udpHandler,
		UDPSize: 65535}

	go s.start(udpResolver)

	go s.startCacheDumping()

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

func (s *Resolver) startCacheDumping() {
	interval := s.cf.Cache.DumpInterval

	for range time.Tick(time.Duration(interval) * time.Second) {
		s.DumpCache()
	}
}
