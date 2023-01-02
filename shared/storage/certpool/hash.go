package certpool

import (
	"crypto/x509"

	"github.com/zeebo/blake3"
)

const (
	// HashSize is the number of bytes of HashCert's output
	HashSize = 32
)

// Hash is a blake3.Sum256 representation of a DER encoded certificate
type Hash [HashSize]byte

// HashCert produces a blake3 unkeyed digest of the DER representation of a Certificate
func HashCert(cert *x509.Certificate) Hash {
	return blake3.Sum256(cert.Raw)
}
