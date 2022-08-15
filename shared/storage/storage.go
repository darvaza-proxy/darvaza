package storage

import (
	"context"
	"crypto/x509"
)

type StoreIterFunc func(x509.Certificate) error

type ReadStore interface {
	Get(ctx context.Context, name string) (x509.Certificate, error)
	ForEach(ctx context.Context, f StoreIterFunc) error
}

type WriteStore interface {
	Put(ctx context.Context, name string, cert x509.Certificate) error
	Delete(ctx context.Context, name string) error
	DeleteCert(ctx context.Context, cert x509.Certificate) error
}
