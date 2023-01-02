// Package x509utils provides abstract access to collections of x509 certificates
package x509utils

import (
	"context"
	"crypto/x509"
)

// StoreIterFunc defines the type of function passed to ReadStore.ForEach
type StoreIterFunc func(*x509.Certificate) error

// ReadStore represents the methods to access a x509 Store
type ReadStore interface {
	Get(ctx context.Context, name string) (*x509.Certificate, error)
	ForEach(ctx context.Context, f StoreIterFunc) error
}

// WriteStore represents the methods to alter a x509 Store
type WriteStore interface {
	Put(ctx context.Context, name string, cert *x509.Certificate) error
	Delete(ctx context.Context, name string) error
	DeleteCert(ctx context.Context, cert *x509.Certificate) error
}

// CertPooler represents the read-only interface of our CertPool
type CertPooler interface {
	ReadStore

	Clone() CertPooler
	Export() *x509.CertPool
}

// CertPoolWriter represents the write-only interface of our CertPool
type CertPoolWriter interface {
	WriteStore

	AddCert(cert *x509.Certificate) bool
	AppendCertsFromPEM(b []byte) bool
}
