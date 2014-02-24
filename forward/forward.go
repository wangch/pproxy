//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"runtime"
	"strings"
)

// privateproxyIP 参数
type PPP struct {
	LocalPort string // 对应的本地端口
	Username  string // 代理认证用户名
	Password  string // 代理认证密码
}

type Config struct {
	HttpConf, SocksConf map[string]PPP // map key is addr, format: IP:PORT
}

var confFile = flag.String("c", "conf.json", "本地配置文件")
var ipsFile = flag.String("ips", "ips.txt", "IP白名单")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

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
		log.Fatal(e)
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

	log.Println("forward server start...")
}
