package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type Crecord struct {
	Value  []string
	Ttl    int
	Rtt    int
	Stored int
}

func (r *Crecord) Expired() bool {
	ttl := r.Ttl - int(time.Now().Unix()-int64(r.Stored))
	return ttl < 0
}

type Cache struct {
	pcache map[string]Crecord
	pcap   int64
	pttl   int32

	ncache map[string]Crecord
	ncap   int64
	nttl   int32
	sync.RWMutex
}

func (c *Crecord) IsEmpty() bool {
	if c.Ttl == 0 && c.Rtt == 0 && c.Stored == 0 && len(c.Value) == 0 {
		return true
	}
	return false
}

func NewCache(size int64, exp int32) *Cache {
	c := new(Cache)
	c.pcache = make(map[string]Crecord)
	c.pcap = size
	c.pttl = exp
	c.ncache = make(map[string]Crecord)
	c.ncap = size
	c.nttl = 60
	c.LoadRoots()
	return c
}

func (c *Cache) Dump(w io.Writer, what bool) error {
	//what bool is true for pcache and false for ncache
	if w != os.Stdout {
		encoder := gob.NewEncoder(w)
		if what {
			errp := encoder.Encode(c.pcache)
			if errp != nil {
				logger.Error("cache %s", errp)
			}

		} else {
			errn := encoder.Encode(c.ncache)
			if errn != nil {
				logger.Error("negative cache %s", errn)
			}

		}
	} else {
		for z, cc := range c.pcache {
			fmt.Printf("%s %+v \n", z, cc)
		}
	}
	return nil
}

func (c *Cache) Get(key string) (Crecord, error) {
	c.RLock()
	centry, ok := c.ncache[key]
	c.RUnlock()

	if !ok {
		c.RLock()
		centry, ok = c.pcache[key]
		c.RUnlock()
		if !ok {
			return Crecord{}, fmt.Errorf("Key %q, not found.", key)
		}
	}

	if centry.Expired() {
		c.Delete(key)
		return Crecord{}, fmt.Errorf("Key %q, expired.", key)
	}

	return centry, nil
}

func (c *Cache) SetVal(key string, mtype string, ttl int, val string) {
	mk := c.MakeKey(key, mtype)
	var mrec Crecord
	mrec.Stored = int(time.Now().Unix())
	mrec.Ttl = ttl
	mrec.Value = append(mrec.Value, val)

	c.pcache[mk] = mrec

}

func (c *Cache) Set(key string, mtype string, d *dns.Msg) {
	mk := c.MakeKey(key, mtype)
	var sk string
	switch {
	case mtype == "NS":
		var mrec Crecord
		var srec Crecord
		for _, q := range d.Ns {
			u, _ := dRRtoRR(q)
			mrec.Stored = int(time.Now().Unix())
			mrec.Ttl = u.Ttl
			mrec.Value = append(mrec.Value, u.Value)
			if len(d.Extra) > 1 {
				for _, s := range d.Extra {
					y, _ := dRRtoRR(s)
					if u.Value == y.Name {
						if y.Type == "A" || y.Type == "AAAA" {
							srec.Value = []string{}
							srec.Stored = int(time.Now().Unix())
							srec.Ttl = y.Ttl
							srec.Value = append(srec.Value, y.Value)
							sk = c.MakeKey(dns.Fqdn(u.Value), y.Type)
							c.pcache[sk] = srec
						}
					}
				}
			}
			c.pcache[mk] = mrec
		}
	case mtype == "A", mtype == "AAAA":
		var rec Crecord
		for _, t := range d.Answer {
			v, _ := dRRtoRR(t)
			rec.Stored = int(time.Now().Unix())
			rec.Ttl = v.Ttl
			rec.Value = append(rec.Value, v.Value)

		}
		c.pcache[mk] = rec
	default:
		logger.Info("%v", mtype)
	}
}

func (c *Cache) Load(r io.Reader, what bool) error {
	//what bool is true for pcache and false for ncache
	var err error
	decoder := gob.NewDecoder(r)
	if what {
		err = decoder.Decode(c.pcache)
	} else {
		err = decoder.Decode(c.pcache)
	}
	return err
}

func (c *Cache) Delete(key string) {
	c.Lock()
	delete(c.pcache, key)
	delete(c.ncache, key)
	c.Unlock()
}

func (c *Cache) Size() int {
	return len(c.pcache)
}

func (c *Cache) LoadRoots() {

	if Config.RootsFile == "" {
		logger.Fatal("Config.RootsFile is empty :(")
	}

	fl, err := os.Open(Config.RootsFile)
	defer fl.Close()

	if err != nil {
		logger.Fatal("Error %s", err)
	}

	reader := bufio.NewReader(fl)
	var nstmp Crecord
	//keep root records for one year.....
	nstmp.Stored = int(time.Now().AddDate(1, 0, 0).Unix())
	nstmp.Ttl = 518400
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		fields := strings.Fields(line)
		nstmp.Value = append(nstmp.Value, fields[0])
	}
	c.Lock()
	c.pcache["./NS"] = nstmp
	c.Unlock()

	fl.Seek(0, 0)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		var v, q Crecord
		//keep root records for one year.....
		v.Stored, q.Stored = int(time.Now().AddDate(1, 0, 0).Unix()), int(time.Now().AddDate(1, 0, 0).Unix())
		v.Ttl, q.Ttl = 518400, 518400
		fields := strings.Fields(line)
		v.Value = append(v.Value, fields[1])
		c.Lock()
		c.pcache[dns.Fqdn(fields[0])+"/A"] = v
		c.Unlock()
		if len(fields) > 2 {
			q.Value = append(q.Value, fields[2])
			c.Lock()
			c.pcache[dns.Fqdn(fields[0])+"/AAAA"] = q
			c.Unlock()

		}
	}
}

func (c *Cache) MakeKey(a string, b string) string {
	return dns.Fqdn(a) + "/" + b
}
