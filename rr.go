package main

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// A DNS resource record.
type RR struct {
	Name  string
	Ttl   int
	Class string
	Type  string
	Value string
}

type RRs []RR

// String returns a string representation of an RR in zone-file format.
func (rr *RR) String() string {
	return rr.Name + "\t      " + fmt.Sprint(rr.Ttl) + "\t" + rr.Class + "\t" + rr.Type + "\t" + rr.Value
}

func sliceToString(sli []string) string {
	result := ""
	if sli != nil {
		for _, s := range sli {
			result += s + "\n"
		}
	}
	return result
}

func dRRtoRR(drr dns.RR) (RR, bool) {
	if drr != nil {
		switch t := drr.(type) {
		case *dns.SOA:
			return RR{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "SOA", dns.Fqdn(strings.ToLower(t.Ns))}, true
		case *dns.NS:
			return RR{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "NS", dns.Fqdn(strings.ToLower(t.Ns))}, true
		case *dns.CNAME:
			return RR{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "CNAME", dns.Fqdn(strings.ToLower(t.Target))}, true
		case *dns.A:
			return RR{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "A", t.A.String()}, true
		case *dns.AAAA:
			return RR{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "AAAA", t.AAAA.String()}, true
		case *dns.TXT:
			return RR{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "TXT", strings.Join(t.Txt, "\t")}, true
		case *dns.MX:
			return RR{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "MX", t.Mx}, true
		}
	}

	return RR{}, false

}

func mdRRtoRRs(mdrr []dns.RR) RRs {
	result := make([]RR, 0)

	if mdrr != nil {
		for _, d := range mdrr {
			if rr, ok := dRRtoRR(d); ok {
				result = append(result, rr)
			}
		}
	}
	return result
}
