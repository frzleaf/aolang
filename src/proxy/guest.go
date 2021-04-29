package proxy

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Guest struct {
	sConn    net.Conn     // server connection
	pConn    net.Conn     // player connection
	pServer  net.Listener // virtual host port
	isClosed bool
	id       int
	sync.Mutex
}

func (c *Guest) SendListGameAndProxyHost(listGameData []byte) error {
	rUdpaddr := net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP(strings.Split(os.Args[2], ":")[0]),
	}
	gameListConn, err := net.DialUDP("udp4", nil, &rUdpaddr)
	if err != nil {
		return err
	}
	gameListConn.Write(listGameData)
	if c.pServer != nil {
		return nil
	}
	//net.Listen("tcp", ":" + os.Args[2])
	c.pServer, err = net.Listen("tcp", os.Args[2])
	//c.pServer, err = net.Listen("tcp", "10.8.0.2:6110")
	go func() {
		for {
			if c.pConn, err = c.pServer.Accept(); err == nil {
				fmt.Println("New connection: ", c.pConn.RemoteAddr())
				//c.sConn.Write(NewInformPacket(c.id, dstClient, nil).ToBytes())

				// new connection
				go func() {
					buf := make([]byte, 1000)
					for {
						read, err := c.pConn.Read(buf)
						if err != nil {
							if err.Error() == "EOF" {
								continue
							}
							fmt.Println("Error on game read ", err)
							break

						} else {
							_, err := c.sConn.Write(NewInformPacket(c.id, dstClient, buf[0:read]).ToBytes())
							if err != nil {
								fmt.Println("Error: ", err)
							}
						}
					}
				}()
			}
		}
	}()
	return err
}
