package gnocco

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type crecord struct {
	Value  []string
	TTL    int
	RTT    int
	Stored int
}

func (c *crecord) expired() bool {
	return c.TTL < int(time.Now().Unix()-int64(c.Stored))
}

type cache struct {
	pcache map[string]crecord
	pcap   int64
	pttl   int32

	ncache map[string]crecord
	ncap   int64
	nttl   int32
	sync.RWMutex
}

func (c *crecord) isEmpty() bool {
	if c.TTL == 0 && c.RTT == 0 && c.Stored == 0 && len(c.Value) == 0 {
		return true
	}
	return false
}

func (c *crecord) String() string {
	result := "Value: "
	for _, z := range c.Value {
		result = result + z
	}
	result = result + " TTL: " + strconv.Itoa(c.TTL)
	result = result + " RTT: " + strconv.Itoa(c.RTT)
	result = result + " Stored: " + time.Unix(int64(c.Stored), 0).String()
	return result
}

func newCache(size int64, exp int32) *cache {
	c := new(cache)
	c.pcache = make(map[string]crecord)
	c.pcap = size
	c.pttl = exp
	c.ncache = make(map[string]crecord)
	c.ncap = size
	c.nttl = 60
	return c
}

func (c *cache) dump(w io.Writer) error {
	if w != os.Stdout {
		encoder := gob.NewEncoder(w)
		errp := encoder.Encode(c.pcache)
		if errp != nil {
			return fmt.Errorf("cache %s", errp)
		}
		errn := encoder.Encode(c.ncache)
		if errn != nil {
			return fmt.Errorf("negative cache %s", errn)
		}
	} else {
		for z, cc := range c.pcache {
			fmt.Printf("%s %+v \n", z, cc)
		}
	}
	return nil
}

func (c *cache) get(key string) (crecord, error) {
	c.RLock()
	centry, ok := c.ncache[key]
	c.RUnlock()

	if !ok {
		c.RLock()
		centry, ok = c.pcache[key]
		c.RUnlock()
		if !ok {
			return crecord{}, fmt.Errorf("key %q, not found", key)
		}
	}

	if centry.expired() {
		c.delete(key)
		return crecord{}, fmt.Errorf("key %q, expired", key)
	}

	return centry, nil
}

func (c *cache) setVal(key string, mtype string, ttl int, val string) {
	mk := c.makeKey(key, mtype)
	var mrec crecord
	mrec.Stored = int(time.Now().Unix())
	mrec.TTL = ttl
	mrec.Value = append(mrec.Value, val)
	c.Lock()
	c.pcache[mk] = mrec
	c.Unlock()
}

func (c *cache) set(key string, mtype string, d *dns.Msg) {
	mk := c.makeKey(key, mtype)
	var sk string
	switch {
	case mtype == "NS":
		var mrec crecord
		var srec crecord
		for _, q := range d.Ns {
			u, _ := dRRtoRR(q)
			mrec.Stored = int(time.Now().Unix())
			mrec.TTL = u.TTL
			mrec.Value = append(mrec.Value, u.Value)
			if len(d.Extra) > 1 {
				for _, s := range d.Extra {
					y, _ := dRRtoRR(s)
					if u.Value == y.Name {
						if y.Type == "A" || y.Type == "AAAA" {
							srec.Value = []string{}
							srec.Stored = int(time.Now().Unix())
							srec.TTL = y.TTL
							srec.Value = append(srec.Value, y.Value)
							sk = c.makeKey(dns.Fqdn(u.Value), y.Type)
							c.Lock()
							c.pcache[sk] = srec
							c.Unlock()
						}
					}
				}
			}
			c.Lock()
			c.pcache[mk] = mrec
			c.Unlock()
		}
	case mtype == "A", mtype == "AAAA", mtype == "CNAME":
		var rec crecord
		for _, t := range d.Answer {
			v, _ := dRRtoRR(t)
			rec.Stored = int(time.Now().Unix())
			rec.TTL = v.TTL
			rec.Value = append(rec.Value, v.Value)
		}
		c.Lock()
		c.pcache[mk] = rec
		c.Unlock()
	default:
		fmt.Printf("%v /n", mtype)
	}
}

func (c *cache) load(r io.Reader) error {
	decoder := gob.NewDecoder(r)
	c.Lock()
	errp := decoder.Decode(&c.pcache)
	errn := decoder.Decode(&c.ncache)
	c.Unlock()

	if errp != nil {
		return errp
	}
	return errn
}

func (c *cache) delete(key string) {
	c.Lock()
	delete(c.pcache, key)
	delete(c.ncache, key)
	c.Unlock()
}

func (c *cache) size() int {
	return len(c.pcache)
}

func (*cache) makeKey(a string, b string) string {
	return dns.Fqdn(a) + "/" + b
}
