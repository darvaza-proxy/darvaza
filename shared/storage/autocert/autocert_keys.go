package autocert

import (
	"darvaza.org/darvaza/shared/x509utils"
)

// AddKey adds a private key to the store after attempting to validate it.
func (s *Store) AddKey(data string) error {
	_, err := s.addKeyString(data, true)
	return err
}

// Keys returns a copy of the slice containing the stored private keys.
func (s *Store) Keys() []x509utils.PrivateKey {
	return s.pool.Keys()
}
