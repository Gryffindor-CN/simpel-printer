package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

func main() {
	//监听控制端口8009
	go makeControl()
	//监听服务端口8007
	go makeAccept()
	//监听转发端口8008
	go makeForward()
	//定时释放连接
	go releaseConnMatch()
	//执行tcp转发
	tcpForward()
}

type ConnMatch struct {
	accept *net.TCPConn
	acceptAddTime int64
	tunnel *net.TCPConn
}

var (
	cache *net.TCPConn = nil
	OUTTER_PORT string = "8007"
	TUNNEL_PORT string = "8008"
	CONTROL_PORT string = "8009"

	connListMap = make(map[string]*ConnMatch)
	lock = sync.Mutex{}

	connListMapUpdate = make(chan int)
)

func makeControl() {
	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", "127.0.0.1:" + CONTROL_PORT)
	tcpListener , err :=net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}
	log.Printf("[√] 正在监听控制端口[%s]...\n", CONTROL_PORT)
	for {
		tcpConn, err :=tcpListener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		if cache != nil {
			log.Println("已经存在一个客户端连接")
			tcpConn.Close()
		} else {
			cache = tcpConn
		}
		go control(tcpConn)
	}
}

func control(conn *net.TCPConn) {
	go func() {
		for {
			_, err := conn.Write([]byte("hi\n"))
			if err != nil {
				cache = nil
			}
			time.Sleep(time.Second * 2)
		}
	}()
}

func makeAccept() {
	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", "127.0.0.1:" + OUTTER_PORT)
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}
	defer tcpListener.Close()
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("已建立连接，监听端口[%s]\n", OUTTER_PORT)
		addConnMatchAccept(tcpConn)
		sendMessage("new\n")
	}
}

func addConnMatchAccept(accept *net.TCPConn) {
	lock.Lock()
	defer lock.Unlock()
	now := time.Now().UnixNano()
	connListMap[strconv.FormatInt(now, 10)] = &ConnMatch{accept, time.Now().Unix(), nil}
}

func sendMessage(msg string) {
	log.Printf("已经发送消息：%s\n", msg)
	if cache != nil {
		_, err := cache.Write([]byte(msg))
		if err != nil {
			log.Println("消息发送异常")
			panic(err)
		}
	} else {
		log.Println("没有客户端连接，无法发送消息")
	}
}

func makeForward() {
	var tcpAddr *net.TCPAddr
	tcpAddr, _ = net.ResolveTCPAddr("tcp", "127.0.0.1:" + TUNNEL_PORT)
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		panic(err)
	}
	defer tcpListener.Close()
	log.Println("等待客户端连接...")
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("%s 已连接端口 %s。。。\n", tcpConn.RemoteAddr().String(), TUNNEL_PORT)
		configConnListTunnel(tcpConn)
	}
}

func configConnListTunnel(tunnel *net.TCPConn) {
	lock.Lock()
	used := false
	for _, connMatch := range connListMap {
		if connMatch.tunnel == nil && connMatch.accept != nil {
			connMatch.tunnel = tunnel
			used = true
			break
		}
	}
	if !used {
		log.Println(len(connListMap))
		_ = tunnel.Close()
		log.Println("正在关闭多余的tunnel")
	}
	lock.Unlock()
	//TODO connListMapUpdate <- UPDATE
}

func tcpForward() {
	for {
		select {
		case <- connListMapUpdate:
			lock.Lock()
			for key, connMatch := range connListMap {
				if connMatch.tunnel != nil && connMatch.accept != nil {
					log.Println("建立 tunnel 连接...")
					go joinConn2(connMatch.accept, connMatch.tunnel)
					delete(connListMap, key)
				}
			}
			lock.Unlock()
		}
	}
}

func joinConn2(conn1 *net.TCPConn, conn2 *net.TCPConn) {
	f := func(local *net.TCPConn, remote *net.TCPConn) {
		//defer保证close
		defer local.Close()
		defer remote.Close()
		//使用io.Copy传输两个tcp连接，
		_, err := io.Copy(local, remote)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println("join Conn2 end")
	}
	go f(conn2, conn1)
	go f(conn1, conn2)
}

func releaseConnMatch() {
	for {
		lock.Lock()
		for key, connMatch := range connListMap {
			//如果在指定时间内没有tunnel的话，则释放该连接
			if connMatch.tunnel == nil && connMatch.accept != nil {
				if time.Now().Unix()-connMatch.acceptAddTime > 5 {
					fmt.Println("释放超时连接")
					err := connMatch.accept.Close()
					if err != nil {
						fmt.Println("释放连接的时候出错了:" + err.Error())
					}
					delete(connListMap, key)
				}
			}
		}
		lock.Unlock()
		time.Sleep(5 * time.Second)
	}
}