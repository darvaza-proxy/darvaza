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
	return fmt.Sprintf("%s \t %s \t %s \t %s \t %s",
		arr.Name, fmt.Sprint(arr.TTL), arr.Class, arr.Type, arr.Value)
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
		rname := dns.Fqdn(strings.ToLower(drr.Header().Name))
		rttl := int(drr.Header().Ttl)
		rcls := dns.ClassToString[drr.Header().Class]
		switch t := drr.(type) {
		case *dns.SOA:
			return rr{rname,
				rttl,
				rcls,
				"SOA",
				dns.Fqdn(strings.ToLower(t.Ns))}, true
		case *dns.NS:
			return rr{rname,
				rttl,
				rcls,
				"NS",
				dns.Fqdn(strings.ToLower(t.Ns))}, true
		case *dns.CNAME:
			return rr{rname,
				rttl,
				rcls,
				"CNAME",
				dns.Fqdn(strings.ToLower(t.Target))}, true
		case *dns.A:
			return rr{rname,
				rttl,
				rcls,
				"A",
				t.A.String()}, true
		case *dns.AAAA:
			return rr{rname,
				rttl,
				rcls,
				"AAAA",
				t.AAAA.String()}, true
		case *dns.TXT:
			return rr{rname,
				rttl,
				rcls,
				"TXT",
				strings.Join(t.Txt, "\t")}, true
		case *dns.MX:
			return rr{rname,
				rttl,
				rcls,
				"MX",
				t.Mx}, true
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
