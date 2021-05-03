package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

func main() {
	ip, _ := findLocalIp("ist02.com:9999")
	fmt.Println(ip)
}

func findLocalIp(targetAddr string) (string, error) {
	if dial, err := net.Dial("tcp", targetAddr); err != nil {
		return "", err
	} else {
		return dial.LocalAddr().String(), nil
	}
}

func testPorts(from, to int) {
	lock := sync.Mutex{}
	lock.Lock()
	for port := from; port <= to; port++ {
		var localPort = port
		go func() {
			server, _ := net.Listen("tcp", ":"+strconv.Itoa(localPort))
			if server != nil {
				defer server.Close()
			} else {
				return
			}
			fmt.Println("Wait on: ", server.Addr())
			if conn, err := server.Accept(); err == nil {
				fmt.Println("Connect from: ", conn.LocalAddr(), " to ", conn.RemoteAddr())
				conn.Close()
				lock.Unlock()
			}
		}()
	}
	lock.Lock()
}
