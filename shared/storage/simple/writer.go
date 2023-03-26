package simple

import (
	"bytes"
	"container/list"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/fs"
	"strings"

	"github.com/darvaza-proxy/core"
	"github.com/darvaza-proxy/darvaza/shared/storage/certpool"
	"github.com/darvaza-proxy/darvaza/shared/x509utils"
)

// Delete removes a certificate by name
func (s *Store) Delete(_ context.Context, name string) error {
	const once = false // all matches
	var certs []*tls.Certificate

	s.mu.Lock()
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
	s.mu.Lock()
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
