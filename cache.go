package main

import (
	"io"
	"time"

	"github.com/miekg/dns"
)

type Item struct {
	Rcode              int
	Authoritative      bool
	AuthenticatedData  bool
	RecursionAvailable bool
	Answer             []dns.RR
	Ns                 []dns.RR
	Extra              []dns.RR

	origTTL int32
	stored  int64
}

type Cache interface {
	Get(key string) (*Item, error)
	Set(key string, value *Item) error
	Delete(key string)
	Size() int
	Dump(io.Writer) error
	Load(io.Reader) error
	Purge() error
	NewItem(*dns.Msg, int32) *Item
}

func (i *Item) Expired() bool {
	ttl := int(i.origTTL) - int(time.Now().Unix()-i.stored)
	return ttl < 0
}
