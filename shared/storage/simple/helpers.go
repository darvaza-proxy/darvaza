package simple

import (
	"bytes"
	"container/list"
	"crypto/tls"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/storage/certpool"
	"darvaza.org/darvaza/shared/x509utils"
)

// FindSupportedInMap attempts to find a matching supported tls.Certificate
// on a MapList
func FindSupportedInMap(chi *tls.ClientHelloInfo,
	name string, m map[string]*list.List) *tls.Certificate {
	//
	var out *tls.Certificate

	if name == "" {
		// no sanitied name provided, produce one
		s, ok := x509utils.SanitiseName(chi.ServerName)
		if !ok {
			return nil
		}
		name = s
	}

	core.MapListForEach(m, name, func(c *tls.Certificate) bool {
		if err := chi.SupportsCertificate(c); err == nil {
			out = c
		}

		// stop on the first supported match
		return out != nil
	})

	return out
}

// FindInMap attempts to find matching [tls.Certificate]s on a MapList
func FindInMap(name string, m map[string]*list.List, once bool) []*tls.Certificate {
	var out []*tls.Certificate

	core.MapListForEach(m, name, func(c *tls.Certificate) bool {
		if c != nil {
			out = append(out, c)
			return once
		}
		core.Panic("unreachable")
		return false
	})

	return out
}

func mapListContainsHash(m map[string]*list.List, name string, hash certpool.Hash) bool {
	var found bool

	core.MapListForEach(m, name, func(c *tls.Certificate) bool {
		h := certpool.HashCert(c.Leaf)
		if bytes.Equal(hash[:], h[:]) {
			found = true
		}
		return found
	})

	return found
}

// PrivateKeyEqual tells if two private keys are the same
func PrivateKeyEqual(a, b x509utils.PrivateKey) bool {
	return a.Equal(b)
}
