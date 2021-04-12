package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 1234,
	})
	if err != nil {
		log.Fatalln(err)
	}
	for {
		data := make([]byte, 4096)
		read, remoteAddr, err := socket.ReadFromUDP(data)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(data[0:read], " from ", remoteAddr)
	}
}
