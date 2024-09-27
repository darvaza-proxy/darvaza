package gnocco

import (
	"github.com/miekg/dns"

	"darvaza.org/slog"

	"darvaza.org/darvaza/shared/version"
)

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

func (h *gnoccoHandler) do(w dns.ResponseWriter, req *dns.Msg) {
	var err error

	if h.Jobs < h.MaxJobs {
		q := req.Question[0]
		qType := dns.TypeToString[q.Qtype]
		qClass := dns.ClassToString[q.Qclass]

		h.Jobs++
		switch {
		case qClass == "IN":
			h.Resolver.Lookup(w, req)
		case qClass == "CH", qType == "TXT":
			m := handleChaos(req)
			err = w.WriteMsg(m)
		default:
		}
		h.Jobs--
	} else {
		m := new(dns.Msg)
		m.SetRcode(req, dns.RcodeServerFailure)
		err = w.WriteMsg(m)
	}

	if err != nil {
		h.logger.Error().Print(err)
	}
}

func handleChaos(req *dns.Msg) (m *dns.Msg) {
	reqQ := req.Question[0]
	qName := reqQ.Name

	m = new(dns.Msg)
	m.SetReply(req)
	hdr := dns.RR_Header{
		Name:   qName,
		Rrtype: dns.TypeTXT,
		Class:  dns.ClassCHAOS,
		Ttl:    0,
	}
	switch qName {
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
