//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// a HTTP proxy server

type httpProxy struct {
	*httputil.ReverseProxy
	ips []net.IP
}

func (p *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip, _, e := net.SplitHostPort(r.RemoteAddr)
	log.Println(ip, r.Method, r.RequestURI)
	if e != nil {
		log.Println(e)
		return
	}
	b := false
	for _, x := range p.ips {
		if x.String() == ip {
			b = true
			break
		}
	}
	if len(p.ips) == 0 {
		b = true
	}
	if !b {
		w.Write([]byte("the proxy not surpport this ip"))
		return
	}

	p.ReverseProxy.ServeHTTP(w, r)
}

func HttpProxy(ips []net.IP, port int) error {
	s := fmt.Sprintf("localhost:%d", port)
	u, e := url.Parse(s)
	if e != nil {
		log.Println(e)
		return e
	}
	rp := httputil.NewSingleHostReverseProxy(u)
	rp.Director = func(r *http.Request) {
		return
	}
	p := &httpProxy{rp, ips}
	e = http.ListenAndServe(s, p)
	if e != nil {
		return e
	}
	return nil
}
