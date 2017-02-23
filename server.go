package main

import (
	"net"
	"os"
	"strconv"

	"github.com/fangdingjun/gpp/util"
	"github.com/miekg/dns"
)

type Server struct {
	host  string
	port  int
	dotcp bool
	user  string
	group string
	cache Cache
}

func (s *Server) Addr() string {
	return net.JoinHostPort(s.host, strconv.Itoa(s.port))
}

func (s *Server) DumpCache() {
	s.cache.Dump(os.Stdout)
}

func (s *Server) Run() {

	Handler := NewHandler()
	s.cache = NewMemory(10000, 60)

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", Handler.DoTCP)
	tcpServer := &dns.Server{Addr: s.Addr(),
		Net:     "tcp",
		Handler: tcpHandler}
	go s.start(tcpServer)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", Handler.DoUDP)

	udpServer := &dns.Server{Addr: s.Addr(),
		Net:     "udp",
		Handler: udpHandler,
		UDPSize: 65535}

	go s.start(udpServer)

	err := util.DropPrivilege(s.user, s.group)
	if err != nil {
		logger.Error("Dropping privileges failed %s", err.Error())
	}
}

func (s *Server) start(ds *dns.Server) {
	logger.Info("Start %s listener", ds.Net)
	err := ds.ListenAndServe()
	if err != nil {
		logger.Fatal("Start %s listener failed:%s", ds.Net, err.Error())
	}
}
