package proxy

import (
	"fmt"
	"net"
)

type Client struct {
	srcId      int
	connection net.Conn
}

func (c *Client) sendMsg(data []byte) error {
	_, err := c.connection.Write(data)
	return err
}

func (c *Client) sendMsg(data []byte) error {
	_, err := c.connection.Write(data)
	return err
}

type Host struct {
	conn     net.Conn
	clients  map[int]*Client
	hostAddr string
}

func (h *Host) run(sAddr string) error {
	connection, err := net.Dial("tcp", sAddr)
	if err != nil {
		return err
	}

	buf := make([]byte, 512)
	if read, err := connection.Read(buf); err != nil {
		return err
	} else {
		fmt.Println(string(buf[0:read]))
	}
	h.conn = connection
}

func (h *Host) host() error {
	buf := make([]byte, 512)
	for {
		read, err := h.conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}
		srcId, data, err := ExtractMessageToConnectorIDAndData(buf[0:read])
		if err != nil {
			continue
		}
		dial, err := net.Dial("tcp", h.hostAddr)
		if err != nil {
			fmt.Println(err)
			continue
		}
		client := Client{
			srcId, dial,
		}
		h.clients[srcId] = &client
	}
}
