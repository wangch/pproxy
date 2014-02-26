//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"bytes"
	"code.google.com/p/go.net/proxy"
	"io"
	"net"
	"testing"
	"time"
)

func TestSocksProxy(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:13000")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	go func() {
		conn, err := l.Accept()
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		defer conn.Close()

		buf := make([]byte, 4)
		if _, err := io.ReadAtLeast(conn, buf, 4); err != nil {
			t.Fatalf("err: %v", err)
		}

		if !bytes.Equal(buf, []byte("ping")) {
			t.Fatalf("bad: %v", buf)
		}
		conn.Write([]byte("pong"))
	}()

	go func() {
		e := SocksProxy(nil, 14000)
		if e != nil {
			t.Fatal(e)
		}
	}()

	dialer, e := proxy.SOCKS5("tcp", "127.0.0.1:14000", nil, proxy.Direct)
	if e != nil {
		t.Error(e)
	}

	conn, e := dialer.Dial("tcp", "127.0.0.1:13000")
	if e != nil {
		t.Error(e)
	}

	conn.Write([]byte("ping"))

	out := make([]byte, 4)

	conn.SetDeadline(time.Now().Add(time.Second))
	if _, err := io.ReadAtLeast(conn, out, len(out)); err != nil {
		t.Fatalf("err: %v", err)
	}

	if !bytes.Equal(out, []byte("pong")) {
		t.Fatalf("bad: %v", out)
	}
}
