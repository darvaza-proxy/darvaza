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

type server struct {
	ip   string
	port int
}

func newServer() *server {
	s := new(server)
	s.ip = ""
	s.port = 8080
	return s
}

func (s *server) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go proxy(conn)
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

	if sn.ServerName != "" {
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
