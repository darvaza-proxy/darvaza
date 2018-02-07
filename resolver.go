package main

import (
	"bufio"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type reser struct {
	ip string //net.IP
}

type resolver struct {
	Resolvers map[int]reser
	Iterative bool
}

func initResolver() *resolver {
	resolver := new(resolver)
	resolvers := make(map[int]reser)

	f, err := os.Open("/etc/resolv.conf")
	defer f.Close()

	if err != nil {
		logger.Warn("Error %s occurred.", err)
	}

	scan := bufio.NewScanner(f)

	i := 0
	for scan.Scan() {
		var re reser
		fields := strings.Fields(scan.Text())
		if len(fields) > 0 && fields[0] == "nameserver" {
			re.ip = net.ParseIP(fields[1]).String()
			i++
			resolvers[i] = re
		}
	}

	resolver.Resolvers = resolvers
	resolver.Iterative = mainconfig.IterateResolv
	return resolver

}

func (r *resolver) Lookup(c *cache, w dns.ResponseWriter, req *dns.Msg) {
	if r.Iterative {
		qn := req.Question[0].Name
		qt := dns.TypeToString[req.Question[0].Qtype]

		result := new(dns.Msg)
		result.SetReply(req)
		slist := newStack()
		slist.push(qn, qt)

		for !slist.isEmpty() {
			r.Iterate(c, qn, qt, slist)

			if rcs, err := c.get(c.makeKey(qn, qt)); err == nil {
				for _, z := range rcs.Value {
					rc, _ := dns.NewRR(qn + " " + qt + " " + z)
					result.Answer = append(result.Answer, rc)
				}
				w.WriteMsg(result)
			} else {
				if crcs, err := c.get(c.makeKey(qn, "CNAME")); err == nil {
					if a, err := c.get(c.makeKey(crcs.Value[0], "A")); err == nil {
						for _, rrc := range a.Value {
							ff, _ := dns.NewRR(qn + " " + qt + " " + rrc)
							result.Answer = append(result.Answer, ff)
						}
						w.WriteMsg(result)
					} else {
						logger.Error("%s", err)
						result.SetRcode(req, 4)
						w.WriteMsg(result)
					}
				}
			}
		}
	} else {
		ip := r.Resolvers[randint(len(r.Resolvers))].ip
		r.lookup(w, req, net.JoinHostPort(ip, "53"))
	}

}

func (r *resolver) lookup(w dns.ResponseWriter, req *dns.Msg, ns string) {

	response, err := dns.Exchange(req, ns)

	if err != nil {
		logger.Error("Error %s", err)

	} else {
		w.WriteMsg(response)
	}
}

func (r *resolver) Iterate(c *cache, qname string, qtype string, slist *stack) {
	//We arrived here because original question was not in cache
	//so we iterate to find a nameserver for our question.

	//Down the rabbit hole!
	ancestor := getParentinCache(qname, c)
	if nstoask, ancerr := c.get(ancestor + "/NS"); ancerr == nil {
		qq := randomfromslice(nstoask.Value)
		if ns, nserr := c.get(dns.Fqdn(qq) + "/A"); nserr == nil {
			ans := getans(qname, qtype, ns.Value[0])

			switch typify(ans) {
			case "Delegation":
				//FIXME: Check bailiwick and observe that Extra can have fewer or more
				//records than NS
				c.set(ans.Ns[0].Header().Name, "NS", ans)

				ex := mdRRtoRRs(ans.Extra)
				for _, x := range ex {
					c.setVal(x.Name, x.Type, x.TTL, x.Value)
				}
				r.Iterate(c, qname, qtype, slist)
			case "Namezone":
				c.set(ans.Ns[0].Header().Name, "NS", ans)
				t := mdRRtoRRs(ans.Ns)
				for _, z := range t {
					slist.push(dns.Fqdn(z.Value), "A")
					r.Iterate(c, dns.Fqdn(z.Value), "A", slist)
				}
			case "Answer":
				c.set(ans.Answer[0].Header().Name, qtype, ans)
				slist.popFor(ans.Answer[0].Header().Name, qtype)
				if slist.isEmpty() {
					break
				} else {
					qn, qt := slist.pop()
					r.Iterate(c, qn, qt, slist)
				}
			case "Cname":
				c.set(ans.Answer[0].Header().Name, "CNAME", ans)
				slist.popFor(ans.Answer[0].Header().Name, qtype)
				v := mdRRtoRRs(ans.Answer)
				for _, g := range v {
					slist.push(dns.Fqdn(g.Value), "A")
					r.Iterate(c, dns.Fqdn(g.Value), "A", slist)
				}

			default:
				logger.Error(typify(ans))
			}
		} else {
			slist.push(qq, "A")
			r.Iterate(c, qq, "A", slist)
		}
	}

}

func getans(qname string, qtype string, nameserver string) (result *dns.Msg) {
	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = false
	m.Question = make([]dns.Question, 1)
	qqt, _ := dns.StringToType[qtype]
	m.Question[0] = dns.Question{qname, qqt, dns.ClassINET}
	result, err := dns.Exchange(m, nameserver+":53")
	if err != nil {
		logger.Error("%s", err)
	}
	return

}

func getParentinCache(domain string, c *cache) string {
	result := domain
	x := dns.SplitDomainName(domain)

	for i := 0; i < len(x); i++ {
		result = strings.TrimPrefix(result, x[i]+".")
		if _, err := c.get(result + "/NS"); err == nil {
			break
		}
	}

	//We ALLWAYS have root.
	if result == "" {
		result = "."
	}
	return result
}

func typify(m *dns.Msg) string {
	if m != nil {
		switch m.Rcode {
		case dns.RcodeSuccess:
			if len(m.Answer) > 0 {
				if m.Answer[0].Header().Rrtype == dns.TypeCNAME {
					return "Cname"
				}
				return "Answer"
			}

			ns := 0
			for _, r := range m.Ns {
				if r.Header().Rrtype == dns.TypeNS {
					ns++
				}
			}
			if ns > 0 && ns == len(m.Ns) {
				if len(m.Extra) < 2 {
					return "Namezone"
				}
				return "Delegation"
			}
		case dns.RcodeRefused:
			return "Refused"
		case dns.RcodeFormatError:
			return "NoEDNS"
		default:
			return "Unknown"
		}
	}
	return "Nil message"
}

func randint(upper int) int {
	var result int
	switch upper {
	case 0, 1:
		result = upper
	default:
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		result = rnd.Intn(upper)
	}
	return result
}

func randomfromslice(s []string) string {
	var result string

	switch len(s) {
	case 0:
		result = ""
	case 1:
		result = s[0]
	default:
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		id := rnd.Intn(len(s))
		result = s[id]
	}
	return result
}
