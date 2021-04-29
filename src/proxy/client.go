package proxy

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	host  *Host
	guest *Guest
	sConn net.Conn
}

var dstClient int

func NewClient() *Client {
	return &Client{}
}

func (c *Client) watchToGame(onExit func()) {
	defer onExit()
	buf := make([]byte, 1000)
	for {
		if r, err1 := c.sConn.Read(buf); err1 != nil {
			if err1.Error() != "EOF" && !c.isClosed {
				fmt.Println("Có lỗi xảy ra:" + err1.Error())
			}
			return
		} else {
			packets := PacketFromBytes(buf[0:r])
			for _, packet := range packets {
				fmt.Printf("\nReceive msg from #%v len %v\n", packet.SrcAddr(), len(packet.Data()))
				switch packet.pkgType {
				case PackageTypeFindHostResponse:
					err1 := c.SendListGameAndOpenVirtualHost(packet.data)
					if err1 != nil {
						fmt.Println(err1)
					}
				case PackageTypeInform:
					var conn net.Conn
					// Host behavior
					if c.pServer == nil {
						conn = c.gConn[packet.SrcAddr()]
						if conn == nil {
							conn, err1 = c.PrepareNewGameConnection(packet.src)
							if err1 != nil {
								fmt.Println("Lỗi khi connect host: ", err1)
								break
							}
							fmt.Println("New client join host: ", packet.SrcAddr())
						}
					} else {
						conn = c.pConn
					}
					if conn != nil {
						if _, err1 = conn.Write(packet.data); err1 != nil {
							fmt.Println("Error on write: ", err1)
						}
					} else {
						fmt.Println("WARN: receive data but pConn is null: ", string(packet.data))
					}
				case PackageTypeConnectHost:
					gConn := c.gConn[packet.SrcAddr()]
					if gConn == nil {
						gConn, err1 = c.PrepareNewGameConnection(packet.src)
						if err1 != nil {
							fmt.Println("Lỗi khi connect host: ", err1)
							break
						}
						fmt.Println("New client join host: ", packet.SrcAddr())
					}
					if _, err := gConn.Write(packet.data); err != nil {
						fmt.Println(err)
					}
				case PackageTypeFindHost:
					if gameData, err := c.GetGameList(); err == nil {
						c.sConn.Write(NewPacket(PackageTypeFindHostResponse, c.id, packet.SrcAddr(), gameData).ToBytes())
					} else {
						fmt.Println("No game found: ", err)
					}
				}
			}
		}
	}
}

func (c *Client) Close() {
	defer c.sConn.Close()
	defer c.sConn.Close()
	defer func() {
		for _, conn := range c.gConn {
			conn.Close()
		}
	}()
	c.isClosed = true
}

func (c *Client) readCommand() {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Gửi lệnh lên server: ")
		line, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println("Lỗi: ", err)
			continue
		}
		lineStr := string(line)
		if strings.HasPrefix(lineStr, "to ") {
			split := strings.Split(lineStr, " ")
			if len(split) == 1 {
				fmt.Println("Missing 2nd argument")
				continue
			} else {
				input2nd, err := strconv.Atoi(split[1])
				if err != nil {
					fmt.Println("Invalid args: ", err)
					continue
				}
				dstClient = input2nd
				fmt.Println("Selected: ", input2nd)
			}
		} else if strings.HasPrefix(lineStr, "find") {
			c.sConn.Write(NewPacket(PackageTypeFindHost, c.id, dstClient, nil).ToBytes())
		} else if strings.HasPrefix(lineStr, "exit") {
			c.Close()
		} else {
			c.sConn.Write(NewSelectPacket(c.id, dstClient, line).ToBytes())
		}
	}
}

func (c *Client) closeGConn(targetId int) {
	if conn := c.gConn[targetId]; conn != nil {
		defer conn.Close()
		delete(c.gConn, targetId)
	}
}

func (c *Client) SendToOther(targetId int, data []byte) {
	c.sConn.Write(NewInformPacket(c.id, targetId, data).ToBytes())
}

func (c *Client) Connect(sAddr string) error {
	connection, err := net.Dial("tcp", sAddr)
	if err != nil {
		return err
	}
	fmt.Println("Connected to: " + sAddr)
	buf := make([]byte, 1000)
	if read, err := connection.Read(buf); err != nil {
		return err
	} else {
		packet := PacketFromBytes(buf[0:read])
		c.id = packet[0].DstAddr()
		fmt.Println("Connection ID: ", c.id)
	}
	c.sConn = connection

	c.Lock()

	go c.watchToGame(func() {
		fmt.Println("End game")
		c.Unlock()
	})

	go c.readCommand()

	c.Lock()
	return nil
}
