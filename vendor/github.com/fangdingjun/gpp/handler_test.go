package gpp

import (
	"bufio"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	//"os"
	"bytes"
	"strconv"
)

// test basic function
func TestBasic(t *testing.T) {
	ts := httptest.NewServer(&Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello"))
		}),
		EnableProxy: false,
	})

	defer ts.Close()

	r, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	defer r.Body.Close()

	got, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Errorf("got %s, want hello", string(got))
	}
}

// test GET through proxy
func TestProxyGET(t *testing.T) {
	ts := httptest.NewServer(&Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello"))
		}),
		EnableProxy:       true,
		EnableProxyHTTP11: true,
	})

	defer ts.Close()

	dstUrl := "http://httpbin.org/get"

	u, _ := url.Parse(ts.URL)

	c, err := net.Dial("tcp", u.Host)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	r := bufio.NewReader(c)

	req, _ := http.NewRequest("GET", dstUrl, nil)
	req.WriteProxy(c)

	res, err := http.ReadResponse(r, req)
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		t.Errorf("proxy get error code %d", res.StatusCode)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("\n%s", string(got))
}

// test POST through proxy
func TestProxyPOST(t *testing.T) {
	ts := httptest.NewServer(&Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello"))
		}),
		EnableProxy:       true,
		EnableProxyHTTP11: true,
	})

	defer ts.Close()

	u, _ := url.Parse(ts.URL)

	body := bytes.NewBufferString("a=b&c=d&e=f")
	req, _ := http.NewRequest("POST", "http://httpbin.org/post", body)

	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	c, err := net.Dial("tcp", u.Host)
	if err != nil {
		t.Fatal(err)
	}

	r := bufio.NewReader(c)

	req.WriteProxy(c)

	res, err := http.ReadResponse(r, req)
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		t.Errorf("post return code %d", res.StatusCode)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\n%s", string(got))
}

// test CONNECT through proxy
func TestProxyCONNECT(t *testing.T) {
	ts := httptest.NewServer(&Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello"))
		}),
		EnableProxy:       true,
		EnableProxyHTTP11: true,
	})

	defer ts.Close()

	u, _ := url.Parse(ts.URL)

	c, err := net.Dial("tcp", u.Host)
	if err != nil {
		t.Fatal(err)
	}

	defer c.Close()

	r := bufio.NewReader(c)

	io.WriteString(c, "CONNECT httpbin.org:80 HTTP/1.0\r\n\r\n")

	// get response line
	l, err := r.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}

	sss := strings.SplitN(l, " ", 3)

	code, err := strconv.Atoi(sss[1])
	if err != nil {
		t.Fatal(err)
	}

	if code != 200 {
		t.Fatal("connect return code %d", code)
	}

	// read response header
	for {
		l, err := r.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}

		l1 := strings.Trim(l, "\r\n")
		if l1 == "" {
			break
		}
	}

	req, _ := http.NewRequest("GET", "http://httpbin.org/get", nil)
	req.Write(c)

	res, err := http.ReadResponse(r, req)
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		t.Errorf("get code %d", res.StatusCode)
	}

	got, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\n%s", string(got))
}
