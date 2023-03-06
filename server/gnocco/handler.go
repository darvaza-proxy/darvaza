package gnocco

import (
	"net"

	"github.com/miekg/dns"

	"github.com/darvaza-proxy/slog"

	"github.com/darvaza-proxy/darvaza/shared/version"
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
	Resolver *resolver
	MaxJobs  int
	Jobs     int
	logger   slog.Logger
}

func (cf *Gnocco) newHandler(m int) *gnoccoHandler {
	r := cf.newResolver()
	return &gnoccoHandler{r, m, 0, cf.Logger()}
}

func getIPFromWriter(w dns.ResponseWriter) net.IP {
	if _, ok := w.RemoteAddr().(*net.UDPAddr); ok {
		return w.RemoteAddr().(*net.UDPAddr).IP
	}
	return w.RemoteAddr().(*net.TCPAddr).IP
}

func (h *gnoccoHandler) do(w dns.ResponseWriter, req *dns.Msg) {
	if h.Jobs < h.MaxJobs {
		q := req.Question[0]
		myQ := question{q.Name, dns.TypeToString[q.Qtype], dns.ClassToString[q.Qclass]}

		h.Jobs++
		switch {
		case myQ.qclass == "IN":
			h.Resolver.Lookup(nil, w, req)
		case myQ.qclass == "CH", myQ.qtype == "TXT":
			m := handleChaos(req)
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

func handleChaos(req *dns.Msg) (m *dns.Msg) {
	reqQ := req.Question[0]
	gq := question{reqQ.Name, dns.TypeToString[reqQ.Qtype], dns.ClassToString[reqQ.Qclass]}

	m = new(dns.Msg)
	m.SetReply(req)
	hdr := dns.RR_Header{
		Name:   gq.qname,
		Rrtype: dns.TypeTXT,
		Class:  dns.ClassCHAOS,
		Ttl:    0,
	}
	switch gq.qname {
	case "authors.bind.":
		m.Answer = append(m.Answer, &dns.TXT{
			Hdr: hdr,
			Txt: []string{"Nagy Kroly Gabriel <k@jpi.io>"}})
	case "version.bind.", "version.server.":
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: hdr,
			Txt: []string{"Version " + version.Version + " built on " + version.BuildDate}}}
	case "hostname.bind.", "id.server.":
		m.Answer = []dns.RR{&dns.TXT{
			Hdr: hdr,
			Txt: []string{"localhost"}}}
	default:
		m.SetRcode(req, dns.RcodeNotImplemented)
	}
	return m
}
