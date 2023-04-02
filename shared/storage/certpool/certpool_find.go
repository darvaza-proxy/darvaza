package certpool

import (
	"container/list"

	"darvaza.org/core"
	"darvaza.org/darvaza/shared/x509utils"
)

func (s *CertPool) getAllHashByName(name string) []Hash {
	if entries, ok := s.getEntriesByName(name); ok {
		out := make([]Hash, 0, len(entries))
		for _, e := range entries {
			if !core.SliceContains(out, e.hash) {
				out = append(out, e.hash)
			}
		}

		return out
	}
	return nil
}

func getFirstInList(l *list.List) (out *certPoolEntry) {
	core.ListForEach(l, func(v *certPoolEntry) bool {
		out = v
		return true
	})
	return out
}

func getEntriesInList(l *list.List) (out []*certPoolEntry, found bool) {
	core.ListForEach(l, func(v *certPoolEntry) bool {
		out = append(out, v)
		return false
	})
	return out, len(out) > 0
}

// revive:disable:cognitive-complexity
func (s *CertPool) getFirstByName(name string) *certPoolEntry {
	name, ok := x509utils.SanitiseName(name)
	if !ok {
		return nil
	}

	// IP
	if ip, ok := x509utils.NameAsIP(name); ok {
		if l, ok := s.names[ip]; ok {
			return getFirstInList(l)
		}
		return nil
	}

	// exact
	if l, ok := s.names[name]; ok {
		out := getFirstInList(l)
		if out != nil {
			return out
		}
	}

	// wildcard
	if suffix, ok := x509utils.NameAsSuffix(name); ok {
		if l, ok := s.patterns[suffix]; ok {
			out := getFirstInList(l)
			if out != nil {
				return out
			}
		}
	}

	return nil
}

func (s *CertPool) getEntriesByName(name string) (out []*certPoolEntry, found bool) {
	name, ok := x509utils.SanitiseName(name)
	if !ok {
		return nil, false
	}

	// IP
	if ip, ok := x509utils.NameAsIP(name); ok {
		if l, ok := s.names[ip]; ok {
			return getEntriesInList(l)
		}
		return nil, false
	}

	// exact
	if l, ok := s.names[name]; ok {
		out, found = getEntriesInList(l)
		if found {
			// don't look for wildcards if we have exact matches
			return out, true
		}
	}

	// wildcard
	if suffix, ok := x509utils.NameAsSuffix(name); ok {
		if l, ok := s.patterns[suffix]; ok {
			m, _ := getEntriesInList(l)
			if len(m) > 0 {
				out = append(out, m...)
				found = true
			}
		}
	}

	return out, found
}
