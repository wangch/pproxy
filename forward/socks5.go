//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"code.google.com/p/go.net/proxy"
	"fmt"
	"github.com/wangch/go-socks5"
	"log"
	"net"
)

func SocksProxy(ips []net.IP, addr, port, user, password string) error {
	auth := &proxy.Auth{user, password}
	dialer, e := proxy.SOCKS5("tcp", addr, auth, proxy.Direct)
	if e != nil {
		log.Println(e)
		return e
	}

	conf := &socks5.Config{Dialer: dialer}
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
