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

type root struct {
	name string
	ip4  string //net.IP
	ip6  string //net.IP
}

var storage string

type Iterator struct {
	RootZone    NSZone
	CurrentZone NSZone
	Question    *dns.Msg
}

func initIterator() *Iterator {
	z := new(Iterator)
	roots := initNSZone()

	if Config.RootsFile == "" {
		logger.Fatal("Config.RootsFile is empty :(")
	}
	fl, err := os.Open(Config.RootsFile)
	defer fl.Close()
	if err != nil {
		logger.Fatal("Error %s", err)
	}
	scanner := bufio.NewScanner(fl)
	for scanner.Scan() {
		flds := strings.Fields(scanner.Text())
		roots.Name = "."

		roots.Nslist[flds[0]] = append(roots.Nslist[flds[0]], net.ParseIP(flds[1]).String())
		if len(flds) > 2 {
			roots.Nslist[flds[0]] = append(roots.Nslist[flds[0]], net.ParseIP(flds[2]).String())
		}
	}
	z.RootZone = roots
	return z
}

func (i *Iterator) Iterate(req *dns.Msg) *dns.Msg {
	if i.CurrentZone.isEmpty() {
		i.CurrentZone = i.RootZone
	}
	tmp := getRandomNsIpFromZone(i.CurrentZone)

	nsrv := net.ParseIP(tmp[0]).To4().String()
	if nsrv != "" {
		nsrv = net.JoinHostPort(nsrv, "53")
	}

	answer, err := dns.Exchange(req, nsrv)
	if err != nil {
		logger.Error("Error %s occured.", err)
	}

	useGlue := true
	t := typify(answer)

	switch {
	case t == "Answer":
		return answer
	case t == "Delegation":
		if useGlue {
			ns := makeNsZone(answer)
			i.CurrentZone = ns
			i.Iterate(req)
		}
	case t == "Namezone":
		s := []string{}
		rs := mdRRtoRRs(answer.Ns)
		for _, r := range rs {
			s = append(s, r.Value)
		}
		ip := i.iterateforip(randomfromslice(s), emptyZone)
		ans, err := dns.Exchange(req, net.JoinHostPort(ip, "53"))
		if err == nil {
			if ans.Rcode == dns.RcodeSuccess {
				return ans
			} else {
				logger.Error("I got response: %s", ans.Rcode)
			}
		} else {
			logger.Error("Error: ", err)
		}

	default:
		logger.Error("Untypyfied message: %s", answer)
	}
	//We should not get here!
	return nil
}

func (i *Iterator) iterateforip(name string, ns NSZone) (ip string) {
	var tmp []string

	if ns.isEmpty() {
		tmp = getRandomNsIpFromZone(i.RootZone)
	} else {
		tmp = getRandomNsIpFromZone(ns)
	}

	nsrv := net.ParseIP(tmp[0]).To4().String()

	if nsrv != "" {
		nsrv = net.JoinHostPort(nsrv, "53")
	}

	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = false
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{name, dns.TypeA, dns.ClassINET}
	answer, err := dns.Exchange(m, nsrv)
	if err != nil {
		logger.Error("Error %s occured.", err)
	}

	t := typify(answer)

	switch {
	case t == "Answer":
		ip = mdRRtoRRs(answer.Answer)[0].Value
	case t == "Nil":
		logger.Error("I sent Nil to typify")
	default:
		ip = i.iterateforip(name, makeNsZone(answer))
	}

	return
}

func typify(m *dns.Msg) string {
	if m != nil {
		if len(m.Answer) > 0 && m.Rcode == dns.RcodeSuccess {
			return "Answer"
		}
		ns := 0
		for _, r := range m.Ns {
			if r.Header().Rrtype == dns.TypeNS {
				ns++
			}
		}
		if ns > 0 && ns == len(m.Ns) && m.Rcode == dns.RcodeSuccess {
			if len(m.Extra) < 2 {
				return "Namezone"
			} else {
				return "Delegation"
			}
		}
		return "Unknown"
	} else {
		return "Nil"
	}
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
