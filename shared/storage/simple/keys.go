package simple

import (
	"darvaza.org/core"
	"darvaza.org/darvaza/shared/x509utils"
)

// validateKey calls the key's Validate() if available
func validateKey(pk x509utils.PrivateKey) error {
	if p, ok := pk.(interface {
		Validate() error
	}); ok {
		return p.Validate()
	}

	return nil
}

// AddKey adds a private key to the store after attempting to validate it.
func (s *Store) AddKey(pk x509utils.PrivateKey) error {
	if err := validateKey(pk); err != nil {
		return err
	}

	s.lockInit()
	defer s.mu.Unlock()

	if !core.SliceContainsFn(s.keys, pk, PrivateKeyEqual) {
		s.keys = append(s.keys, pk)
	}

	return nil
}

// Keys returns a copy of the slice containing the stored private keys.
func (s *Store) Keys() []x509utils.PrivateKey {
	s.lockInit()
	defer s.mu.Unlock()

	out := make([]x509utils.PrivateKey, len(s.keys))
	copy(out, s.keys)
	return out
}
