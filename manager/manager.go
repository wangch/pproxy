//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"runtime"
	"strings"
)

type Config struct {
	Http, Socks5 PortRange
}

type Manager struct {
	ips      []net.IP
	hpr, spr PortRange
	ss       map[string]int
}

func (m *Manager) GetIPs(id string, ips *[]net.IP) error {
	log.Println("Manager.GetIPs")
	*ips = m.ips
	m.ss[id] = 1
	return nil
}

type PortRange struct {
	Min, Max int
}

var defaultHttpPortRange = PortRange{40001, 40050}
var defaultSocks5PortRange = PortRange{41001, 41050}

func (m *Manager) GetHttpPortRange(id string, r *PortRange) error {
	log.Println("Manager.GetHttpPortRange", id)
	*r = m.hpr
	m.ss[id] = 1
	return nil
}

func (m *Manager) GetSocksPortRange(id string, r *PortRange) error {
	log.Println("Manager.GetSocksPortRange", id)
	*r = m.spr
	m.ss[id] = 1
	return nil
}

func (m *Manager) Heartbeat(id string, r *int) error {
	log.Println("Manager.Heartbeat", id)
	*r = m.ss[id]
	return nil
}

var port = flag.String("port", ":15926", "管理服务器的服务端口")
var ipsFile = flag.String("ips", "ips.txt", "IP白名单文件")
var confFile = flag.String("conf", "conf.json", "配置文件")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	// 读取本地IP白名单
	var ips []net.IP
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
	// 读取本地数据库配置参数
	var conf Config
	hpr, spr := defaultHttpPortRange, defaultSocks5PortRange
	b, e = ioutil.ReadFile(*confFile)
	if e != nil {
		log.Println(e)
		conf.Http = hpr
		conf.Socks5 = spr
		b, e = json.MarshalIndent(&conf, "", " ")
		if e != nil {
			log.Fatal(e)
		}
		ioutil.WriteFile(*confFile, b, os.ModePerm)
	} else {
		e = json.Unmarshal(b, &conf)
		if e != nil {
			log.Println(e)
		}
		hpr = conf.Http
		spr = conf.Socks5
	}

	m := &Manager{ips, hpr, spr, make(map[string]int)}
	rpc.Register(m)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", *port)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	log.Println("manager server start...")
	http.Serve(l, nil)
}
