package simple

import (
	"bytes"
	"container/list"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/fs"

	"github.com/darvaza-proxy/core"
	"github.com/darvaza-proxy/darvaza/shared/storage/certpool"
)

// DeleteCert removes a certificate from the store
func (s *Store) DeleteCert(_ context.Context, cert *x509.Certificate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := certpool.HashCert(cert)
	if ci, ok := s.hashed[hash]; ok {
		s.deleteByCertInfo(ci)
		return nil
	}
	return fs.ErrNotExist
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
