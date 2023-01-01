package certpool

import (
	"encoding/pem"
	"io/fs"
	"sync"

	"github.com/darvaza-proxy/darvaza/shared/x509utils"
)

var (
	rootsMutex     sync.Mutex
	systemRoots    *CertPool
	systemRootsErr error
)

// SystemCertPool returns a Pool populated with the
// system's valid CA certificates
func SystemCertPool() (*CertPool, error) {
	rootsMutex.Lock()
	defer rootsMutex.Unlock()

	if systemRootsErr != nil {
		return nil, systemRootsErr
	} else if systemRoots != nil {
		return systemRoots.Clone().(*CertPool), nil
	} else {
		// first call
		roots, err := loadSystemRoots()
		if err != nil {
			// memoize error
			systemRootsErr = err
			return nil, err
		}

		// memoize roots
		systemRoots = roots
		return roots.Clone().(*CertPool), nil
	}
}

// revive:disable:cognitive-complexity
func loadSystemRoots() (*CertPool, error) {
	var pool CertPool
	var count int
	var err error

	fn := func(_ string, block *pem.Block) bool {
		cert, err := x509utils.BlockToCertificate(block)
		if err == nil && pool.AddCert(cert) {
			count++
		}
		return false
	}

	for _, f := range certFiles {
		err = x509utils.ReadStringPEM(f, fn)
		if count > 0 {
			// stop after finding one
			break
		}
	}

	for _, d := range certDirectories {
		x509utils.ReadStringPEM(d, fn)
	}

	if count > 0 {
		return &pool, nil
	}

	if err == nil {
		err = fs.ErrNotExist
	}
	return nil, err
}
