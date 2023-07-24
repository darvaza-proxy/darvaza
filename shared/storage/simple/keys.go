package simple

import (
	"crypto"

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

// HasKey checks if the store contains the specified PrivateKey.
func (s *Store) HasKey(pk crypto.PrivateKey) bool {
	if pk != nil {
		s.lockInit()
		defer s.mu.Unlock()

		for _, pk1 := range s.keys {
			if pk1.Equal(pk) {
				return true
			}
		}
	}

	return false
}

// HasPublicKey checks if the store contains a PrivateKey matching
// the given PublicKey.
func (s *Store) HasPublicKey(pub crypto.PublicKey) bool {
	if pub != nil {
		s.lockInit()
		defer s.mu.Unlock()

		for _, pk1 := range s.keys {
			pub1, ok := pk1.Public().(x509utils.PublicKey)
			if ok && pub1.Equal(pub) {
				return true
			}
		}
	}

	return false
}

// Keys returns a copy of the slice containing the stored private keys.
func (s *Store) Keys() []x509utils.PrivateKey {
	s.lockInit()
	defer s.mu.Unlock()

	out := make([]x509utils.PrivateKey, len(s.keys))
	copy(out, s.keys)
	return out
}
