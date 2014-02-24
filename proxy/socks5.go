//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"fmt"
	"github.com/armon/go-socks5"
	"log"
	"net"
)

// a SOCKS5 proxy server

func SocksProxy(ips []net.IP, port int) error {
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
