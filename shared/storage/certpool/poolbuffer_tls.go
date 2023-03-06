package certpool

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/darvaza-proxy/core"
	"github.com/darvaza-proxy/darvaza/shared/x509utils"
	"github.com/darvaza-proxy/slog"
)

func (pb *PoolBuffer) newBundler(base x509utils.CertPooler) (*Bundler, error) {
	if base == nil {
		var err error
		base, err = SystemCertPool()
		if err != nil {
			return nil, err
		}
	}

	b := &Bundler{
		Roots: base,
		Inter: pb.roots.Pool(),
	}

	return b, nil
}

type pbPair struct {
	cert *pbCertData
	key  *pbKeyData
	err  error
}

func (pb *PoolBuffer) warnPair(p pbPair, msg string) {
	if log, ok := pb.warn(); ok {
		fields := slog.Fields{
			"public": p.key.Public(),
		}

		if p.key.filename != "" {
			fields["filename"] = p.key.filename
		}

		log.WithFields(fields).Print(msg)
	}
}

func (pb *PoolBuffer) errCertPair(cd *pbCertData, err error, msg string) error {
	var log slog.Logger
	var ok bool

	if err != nil {
		log, ok = pb.error(err)
		err = core.Wrapf(err, "%q: %s", cd.Filename, msg)
	} else {
		log, ok = pb.warn()
		err = fmt.Errorf("%q: %s", cd.Filename, msg)
	}

	if ok {
		fields := slog.Fields{
			"subject-id": string(cd.Cert.SubjectKeyId),
		}
		if cd.Filename != "" {
			fields["filename"] = cd.Filename
		}

		log.WithFields(fields).Print(msg)
	}

	return err
}

func (pb *PoolBuffer) errKeyPair(pk *pbKeyData, err error, msg string) pbPair {
	var log slog.Logger
	var ok bool

	if err != nil {
		log, ok = pb.error(err)
	} else {
		log, ok = pb.warn()
	}

	if ok {
		fields := slog.Fields{
			"public": pk.Public(),
		}

		if pk.filename != "" {
			fields["filename"] = pk.filename
		}

		log.WithFields(fields).Print(msg)
	}

	if err != nil {
		err = fmt.Errorf("%q: %s", pk.filename, msg)
	}

	return pbPair{
		key: pk,
		err: err,
	}
}

func (pb *PoolBuffer) appendErrKeyNoCerts(out []pbPair, pk *pbKeyData) []pbPair {
	p := pb.errKeyPair(pk, nil, "certificate not found")
	return append(out, p)
}

func (pb *PoolBuffer) appendErrKeyValidate(out []pbPair, pk *pbKeyData, err error) []pbPair {
	p := pb.errKeyPair(pk, err, "invalid key")
	return append(out, p)
}

func (*PoolBuffer) appendPairs(out []pbPair, pk *pbKeyData, certs []*pbCertData) []pbPair {
	for _, cert := range certs {
		out = append(out, pbPair{
			key:  pk,
			cert: cert,
		})
	}
	return out
}

func (pb *PoolBuffer) pairs() []pbPair {
	var out []pbPair

	core.ListForEach(pb.keys.keys, func(pk *pbKeyData) bool {
		pub := pk.Public()

		if err := pk.Validate(); err != nil {
			// invalid key
			out = pb.appendErrKeyValidate(out, pk, err)
			return false
		}

		// Certificates with matching Public Key
		certs := pb.certs.findByPublic(pub)
		if len(certs) == 0 {
			// try CAs
			certs = pb.roots.findByPublic(pub)
		}

		if len(certs) == 0 {
			// certificate not found
			out = pb.appendErrKeyNoCerts(out, pk)
		}

		// append pairs
		out = pb.appendPairs(out, pk, certs)
		return false
	})

	return out
}

// Certificates exports all the Certificates it contains bundled considering
// a given base
func (pb *PoolBuffer) Certificates(base x509utils.CertPooler) ([]*tls.Certificate, error) {
	certs, errors := pb.CertificatesErrors(base)
	if len(errors) > 0 {
		// return first error and bundled certs
		return certs, errors[0]
	}
	return certs, nil
}

// revive:disable:cognitive-complexity
// revive:disable:cyclomatic

// CertificatesErrors exports all the Certificates it contains bundled considering
// a given base, but collecting all errors
func (pb *PoolBuffer) CertificatesErrors(base x509utils.CertPooler) (
	out []*tls.Certificate, errors []error) {
	// revive:enable:cognitive-complexity
	// revive:enable:cyclomatic

	b, err := pb.newBundler(base)
	if err != nil {
		return nil, []error{err}
	}

	// Pairs
	pairs := pb.pairs()

	// deduplication
	certs := make(map[Hash]bool)

	// pairs
	for _, pair := range pairs {
		var err error

		switch {
		case pair.err != nil:
			// invalid key
			err = pair.err
		case pair.cert == nil:
			// missing cert
			panic("unreachable")
		case certs[pair.cert.Hash]:
			// duplicate
			pb.warnPair(pair, "duplicated key")
		default:
			var crt *tls.Certificate

			crt, err = pb.bundlePair(b, pair.cert, pair.key)
			if crt != nil {
				// success
				certs[pair.cert.Hash] = true
				out = append(out, crt)
			}
		}

		if err != nil {
			errors = append(errors, err)
		}
	}

	// keyless certificates
	for _, cert := range pb.certs.certs {
		if _, known := certs[cert.Hash]; !known {
			crt, err := pb.bundlePair(b, cert, nil)
			if crt != nil {
				out = append(out, crt)
			}
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	return out, errors
}

func (pb *PoolBuffer) bundlePair(b *Bundler, cd *pbCertData, kd *pbKeyData) (
	*tls.Certificate, error) {
	//
	var cert *x509.Certificate
	var pk x509utils.PrivateKey

	if cd != nil {
		cert = cd.Cert
	}
	if kd != nil {
		pk = kd.pk
	}

	crt, err := b.Bundle(cert, pk)
	if err == nil {
		return crt, nil
	}
	// failed to bundle
	err = pb.errCertPair(cd, err, "failed to bundle")
	return nil, err
}

// Bundle verifies a leaf x509.Certificate and return a tls.Certificate
func (pb *PoolBuffer) Bundle(cert *x509.Certificate, key x509utils.PrivateKey,
	roots x509utils.CertPooler) (*tls.Certificate, error) {
	return pb.roots.Pool().Bundle(cert, key, roots)
}
