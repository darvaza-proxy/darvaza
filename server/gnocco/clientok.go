package gnocco

import (
	"net"
	"os"
	"strings"
)

// ClientOK checks if the given IP address is good
func (cf *Gnocco) ClientOK(ip net.IP) bool {
	result := false
	permdir := cf.PermissionsDir
	// if we do not have the permissions directory than
	// everybody is allowed
	if _, err := os.Stat(permdir); os.IsNotExist(err) {
		result = true
		return result
	}
	var ipsep string
	if ip.To4() != nil {
		ipsep = "."
	} else {
		ipsep = ":"
	}
	ipsslc := strings.Split(ip.String(), ipsep)

	psep := os.PathSeparator
	tail := permdir + string(psep)

	for i, v := range ipsslc {
		if i == 0 {
			tail = tail + v
		} else {
			tail = tail + ipsep + v
		}
		if _, err := os.Stat(tail); !os.IsNotExist(err) {
			result = true
			break
		}
	}

	return result
}
