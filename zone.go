package main

import (
	"math/rand"
	"net"
	"time"

	"github.com/miekg/dns"
)

var emptyZone = Zone{}

type Zone struct {
	Name   string
	Nslist map[string][]string
}

func (z *Zone) isEmpty() bool {
	if z.Name == "" && len(z.Nslist) == 0 {
		return true
	}
	return false
}

func (z *Zone) addNs(ns string, ip string) {
	s := net.ParseIP(ip)
	if s != nil && ns != "" && !z.isEmpty() {
		z.Nslist[ns] = append(z.Nslist[ns], ip)
	}
}

func initZone() Zone {
	var result Zone
	result.Nslist = make(map[string][]string)
	return result
}
func makeNsZone(msg *dns.Msg) Zone {
	if msg != nil {
		rs := mdRRtoRRs(msg.Ns)

		result := Zone{msg.Ns[0].Header().Name, make(map[string][]string)}
		for _, r := range rs {
			result.Nslist[r.Value] = make([]string, 0)
		}
		if len(msg.Extra) > 1 {
			ex := mdRRtoRRs(msg.Extra)
			for _, x := range ex {
				result.addNs(x.Name, x.Value)
			}
		} else {
		}
		return result
	}
	return emptyZone
}

func getRandomNsIpFromZone(zn Zone) []string {
	rand.Seed(time.Now().UnixNano())
	var result []string
	i := int(float32(len(zn.Nslist)) * rand.Float32())
	for _, v := range zn.Nslist {
		if i != 0 {
			i--
		}
		result = v
	}
	return result
}
