// Package tlsfwd is this and that
package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/darvaza-proxy/darvaza/agent/tlsfwd"
)

// PortsFlag is a storage for port related flag
type PortsFlag struct {
	Ports []int
}

// String returns the string representaion of the PortsFlag
func (p *PortsFlag) String() string {
	return fmt.Sprint(p.Ports)
}

// Set will set a value on the PortsFlag
func (p *PortsFlag) Set(v string) error {
	vv := strings.Split(v, ",")
	for _, vg := range vv {
		pv, err := strconv.Atoi(vg)
		if err != nil {
			return err
		}
		p.Ports = append(p.Ports, pv)
	}
	return nil
}

func main() {
	// By default we listen on port 443 on TCP and UDP
	var ports PortsFlag
	flag.Var(&ports, "p", "Comma separated list of additional ports -p 8080,9090")

	usRedirect := flag.Bool("u", false, "Redirect port HTTP on 80 to HTTPS on 443")
	flag.Parse()

	ports.Ports = append(ports.Ports, 443)
	ports.Ports = unique(ports.Ports)

	s, err := tlsfwd.NewTLSFwd(ports.Ports)
	if err != nil {
		log.Fatal(err)
	}
	if *usRedirect {
		s, err = tlsfwd.NewTLSFwd(ports.Ports, tlsfwd.SetRedir)
		if err != nil {
			log.Fatal(err)
		}
	}
	s.Serve()
}

func unique(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
