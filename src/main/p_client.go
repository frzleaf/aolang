package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	raddr := net.TCPAddr{
		//IP:   net.ParseIP("127.0.0.1"),
		IP:   net.ParseIP("115.146.126.31"),
		Port: 7128,
	}
	tcp, err := net.DialTCP("tcp", nil, &raddr)
	if err != nil {
		log.Fatal(err)
	}
	defer tcp.Close()
	response := make([]byte, 1000)
	if read, err := tcp.Read(response); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Listen at: ", tcp.LocalAddr())
		fmt.Println(string(response[0:read]))
	}
	if _, err := tcp.Write([]byte("test")); err != nil {
		log.Fatal(err)
	} else {
		if receiverTcp, err := net.Listen("tcp", tcp.LocalAddr().String()); err != nil {
			log.Fatal(err)
		} else {
			defer receiverTcp.Close()
			for {
				if accept, err := receiverTcp.Accept(); err != nil {
					log.Fatal(err)
				} else {
					go func() {
						defer accept.Close()
						buf := make([]byte, 1000)
						name := accept.RemoteAddr()
						fmt.Println("Connection from: ", name)
						for {
							if read, err := accept.Read(buf); err != nil {
								fmt.Println(err)
								break
							} else {
								fmt.Println(name, string(buf[0:read]))
								tcp.Write(buf[0:read])
							}
						}
					}()
				}
			}
		}
	}
}
