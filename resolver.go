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

type Resolver struct {
	Iterator  *Iterator
	Resolvers map[int]reser
	Safe      bool
}

func initResolver() *Resolver {
	Resolver := new(Resolver)
	resolvers := make(map[int]reser)

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
	Resolver.Iterator = initIterator()
	Resolver.Safe = Config.SafeResolv
	return Resolver

}

func (r *Resolver) LookupGen(w dns.ResponseWriter, req *dns.Msg) {
	if r.Safe {
		z := r.Iterator.Iterate(req)
		if z != nil {
			w.WriteMsg(z)
		} else {
			logger.Error("Iterator resulted in nil response")
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
