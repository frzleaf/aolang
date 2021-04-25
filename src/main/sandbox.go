package main

import (
	"fmt"
	"net"
	"strconv"
	"sync"
)

func main() {
	testPorts(6110, 6112)
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
