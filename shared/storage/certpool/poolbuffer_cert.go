package certpool

import (
	"crypto/x509"

	"darvaza.org/core"
	"darvaza.org/x/tls/x509utils"
	"darvaza.org/x/tls/x509utils/certpool"
)

type pbCertData struct {
	Filename string
	Cert     *x509.Certificate

	Hash certpool.Hash
	Pub  x509utils.PublicKey
}

func (pb *PoolBuffer) addCertUnlocked(fn string, cert *x509.Certificate) error {
	if pb.index == nil {
		pb.index = make(map[certpool.Hash]*pbCertData)
	}

	if cert == nil {
		return nil
	}

	if err := pb.printCert(fn, cert); err != nil {
		return err
	}

	hash, ok := certpool.HashCert(cert)
	if !ok {
		return core.Wrap(core.ErrInvalid, "bad cert")
	}

	if _, ok := pb.index[hash]; !ok {
		// new cert
		pb.index[hash] = &pbCertData{
			Filename: fn,
			Cert:     cert,
			Pub:      cert.PublicKey.(x509utils.PublicKey),
		}

		pb.addCertToPools(cert)
	}

	return nil
}

func (pb *PoolBuffer) addCertToPools(cert *x509.Certificate) {
	if cert.IsCA {
		if x509utils.IsSelfSigned(cert) {
			pb.roots.AddCert(cert)
		}
		pb.inter.AddCert(cert)
	} else {
		pb.certs.AddCert(cert)
	}
}

func (pb *PoolBuffer) findByPublic(pub x509utils.PublicKey) []*pbCertData {
	var out []*pbCertData
	for _, cd := range pb.index {
		if pub.Equal(cd.Pub) {
			out = append(out, cd)
		}
	}
	return out
}
