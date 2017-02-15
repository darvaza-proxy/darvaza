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
	ip4  net.IP
	ip6  net.IP
}

type reser struct {
	ip net.IP
}

var Roots map[int]root

type Resolver struct {
	Resolvers map[int]reser
	Safe      bool
}

func initResolver() {
	if Config.RootsFile == "" {
		logger.Fatal("Config.RootsFile is empty :(")
	}
	fl, err := os.Open(Config.RootsFile)
	defer fl.Close()
	if err != nil {
		logger.Fatal("Error %s occured.", err)
	}
	scanner := bufio.NewScanner(fl)
	Roots = make(map[int]root)
	for scanner.Scan() {
		var rt root
		flds := strings.Fields(scanner.Text())
		rt.name = flds[0]
		rt.ip4 = net.ParseIP(flds[1])
		if len(flds) > 2 {
			rt.ip6 = net.ParseIP(flds[2])
		} else {
			rt.ip6 = nil
		}
		// A is ascii 65, B is 66 etc.
		id := int([]byte(rt.name)[0] - 65)
		Roots[id] = rt
	}
	Resolver := new(Resolver)

	f, err := os.Open("/etc/resolv.conf")
	defer f.Close()

	if err != nil {
		logger.Warn("Error %s occured.", err)
	}

	scan := bufio.NewScanner(f)
	resolvers := make(map[int]reser)

	i := 0
	for scan.Scan() {
		var re reser
		fields := strings.Fields(scan.Text())
		if fields[0] == "nameserver" {
			re.ip = net.ParseIP(fields[1])
			i++
			resolvers[i] = re
		}
	}
	Resolver.Resolvers = resolvers
	Resolver.Safe = Config.SafeResolv

}

func lookup(w dns.ResponseWriter, req *dns.Msg, ns string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if ns == "" {
		ns = net.JoinHostPort(Roots[r.Intn(len(Roots))].ip4.String(), "53")
	}
	cl := new(dns.Client)

	response, _, _ := cl.Exchange(req, ns)

	if len(response.Answer) == 0 {
		if len(response.Ns) != 0 {
			ns, _ := response.Ns[r.Intn(len(response.Ns))].(*dns.NS)
			lookup(w, req, fmt.Sprintf("%s:53", ns.Ns[0:len(ns.Ns)-1]))
		}
	} else {
		w.WriteMsg(response)
	}
}
