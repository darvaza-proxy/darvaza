// Package file provides a Storage implementation for storing x509 certificates as files
package file

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/fs"

	"darvaza.org/darvaza/shared/os"
	"darvaza.org/darvaza/shared/os/flock"
	"darvaza.org/darvaza/shared/x509utils"
)

var (
	_ x509utils.ReadStore  = (*Store)(nil)
	_ x509utils.WriteStore = (*Store)(nil)
)

// Store is a darvaza Storage implementation for storing x509 certificates as files
type Store struct {
	fl flock.Options
}

// Get will return the first x509 certificate and an error, the certificate having the same
// common name as the name parameter
func (m *Store) Get(_ context.Context, name string) (*x509.Certificate, error) {
	_, cert, err := m.fileCertFromName(name)
	return cert, err
}

// ForEach will walk the store and ececute the StoreIterFunc for each certificate
// it can decode
func (m *Store) ForEach(_ context.Context, f x509utils.StoreIterFunc) error {
	return m.forEachCert(func(_ string, cert *x509.Certificate) bool {
		if err := f(cert); err != nil {
			// TODO: terminate or just continue?
			return true
		}

		// next
		return false
	})
}

// Put will create a file in the store with the given name and the given certificate
// it will write in the fil eteh content of cert.Raw field
func (m *Store) Put(_ context.Context, name string, cert *x509.Certificate) error {
	return m.fl.WriteFile(name, cert.Raw, 0666)
}

// Delete will delete the first certificate with the same common name as
// the given parameter
func (m *Store) Delete(_ context.Context, name string) error {
	file, _, err := m.fileCertFromName(name)
	if err != nil {
		return err
	}

	err = os.Remove(file)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return err
}

// DeleteCert will delete from the store the certificate given as parameter
func (m *Store) DeleteCert(_ context.Context, cert *x509.Certificate) error {
	name := cert.Subject.CommonName
	file, _, err := m.fileCertFromName(name)
	if err != nil {
		return err
	}

	err = os.Remove(file)
	if err == nil || os.IsNotExist(err) {
		return nil
	}
	return err
}

// Options is a struct containing the storage options
// in the case of FileStorage it only contains a directory name
// and the file mode
type Options struct {
	Directory string
	DirMode   fs.FileMode
}

// DefaultOptions for FileStorage Options
var DefaultOptions = Options{
	Directory: "darvaza_store",
	DirMode:   0700,
}

// NewStore will create a new File Storage. If no options
// are given it will use the DefaultOptions
func NewStore(options Options) (Store, error) {
	result := Store{}

	if options.Directory == "" {
		options.Directory = DefaultOptions.Directory
	}

	if options.DirMode == 0 {
		options.DirMode = DefaultOptions.DirMode
	}

	fl := flock.Options{
		Base:    options.Directory,
		Create:  true,
		DirMode: options.DirMode,
	}

	err := fl.MkdirBase(0)
	if err != nil {
		return result, err
	}

	result.fl = fl
	return result, nil
}

func (m *Store) fileCertFromName(name string) (string, *x509.Certificate, error) {
	var match *x509.Certificate
	var filename string

	fn := func(fl string, cert *x509.Certificate) bool {
		if verifyHostname(cert, name) {
			match = cert
			filename = fl
			return true // term
		}

		return false
	}

	_ = m.forEachCert(fn)
	if match != nil {
		return filename, match, nil
	}
	return "", nil, fmt.Errorf("certificate not found")
}

func verifyHostname(cert *x509.Certificate, name string) bool {
	if len(cert.URIs) == 0 {
		// an "old" certificate, no SAN
		return cert.Subject.CommonName == name
	}

	return cert.VerifyHostname(name) == nil
}

func (m *Store) forEachCert(fn func(string, *x509.Certificate) bool) error {
	if fn == nil {
		return nil
	}

	return x509utils.ReadStringPEM(m.fl.Base, func(fl string, block *pem.Block) bool {
		if cert, _ := x509utils.BlockToCertificate(block); cert != nil {
			return fn(fl, cert)
		}

		return false
	})
}
