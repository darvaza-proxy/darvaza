package main

import (
	//	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/fangdingjun/gpp/util"
	"github.com/miekg/dns"
)

type Server struct {
	host       string
	port       int
	user       string
	group      string
	maxjobs    int
	maxqueries int
	handler    *GnoccoHandler
}

func (s *Server) Addr() string {
	return net.JoinHostPort(s.host, strconv.Itoa(s.port))
}

func (s *Server) DumpCache() {
	files := []string{Config.Cache.CachePath + "/pcache", Config.Cache.CachePath + "/ncache"}

	for _, file := range files {
		_ = os.MkdirAll(Config.Cache.CachePath, 0755)
		if fl, err := os.Create(file); err == nil {

			defer fl.Close()
			if strings.HasSuffix(file, "pcache") {
				s.handler.Cache.Dump(fl, true)
			} else {
				s.handler.Cache.Dump(fl, false)
			}

		} else {
			logger.Error("%s", err)
		}
	}

}

func (s *Server) Run() {

	s.handler = NewHandler(s.maxjobs)
	if _, err := os.Stat(Config.Cache.CachePath + "/pcache"); err == nil {
		var file, err = os.OpenFile(Config.Cache.CachePath+"/pcache", os.O_RDWR, 0644)
		if err == nil {
			s.handler.Cache.Load(file, true)
		}
	}
	if _, err := os.Stat(Config.Cache.CachePath + "/ncache"); err == nil {
		var file, err = os.OpenFile(Config.Cache.CachePath+"/ncache", os.O_RDWR, 0644)
		if err == nil {
			s.handler.Cache.Load(file, false)
		}

	}

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", s.handler.DoTCP)
	tcpServer := &dns.Server{Addr: s.Addr(),
		Net:     "tcp",
		Handler: tcpHandler}
	go s.start(tcpServer)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", s.handler.DoUDP)

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

func (s *Server) ShutDown() {
	logger.Info("Shutdown called.")
}
