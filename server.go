package main

import (
	//	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fangdingjun/gpp/util"
	"github.com/miekg/dns"
)

type server struct {
	host       string
	port       int
	user       string
	group      string
	maxjobs    int
	maxqueries int
	handler    *gnoccoHandler
}

func (s *server) Addr() string {
	return net.JoinHostPort(s.host, strconv.Itoa(s.port))
}

func (s *server) dumpCache() {
	files := []string{mainconfig.Cache.CachePath + "/pcache", mainconfig.Cache.CachePath + "/ncache"}

	for _, file := range files {
		_ = os.MkdirAll(mainconfig.Cache.CachePath, 0755)
		if fl, err := os.Create(file); err == nil {

			defer fl.Close()
			if strings.HasSuffix(file, "pcache") {
				s.handler.Cache.dump(fl, true)
			} else {
				s.handler.Cache.dump(fl, false)
			}

		} else {
			logger.error("%s", err)
		}
	}

}

func (s *server) run() {

	s.handler = newHandler(s.maxjobs)
	if _, err := os.Stat(mainconfig.Cache.CachePath + "/pcache"); err == nil {
		var file, err = os.OpenFile(mainconfig.Cache.CachePath+"/pcache", os.O_RDWR, 0644)
		if err == nil {
			s.handler.Cache.load(file, true)
		}
	}
	if _, err := os.Stat(mainconfig.Cache.CachePath + "/ncache"); err == nil {
		var file, err = os.OpenFile(mainconfig.Cache.CachePath+"/ncache", os.O_RDWR, 0644)
		if err == nil {
			s.handler.Cache.load(file, false)
		}

	}

	tcpHandler := dns.NewServeMux()
	tcpHandler.HandleFunc(".", s.handler.doTCP)
	tcpServer := &dns.Server{Addr: s.Addr(),
		Net:     "tcp",
		Handler: tcpHandler}
	go s.start(tcpServer)

	udpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", s.handler.doUDP)

	udpServer := &dns.Server{Addr: s.Addr(),
		Net:     "udp",
		Handler: udpHandler,
		UDPSize: 65535}

	go s.start(udpServer)

	go s.startCacheDumping()

	err := util.DropPrivilege(s.user, s.group)
	if err != nil {
		logger.error("Dropping privileges failed %s", err.Error())
	}
}

func (s *server) start(ds *dns.Server) {
	logger.info("Start %s listener", ds.Net)
	err := ds.ListenAndServe()
	if err != nil {
		logger.fatal("Start %s listener failed:%s", ds.Net, err.Error())
	}
}

func (s *server) shutDown() {
	logger.info("Shutdown called.")
}

func (s *server) startCacheDumping() {
	interval := mainconfig.Cache.DumpInterval
	for {
		<-time.After(time.Duration(interval) * time.Second)
		go s.dumpCache()
	}
}
