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
	encoder := gob.NewEncoder(w)
	if what {
		errp := encoder.Encode(c.pcache)
		if errp != nil {
			fmt.Println("encode:", errp)
		}

	} else {
		errn := encoder.Encode(c.ncache)
		if errn != nil {
			fmt.Println("encode:", errn)
		}

	}
	return nil
}

func (c *Cache) Get(key string, req *dns.Msg) (*dns.Msg, error) {
	c.RLock()
	centry, ok := c.ncache[key]
	c.RUnlock()

	if !ok {
		c.RLock()
		centry, ok = c.pcache[key]
		c.RUnlock()
		if !ok {

			return nil, fmt.Errorf("Key %q, not found.", key)
		}
	}

	if centry.Expired() {
		c.Delete(key)
		return nil, fmt.Errorf("Key %q, expired.", key)
	}

	//we have an answer now construct a dns.Msg
	result := new(dns.Msg)
	result.SetReply(req)
	qname := strings.Split(key, "/")[0]
	qtype := strings.Split(key, "/")[1]
	for _, z := range centry.Value {
		rec, _ := dns.NewRR(dns.Fqdn(qname) + " " + qtype + " " + z)
		result.Answer = append(result.Answer, rec)
	}

	return result, nil
}

func (c *Cache) Set(key string, d *dns.Msg) error {

	return nil
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
	nstmp.Ttl = 32767
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
		v.Stored, q.Stored = int(time.Now().AddDate(1, 0, 0).Unix()), int(time.Now().AddDate(1, 0, 0).Unix())
		v.Ttl, q.Ttl = 32767, 32767
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

func (c *Cache) makeKey(a string, b string) string {
	return a + "/" + b
}
