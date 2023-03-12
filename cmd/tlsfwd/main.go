// Package tlsfwd is this and that
package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/darvaza-proxy/core"
	"github.com/darvaza-proxy/darvaza/agent/tlsfwd"
)

// PortsFlag is a storage for secure port related flag
type PortsFlag struct {
	Ports []uint16
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
		p.Ports = append(p.Ports, uint16(pv))
	}
	return nil
}

func main() {
	var securePorts PortsFlag
	flag.Var(&securePorts, "s", "Comma separated list of secure ports -s 443,8080,9090")

	usRedirect := flag.Int("u", -1, "Redirect this HTTP port to HTTPS on 443")
	flag.Parse()

	securePorts.Ports = core.SliceUniquify(&securePorts.Ports)

	s, err := tlsfwd.NewTLSFwd(securePorts.Ports)
	if err != nil {
		log.Fatal(err)
	}
	if *usRedirect >= 0 {
		s, err = tlsfwd.NewTLSFwd(securePorts.Ports, tlsfwd.Redir(*usRedirect))
		if err != nil {
			log.Fatal(err)
		}
	}
	s.Serve()
}
