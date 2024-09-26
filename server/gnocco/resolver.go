package gnocco

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"

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

	ccfg, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		cf.logger.Warn().Print("attempting to parse resolv.conf resulted in %s", err)
	}
	for _, s := range ccfg.Servers {
		myIP := net.ParseIP(s)
		if myIP != nil {
			resolvers = append(resolvers, myIP)
		}
	}

	// load the roots in the slice
	rrs, err := loadFileColumn(cf.RootsFile, 1)
	if err != nil {
		cf.logger.Fatal().Print("attempting to parse %s resulted in %s", cf.RootsFile, err)
	}
	resolver := &resolver{
		Resolvers: resolvers,
		Iterative: cf.IterateResolv,
		Logger:    cf.logger,
		roots:     rrs,
	}
	return resolver
}

func loadFileColumn(file string, column int) ([]string, error) {
	result := make([]string, 0)
	f, err := os.Open(file)
	if err != nil {
		return result, nil
	}

	defer unsafeClose(f)

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		lf := strings.TrimSpace(scan.Text())
		if len(lf) > 0 {
			ff := strings.Fields(lf)
			if len(ff) > 0 {
				result = append(result, ff[column])
			}
		}
	}
	return result, nil
}

func (r *resolver) Lookup(_ *cache, w dns.ResponseWriter, req *dns.Msg) {
	var err error

	if r.Iterative {
		err = r.doInteractiveLookup(w, req)
	} else {
		err = r.doLookup(w, req)
	}

	if err != nil {
		r.Logger.Error().Print(err)
	}
}

func (r *resolver) doInteractiveLookup(w dns.ResponseWriter, req *dns.Msg) error {
	root, _ := core.SliceRandom(r.roots)
	r.Logger.Info().Printf("using root %s", root)
	root = root + ":53"
	qn := dns.Fqdn(req.Question[0].Name)
	qt := req.Question[0].Qtype

	resp, err := r.Iterate(qn, qt, root)
	if err != nil {
		return err
	}

	resp.SetReply(req)
	return w.WriteMsg(resp)
}

func (r *resolver) doLookup(w dns.ResponseWriter, req *dns.Msg) error {
	var ip net.IP
	if len(r.Resolvers) > 1 {
		ip = r.Resolvers[randint(len(r.Resolvers))]
	} else {
		ip = r.Resolvers[0]
	}
	resp, err := dns.Exchange(req, net.JoinHostPort(ip.String(), "53"))
	if err != nil {
		return err
	}

	return w.WriteMsg(resp)
}

// revive:disable:cognitive-complexity
// Iterate will perform an itarative query with given name ant type
// starting with given nameserver
func (r *resolver) Iterate(name string, qtype uint16, server string) (*dns.Msg, error) {
	// revive:enable:cognitive-complexity
	msg := newMsgFromParts(name, qtype)
	resp, _, err := clientTalk(msg, server)
	if werr := validateResp(resp, err); werr != nil {
		return nil, err
	}

	if len(resp.Answer) == 0 && len(resp.Ns) > 0 {
		// We are in referral mode, get the next server
		nextServer := make([]string, 0)
		for _, ns := range resp.Ns {
			if ns.Header().Rrtype == dns.TypeNS {
				nextServer = append(nextServer, strings.TrimSuffix(ns.(*dns.NS).Ns, "."))
			}
		}
		if len(nextServer) == 0 {
			return nil, fmt.Errorf("no authoritative server found in referral")
		}
		// Begin again with new forces
		newMsg := newMsgFromParts(name, dns.TypeNS)
		newMsg.Question[0].Qclass = dns.ClassINET

		nns, _ := core.SliceRandom(nextServer)
		r.Logger.Info().Printf("using server %s", nns)
		rsp, _, err := clientTalk(newMsg, nns+":53")
		if wwerr := validateResp(rsp, err); wwerr != nil {
			return nil, err
		}
		return r.Iterate(name, qtype, nns+":53")
	}
	// We got an answer
	return resp, nil
}

func newMsgFromParts(qName string, qType uint16) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(qName, qType)
	msg.RecursionDesired = false
	return msg
}

func validateResp(r *dns.Msg, err error) error {
	if err != nil {
		return err
	}
	if r.Truncated {
		return fmt.Errorf("dns response was truncated")
	}
	if r.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("dns response error: %s", dns.RcodeToString[r.Rcode])
	}
	return nil
}

func clientTalk(msg *dns.Msg, server string) (r *dns.Msg, rtt time.Duration, err error) {
	client := &dns.Client{}
	client.Net = "tcp"

	return client.Exchange(msg, server)
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

func unsafeClose(f io.Closer) {
	_ = f.Close()
}
