package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	//raddr := net.UDPAddr{
	//	Port: 6112,
	//	//IP: net.ParseIP("0.0.0.0"),
	//	IP: net.ParseIP("255.255.255.255"),
	//}
	//addr, err := net.ResolveUDPAddr("udp", ":0")
	//if err != nil {
	//	log.Fatal(err)
	//}
	laddr := net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP("0.0.0.0"),
		//IP:   net.ParseIP("10.97.72.86"),
	}
	ln, err := net.ListenUDP("udp4", &laddr)
	//ln, err := net.DialUDP("udp4", addr, &raddr)
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
