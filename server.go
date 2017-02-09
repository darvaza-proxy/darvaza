package main

import (
	"net"
	"strconv"

	"github.com/fangdingjun/gpp/util"
	"github.com/miekg/dns"
)

type Server struct {
	thost string
	tport int
	uhost string
	uport int
	user  string
	group string
}

func (s *Server) UAddr() string {
	return net.JoinHostPort(s.uhost, strconv.Itoa(s.uport))
}

func (s *Server) TAddr() string {
	return net.JoinHostPort(s.thost, strconv.Itoa(s.tport))
}
func (s *Server) Run() {

	Handler := NewHandler()

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", Handler.DoTCP)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", Handler.DoUDP)
	tcpServer := &dns.Server{Addr: s.TAddr(),
		Net:     "tcp",
		Handler: tcpHandler}

	udpServer := &dns.Server{Addr: s.UAddr(),
		Net:     "udp",
		Handler: udpHandler,
		UDPSize: 65535}

	go s.start(udpServer)
	go s.start(tcpServer)

	err := util.DropPrivilege(s.user, s.group)
	if err != nil {
		logger.Error("Dropping privileges failed %s", err.Error())
	}
}

func (s *Server) start(ds *dns.Server) {
	logger.Info("Start %s listener", ds.Net)
	err := ds.ListenAndServe()
	if err != nil {
		logger.Error("Start %s listener failed:%s", ds.Net, err.Error())
	}
}
