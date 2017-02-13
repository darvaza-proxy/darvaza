package main

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type Mesg struct {
	Msg    *dns.Msg
	Expire time.Time
}

type MCache struct {
	Backend  map[string]Mesg
	Expire   time.Duration
	Maxcount int
	mu       sync.RWMutex
}

type keyNotFound struct {
	key string
}

func (k keyNotFound) Error() string {
	return "Key " + k.key + " not found."
}

type keyExpired struct {
	key string
}

func (k keyExpired) Error() string {
	return "Key " + k.key + " expired."
}

type CacheIsFull struct {
}

func (c CacheIsFull) Error() string {
	return "Cache is Full"
}

func (m *MCache) Get(key string) (*dns.Msg, error) {
	m.mu.RLock()
	mesg, ok := m.Backend[key]
	m.mu.RUnlock()

	if !ok {
		return nil, keyNotFound{key}
	}

	if mesg.Expire.Before(time.Now()) {
		m.Remove(key)
		return nil, keyExpired{key}
	}

	return mesg.Msg, nil
}

func (m *MCache) Set(key string, msg *dns.Msg) error {
	if m.Full() && !m.Exists(key) {
		return CacheIsFull{}
	}

	expire := time.Now().Add(m.Expire)
	mesg := Mesg{msg, expire}
	m.mu.Lock()
	m.Backend[key] = mesg
	m.mu.Unlock()

	return nil
}

func (m *MCache) Remove(key string) {
	m.mu.Lock()
	delete(m.Backend, key)
	m.mu.Unlock()
}

func (m *MCache) Exists(key string) bool {
	m.mu.RLock()
	_, ok := m.Backend[key]
	m.mu.RUnlock()
	return ok
}

func (m *MCache) Length() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Backend)
}

func (m *MCache) Full() bool {
	if m.Maxcount == 0 {
		return false
	}
	return m.Length() >= m.Maxcount
}

func (m *MCache) Load() error {
	return nil
}

func (m *MCache) Save() error {
	return nil
}

func (m *MCache) Dump() {
	fmt.Println(m)
}

func KeyGen(q Question) string {
	hash := md5.New()
	hash.Write([]byte(q.String()))
	hsum := hash.Sum(nil)
	key := fmt.Sprintf("%x", hsum)
	return key
}
