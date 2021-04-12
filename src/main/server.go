package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	raddr := net.UDPAddr{
		Port: 6112,
		//IP: net.ParseIP("0.0.0.0"),
		IP: net.ParseIP("255.255.255.255"),
	}
	laddr := net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP("0.0.0.0"),
		//IP:   net.ParseIP("255.255.255.255"),
	}
	//laddr := net.UDPAddr{
	//	Port: 11260,
	//	IP: net.ParseIP("0.0.0.0"),
	//}
	ln, err := net.DialUDP("udp4", &laddr, &raddr)
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 16)
	for {
		read, client, err := ln.ReadFromUDP(buf)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(buf[0:read], " from ", client)
	}

}
