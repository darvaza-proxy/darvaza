package certpool

import (
	"container/list"
	"fmt"
	"strings"

	"github.com/darvaza-proxy/core"
	"github.com/darvaza-proxy/darvaza/shared/x509utils"
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

func nameAsIP(name string) (string, bool) {
	if addr, err := core.ParseAddr(name); err == nil {
		return fmt.Sprintf("[%s]", addr.String()), true
	}
	return "", false
}

func nameAsSuffix(name string) (string, bool) {
	if idx := strings.IndexRune(name, '.'); idx > 0 {
		name = name[idx:]
		return name, len(name) > 2
	}
	return "", false
}

// revive:disable:cognitive-complexity
func (s *CertPool) getFirstByName(name string) *certPoolEntry {
	name, ok := x509utils.SanitiseName(name)
	if !ok {
		return nil
	}

	// IP
	if ip, ok := nameAsIP(name); ok {
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
	if suffix, ok := nameAsSuffix(name); ok {
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
	if ip, ok := nameAsIP(name); ok {
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
	if suffix, ok := nameAsSuffix(name); ok {
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
