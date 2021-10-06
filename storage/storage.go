package storage

import (
	"context"
	"crypto/x509"
	"errors"
)

// ErrStorageMiss is returned when a certificate is not found in the storage.
var ErrStorageMiss = errors.New("darvaza/: certificate storage miss")

type StoreIterFunc func(x509.Certificate) error

type ReadStore interface {
	Get(ctx context.Context, name string) ([]byte, error)
	ForEach(ctx context.Context, f StoreIterFunc) error
}

type WriteStore interface {
	Put(ctx context.Context, name string, cert []byte) error
	Delete(ctx context.Context, name string) error
	DeleteCert(ctx context.Context, cert []byte) error
}
