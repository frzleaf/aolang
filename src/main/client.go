package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	ln, _ := net.ListenUDP("udp4", &net.UDPAddr{Port: 11260})
	raddr := &net.UDPAddr{IP: []byte{255, 255, 255, 255}, Port: 6112}
	b := make([]byte, 128)
	for {
		n, _ := ln.Read(b)
		fmt.Println("Echoing", string(b[:n]))
		_, err := ln.WriteToUDP(b[:n], raddr)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
