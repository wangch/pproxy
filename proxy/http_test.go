//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpProxy(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	defer ts.Close()
	go ts.Start()

	go func() {
		e := HttpProxy(nil, 8000)
		if e != nil {
			t.Fatal(e)
		}
	}()

	// connect forward server
	conn, e := net.Dial("tcp", "127.0.0.1:8000")
	if e != nil {
		t.Fatal(e)
	}
	defer conn.Close()

	// send request
	conn.Write([]byte("GET http://127.0.0.1/ HTTP/1.1\r\n\r\n"))

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
		t.Error(buf)
	}
}
