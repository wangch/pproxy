//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"encoding/base64"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// onExitFlushLoop is a callback set by tests to detect the state of the
// flushLoop() goroutine.
var onExitFlushLoop func()

// ReverseProxy is an HTTP Handler that takes an incoming request and
// sends it to another server, proxying the response back to the
// client.
type ReverseProxy struct {
	// The transport used to perform proxy requests.
	// If nil, http.DefaultTransport is used.
	Transport *http.Transport

	// FlushInterval specifies the flush interval
	// to flush to the client while copying the
	// response body.
	// If zero, no periodic flushing is done.
	FlushInterval time.Duration

	ips            []net.IP
	user, password string
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func (p *ReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	b := false
	clientIP, _, e := net.SplitHostPort(req.RemoteAddr)
	if e == nil {
		ip := net.ParseIP(clientIP)
		for _, x := range p.ips {
			if ip.Equal(x) {
				b = true
				break
			}
		}
		if prior, ok := req.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		req.Header.Set("X-Forwarded-For", clientIP)
	}
	if len(p.ips) == 0 {
		b = true
	}
	if !b {
		rw.Write([]byte("the proxy not surpport this ip"))
		return
	}

	// log.Printf("recv http request: %#v", req.URL)

	req.URL.Scheme = "http"

	p.Transport.Proxy = func(r *http.Request) (*url.URL, error) {
		return req.URL, nil
	}

	res, err := p.Transport.RoundTrip(req)
	if err != nil {
		log.Printf("http: proxy error: %v", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	if res.StatusCode == 407 { // 需要认证
		req.Header.Set("Proxy-Authorization", base64.StdEncoding.EncodeToString([]byte(p.user+":"+p.password)))
		res.Body.Close()
		res, err = p.Transport.RoundTrip(req)
		if err != nil {
			log.Printf("http: proxy error: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	defer res.Body.Close()

	copyHeader(rw.Header(), res.Header)

	rw.WriteHeader(res.StatusCode)
	p.copyResponse(rw, res.Body)
}

func (p *ReverseProxy) copyResponse(dst io.Writer, src io.Reader) {
	if p.FlushInterval != 0 {
		if wf, ok := dst.(writeFlusher); ok {
			mlw := &maxLatencyWriter{
				dst:     wf,
				latency: p.FlushInterval,
				done:    make(chan bool),
			}
			go mlw.flushLoop()
			defer mlw.stop()
			dst = mlw
		}
	}

	io.Copy(dst, src)
}

type writeFlusher interface {
	io.Writer
	http.Flusher
}

type maxLatencyWriter struct {
	dst     writeFlusher
	latency time.Duration

	lk   sync.Mutex // protects Write + Flush
	done chan bool
}

func (m *maxLatencyWriter) Write(p []byte) (int, error) {
	m.lk.Lock()
	defer m.lk.Unlock()
	return m.dst.Write(p)
}

func (m *maxLatencyWriter) flushLoop() {
	t := time.NewTicker(m.latency)
	defer t.Stop()
	for {
		select {
		case <-m.done:
			if onExitFlushLoop != nil {
				onExitFlushLoop()
			}
			return
		case <-t.C:
			m.lk.Lock()
			m.dst.Flush()
			m.lk.Unlock()
		}
	}
}

func (m *maxLatencyWriter) stop() {
	m.done <- true
}

type conn struct {
	net.Conn
}

func (c *conn) Serve(addr string, ips []net.IP) {
	b := false
	if len(ips) == 0 {
		b = true
	}
	ip := net.ParseIP(c.RemoteAddr().String())
	for _, x := range ips {
		if ip.Equal(x) {
			b = true
			break
		}
	}
	if !b {
		buf := []byte("HTTP/1.1 200 OK\r\n\r\n\r\n\r\n not surpport this ip:" + ip.String())
		c.Write(buf)
		return
	}

	cc, e := net.Dial("tcp", addr)
	if e != nil {
		log.Println(e)
		return
	}
	defer cc.Close()
	defer c.Close()
	go func() {
		_, e := io.Copy(c, cc)
		if e != nil {
			log.Println(e)
			c.Close()
		}
	}()
	_, e = io.Copy(cc, c)
	if e != nil {
		log.Println(e)
		c.Close()
	}
	return
}

type httpForward struct {
	ips   []net.IP
	raddr string
}

func (s *httpForward) Serve(l net.Listener) error {
	for {
		c, e := l.Accept()
		if e != nil {
			log.Println(e)
			return e
		}
		cc := &conn{c}
		go cc.Serve(s.raddr, s.ips)
	}
	return nil
}

func (s *httpForward) ListenAndServe(network string, laddr string) error {
	l, e := net.Listen(network, laddr)
	if e != nil {
		log.Println(e)
		return e
	}
	return s.Serve(l)
}

func HttpProxy2(ips []net.IP, laddr, raddr, user, password string) error {
	s := &httpForward{ips, raddr}
	e := s.ListenAndServe("tcp", laddr)
	if e != nil {
		log.Println(e)
		return e
	}
	return nil
}

func HttpProxy(ips []net.IP, laddr, raddr, user, password string) error {
	f := func(network string, addr string) (net.Conn, error) {
		return net.Dial("tcp", raddr)
	}
	t := &http.Transport{Dial: f, Proxy: http.ProxyFromEnvironment}
	p := &ReverseProxy{Transport: t, ips: ips, user: user, password: password}

	e := http.ListenAndServe(laddr, p)
	if e != nil {
		log.Println(e)
		return e
	}
	return nil
}
