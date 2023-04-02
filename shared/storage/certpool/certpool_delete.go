package certpool

import (
	"container/list"
	"context"
	"crypto/x509"
	"io/fs"

	"darvaza.org/core"
)

// Delete removes a certificate by name
func (s *CertPool) Delete(_ context.Context, name string) error {
	if name != "" {
		s.mu.Lock()
		defer s.mu.Unlock()

		hashes := s.getAllHashByName(name)
		if len(hashes) > 0 {
			for _, hash := range hashes {
				_ = s.deleteHash(hash)
			}
			return nil
		}
	}
	return fs.ErrNotExist
}

// DeleteCert removes a given certificate
func (s *CertPool) DeleteCert(_ context.Context, cert *x509.Certificate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.deleteHash(HashCert(cert))
}

func (s *CertPool) deleteHash(hash [HashSize]byte) error {
	if p, ok := s.hashed[hash]; ok {
		s.cached = nil // invalidate cache
		deleteHashFromNames(s.names, hash, p.names...)
		deleteHashFromNames(s.patterns, hash, p.patterns...)
		if skid := string(p.cert.SubjectKeyId); len(skid) > 0 {
			deleteHashFromNames(s.subjects, hash, skid)
		}
		delete(s.hashed, hash)
		return nil
	}

	return fs.ErrNotExist
}

func deleteHashFromNames(m map[string]*list.List, hash Hash, names ...string) {
	for _, name := range names {
		if l, ok := m[name]; ok {
			deleteHashFromList(l, hash)
		}
	}
}

func deleteHashFromList(l *list.List, hash Hash) {
	core.ListForEachElement(l, func(e *list.Element) bool {
		if p, ok := e.Value.(*certPoolEntry); ok {
			if p.hash == hash {
				l.Remove(e)
			}
		}
		return false // continue
	})
}
