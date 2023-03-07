package certpool

import (
	"crypto/x509"

	"github.com/darvaza-proxy/darvaza/shared/x509utils"
)

type pbCertData struct {
	Filename string
	Cert     *x509.Certificate

	Hash Hash
	Pub  x509utils.PublicKey
}

func (pb *PoolBuffer) addCertUnlocked(fn string, cert *x509.Certificate) error {
	if pb.index == nil {
		pb.index = make(map[Hash]*pbCertData)
	}

	if cert == nil {
		return nil
	}

	if err := pb.printCert(fn, cert); err != nil {
		return err
	}

	hash := HashCert(cert)
	if _, ok := pb.index[hash]; !ok {
		// new cert
		pb.index[hash] = &pbCertData{
			Filename: fn,
			Cert:     cert,
			Pub:      cert.PublicKey.(x509utils.PublicKey),
		}

		pb.addCertToPools(hash, cert)
	}

	return nil
}

func (pb *PoolBuffer) addCertToPools(hash Hash, cert *x509.Certificate) {
	if cert.IsCA {
		if x509utils.IsSelfSigned(cert) {
			pb.roots.addCertUnsafe(hash, "", cert)
		}
		pb.inter.addCertUnsafe(hash, "", cert)
	} else {
		pb.certs.addCertUnsafe(hash, "", cert)
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
