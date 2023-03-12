package certpool

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

	"github.com/darvaza-proxy/darvaza/shared/x509utils"
)

var (
	_ x509utils.Bundler = (*CertPool)(nil)
)

// Bundler uses two CertPoolers to bundler keys and certificates
type Bundler struct {
	Roots x509utils.CertPooler
	Inter x509utils.CertPooler
}

// Bundle bundles a key and a certificate into a *tls.Certificate
func (b *Bundler) Bundle(cert *x509.Certificate, key x509utils.PrivateKey) (
	*tls.Certificate, error) {
	var opts x509.VerifyOptions
	if b.Roots != nil {
		opts.Roots = b.Roots.Export()
	}
	if b.Inter != nil {
		opts.Intermediates = b.Inter.Export()
	}

	chains, err := cert.Verify(opts)
	if err != nil {
		return nil, err
	}
	chain := rawBestChain(chains)
	if len(chain) == 0 {
		err := errors.New("couldn't verify")
		return nil, err
	}

	out := &tls.Certificate{
		Certificate: chain,
		PrivateKey:  key,
		Leaf:        cert,
	}

	return out, nil
}

func rawBestChain(chain [][]*x509.Certificate) [][]byte {
	var best [][]byte

	for _, ch := range chain {
		// build list of raw certificates
		if option, ok := copyRawChain(ch); ok {
			if l := len(best); l == 0 || l > len(option) {
				best = option
			}
		}
	}
	return best
}

func copyRawChain(chain []*x509.Certificate) ([][]byte, bool) {
	out := make([][]byte, 0, len(chain))
	for _, cert := range chain {
		if len(cert.Raw) == 0 {
			// bad chain
			return nil, false
		}
		out = append(out, cert.Raw)
	}
	return out, true
}

// Bundle verifies a leaf x509.Certificate and return a tls.Certificate
func (pool *CertPool) Bundle(cert *x509.Certificate, key x509utils.PrivateKey,
	roots x509utils.CertPooler) (*tls.Certificate, error) {
	if roots == nil {
		var err error
		roots, err = SystemCertPool()
		if err != nil {
			return nil, err
		}
	}

	b := Bundler{
		Roots: roots,
		Inter: pool,
	}

	return b.Bundle(cert, key)
}
