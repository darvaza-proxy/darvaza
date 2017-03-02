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
	Cache    *Cache
	Resolver *Resolver
	MaxJobs  int
	Jobs     int
}

func NewHandler(m int) *GnoccoHandler {
	c := NewCache(int64(Config.Cache.MaxCount), int32(Config.Cache.Expire))
	r := initResolver()
	return &GnoccoHandler{c, r, m, 0}
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
		switch {
		case Q.qclass == "IN":
			if r, err := h.Cache.Get(h.Cache.makeKey(Q.qname, Q.qtype), req); err == nil {
				w.WriteMsg(r)
			} else {
				logger.Debug("%s", err)
				h.Resolver.LookupGen(w, req)
			}
		case Q.qclass == "CH", Q.qtype == "TXT":
			m := new(dns.Msg)
			m.SetReply(req)
			hdr := dns.RR_Header{Name: Q.qname, Rrtype: dns.TypeTXT, Class: dns.ClassCHAOS, Ttl: 0}
			switch Q.qname {
			default:
				m.SetRcode(req, 4)
				w.WriteMsg(m)
			case "authors.bind.":
				m.Answer = append(m.Answer, &dns.TXT{Hdr: hdr, Txt: []string{"Nagy Karoly Gabriel <k@jpi.io>"}})
			case "version.bind.", "version.server.":
				m.Answer = []dns.RR{&dns.TXT{Hdr: hdr, Txt: []string{"gnocco-alpha"}}}
			case "hostname.bind.", "id.server.":
				m.Answer = []dns.RR{&dns.TXT{Hdr: hdr, Txt: []string{"localhost"}}}
			}
			w.WriteMsg(m)

		default:
		}
		h.Jobs--
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
