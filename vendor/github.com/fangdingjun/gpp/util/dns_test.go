package util

import (
	"testing"
)

func TestResolveAAAA(t *testing.T) {
	res, err := ResolveAAAA("mail.google.com")
	if err != nil {
		t.Error(err)
	}

	if len(res) < 1 {
		t.Error("resolve a error")
	}

	for _, ip := range res {
		t.Logf("ip: %s\n", ip.String())
	}
}

func TestResolveA(t *testing.T) {
	res, err := ResolveA("mail.google.com")
	if err != nil {
		t.Error(err)
	}

	if len(res) < 1 {
		t.Error("resolve a error")
	}

	for _, ip := range res {
		t.Logf("ip: %s\n", ip.String())
	}
}

func TestResolveDNS(t *testing.T) {
	res, err := ResolveDNS("mail.google.com")
	if err != nil {
		t.Error(err)
	}

	if len(res) < 1 {
		t.Error("resolve a error")
	}

	for _, ip := range res {
		t.Logf("ip: %s\n", ip.String())
	}
}
