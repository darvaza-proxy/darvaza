// This program generates docs/roots. It can be invoked by running
// go generate
package main

import (
	"io"
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

const url = "https://www.internic.net/domain/named.root"

func trimDot(s string) string {
	return strings.TrimSuffix(s, ".")
}

func main() {
	var out *os.File

	if len(os.Args) > 1 && os.Args[1] != "-" {
		var err error
		fileName := os.Args[1]
		out, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
	} else {
		out = os.Stdout
	}

	getRoots(out)

}

func getRoots(rootsfile *os.File) {

	var myroots = map[string]*ips{}

	rsp, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	defer rsp.Body.Close()

	xroots, err := io.ReadAll(rsp.Body)

	if err != nil {
		panic(err)
	}

	if len(xroots) == 0 {
		panic("no roots found in the download")
	}

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
	sortedKeys := make([]string, 0, len(myroots))

	for k := range myroots {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		_, err := rootsfile.WriteString(k + " " + myroots[k].ip4.String() + " " + myroots[k].ip6.String() + "\n")
		if err != nil {
			panic(err)
		}
	}
	if rootsfile.Name() != os.Stdout.Name() {
		if err = rootsfile.Sync(); err != nil {
			panic(err)
		}
	}
}
