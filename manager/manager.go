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
	"runtime"
	"strings"
)

type DbParam struct {
	User, Password string
	Addr           string // ip:port
	DbName         string
	TableName      string
}

type Manager struct {
	ips      []net.IP
	hpr, spr map[string]PortRange
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

func (m *Manager) GetHttpPortRange(id string, r *PortRange) error {
	log.Println("Manager.GetHttpPortRange")
	_, ok := m.hpr[id]
	if !ok {
		m.hpr[id] = PortRange{8000, 8100}
	}
	*r = m.hpr[id]
	m.ss[id] = 1
	return nil
}

func (m *Manager) GetSocksPortRange(id string, r *PortRange) error {
	log.Println("Manager.GetSocksPortRange")
	_, ok := m.spr[id]
	if !ok {
		m.spr[id] = PortRange{10000, 10100}
	}
	*r = m.spr[id]
	m.ss[id] = 1
	return nil
}

func (m *Manager) Heartbeat(id string, r *int) error {
	log.Println("Manager.Heartbeat")
	*r = m.ss[id]
	return nil
}

var port = flag.String("port", ":15926", "管理服务器的服务端口")
var ipsFile = flag.String("ips", "ips.txt", "IP白名单")
var db = flag.String("db", "db.json", "数据库参数")

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
	b, e = ioutil.ReadFile(*db)
	if e != nil {
		log.Println(e)
	}
	var db DbParam
	e = json.Unmarshal(b, &db)
	if e != nil {
		log.Println(e)
	}
	// 读取数据库端口范围
	// hpr := PortRange{8000, 8100}
	// spr := PortRange{10000, 10100}

	m := &Manager{ips, make(map[string]PortRange), make(map[string]PortRange), make(map[string]int)}
	rpc.Register(m)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", *port)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	log.Println("manager server start...")
	http.Serve(l, nil)
}
