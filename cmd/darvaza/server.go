package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/darvaza-proxy/darvaza/tls/sni"
)

type Proxy struct {
	Protocol   string   `default:"http" hcl:"protocol,label"`
	ListenAddr []string `default:"[\":8080\"]" hcl:"listen"`
}

func (p *Proxy) listeners() []net.Listener {
	ls := make([]net.Listener, 0)
	for _, laddr := range p.ListenAddr {
		//TODO do we want UDP/IP and or others?
		l, err := net.Listen("tcp", laddr)
		if err != nil {
			log.Printf("cannot listen on %s.\n %q\n", laddr, err)
			continue
		}
		ls = append(ls, l)
	}
	return ls
}

func (p *Proxy) Run() {
	defer cfg.Done()
	ls := p.listeners()
	for {
		for _, l := range ls {
			conn, err := l.Accept()
			if err != nil {
				log.Println(err)
			}
			go proxy(conn)
		}
	}
}

func proxy(conn net.Conn) {
	defer conn.Close()

	var upstream net.Conn
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	var b bytes.Buffer
	n, err := conn.Read(b.Bytes())

	if err != nil {
		fmt.Println(err)
		return
	}
	sn := sni.GetInfo(b.Bytes()[:n])

	if sn != nil && sn.ServerName != "" {
		upstream, err = net.Dial("tcp", fmt.Sprintf("%s:%d", sn.ServerName, 443))
		if err != nil {
			// TODO: This should not be fatal, need to retry
			log.Fatal(err)
			return
		}
	} else {
		return
	}

	defer upstream.Close()

	go io.Copy(upstream, io.MultiReader(bytes.NewReader(b.Bytes()), conn))
	io.Copy(conn, upstream)
}
