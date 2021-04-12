package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	addr := net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP("127.0.0.1"),
	}
	ln, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Fatal(err)
	}

	var buf [50]byte
	counter := 0
	for {
		read, client, err := ln.ReadFromUDP(buf[:])
		if err != nil {
			log.Fatal(err, read)
		}
		fmt.Println(buf[0:read])
		fmt.Println(counter)
		_, _, e := ln.WriteMsgUDP([]byte("tesst"), nil, client)
		if e != nil {
			fmt.Errorf("Can not send %v", err)
		} else {
			fmt.Println("send ok")
			if counter == 10 {
				log.Fatal("end")
			}
		}
		counter++
	}

}