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
	sConn    net.Conn         // server connection
	gConn    map[int]net.Conn // game connections
	pConn    net.Conn         // player connection
	pServer  net.Listener     // virtual host port
	isClosed bool
	id       int
	sync.Mutex
}

var dstClient int

func NewClient() *Client {
	return &Client{
		gConn: make(map[int]net.Conn),
		id:    -1,
	}
}

func (c *Client) watchToGame(onExit func()) {
	defer onExit()
	buf := make([]byte, 1000)
	var trickyBytes []byte
	for {
		if r, err1 := c.sConn.Read(buf); err1 != nil {
			if err1.Error() != "EOF" && !c.isClosed {
				fmt.Println("Có lỗi xảy ra:" + err1.Error())
			}
			return
		} else {
			packet := PacketFromBytes(buf[0:r])
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
					fmt.Println("On data ", string(packet.data))
					if trickyBytes == nil && len(packet.data) > 200 {
						trickyBytes = packet.data
					} else {
						if len(trickyBytes) > 200 {
							conn.Write(trickyBytes)
							trickyBytes = make([]byte, 1)
						}
						conn.Write(packet.data)
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

func (c *Client) GetGameList() (data []byte, err error) {
	raddr := net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP("127.0.0.1"),
	}

	scannerConn, err := net.DialUDP("udp4", nil, &raddr)
	if err != nil {
		log.Fatal(err)
	}

	defer scannerConn.Close()
	buf := make([]byte, 1000)
	var scanCounter byte = 0

	for {
		_, err = scannerConn.Write(
			[]byte{0xf7, 0x2f, 0x10, 0x00, 0x50, 0x58, 0x33, 0x57, 0x18, 0x00, 0x00, 0x00, scanCounter, 0x00, 0x00, 0x00},
		)
		scanCounter = scanCounter + 1
		scannerConn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
		read, _, err1 := scannerConn.ReadFromUDP(buf)
		if err1 != nil {
			return nil, err1

		} else {
			return buf[0:read], nil
		}
	}
}

func (c *Client) PrepareNewGameConnection(srcId int) (gConn net.Conn, err error) {
	gConn, err = net.Dial("tcp", os.Args[2])
	if err != nil {
		return
	}
	c.gConn[srcId] = gConn

	go func() {
		buf := make([]byte, 1000)
		for {
			read, err := gConn.Read(buf)
			if err != nil {
				if err.Error() == "EOF" {
					continue
				}
				fmt.Println("Error on game send ", err)
				gConn.Close()
				delete(c.gConn, srcId)
				break
			} else {
				_, err := c.sConn.Write(NewInformPacket(c.id, srcId, buf[0:read]).ToBytes())
				if err != nil {
					fmt.Println("Error: ", err)
				}
			}
		}
	}()
	return gConn, nil
}

func (c *Client) SendListGameAndOpenVirtualHost(listGameData []byte) error {
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
		//for {
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
						fmt.Println("Error on game send ", err)
						//err.Error()
						break

					} else {
						fmt.Println("Game to local: ", string(buf[0:read]))
						_, err := c.sConn.Write(NewInformPacket(c.id, dstClient, buf[0:read]).ToBytes())
						if err != nil {
							fmt.Println("Error: ", err)
						}
					}
				}
			}()
		}
		//}
	}()
	return err
}

func (c *Client) Close() {
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

func (c *Client) watchToServer(targetId int) {
	defer c.closeGConn(targetId)
	if conn := c.gConn[targetId]; conn != nil {
		buf := make([]byte, 1000)
		for {
			if r, err1 := conn.Read(buf); err1 != nil {
				if err1.Error() != "EOF" {
					fmt.Println("Có lỗi xảy ra:" + err1.Error())
				}
				return
			} else {
				c.SendToOther(targetId, buf[0:r])
			}
		}
	}
}

//func (c *Client) listenToGame()  {
//	for {
//		gameConn, err := net.Listen("tcp", "localhost:1993")
//	}
//}

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
		c.id = packet.dst
		fmt.Println("Connection ID: ", c.id)
	}
	c.sConn = connection

	c.Lock()

	go c.watchToGame(func() {
		fmt.Println("End game")
		c.Unlock()
	})

	go c.readCommand()

	go c.watchToServer(0)

	//go c.listenToGame()

	c.Lock()
	return nil
}
