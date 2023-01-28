package certpool

import (
	"crypto/x509"
	"encoding/pem"
	"sync"

	"github.com/darvaza-proxy/darvaza/shared/x509utils"
	"github.com/darvaza-proxy/slog"
)

// PoolBuffer is a CertPool in the making
type PoolBuffer struct {
	mu     sync.Mutex
	logger slog.Logger

	roots pbCerts
	certs pbCerts
	keys  pbKeys
}

// Reset makes the PoolBuffer go back to its initial state, empty
func (pb *PoolBuffer) Reset() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.roots.Reset()
	pb.certs.Reset()
	pb.keys.Reset()
}

// AddKey adds a PrivateKey to the PoolBuffer
func (pb *PoolBuffer) AddKey(fn string, pk x509utils.PrivateKey) error {
	var err error

	if pk != nil {
		pb.mu.Lock()
		defer pb.mu.Unlock()

		err = pb.addKeyUnlocked(fn, pk)
	}
	return err
}

// AddCert adds a Certificate to the PoolBuffer
func (pb *PoolBuffer) AddCert(fn string, cert *x509.Certificate) error {
	var err error

	if cert != nil {
		pb.mu.Lock()
		defer pb.mu.Unlock()

		err = pb.addCertUnlocked(fn, cert)
	}
	return err
}

// Add loads private keys and certificates from PEM files, directories, and direct text
func (pb *PoolBuffer) Add(s string) error {
	var readErr, addErr error

	pb.mu.Lock()
	defer pb.mu.Unlock()

	readErr = x509utils.ReadStringPEM(s, func(fn string, block *pem.Block) bool {
		if err := pb.addBlock(fn, block); err != nil {
			addErr = err
			return true // abort
		}

		return false // continue
	})

	if readErr != nil {
		return readErr
	}
	return addErr
}

func (pb *PoolBuffer) addBlock(fn string, block *pem.Block) error {
	if pk, err := x509utils.BlockToPrivateKey(block); pk != nil {
		// PrivateKey
		return pb.addKeyUnlocked(fn, pk)
	} else if err != x509utils.ErrIgnored {
		// Bad PrivateKey
		return err
	}

	if cert, err := x509utils.BlockToCertificate(block); cert != nil {
		// Certificate
		return pb.addCertUnlocked(fn, cert)
	} else if err != x509utils.ErrIgnored {
		// Bad Certificate
		return err
	}

	// Ignore other blocks
	return nil
}

// Count returns how many certificates are in the buffer
func (pb *PoolBuffer) Count() int {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	return pb.roots.Count() + pb.certs.Count()
}

// Export returns a new x509.CertPool with the CA certificates
func (pb *PoolBuffer) Export() *x509.CertPool {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	return pb.roots.Export()
}

// Pool returns a new CertPool with the CA certificates
func (pb *PoolBuffer) Pool() *CertPool {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	return pb.roots.Pool().Clone().(*CertPool)
}
