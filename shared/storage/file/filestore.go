package file

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/darvaza-proxy/darvaza/shared/x509utils"
)

// FileStore is a darvaza Storage implementation for storing x509 certificates as files
type FileStore struct {
	locksLock *sync.Mutex
	fileLocks map[string]*sync.RWMutex
	directory string
}

// Get will return the first x509 certificate and an error, the certificate having the same
// common name as the name parameter
func (fs *FileStore) Get(ctx context.Context, name string) (*x509.Certificate, error) {
	_, cert, err := fs.fileCertFromName(name)
	return cert, err

}

// ForEach will walk the store and ececute the StoreIterFunc for each certificate
// it can decode
func (fs *FileStore) ForEach(ctx context.Context, f x509utils.StoreIterFunc) error {
	files, err := ioutil.ReadDir(fs.directory)
	if err != nil {
		return err
	}
	fs.locksLock.Lock()
	defer fs.locksLock.Unlock()
	for _, file := range files {
		fl := filepath.Join(fs.directory, file.Name())
		lock := fs.fsLock(fl)
		lock.RLock()
		content, err := os.ReadFile(fl)
		lock.RUnlock()
		if err != nil {
			return err
		}
		block, _ := pem.Decode(content)
		if block == nil {
			return fmt.Errorf("failed to decode data")
		}
		if block.Type == "CERTIFICATE" {
			x, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return err
			}
			err = f(x)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Put will create a file in the store with the given name and the given certificate
// it will write in the fil eteh content of cert.Raw field
func (fs *FileStore) Put(ctx context.Context, name string, cert *x509.Certificate) error {
	file := filepath.Join(fs.directory, name)
	lock := fs.fsLock(file)
	lock.Lock()
	defer lock.Unlock()

	err := os.WriteFile(file, cert.Raw, 0666)
	if err != nil {
		return err
	}
	return nil
}

// Delete will delete the first certificate with the same common name as
// the given parameter
func (fs *FileStore) Delete(ctx context.Context, name string) error {
	file, _, err := fs.fileCertFromName(name)
	if err != nil {
		return err
	}
	lock := fs.fsLock(file)
	lock.Lock()
	defer lock.Unlock()
	err = os.Remove(file)
	if os.IsNotExist(err) {
		return nil
	}
	return err

}

// DeleteCert will delete from the store the certificate given as parameter
func (fs *FileStore) DeleteCert(ctx context.Context, cert *x509.Certificate) error {
	name := cert.Subject.CommonName
	file, _, err := fs.fileCertFromName(name)
	if err != nil {
		return err
	}
	lock := fs.fsLock(file)
	lock.Lock()
	defer lock.Unlock()
	err = os.Remove(file)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Options is a struct containing the storage options
// in the case of FileStorage it only contains a directory name
// and the file mode
type Options struct {
	Directory string
	FMode     os.FileMode
}

// DefaultOptions for FileStorage Options
var DefaultOptions = Options{
	Directory: "darvaza_store",
	FMode:     0700,
}

// NewStore will create a new File Storage. If no options
// are given it will use the DefaultOptions
func NewStore(options Options) (FileStore, error) {
	result := FileStore{}

	if options.Directory == "" {
		options.Directory = DefaultOptions.Directory
	}
	if options.FMode == 0 {
		options.FMode = DefaultOptions.FMode
	}

	err := os.MkdirAll(options.Directory, options.FMode)
	if err != nil {
		return result, err
	}

	result.directory = options.Directory
	result.locksLock = new(sync.Mutex)
	result.fileLocks = make(map[string]*sync.RWMutex)
	return result, nil
}

func (fs FileStore) fsLock(filename string) *sync.RWMutex {
	fs.locksLock.Lock()
	lock, found := fs.fileLocks[filename]
	if !found {
		lock = new(sync.RWMutex)
		fs.fileLocks[filename] = lock
	}
	fs.locksLock.Unlock()
	return lock
}

func (fs FileStore) fileCertFromName(name string) (string, *x509.Certificate, error) {
	files, err := ioutil.ReadDir(fs.directory)
	if err != nil {
		return "", nil, err
	}

	for _, file := range files {
		fl := filepath.Join(fs.directory, file.Name())
		lock := fs.fsLock(fl)
		lock.RLock()
		content, err := os.ReadFile(fl)
		lock.RUnlock()
		if err != nil {
			return "", nil, err
		}
		block, _ := pem.Decode(content)
		if block == nil {
			return "", nil, fmt.Errorf("failed to decode data")
		}
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				continue
			}
			switch len(cert.URIs) {
			case 0:
				//an "old" certificate, no SAN
				if cert.Subject.CommonName == name {
					return fl, cert, nil

				}
			default:
				//normal "modern" certificate uses SAN
				err := cert.VerifyHostname(name)
				if err == nil {
					return fl, cert, nil
				}

			}
		}
	}
	return "", nil, fmt.Errorf("certificate not found")

}
