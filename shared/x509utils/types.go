package x509utils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
)

var (
	_ PrivateKey = (*rsa.PrivateKey)(nil)
	_ PrivateKey = (*ecdsa.PrivateKey)(nil)
	_ PrivateKey = (*ed25519.PrivateKey)(nil)

	_ PublicKey = (*rsa.PublicKey)(nil)
	_ PublicKey = (*ecdsa.PublicKey)(nil)
	_ PublicKey = (*ed25519.PublicKey)(nil)
)

var (
	// ErrIgnored is used when we ask the user to try a different function instead
	ErrIgnored = errors.New("type of value out of scope")
)

// PrivateKey implements what crypto.PrivateKey should have
type PrivateKey interface {
	Public() crypto.PublicKey
	Equal(x crypto.PrivateKey) bool
}

// PublicKey implements what crypto.PublicKey should have
type PublicKey interface {
	Equal(x crypto.PublicKey) bool
}

// BlockToPrivateKey parses a pem Block looking for rsa, ecdsa or ed25519 Private Keys
func BlockToPrivateKey(block *pem.Block) (PrivateKey, error) {
	if block.Type == "PRIVATE KEY" || strings.HasSuffix(block.Type, " PRIVATE KEY") {
		if pk, _ := x509.ParsePKCS1PrivateKey(block.Bytes); pk != nil {
			// *rsa.PrivateKey
			return pk, nil
		}

		pk, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		return pk.(PrivateKey), nil
	}

	return nil, ErrIgnored
}

// BlockToRSAPrivateKey attempts to parse a pem.Block to extract an rsa.PrivateKey
func BlockToRSAPrivateKey(block *pem.Block) (*rsa.PrivateKey, error) {
	pk, err := BlockToPrivateKey(block)
	if err != nil {
		return nil, err
	}

	if key, ok := pk.(*rsa.PrivateKey); ok {
		return key, nil
	}

	return nil, ErrIgnored
}

// BlockToCertificate attempts to parse a pem.Block to extract a x509.Certificate
func BlockToCertificate(block *pem.Block) (*x509.Certificate, error) {
	if block.Type == "CERTIFICATE" {
		if cert, err := x509.ParseCertificate(block.Bytes); err != nil {
			return nil, err
		} else if cert != nil {
			return cert, nil
		}

		panic("unreachable")
	}
	return nil, ErrIgnored
}
