package gnocco

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// A DNS resource record.
type rr struct {
	Name  string
	TTL   int
	Class string
	Type  string
	Value string
}

type rrs []rr

// String returns a string representation of an RR in zone-file format.
func (arr *rr) String() string {
	return arr.Name + "\t      " + fmt.Sprint(arr.TTL) + "\t" + arr.Class + "\t" + arr.Type + "\t" + arr.Value
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

func dRRtoRR(drr dns.RR) (rr, bool) {
	if drr != nil {
		switch t := drr.(type) {
		case *dns.SOA:
			return rr{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "SOA", dns.Fqdn(strings.ToLower(t.Ns))}, true
		case *dns.NS:
			return rr{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "NS", dns.Fqdn(strings.ToLower(t.Ns))}, true
		case *dns.CNAME:
			return rr{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "CNAME", dns.Fqdn(strings.ToLower(t.Target))}, true
		case *dns.A:
			return rr{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "A", t.A.String()}, true
		case *dns.AAAA:
			return rr{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "AAAA", t.AAAA.String()}, true
		case *dns.TXT:
			return rr{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "TXT", strings.Join(t.Txt, "\t")}, true
		case *dns.MX:
			return rr{dns.Fqdn(strings.ToLower(t.Hdr.Name)), int(t.Hdr.Ttl), dns.ClassToString[t.Hdr.Class], "MX", t.Mx}, true
		}
	}

	return rr{}, false

}

func mdRRtoRRs(mdrr []dns.RR) rrs {
	result := make([]rr, 0)

	if mdrr != nil {
		for _, d := range mdrr {
			if rr, ok := dRRtoRR(d); ok {
				result = append(result, rr)
			}
		}
	}
	return result
}
