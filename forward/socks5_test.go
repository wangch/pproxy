//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"bytes"
	"code.google.com/p/go.net/proxy"
	"fmt"
	"github.com/wangch/go-socks5"
	"io"
	"log"
	"net"
	"testing"
	"time"
)

// a SOCKS5 proxy server
func SocksProxyServer(ips []net.IP, port int) error {
	conf := &socks5.Config{AuthMethods: []socks5.Authenticator{socks5.NoAuthAuthenticator{}}}
	s, e := socks5.New(conf)
	if e != nil {
		log.Fatal(e)
	}
	l, e := s.Listen("tcp", fmt.Sprintf(":%d", port))
	if e != nil {
		return e
	}
	log.Println(s.ServeF(l, func(addr net.Addr) bool {
		sip, _, e := net.SplitHostPort(addr.String())
		if e != nil {
			return false
		}
		if len(ips) == 0 {
			return true
		}
		ip := net.ParseIP(sip)
		for _, x := range ips {
			if ip.Equal(x) {
				return true
			}
		}
		return false
	}))
	return nil
}

func TestSocksProxy(t *testing.T) {
	// create tcp server
	l, err := net.Listen("tcp", "127.0.0.1:11000")
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

	// create socks5 proxy server
	go func() {
		e := SocksProxyServer(nil, 12000)
		if e != nil {
			t.Fatal(e)
		}
	}()

	// create socks5 forward server
	go func() {
		e := SocksProxy(nil, "127.0.0.1:12000", ":12001", "user", "password")
		if e != nil {
			t.Fatal(e)
		}
	}()

	// soccks 5 client
	dialer, e := proxy.SOCKS5("tcp", "127.0.0.1:12001", nil, proxy.Direct)
	if e != nil {
		t.Error(e)
	}

	conn, e := dialer.Dial("tcp", "127.0.0.1:11000")
	if e != nil {
		t.Error(e)
	}

	conn.Write([]byte("ping"))

	out := make([]byte, 4)
	conn.SetDeadline(time.Now().Add(time.Second))
	if _, err := conn.Read(out); err != nil {
		t.Fatalf("err: %v", err)
	}

	if !bytes.Equal(out, []byte("pong")) {
		t.Fatalf("bad: %v", out)
	}
}
