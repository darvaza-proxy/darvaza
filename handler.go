package main

import (
	"net"

	"github.com/miekg/dns"
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
	Resolver *Resolver
	MaxJobs  int
	Jobs     int
}

func NewHandler(m int) *GnoccoHandler {
	res := initResolver()
	return &GnoccoHandler{res, m, 0}
}

func (h *GnoccoHandler) do(Net string, w dns.ResponseWriter, req *dns.Msg) {
	if h.Jobs < h.MaxJobs {
		q := req.Question[0]
		Q := Question{q.Name, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

		var remote net.IP
		if Net == "tcp" {
			remote = w.RemoteAddr().(*net.TCPAddr).IP
		} else {
			remote = w.RemoteAddr().(*net.UDPAddr).IP
		}

		logger.Info("%s lookup　%s", remote, Q.String())
		h.Jobs++
		h.Resolver.LookupGen(w, req)
	} else {
		m := new(dns.Msg)
		m.SetRcode(req, 2)
		w.WriteMsg(m)
	}
}

func (h *GnoccoHandler) DoTCP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("tcp", w, req)
}

func (h *GnoccoHandler) DoUDP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("udp", w, req)
}
