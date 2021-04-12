package main

import (
	"log"
	"net"
)

func main() {
	BROADCAST_IPv4 := net.IPv4(255, 255, 255, 255)
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   BROADCAST_IPv4,
		Port: 6112,
	})
	if err != nil {
		log.Fatalln(err)
	}
	socket.Write([]byte("test"))
}
