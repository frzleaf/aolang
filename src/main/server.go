package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Không điền đủ tham số, vui lòng nhập lại theo mẫu: ")
		fmt.Println("warlan 10.0.0.1 10.0.0.2")
		os.Exit(0)
	}

	lAddr := os.Args[1]
	rAddr := os.Args[2]

	for {
		watchAndForward(lAddr, rAddr)
	}
}

func watchAndForward(lAddr, rAddr string) {
	lUdpaddr := net.UDPAddr{
		Port: 6110,
		IP:   net.ParseIP(lAddr),
	}

	l2Udpaddr := net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP(lAddr),
	}

	raddr := net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP(rAddr),
	}

	scannerConn, err := net.DialUDP("udp4", &lUdpaddr, &raddr)
	if err != nil {
		log.Fatal(err)
	}

	defer scannerConn.Close()
	buf := make([]byte, 1000)
	var scanCounter byte = 0

	joinedGame := false
	gameListConn, err := net.DialUDP("udp4", nil, &l2Udpaddr)
	for {
		_, err = scannerConn.Write(
			[]byte{0xf7, 0x2f, 0x10, 0x00, 0x50, 0x58, 0x33, 0x57, 0x18, 0x00, 0x00, 0x00, scanCounter, 0x00, 0x00, 0x00},
		)
		scanCounter = scanCounter + 1
		scannerConn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
		read, client, err := scannerConn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		if err != nil {
			log.Println(err)
		} else {
			fmt.Println("Host " + client.IP.String() + "     : " + string(buf[0:read]))
			go func() {
				defer gameListConn.Close()
				for !joinedGame {
					_, err = gameListConn.Write(buf[0:read])
					time.Sleep(time.Second)
				}
			}()
			break
		}
	}

	lTcpAddr := net.TCPAddr{
		IP:   net.ParseIP(lAddr),
		Port: 6110,
	}
	rTcpAddr := net.TCPAddr{
		IP:   net.ParseIP(rAddr),
		Port: 6110,
	}
	ltcp, err := net.ListenTCP("tcp", &lTcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer ltcp.Close()
	tcpConn, err := ltcp.AcceptTCP()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Đã join game!")
	joinedGame = true
	outGame := false
	go intermediate(tcpConn, rTcpAddr, func() {
		fmt.Println("Đã out game")
		outGame = true
	})
	for !outGame {
		time.Sleep(time.Millisecond * 500)
	}
}

func intermediate(tcpConn net.Conn, rTcpAddr net.TCPAddr, onExit func()) {
	rtcp, err := net.DialTCP("tcp", nil, &rTcpAddr)
	if err != nil {
		log.Fatal(err)
	}

	buf1 := make([]byte, 1000)
	buf2 := make([]byte, 1000)

	go func() {
		defer tcpConn.Close()
		defer onExit()
		for {
			if r, err1 := tcpConn.Read(buf1); err1 != nil {
				if err1.Error() != "EOF" {
					fmt.Println("Có lỗi xảy ra:" + err1.Error())
				}
				break
			} else {
				rtcp.Write(buf1[0:r])
			}
		}
	}()

	go func() {
		defer rtcp.Close()
		for {
			if r, err1 := rtcp.Read(buf2); err1 != nil {
				if err1.Error() != "EOF" {
					fmt.Println("Có lỗi xảy ra:" + err1.Error())
				}
				break
			} else {
				tcpConn.Write(buf2[0:r])
			}
		}
	}()
}
