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
			if recs, err := h.Cache.Get(h.Cache.MakeKey(Q.qname, Q.qtype)); err == nil {
				//we have an answer now construct a dns.Msg
				result := new(dns.Msg)
				result.SetReply(req)
				for _, z := range recs.Value {
					rec, _ := dns.NewRR(dns.Fqdn(Q.qname) + " " + Q.qtype + " " + z)
					result.Answer = append(result.Answer, rec)
				}
				w.WriteMsg(result)
			} else {
				if rcs, err := h.Cache.Get(h.Cache.MakeKey(Q.qname, "CNAME")); err == nil {
					logger.Info("Found CNAME %s", rcs.String())
					result := new(dns.Msg)
					result.SetReply(req)
					for _, z := range rcs.Value {
						rc, _ := dns.NewRR(dns.Fqdn(Q.qname) + " " + "CNAME" + " " + z)
						result.Answer = append(result.Answer, rc)
						if rt, err := h.Cache.Get(h.Cache.MakeKey(z, Q.qtype)); err == nil {
							for _, ey := range rt.Value {
								gg, _ := dns.NewRR(z + " " + Q.qtype + " " + ey)
								result.Answer = append(result.Answer, gg)
							}
						}
					}
					w.WriteMsg(result)
				} else {
					h.Resolver.Lookup(h.Cache, w, req)
				}
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
