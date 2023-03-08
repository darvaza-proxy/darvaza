package x509utils

import (
	"bytes"
	"crypto/x509"
)

// IsSelfSigned tests if a certificate corresponds to a self-signed CA
func IsSelfSigned(c *x509.Certificate) bool {
	if c != nil && c.IsCA && bytes.Equal(c.RawSubject, c.RawIssuer) {
		pool := x509.NewCertPool()

		pool.AddCert(c)
		_, err := c.Verify(x509.VerifyOptions{
			Roots: pool,
		})

		return err == nil
	}
	return false
}
