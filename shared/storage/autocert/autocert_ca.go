package autocert

import (
	"context"
	"crypto/tls"

	"darvaza.org/darvaza/shared/storage/simple"
	"darvaza.org/darvaza/shared/x509utils"
)

func (*Store) issueCertificate(context.Context,
	x509utils.PrivateKey, string) (*tls.Certificate, error) {
	// TODO: Implement
	return nil, simple.ErrNotImplemented
}
