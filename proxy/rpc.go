//  Copyright 2014 wangox@gmail.com. All rights reserved.

package main

import (
	// "log"
	"net"
	"net/rpc"
)

type client struct {
	*rpc.Client
}

func NewClient(address string) (*client, error) {
	clt, e := rpc.DialHTTP("tcp", address)
	if e != nil {
		return nil, e
	}
	return &client{clt}, nil
}

// 和manager server之间的rpc接口

// 获取IP白名单接口
func (c *client) GetIPs() ([]net.IP, error) {
	var ips []net.IP
	e := c.Call("Manager.GetIPs", id, &ips)
	if e != nil {
		return nil, e
	}
	return ips, nil
}

type PortRange struct {
	Min, Max int
}

// 获取端口范围接口
func (c *client) GetHttpPortRange() (int, int, error) {
	var hpr PortRange
	e := c.Call("Manager.GetHttpPortRange", id, &hpr)
	if e != nil {
		return 0, 0, e
	}
	return hpr.Min, hpr.Max, nil
}

func (c *client) GetSocksPortRange() (int, int, error) {
	var spr PortRange
	e := c.Call("Manager.GetSocksPortRange", id, &spr)
	if e != nil {
		return 0, 0, e
	}
	return spr.Min, spr.Max, nil
}

// 和manager server之间心跳接口
func (c *client) Heartbeat() (int, error) {
	var r int
	e := c.Call("Manager.Heartbeat", id, &r)
	if e != nil {
		return 1, e
	}
	return r, nil
}
