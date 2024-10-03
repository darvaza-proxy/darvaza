package simple

import (
	"container/list"
	"crypto/tls"
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

// FindSupportedInMap attempts to find a matching supported tls.Certificate
// on a MapList
func FindSupportedInMap(chi *tls.ClientHelloInfo,
	name string, m map[string]*list.List) *tls.Certificate {
	//
	var out *tls.Certificate

	if name == "" {
		// no sanitized name provided, produce one
		s, ok := x509utils.SanitizeName(chi.ServerName)
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
		if hash.EqualCert(c.Leaf) {
			found = true
		}
		return found
	})

	return found
}

// PairMatch tells if the public key of a PrivateKey is the
// same as included in a *x509.Certificate
func PairMatch(cert *x509.Certificate, pk x509utils.PrivateKey) bool {
	if pub, ok := pk.Public().(x509utils.PublicKey); ok {
		return pub.Equal(cert.PublicKey)
	}
	return false
}

// PrivateKeyEqual tells if two private keys are the same
func PrivateKeyEqual(a, b x509utils.PrivateKey) bool {
	return a.Equal(b)
}
