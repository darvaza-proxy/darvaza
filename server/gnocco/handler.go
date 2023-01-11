package gnocco

import (
	"net"

	"github.com/miekg/dns"

	"github.com/darvaza-proxy/slog"

	"github.com/darvaza-proxy/gnocco/shared/version"
)

type question struct {
	qname  string
	qtype  string
	qclass string
}

func (q *question) String() string {
	return q.qname + " " + q.qclass + " " + q.qtype
}

type gnoccoHandler struct {
	Cache    *cache
	Resolver *resolver
	MaxJobs  int
	Jobs     int
	logger   slog.Logger
}

func (cf *Gnocco) newHandler(m int) *gnoccoHandler {
	c := cf.newCache(int64(cf.Cache.MaxCount), int32(cf.Cache.Expire))
	r := cf.newResolver()
	return &gnoccoHandler{c, r, m, 0, cf.Logger()}
}

func (h *gnoccoHandler) do(Net string, w dns.ResponseWriter, req *dns.Msg) {
	if h.Jobs < h.MaxJobs {
		q := req.Question[0]
		Q := question{q.Name, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

		var remote net.IP
		if Net == "tcp" {
			remote = w.RemoteAddr().(*net.TCPAddr).IP
		} else {
			remote = w.RemoteAddr().(*net.UDPAddr).IP
		}

		h.logger.Info().Printf("%s lookup　%s", remote, Q.String())
		h.Jobs++
		switch {
		case Q.qclass == "IN":
			if recs, err := h.Cache.get(h.Cache.makeKey(Q.qname, Q.qtype)); err == nil {
				//we have an answer now construct a dns.Msg
				result := new(dns.Msg)
				result.SetReply(req)
				for _, z := range recs.Value {
					rec, _ := dns.NewRR(dns.Fqdn(Q.qname) + " " + Q.qtype + " " + z)
					result.Answer = append(result.Answer, rec)
				}
				w.WriteMsg(result)
			} else {
				if rcs, err := h.Cache.get(h.Cache.makeKey(Q.qname, "CNAME")); err == nil {
					h.logger.Info().Printf("Found CNAME %s", rcs.String())
					result := new(dns.Msg)
					result.SetReply(req)
					for _, z := range rcs.Value {
						rc, _ := dns.NewRR(dns.Fqdn(Q.qname) + " " + "CNAME" + " " + z)
						result.Answer = append(result.Answer, rc)
						if rt, err := h.Cache.get(h.Cache.makeKey(z, Q.qtype)); err == nil {
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
			case "authors.bind.":
				m.Answer = append(m.Answer, &dns.TXT{Hdr: hdr, Txt: []string{"Nagy Karoly Gabriel <k@jpi.io>"}})
			case "version.bind.", "version.server.":
				m.Answer = []dns.RR{&dns.TXT{Hdr: hdr, Txt: []string{"Version " + version.Version + " built on " + version.BuildDate}}}
			case "hostname.bind.", "id.server.":
				m.Answer = []dns.RR{&dns.TXT{Hdr: hdr, Txt: []string{"localhost"}}}
			default:
				m.SetRcode(req, dns.RcodeNotImplemented)
				w.WriteMsg(m)
			}
			w.WriteMsg(m)

		default:
		}
		h.Jobs--
	} else {
		m := new(dns.Msg)
		m.SetRcode(req, dns.RcodeServerFailure)
		w.WriteMsg(m)
	}
}

func (h *gnoccoHandler) doTCP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("tcp", w, req)
}

func (h *gnoccoHandler) doUDP(w dns.ResponseWriter, req *dns.Msg) {
	h.do("udp", w, req)
}
