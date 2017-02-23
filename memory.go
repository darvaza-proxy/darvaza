package main

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type Memory struct {
	backend map[string]*Item
	ttl     time.Duration
	cap     int
	sync.RWMutex
}

func NewMemory(size, exp int) *Memory {
	m := new(Memory)
	m.backend = make(map[string]*Item)
	m.cap = size
	m.ttl = time.Duration(exp) * time.Second
	return m
}

func (m *Memory) NewItem(dm *dns.Msg, d int32) *Item {
	i := new(Item)
	i.Rcode = dm.Rcode
	i.Authoritative = dm.Authoritative
	i.AuthenticatedData = dm.AuthenticatedData
	i.RecursionAvailable = dm.RecursionAvailable
	i.Answer = dm.Answer
	i.Ns = dm.Ns
	i.Extra = make([]dns.RR, len(dm.Extra))
	// Don't copy OPT record as these are hop-by-hop.
	j := 0
	for _, e := range dm.Extra {
		if e.Header().Rrtype == dns.TypeOPT {
			continue
		}
		i.Extra[j] = e
		j++
	}
	i.Extra = i.Extra[:j]

	i.origTTL = d
	i.stored = time.Now().Unix()

	return i
}
func (m *Memory) Get(key string) (*Item, error) {
	m.RLock()
	item, ok := m.backend[key]
	m.RUnlock()

	if !ok {
		return nil, fmt.Errorf("Key %q, not found.", key)
	}

	if item.Expired() {
		m.Delete(key)
		return nil, fmt.Errorf("Key %q, expired.", key)
	}

	return item, nil
}

func (m *Memory) Set(key string, item *Item) error {
	return nil
}

func (m *Memory) Delete(key string) {
	m.Lock()
	delete(m.backend, key)
	m.Unlock()
}

func (m *Memory) Size() int {
	m.RLock()
	defer m.RUnlock()
	return int(0)
}

func (m *Memory) Load(r io.Reader) error {
	return nil
}

func (m *Memory) Dump(w io.Writer) error {
	return nil
}

func (m *Memory) Purge() error {
	return nil
}
