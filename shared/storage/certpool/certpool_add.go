package certpool

import (
	"container/list"
	"context"
	"crypto/x509"
	"encoding/pem"
	"os"
	"strings"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/x509utils"
)

// Put adds a certificate by name
func (s *CertPool) Put(_ context.Context, name string, cert *x509.Certificate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.addCertUnsafe(HashCert(cert), name, cert) {
		return nil
	}

	return os.ErrExist
}

// AppendCertsFromPEM adds certificates to the Pool from a PEM encoded blob,
// and returns true if a new Certificate was effectivelt added
func (s *CertPool) AppendCertsFromPEM(b []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	var added bool

	x509utils.ReadPEM(b, func(_ string, block *pem.Block) bool {
		if cert, _ := x509utils.BlockToCertificate(block); cert != nil && cert.IsCA {
			if s.addCertUnsafe(HashCert(cert), "", cert) {
				added = true
			}
		}
		return false // continue
	})

	return added
}

// AddCert adds parsed CA certificates to the CertPool
func (s *CertPool) AddCert(cert *x509.Certificate) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cert != nil && cert.IsCA {
		return s.addCertUnsafe(HashCert(cert), "", cert)
	}
	return false
}

func (s *CertPool) addCertUnsafe(hash Hash, name string, cert *x509.Certificate) bool {
	var added bool

	if s.hashed == nil {
		s.init()
	}

	p, ok := s.hashed[hash]
	if !ok {
		names, patterns := x509utils.Names(cert)

		p = &certPoolEntry{
			hash:     hash,
			cert:     cert,
			names:    names,
			patterns: patterns,
		}

		s.cached = nil // invalidate cache
		s.hashed[hash] = p
		s.setNames(p, names)
		s.setPatterns(p, patterns)
		s.setSubjectKeyID(p, cert.SubjectKeyId)
		added = true
	}

	if name != "" {
		name = strings.ToLower(name)
		if !core.SliceContains(p.names, name) {
			added = true
			p.names = append(p.names, name)
			s.setNames(p, []string{name})
		}
	}

	return added
}

func (s *CertPool) setNames(p *certPoolEntry, names []string) {
	for _, name := range names {
		s.setListItem(s.names, p, name)
	}
}

func (s *CertPool) setPatterns(p *certPoolEntry, patterns []string) {
	for _, pat := range patterns {
		s.setListItem(s.patterns, p, pat)
	}
}

func (s *CertPool) setSubjectKeyID(p *certPoolEntry, skid []byte) {
	s.setListItem(s.subjects, p, string(skid))
}

func (*CertPool) setListItem(m map[string]*list.List, p *certPoolEntry, name string) {
	if len(name) > 0 {
		core.MapListInsertUnique(m, name, p)
	}
}
