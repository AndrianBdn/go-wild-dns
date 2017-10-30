package main

import (
	"testing"
	"net"
)

func TestBasic(t *testing.T) {

	var ip net.IP
	lh := net.ParseIP("127.0.0.1")

	ip = ipFromHost("127.0.0.1.test.com", nil)
	t.Log(ip)
	if !ip.Equal(lh) {
		t.Error("fail dot test")
	}

	ip = ipFromHost("this-is-good-127-0-0-1.test.com", nil)
	if !ip.Equal(lh) {
		t.Error("fail dash test")
	}

	ip = ipFromHost("bad-555-0-0-1.test.com", lh)
	if !ip.Equal(lh) {
		t.Error("fail default test")
	}

}