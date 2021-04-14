package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	laddr := net.UDPAddr{
		Port: 6110,
		IP:   net.ParseIP("10.8.0.2"),
	}

	l2addr := net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP("10.8.0.2"),
	}

	//listen, err2 := net.ListenUDP("udp", &l2addr)
	//if err2 != nil {
	//	log.Fatal(err2)
	//}
	//defer listen.Close()
	//go func() {
	//	for {
	//		var inputData []byte
	//		udp, addr, err := listen.ReadFromUDP(inputData)
	//		if err != nil {
	//			fmt.Println(err)
	//		} else {
	//			fmt.Println(inputData[0:udp], " from ", addr)
	//		}
	//		time.Sleep(time.Second * 1)
	//	}
	//}()

	raddr := net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP("10.8.0.6"),
	}
	ln, err := net.DialUDP("udp4", &laddr, &raddr)
	if err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 1000)
	//for {
	_, err = ln.Write(
		[]byte{0xf7, 0x2f, 0x10, 0x00, 0x50, 0x58, 0x33, 0x57, 0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		//&raddr,
	)
	read, client, err := ln.ReadFromUDP(buf)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf[0:read], " from ", client)

	dial, err := net.DialUDP("udp4", nil, &l2addr)
	//dial, err := net.DialUDP("udp4", &laddr, &l2addr)
	if err != nil {
		log.Println(err)
	} else {
		_, err = dial.Write(buf[0:read])
		if err != nil {
			log.Println(err)
		}
	}
	tcpAddr := net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 6112,
	}
	tcp, err := net.ListenTCP("tcp", &tcpAddr)
	if err != nil {
		fmt.Println(err)
	} else {
		acceptTCP, _ := tcp.AcceptTCP()
		for {
			if i, err := acceptTCP.Read(buf); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Receive", buf[0:i], acceptTCP.RemoteAddr())
			}
		}
	}
	//time.Sleep(time.Second * 1)
	//}

}
