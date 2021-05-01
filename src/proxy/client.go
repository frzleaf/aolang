package proxy

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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
			if err1.Error() != "EOF" {
				LOG.Error("Có lỗi xảy ra:", err1)
			}
			return
		} else {
			if c.host.IsOn() {
				c.host.serverToGame(buf[0:r])
			}
			if c.guest.IsOn() {
				c.guest.serverToGame(buf[0:r])
			}
		}
	}
}

func (c *Client) readCommand() {
	for {
		reader := bufio.NewReader(os.Stdin)
		LOG.Info("Gửi lệnh lên server: ")
		line, _, err := reader.ReadLine()
		if err != nil {
			LOG.Info("Command error: ", err)
			continue
		}
		lineStr := string(line)
		if strings.HasPrefix(lineStr, "to ") {
			split := strings.Split(lineStr, " ")
			if len(split) == 1 {
				LOG.Info("Missing 2nd argument")
				continue
			} else {
				input2nd, err := strconv.Atoi(split[1])
				if err != nil {
					LOG.Info("Invalid args: ", err)
					continue
				}
				dstClient = input2nd
				LOG.Info("Selected: ", input2nd)
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
	LOG.Info("Connected to: " + sAddr)
	buf := make([]byte, 1000)
	if read, err := connection.Read(buf); err != nil {
		return err
	} else {
		packet := PacketFromBytes(buf[0:read])
		c.id = packet[0].DstAddr()
		LOG.Info("Connection ID: ", c.id)
	}
	c.sConn = connection

	c.Lock()

	go c.watchToGame(func() {
		LOG.Info("End game")
		c.Unlock()
	})

	go c.readCommand()

	c.Lock()
	return nil
}
