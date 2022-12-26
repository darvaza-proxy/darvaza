package x509utils

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/fs"
	"path/filepath"

	"github.com/darvaza-proxy/darvaza/shared/os"
	"github.com/pkg/errors"
)

// DecodePEMBlockFunc is called for each PEM block coded. it returns true
// to terminate the loop
type DecodePEMBlockFunc func(filename string, block *pem.Block) bool

// ReadPEM invoques a callback for each PEM block found
// it can receive raw PEM data, a filename or a directory to scan
func ReadPEM(s string, cb DecodePEMBlockFunc) error {
	if s == "" {
		// empty
		return nil
	} else if block, rest := pem.Decode([]byte(s)); block != nil {
		// PEM chain
		_ = readPEM("", block, rest, cb)
		return nil
	} else if st, err := os.Stat(s); err != nil {
		// Unknown
		return fs.ErrNotExist
	} else if st.IsDir() {
		// Directory
		_, err := dirReadPEM(s, cb)
		return err
	} else if !st.Mode().IsRegular() {
		// Invalid file type
		return fs.ErrInvalid
	} else if st.Size() > 0 {
		// Non-Empty File
		_, err := fileReadPEM(s, cb)
		return err
	} else {
		// Empty File
		return nil
	}
}

func readPEM(filename string, block *pem.Block, rest []byte, cb DecodePEMBlockFunc) bool {
	for block != nil {
		if cb(filename, block) {
			// cascade termination request
			return true
		} else if len(rest) == 0 {
			// EOF
			break
		} else {
			// next
			block, rest = pem.Decode(rest)
		}
	}

	return false
}

func dirReadPEM(dirname string, cb DecodePEMBlockFunc) (bool, error) {
	files, err := os.ReadDirWithLock(dirname)
	if err != nil {
		return false, err
	}

	for _, file := range files {
		fl := filepath.Join(dirname, file.Name())

		if st, err := os.Stat(fl); err != nil {
			// skip bad file
		} else if st.IsDir() {
			if term, _ := dirReadPEM(fl, cb); term {
				// cascade termination request
				return true, nil
			}
		} else if !st.Mode().IsRegular() {
			// skip unknown file type
		} else if st.Size() > 0 {
			if term, _ := fileReadPEM(fl, cb); term {
				// cascade termination request
				return true, nil
			}
		} else {
			// skip empty file
		}
	}

	return false, nil
}

func fileReadPEM(filename string, cb DecodePEMBlockFunc) (bool, error) {
	if b, err := os.ReadFileWithLock(filename); err != nil {
		// read error
		return false, err
	} else if len(b) > 0 {
		block, rest := pem.Decode(b)
		if block != nil {
			// process PEM file and propagate termination if needed
			term := readPEM(filename, block, rest, cb)
			return term, nil
		}
	}

	// skip non-PEM files
	return false, nil
}

// EncodeBytes produces a PEM encoded block
func EncodeBytes(label string, body []byte, headers map[string]string) []byte {
	var b bytes.Buffer
	pem.Encode(&b, &pem.Block{
		Type:    label,
		Bytes:   body,
		Headers: headers,
	})
	return b.Bytes()
}

// EncodePKCS1PrivateKey produces a PEM encoded RSA Private Key
func EncodePKCS1PrivateKey(key *rsa.PrivateKey) []byte {
	var out []byte
	if key != nil {
		body := x509.MarshalPKCS1PrivateKey(key)
		out = EncodeBytes("RSA PRIVATE KEY", body, nil)
	}
	return out
}

// EncodePKCS8PrivateKey produces a PEM encoded Private Key
func EncodePKCS8PrivateKey(key PrivateKey) []byte {
	var out []byte
	if key != nil {
		body, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			panic(errors.Wrap(err, "unreachable"))
		}
		out = EncodeBytes("PRIVATE KEY", body, nil)
	}
	return out
}

// EncodeCertificate produces a PEM encoded x509 Certificate
// without optional headers
func EncodeCertificate(der []byte) []byte {
	return EncodeBytes("CERTIFICATE", der, nil)
}
