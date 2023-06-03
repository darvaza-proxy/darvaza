package simple

import (
	"bytes"
	"container/list"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/fs"
	"strings"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/storage/certpool"
	"darvaza.org/darvaza/shared/x509utils"
)

var (
	_ x509utils.WriteStore = (*Store)(nil)
)

// Put adds a certificate to the store
func (s *Store) Put(_ context.Context, name string, cert *x509.Certificate) error {
	s.lockInit()
	defer s.mu.Unlock()

	hash := certpool.HashCert(cert)
	if ci, ok := s.hashed[hash]; ok {
		// known
		if name != "" {
			// maybe new name?
			return s.appendName(ci, name)
		}
		return fs.ErrExist
	}

	key := s.findMatchingKey(cert)
	if key == nil {
		err := core.Wrap(fs.ErrNotExist, "no suitable key available")
		return err
	}

	c, err := s.bundler.Bundle(cert, key)
	if err != nil {
		err = core.Wrap(err, "failed to bundle certificate")
		return err
	}

	addCerts(s, c)
	return nil
}

func (s *Store) appendName(ci *certInfo, name string) error {
	if strings.HasPrefix(name, "*.") {
		// pattern
		k := name[1:]
		if mapListContainsHash(s.patterns, k, ci.hash) {
			return fs.ErrExist
		}

		ci.patterns = append(ci.patterns, k)
		core.MapListAppend(s.patterns, k, ci.c)
		return nil
	}

	if n, ok := x509utils.NameAsIP(name); ok {
		// IP
		name = n
	}

	if mapListContainsHash(s.names, name, ci.hash) {
		// already set
		return fs.ErrExist
	}

	ci.names = append(ci.names, name)
	core.MapListAppend(s.names, name, ci.c)
	return nil
}

func (s *Store) findMatchingKey(cert *x509.Certificate) x509utils.PrivateKey {
	for _, key := range s.keys {
		if PairMatch(cert, key) {
			return key
		}
	}
	return nil
}

// Delete removes a certificate by name
func (s *Store) Delete(_ context.Context, name string) error {
	const once = false // all matches
	var certs []*tls.Certificate

	s.lockInit()
	defer s.mu.Unlock()

	// find by name
	if strings.HasPrefix(name, "*.") {
		// pattern
		certs = FindInMap(name[1:], s.patterns, once)
	} else {
		if n, ok := x509utils.NameAsIP(name); ok {
			// exact match for IP
			name = n
		}

		certs = FindInMap(name, s.names, once)
	}

	if len(certs) == 0 {
		// none found
		return fs.ErrNotExist
	}

	// delete all matches
	for _, c := range certs {
		if ci := s.findCertInfo(c.Leaf); ci != nil {
			s.deleteByCertInfo(ci)
		}
	}

	return nil
}

// DeleteCert removes a certificate from the store
func (s *Store) DeleteCert(_ context.Context, cert *x509.Certificate) error {
	s.lockInit()
	defer s.mu.Unlock()

	if ci := s.findCertInfo(cert); ci != nil {
		s.deleteByCertInfo(ci)
		return nil
	}
	return fs.ErrNotExist
}

func (s *Store) findCertInfo(cert *x509.Certificate) *certInfo {
	hash := certpool.HashCert(cert)
	if ci, ok := s.hashed[hash]; ok {
		return ci
	}
	return nil
}

func (s *Store) deleteByCertInfo(ci *certInfo) {
	deleteMapListElementByHash(s.names, ci.names, ci.hash)
	deleteMapListElementByHash(s.patterns, ci.patterns, ci.hash)
	deleteListElementByPointer(s.certs, ci, true)
	delete(s.hashed, ci.hash)
}

func deleteMapListElementByHash(m map[string]*list.List,
	keys []string, hash certpool.Hash) int {
	// for each name we can only have one match
	const once = true

	// leaf has the hash we need
	match := func(c *tls.Certificate) bool {
		h := certpool.HashCert(c.Leaf)
		return bytes.Equal(hash[:], h[:])
	}

	return deleteMapListElementByMatcher(m, once, keys, match)
}

func deleteMapListElementByMatcher[T any, K comparable](m map[K]*list.List, once bool, keys []K,
	match func(v T) bool) int {
	var count int

	fn := func(key K, el *list.Element) bool {
		if v, ok := el.Value.(T); ok {
			if match(v) {
				m[key].Remove(el)
				count++
				return once
			}
		}

		return false // continue
	}

	for _, key := range keys {
		core.MapListForEachElement(m, key, func(el *list.Element) bool {
			return fn(key, el)
		})
	}

	return count
}

func deleteListElementByPointer[T any](l *list.List, ptr *T, once bool) int {
	var count int

	fn := func(el *list.Element) bool {
		if p, ok := el.Value.(*T); ok {
			if ptr == p {
				// match
				l.Remove(el)
				count++
				return once
			}
		}

		return false // continue
	}

	core.ListForEachElement(l, fn)
	return count
}
