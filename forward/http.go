//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// a HTTP proxy server

type httpProxy struct {
	*httputil.ReverseProxy
	ips            []net.IP
	user, password string
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

	t, _ := p.Transport.(*http.Transport)
	t.Proxy = func(req *http.Request) (*url.URL, error) {
		req.URL.User = url.UserPassword(p.user, p.password)
		return req.URL, nil
	}

	p.ReverseProxy.ServeHTTP(w, r)
}

func HttpProxy(ips []net.IP, raddr, laddr, user, password string) error {
	u := &url.URL{}
	rp := httputil.NewSingleHostReverseProxy(u)
	rp.Director = func(r *http.Request) {
		return
	}
	f := func(network string, addr string) (net.Conn, error) {
		return net.Dial("tcp", raddr)
	}
	t := &http.Transport{Dial: f}
	p := &httpProxy{rp, ips, user, password}
	p.Transport = t

	e := http.ListenAndServe(laddr, p)
	if e != nil {
		return e
	}
	return nil
}
