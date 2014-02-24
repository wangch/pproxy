//  Copyright 2014 wangox@gmail.com. All rights reserved.

// a HTTP proxy server and SOCKS5 proxy server

package main

import (
	"flag"
	"log"
	"runtime"
	"time"
)

var manger = flag.String("m", "", "管理服务器的地址, 格式为 IP:PORT, 如:211.323.197.314:10015")
var local = flag.String("l", "", "本地IP")

var id string

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	if len(*manger) == 0 {
		log.Println("必须设置管理服务器IP和端口")
		flag.Usage()
		return
	}
	if len(*local) == 0 {
		log.Println("必须设置本地IP地址")
		flag.Usage()
		return
	}

	id = *local

	for {
		clt, e := NewClient(*manger)
		if e != nil {
			log.Println("管理服务器连接不上", e)
			time.Sleep(10 * time.Second)
			continue
		}

		// 获取manager server 的ip白名单和端口范围
		ips, e := clt.GetIPs()
		if e != nil {
			log.Println(e)
		}
		log.Println("ips:", ips)

		minh, maxh, e := clt.GetHttpPortRange()
		if e != nil {
			log.Println(e)
			minh = 8000
			maxh = 8100
		}
		log.Println("http port range:", minh, maxh)

		mins, maxs, e := clt.GetSocksPortRange()
		if e != nil {
			log.Println(e)
			minh = 10000
			maxh = 10100
		}
		log.Println("socks port range:", mins, maxs)

		// 建立不同端口的http proxy 和socks5 proxy
		for i := minh; i <= maxh; i++ {
			go HttpProxy(ips, i)
		}
		for i := mins; i <= maxs; i++ {
			go SocksProxy(ips, i)
		}
		// 和manager server 心跳循环, 当返回标志位RESET时，跳出循环
		tch := time.Tick(time.Second * 5)
		en := 0
		for {
			select {
			case <-tch:
				status, e := clt.Heartbeat()
				if e != nil {
					log.Println(e)
					en++
				}
				if status == 0 {
					log.Println("proxy reset")
					break
				}
			}
			if en >= 3 {
				log.Println("proxy reset")
				break
			}
		}
	}
}
