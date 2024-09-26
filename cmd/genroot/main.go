// This program generates docs/roots. It can be invoked by running
// go generate
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"darvaza.org/core"

	"github.com/miekg/dns"
)

type ips struct {
	ip4 string
	ip6 string
}

const url = "https://www.internic.net/domain/named.root"

func trimDot(s string) string {
	return strings.TrimSuffix(s, ".")
}

func main() {
	var out *os.File
	var err error

	if len(os.Args) > 1 && os.Args[1] != "-" {
		fileName := os.Args[1]
		out, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
	} else {
		out = os.Stdout
	}

	if err == nil {
		err = genRoots(out)
	}

	if err != nil {
		panic(err)
	}
}

// revive:disable:cognitive-complexity
func genRoots(rootsfile *os.File) error {
	rsp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer rsp.Body.Close()

	xroots, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	if len(xroots) == 0 {
		_, _ = fmt.Println("no roots found in the download")
	}

	myroots := buildMyRoots(xroots)

	for _, k := range core.SortedKeys(myroots) {
		strToWrite := fmt.Sprintf("%s %s %s \n", k, myroots[k].ip4, myroots[k].ip6)
		if _, err2 := rootsfile.WriteString(strToWrite); err2 != nil {
			return err
		}
	}
	if rootsfile.Name() != os.Stdout.Name() {
		if err = rootsfile.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func buildMyRoots(b []byte) map[string]ips {
	myroots := make(map[string]ips)
	zp := dns.NewZoneParser(strings.NewReader(string(b)), "", "")
	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		dom := trimDot(strings.ToLower(rr.Header().Name))
		if dom != "" {
			switch tt := rr.(type) {
			case *dns.A:
				z4, ok := myroots[dom]
				if !ok {
					z4 = ips{}
				}
				z4.ip4 = tt.A.String()
				myroots[dom] = z4
			case *dns.AAAA:
				z6, ok := myroots[dom]
				if !ok {
					z6 = ips{}
				}
				z6.ip6 = tt.AAAA.String()
				myroots[dom] = z6
			}
		}
	}
	return myroots
}

//revive:enable:cognitive-complexity
