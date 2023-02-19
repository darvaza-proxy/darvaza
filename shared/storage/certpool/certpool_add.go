package certpool

import (
	"container/list"
	"crypto/x509"
	"encoding/pem"

	"github.com/darvaza-proxy/darvaza/shared/data"
	"github.com/darvaza-proxy/darvaza/shared/x509utils"
)

// AppendCertsFromPEM adds certificates to the Pool from a PEM encoded blob,
// and returns true if a new Certificate was effectivelt added
func (s *CertPool) AppendCertsFromPEM(b []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	var added bool

	x509utils.ReadPEM(b, func(_ string, block *pem.Block) bool {
		if cert, _ := x509utils.BlockToCertificate(block); cert != nil && cert.IsCA {
			if s.addCertUnsafe(HashCert(cert), cert) {
				added = true
			}
		}
		return false // continue
	})

	return added
}

// AddCert adds parsed certificates to the CertPool
func (s *CertPool) AddCert(cert *x509.Certificate) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cert != nil && cert.IsCA {
		return s.addCertUnsafe(HashCert(cert), cert)
	}
	return false
}

func (s *CertPool) addCertUnsafe(hash Hash, cert *x509.Certificate) bool {
	var added bool

	if s.hashed == nil {
		s.init()
	}

	if _, ok := s.hashed[hash]; !ok {
		names, patterns := x509utils.Names(cert)

		p := &certPoolEntry{
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
		l, ok := m[name]
		if !ok {
			l = list.New()
			m[name] = l
		}
		if !data.ListContains(l, p) {
			l.PushFront(p)
		}
	}
}
