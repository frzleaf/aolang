package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	//BROADCAST_IPv4 := net.IPv4(255, 255, 255, 255)
	BROADCAST_IPv4 := net.IPv4(10, 97, 72, 86)
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   BROADCAST_IPv4,
		Port: 6112,
	})
	if err != nil {
		log.Fatalln(err)
	}
	_, err = socket.Write([]byte{247, 47, 16, 0, 80, 88, 51, 87, 24, 0, 0, 0, 0, 0, 0, 0})
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 1000)
	udp, addr, err := socket.ReadFromUDP(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf[0:udp], " from ", addr)
}
