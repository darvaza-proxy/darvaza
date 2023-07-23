package autocert

import (
	"encoding/pem"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/x509utils"
)

func (s *Store) addCertString(data string, ca bool) (bool, error) {
	var addErr error
	var added bool

	readErr := x509utils.ReadStringPEM(data, func(origin string, block *pem.Block) bool {
		if ok, err := s.addCertBlock(origin, ca, block); err != nil {
			addErr = err
		} else if ok {
			added = true
		}

		return addErr != nil // stop on error
	})

	return added, core.CoalesceError(addErr, readErr)
}

func (s *Store) addKeyString(data string, ca bool) (bool, error) {
	var addErr error
	var added bool

	readErr := x509utils.ReadStringPEM(data, func(origin string, block *pem.Block) bool {
		if ok, err := s.addKeyBlock(origin, ca, block); err != nil {
			addErr = err
		} else if ok {
			added = true
		}

		return addErr != nil // stop on error
	})

	return added, core.CoalesceError(addErr, readErr)
}

// revive:disable:flag-parameter

func (s *Store) addCertBlock(_ string, ca bool, block *pem.Block) (bool, error) {
	// revive:enable:flag-parameter
	cert, err := x509utils.BlockToCertificate(block)
	switch {
	case cert != nil:
		// Certificate
		switch {
		case ca:
			err = s.pool.AddCACert(cert)
		case !ca:
			err = s.pool.AddCert("", cert)
		}
		return true, err
	case err != x509utils.ErrIgnored:
		// Bad Certificate
		return false, err
	default:
		// Not a Certificate
		return false, nil
	}
}

func (s *Store) addKeyBlock(_ string, _ bool, block *pem.Block) (bool, error) {
	pk, err := x509utils.BlockToPrivateKey(block)
	switch {
	case pk != nil:
		// PrivateKey
		return true, s.pool.AddKey(pk)
	case err != x509utils.ErrIgnored:
		// Bad PrivateKey
		return false, err
	default:
		// Not a PrivateKey
		return false, nil
	}
}
