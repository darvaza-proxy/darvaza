package sni

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net"
	"time"
)

// Conn is a net.Conn with custom Reader
type Conn struct {
	net.Conn
	io.Reader
}

func (c *Conn) Read(b []byte) (int, error) {
	return c.Reader.Read(b)
}

// PeekClientHelloInfo extracts the ClientHelloInfo from a connection
// still allowing a future handler have complete untouched access to
// the stream
func PeekClientHelloInfo(ctx context.Context,
	conn net.Conn) (*tls.ClientHelloInfo, net.Conn, error) {
	//
	var buf bytes.Buffer

	chi, err := ReadClientHelloInfo(ctx, io.TeeReader(conn, &buf))
	if err != nil {
		return nil, nil, err
	}

	conn2 := &Conn{
		Conn:   conn,
		Reader: io.MultiReader(&buf, conn),
	}

	return chi, conn2, nil
}

// ReadClientHelloInfo mimics a TLS connection to let Go's tls.Server parse the
// ClientHelloInfo for us
// - https://www.agwa.name/blog/post/writing_an_sni_proxy_in_go
func ReadClientHelloInfo(ctx context.Context,
	f io.Reader) (*tls.ClientHelloInfo, error) {
	//
	var out *tls.ClientHelloInfo

	conn := readOnlyConn{reader: f}
	conf := &tls.Config{
		GetConfigForClient: func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
			// copy
			out = new(tls.ClientHelloInfo)
			*out = *chi
			return nil, nil
		},
	}

	err := tls.Server(conn, conf).HandshakeContext(ctx)
	if out != nil {
		// no error if we got the ClientHelloInfo
		return out, nil
	}

	// otherwise the error is from the actual connection
	return nil, err
}

// readOnlyConn forwards reads to the reader and simulates a broken pipe when written to
// (as if the client closed the connection before the server could reply).
// All other operations are a no-op.
// - https://www.agwa.name/blog/post/writing_an_sni_proxy_in_go
type readOnlyConn struct {
	reader io.Reader
}

func (conn readOnlyConn) Read(b []byte) (int, error)    { return conn.reader.Read(b) }
func (readOnlyConn) Write(_ []byte) (int, error)        { return 0, io.ErrClosedPipe }
func (readOnlyConn) Close() error                       { return nil }
func (readOnlyConn) LocalAddr() net.Addr                { return nil }
func (readOnlyConn) RemoteAddr() net.Addr               { return nil }
func (readOnlyConn) SetDeadline(_ time.Time) error      { return nil }
func (readOnlyConn) SetReadDeadline(_ time.Time) error  { return nil }
func (readOnlyConn) SetWriteDeadline(_ time.Time) error { return nil }
