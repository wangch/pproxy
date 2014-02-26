//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
)

// a HTTP proxy server

type httpProxys struct {
	*httputil.ReverseProxy
	ips []net.IP
}

func (p *httpProxys) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.ReverseProxy.ServeHTTP(w, r)
}

func HttpProxyServe(ips []net.IP, port int) error {
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
	p := &httpProxys{rp, ips}
	e = http.ListenAndServe(s, p)
	if e != nil {
		return e
	}
	return nil
}

func TestHttpProxy(t *testing.T) {
	// create http server
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	defer ts.Close()
	go ts.Start()

	// create remote http proxy server
	go func() {
		e := HttpProxyServe(nil, 9091)
		if e != nil {
			t.Fatal(e)
		}
	}()

	// create forward server
	go func() {
		e := HttpProxy3(nil, "localhost:9091", ":9002", "user", "password")
		if e != nil {
			t.Fatal(e)
		}
	}()

	// connect forward server
	conn, e := net.Dial("tcp", "127.0.0.1:9002")
	if e != nil {
		t.Fatal(e)
	}
	defer conn.Close()

	// send request
	conn.Write([]byte("GET http://127.0.0.1:8080 HTTP/1.1\r\n\r\n"))

	// read respinse
	r := bufio.NewReader(conn)
	resp, e := http.ReadResponse(r, nil)
	if e != nil {
		t.Error(e)
	}
	defer resp.Body.Close()

	buf := make([]byte, 128)
	n, e := resp.Body.Read(buf)
	if string(buf[:n]) != "hello" {
		t.Error(string(buf))
	}
}
