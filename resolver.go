package main

import (
	"bufio"
	"fmt"
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

type reser struct {
	ip string //net.IP
}

type Resolver struct {
	RootZone  NSZone
	Resolvers map[int]reser
	Safe      bool
}

var ips []string

func initResolver() *Resolver {
	Resolver := new(Resolver)
	resolvers := make(map[int]reser)
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

	f, err := os.Open("/etc/resolv.conf")
	defer f.Close()

	if err != nil {
		logger.Warn("Error %s occured.", err)
	}

	scan := bufio.NewScanner(f)

	i := 0
	for scan.Scan() {
		var re reser
		fields := strings.Fields(scan.Text())
		if fields[0] == "nameserver" {
			re.ip = net.ParseIP(fields[1]).String()
			i++
			resolvers[i] = re
		}
	}

	Resolver.Resolvers = resolvers
	Resolver.RootZone = roots
	Resolver.Safe = Config.SafeResolv
	return Resolver

}

func (r *Resolver) LookupGen(w dns.ResponseWriter, req *dns.Msg) {
	if r.Safe {
		ips = []string{}
		ips = r.getIp(req.Question[0].Name, emptyZone, "NS")
		var ip string
		if len(ips) > 0 {
			ip = randomfromslice(ips)
		}
		if ip != "" {
			r.lookup(w, req, net.JoinHostPort(ip, "53"))
		}
	} else {
		ip := r.Resolvers[randint(len(r.Resolvers))].ip
		r.lookup(w, req, net.JoinHostPort(ip, "53"))
	}
}

func (r *Resolver) lookup(w dns.ResponseWriter, req *dns.Msg, ns string) {
	cl := new(dns.Client)
	req.RecursionDesired = !r.Safe

	response, _, err := cl.Exchange(req, ns)

	if err != nil {
		logger.Error("Error %s", err)

	} else {

		w.WriteMsg(response)
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

func (r *Resolver) getIp(host string, zn NSZone, tip string) []string {
	if zn.isEmpty() {
		zn = r.RootZone
	}
	tmp := getRandomNsIpFromZone(zn)

	nsrv := net.ParseIP(tmp[0]).To4().String()
	if nsrv != "" {
		nsrv = net.JoinHostPort(nsrv, "53")
	}

	m1 := new(dns.Msg)
	m1.Id = dns.Id()
	m1.RecursionDesired = false
	m1.Question = make([]dns.Question, 1)

	switch {
	case tip == "NS":
		m1.Question[0] = dns.Question{dns.Fqdn(host), dns.TypeNS, dns.ClassINET}
	case tip == "A":
		m1.Question[0] = dns.Question{dns.Fqdn(host), dns.TypeA, dns.ClassINET}
	case tip == "AAAA":
		m1.Question[0] = dns.Question{dns.Fqdn(host), dns.TypeAAAA, dns.ClassINET}
	}

	in, err := dns.Exchange(m1, nsrv)

	if err != nil {
		logger.Error("Error %s occured.", err)
	}

	result := make([]string, 0)

	ans, ref, ns := processResponse(in)

	switch {
	case len(ans) > 0:
		switch tip {
		case "A", "AAAA":
			for _, a := range ans {
				ips = append(result, a.Value)
			}
		case "NS":
			//we got an authoritative NS with no glue
			for _, q := range ans {
				_ = r.getIp(q.Value, emptyZone, "A")

			}
		}

	case len(ns) == 0 && !ref.isEmpty():
		_ = r.getIp(host, ref, tip)

	//case len(ans) == 0 && ref.isEmpty() && len(ns) > 0:
	case len(ns) > 0:
		if ref.isEmpty() {
			for _, n := range ns {
				nip := r.getIp(n.Value, emptyZone, "A")
				if len(nip) > 0 {
					ips = append(result, nip[0])
				}
			}
		} else {
			//create NSzone from ref and NS then return NSips.
		}
	default:
		fmt.Println("Unexpected")
	}
	return ips

}

func processResponse(resp *dns.Msg) (RRs, NSZone, RRs) {
	var answer RRs
	var referal NSZone
	var ns RRs

	if resp == nil {
		return nil, NSZone{}, nil
	}

	rCode := resp.Rcode
	switch rCode {
	case dns.RcodeSuccess:
		switch getMsgType(resp) {
		case "Referral":
			referal = processReferal(resp)
		case "Answer":
			answer = processAnswer(resp)
		case "Nameserver":
			ns = processNS(resp)
		default:
			fmt.Println(resp)
		}
	case dns.RcodeNameError: //NXDOMAIN
	default:
		fmt.Println("Unexpected DNS code", dns.RcodeToString[rCode])

	}
	return answer, referal, ns
}

func processReferal(res *dns.Msg) NSZone {
	var result NSZone
	result = makeNsZone(res)
	return result
}

func processAnswer(res *dns.Msg) RRs {
	result := mdRRtoRRs(res.Answer)
	return result
}

func processNS(res *dns.Msg) RRs {
	var result RRs
	result = mdRRtoRRs(res.Ns)
	return result
}
func getMsgType(m *dns.Msg) string {
	result := ""
	if m != nil {
		switch {
		case len(m.Answer) == 0 && len(m.Ns) > 0 && len(m.Extra) > 1: //Referral
			result = "Referral"
		case len(m.Answer) == 0 && len(m.Extra) > 1: //Nameserver
			result = "Nameserver"
		case len(m.Answer) > 0: //Answer
			result = "Answer"
		default:
			result = "Unknown"
		}
	}
	return result
}
