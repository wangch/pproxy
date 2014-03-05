//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
)

// privateproxyIP 参数
type PPP struct {
	LocalPort string // 对应的本地端口 ":port"
	Username  string // 代理认证用户名
	Password  string // 代理认证密码
}

type Config struct {
	HttpConf, SocksConf map[string]PPP // map key is addr, format: IP:PORT
}

var confFile = flag.String("c", "fconf.json", "本地配置文件")
var ipsFile = flag.String("ips", "ips.txt", "IP白名单")
var debug = flag.Bool("d", false, "打印调试日志")

var traceLog *log.Logger //log.New(ioutil.Discard, "[trace]", log.Ltime)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	if *debug {
		w, e := os.Create("forwardLog.txt")
		if e != nil {
			panic(e)
		}
		defer w.Close()
		traceLog = log.New(w, "@@@", log.Ldate|log.Ltime)
	}

	var conf Config
	var ips []net.IP

	// 读取本地IP白名单
	b, e := ioutil.ReadFile(*ipsFile)
	if e != nil {
		log.Println(e)
	}
	ss := strings.Split(string(b), "\r\n")
	for _, s := range ss {
		ip := net.ParseIP(s)
		if ip != nil {
			ips = append(ips, ip)
		}
	}

	// 读取本地配置文件
	data, e := ioutil.ReadFile(*confFile)
	if e != nil {
		httpconf := map[string]PPP{
			"127.0.0.1:12345": PPP{":23456", "username", "password"},
		}
		socksconf := map[string]PPP{
			"127.0.0.1:34567": PPP{":45678", "username", "password"},
		}
		conf := &Config{httpconf, socksconf}
		b, e := json.MarshalIndent(conf, "", "    ")
		if e != nil {
			panic(e)
		}
		e = ioutil.WriteFile(*confFile, b, os.ModePerm)
		if e != nil {
			panic(e)
		}
		log.Println("Please edit config file for port maping")
		return
	}

	e = json.Unmarshal(data, &conf)
	if e != nil {
		log.Fatal(e)
	}

	// 建立http代理
	for k, v := range conf.HttpConf {
		go HttpProxy(ips, k, v.LocalPort, v.Username, v.Password)
	}
	// 建立socks5代理
	for k, v := range conf.SocksConf {
		go SocksProxy(ips, k, v.LocalPort, v.Username, v.Password)
	}

	log.Println("forward server start... and type 'q' to exit")

	b = make([]byte, 1)
	for {
		_, e := os.Stdin.Read(b)
		if e != nil {
			panic(e)
		}

		switch b[0] {
		case 'q':
			return
		}
	}
}
