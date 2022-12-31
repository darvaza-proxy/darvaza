package x509utils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
)

var (
	_ PrivateKey = (*rsa.PrivateKey)(nil)
	_ PrivateKey = (*ecdsa.PrivateKey)(nil)
	_ PrivateKey = (*ed25519.PrivateKey)(nil)

	_ PublicKey = (*rsa.PublicKey)(nil)
	_ PublicKey = (*ecdsa.PublicKey)(nil)
	_ PublicKey = (*ed25519.PublicKey)(nil)
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
