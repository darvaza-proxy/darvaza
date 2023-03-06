package gnocco

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/darvaza-proxy/slog"
	"github.com/miekg/dns"
)

type resolver struct {
	Resolvers []net.IP
	Iterative bool
	roots     []string
	Logger    slog.Logger
}

func (cf *Gnocco) newResolver() *resolver {
	resolvers := make([]net.IP, 0)

	f, err := os.Open("/etc/resolv.conf")
	defer f.Close()

	if err != nil {
		cf.logger.Warn().Printf("Error %s occurred.", err)
	}

	scan := bufio.NewScanner(f)

	for scan.Scan() {
		line := scan.Text()
		line = strings.TrimSpace(line)
		if len(line) > 0 && line[1] != '#' {
			fields := strings.Fields(line)
			if len(fields) > 0 && fields[0] == "nameserver" {
				myIP := net.ParseIP(fields[1])
				if myIP != nil {
					resolvers = append(resolvers, myIP)
				}
			}
		}
	}

	// load the roots in the slice
	fr, err := os.Open(cf.RootsFile)
	defer fr.Close()
	if err != nil {
		cf.logger.Fatal().Printf("cannot load roots file %s", cf.RootsFile)
	}
	rrs := make([]string, 0)
	scan = bufio.NewScanner(fr)
	for scan.Scan() {
		lf := strings.TrimSpace(scan.Text())
		if len(lf) > 0 {
			ff := strings.Fields(lf)
			if len(ff) > 0 {
				rrs = append(rrs, ff[1])
			}
		}
	}

	resolver := &resolver{
		Resolvers: resolvers,
		Iterative: cf.IterateResolv,
		Logger:    cf.logger,
		roots:     rrs,
	}
	return resolver
}

func (r *resolver) Lookup(_ *cache, w dns.ResponseWriter, req *dns.Msg) {
	if r.Iterative {
		root := randomfromslice(r.roots)
		r.Logger.Info().Printf("using root %s", root)
		root = root + ":53"
		qn := dns.Fqdn(req.Question[0].Name)
		qt := req.Question[0].Qtype
		resp, err := Iterate(qn, qt, root)
		if err == nil {
			resp.SetReply(req)
			w.WriteMsg(resp)
		}
	} else {
		var ip net.IP
		if len(r.Resolvers) > 1 {
			ip = r.Resolvers[randint(len(r.Resolvers))]
		} else {
			ip = r.Resolvers[0]
		}
		resp, err := dns.Exchange(req, net.JoinHostPort(ip.String(), "53"))
		if err != nil {
			r.Logger.Error().Print(err)
		} else {
			w.WriteMsg(resp)
		}
	}
}

// Iterate will perform an itarative query with given name ant type
// starting with given nameserver
func Iterate(name string, qtype uint16, server string) (*dns.Msg, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(name, qtype)
	msg.RecursionDesired = false
	client := &dns.Client{Timeout: 5 * time.Second}
	client.Net = "tcp"
	resp, _, err := client.Exchange(msg, server)
	if err != nil {
		return nil, err
	}
	if resp.Truncated {
		return nil, fmt.Errorf("dns response was truncated")
	}
	if resp.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("dns response error: %s", dns.RcodeToString[resp.Rcode])
	}

	if len(resp.Answer) == 0 && len(resp.Ns) > 0 {
		// We are in referal mode, get the next server
		nextServer := ""
		for _, ns := range resp.Ns {
			if ns.Header().Rrtype == dns.TypeNS {
				nextServer = strings.TrimSuffix(ns.(*dns.NS).Ns, ".")
				break
			}
		}
		if nextServer == "" {
			return nil, fmt.Errorf("no authoritative server found in referral")
		}
		// Begin again with new forces
		newMsg := new(dns.Msg)
		newMsg.SetQuestion(name, dns.TypeNS)
		newMsg.RecursionDesired = false
		newMsg.Question[0].Qclass = dns.ClassINET
		newMsg.Question[0].Name = name

		client := &dns.Client{Timeout: 5 * time.Second}
		client.Net = "tcp"
		rsp, _, err := client.Exchange(newMsg, nextServer+":53")
		if err != nil {
			return nil, err
		}
		if rsp.Truncated {
			return nil, fmt.Errorf("dns response was truncated")
		}
		if rsp.Rcode != dns.RcodeSuccess {
			return nil, fmt.Errorf("dns response error %s", dns.RcodeToString[rsp.Rcode])
		}
		return Iterate(name, qtype, nextServer+":53")
	}
	// We got an answer
	return resp, nil
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
