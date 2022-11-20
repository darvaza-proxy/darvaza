// This program generates docs/roots. It can be invoked by running
// go generate
package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/miekg/dns"
)

type ips struct {
	ip4 net.IP
	ip6 net.IP
}

var myroots = map[string]*ips{}

const (
	url       = "https://www.internic.net/domain/named.root"
	rootsfile = "doc/roots"
)

func trimDot(s string) string {
	if strings.HasSuffix(s, ".") {
		s = s[:len(s)-1]
	}
	return s
}

func main() {
	rsp, _ := http.Get(url)
	defer rsp.Body.Close()

	xroots, _ := ioutil.ReadAll(rsp.Body)

	zp := dns.NewZoneParser(strings.NewReader(string(xroots)), "", "")

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		dom := trimDot(strings.ToLower(rr.Header().Name))
		if dom != "" {
			switch tt := rr.(type) {
			case *dns.A:
				z4, ok := myroots[dom]
				if !ok {
					z4 = &ips{}
				}
				z4.ip4 = tt.A
				myroots[dom] = z4
			case *dns.AAAA:
				z6, ok := myroots[dom]
				if !ok {
					z6 = &ips{}
				}
				z6.ip6 = tt.AAAA
				myroots[dom] = z6
			}
		}
	}
	f, err := os.Create(rootsfile)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	sortedKeys := make([]string, 0, len(myroots))

	for k := range myroots {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		f.WriteString(k + " " + myroots[k].ip4.String() + " " + myroots[k].ip6.String() + "\n")
	}
	f.Sync()
}
