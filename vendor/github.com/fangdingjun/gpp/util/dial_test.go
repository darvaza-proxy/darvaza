package util

import (
	"testing"
)

func TestDial(t *testing.T) {
	conn, err := Dial("tcp", "www.baidu.com:80")
	if err != nil {
		t.Error(err)
	}
	_, err = conn.Write([]byte("GET /aasdf/asdfasf HTTP/1.0\r\nHost: www.baidu.com\r\nUser-Agent: ffdasdfsf/1.0\r\n\r\n"))
	if err != nil {
		t.Error(err)
	}
	buf := make([]byte, 10240)
	n, err := conn.Read(buf[0:])
	if err != nil {
		t.Error(err)
	}
	t.Logf("receive %d bytes from network\n", n)
	t.Logf("%s\n", string(buf[:n]))
	conn.Close()
}
