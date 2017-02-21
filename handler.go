package main

import (
	"net"
	"time"

	"github.com/miekg/dns"
)

const (
	notIPQuery = 0
	_IP4Query  = 4
	_IP6Query  = 6
)

type Question struct {
	qname  string
	qtype  string
	qclass string
}

func (q *Question) String() string {
	return q.qname + " " + q.qclass + " " + q.qtype
}

type GnoccoHandler struct {
	Cache    *MCache
	Resolver *Resolver
}

func NewHandler() *GnoccoHandler {
	cache := &MCache{
		Backend:  make(map[string]Mesg, Config.Cache.MaxCount),
		Expire:   time.Duration(Config.Cache.Expire) * time.Second,
		Maxcount: Config.Cache.MaxCount,
	}
	res := initResolver()
	return &GnoccoHandler{cache, res}
}

func (h *GnoccoHandler) do(Net string, w dns.ResponseWriter, req *dns.Msg) {
	q := req.Question[0]
	Q := Question{q.Name, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

	var remote net.IP
	if Net == "tcp" {
		remote = w.RemoteAddr().(*net.TCPAddr).IP
	} else {
		remote = w.RemoteAddr().(*net.UDPAddr).IP
	}

	logger.Info("%s lookup　%s", remote, Q.String())
	h.Resolver.LookupGen(w, req)
}

func (h *GnoccoHandler) DoTCP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("tcp", w, req)
}

func (h *GnoccoHandler) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("udp", w, req)
}
