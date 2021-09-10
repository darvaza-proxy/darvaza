package main

import (
	"bytes"
	"encoding/binary"
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

type prefixConn struct {
	net.Conn
	io.Reader
}

func (c prefixConn) Read(p []byte) (int, error) {
	return c.Reader.Read(p)
}

func proxy(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	var buf bytes.Buffer
	if _, err := io.CopyN(&buf, conn, 1+2+2); err != nil {
		log.Println(err)
		return
	}
	length := binary.BigEndian.Uint16(buf.Bytes()[3:5])
	if _, err := io.CopyN(&buf, conn, int64(length)); err != nil {
		log.Println(err)
		return
	}
	sn := sni.GetInfo(buf.Bytes())
	//TODO Deal with non TLS connections
	if sn != nil && sn.ServerName != "" {
		c := prefixConn{
			Conn:   conn,
			Reader: io.MultiReader(&buf, conn),
		}
		conn.SetReadDeadline(time.Time{})
		defer c.Close()
		var upstream net.Conn
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		//TODO after we will have backends we can drop the hardcoded 443
		upstream, err := net.Dial("tcp", fmt.Sprintf("%s:%d", sn.ServerName, 443))
		if err != nil {
			// TODO: This should not be fatal, need to retry
			log.Fatal(err)
			return
		}
		defer upstream.Close()

		go io.Copy(upstream, io.MultiReader(bytes.NewReader(buf.Bytes()), c))
		io.Copy(c, upstream)
	}
}
