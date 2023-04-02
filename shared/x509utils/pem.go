package x509utils

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"io/fs"
	"path/filepath"

	"darvaza.org/core"
	"github.com/darvaza-proxy/darvaza/shared/os"
)

// DecodePEMBlockFunc is called for each PEM block coded. it returns true
// to terminate the loop
type DecodePEMBlockFunc func(filename string, block *pem.Block) bool

// ReadPEM invoques a callback for each PEM block found
// it can receive raw PEM data
func ReadPEM(b []byte, cb DecodePEMBlockFunc) error {
	if len(b) == 0 {
		// empty
		return nil
	} else if block, rest := pem.Decode(b); block != nil {
		// PEM chain
		_ = readBlock("", block, rest, cb)
		return nil
	} else {
		// Not PEM
		return fs.ErrInvalid
	}
}

// ReadStringPEM invoques a callback for each PEM block found
// it can receive raw PEM data, a filename or a directory to scan
func ReadStringPEM(s string, cb DecodePEMBlockFunc) error {
	if ReadPEM([]byte(s), cb) == nil {
		// done
		return nil
	}

	if st, _ := os.Stat(s); st != nil {
		switch {
		case st.IsDir():
			// Directory
			_, err := dirReadPEM(s, cb)
			return err
		case !st.Mode().IsRegular():
			// Invalid file type
			return &fs.PathError{
				Op:   "read",
				Path: s,
				Err:  fs.ErrInvalid,
			}
		case st.Size() == 0:
			// Empty File
			return nil
		default:
			// Non-Empty File
			_, err := fileReadPEM(s, cb)
			return err
		}
	}
	return fs.ErrNotExist
}

// ReadFilePEM reads a PEM file calling cb for each block
func ReadFilePEM(filename string, cb DecodePEMBlockFunc) error {
	b, err := os.ReadFileWithLock(filename)
	if err != nil {
		// read error
		return err
	}
	if len(b) > 0 {
		block, rest := pem.Decode(b)
		if block != nil {
			readBlock(filename, block, rest, cb)
			return nil
		}
	}

	err = &fs.PathError{
		Op:   "pem.Decode",
		Path: filename,
		Err:  fs.ErrInvalid,
	}

	return err
}

func readBlock(filename string, block *pem.Block, rest []byte, cb DecodePEMBlockFunc) bool {
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
		term, _ := dirReadFilePEM(filepath.Join(dirname, file.Name()), cb)
		if term {
			// cascade termination request
			return true, nil
		}
	}

	return false, nil
}

func dirReadFilePEM(filename string, cb DecodePEMBlockFunc) (bool, error) {
	st, err := os.Stat(filename)

	switch {
	case err != nil:
		// file not found
		return false, err
	case st.IsDir():
		if term, _ := dirReadPEM(filename, cb); term {
			// cascade termination request
			return true, nil
		}
	case st.Mode().IsRegular() && st.Size() > 0:
		if term, _ := fileReadPEM(filename, cb); term {
			// cascade termination request
			return true, nil
		}
	}

	// continue
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
			term := readBlock(filename, block, rest, cb)
			return term, nil
		}
	}

	// skip non-PEM files
	return false, nil
}

// EncodeBytes produces a PEM encoded block
func EncodeBytes(label string, body []byte, headers map[string]string) []byte {
	var b bytes.Buffer
	_ = pem.Encode(&b, &pem.Block{
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
			panic(core.Wrap(err, "unreachable"))
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

// WriteKey writes a PEM encoded private key
func WriteKey(w io.Writer, key PrivateKey) (int64, error) {
	var buf bytes.Buffer

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err == nil {
		err = pem.Encode(&buf, &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyDER,
		})

		if err == nil {
			return buf.WriteTo(w)
		}
	}

	err = core.Wrap(err, "failed to encode certificate key")
	return 0, err
}

// WriteCert writes a PEM encoded certificate
func WriteCert(w io.Writer, cert *x509.Certificate) (int64, error) {
	var buf bytes.Buffer

	if len(cert.Raw) == 0 {
		err := errors.New("missing Raw DER certificate")
		return 0, err
	}

	if err := pem.Encode(&buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}); err != nil {
		err = core.Wrap(err, "failed to encode certificate")
		return 0, err
	}

	return buf.WriteTo(w)
}
