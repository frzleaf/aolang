package proxy

import (
	"fmt"
	"log"
	"net"
)

type Client struct {
	sConn net.Conn
	gConn map[int] net.Conn
}

func (c *Client) sendMsgToGame(srcId int, data []byte) error {
	if conn := c.gConn[srcId]; conn != nil {
		_, err := conn.Write(data)
		return err
	}
	return nil
}

func (c *Client) sendMsgToClient(targetId int, data []byte) error {
	_, err := c.sConn.Write(WrapMessage(targetId, data))
	return err
}

func (c *Client) watchAndForward(onExit func()) {
	buf1 := make([]byte, 1000)

	go func() {
		defer c.sConn.Close()
		defer onExit()
		for {
			if r, err1 := c.sConn.Read(buf1); err1 != nil {
				if err1.Error() != "EOF" {
					fmt.Println("Có lỗi xảy ra:" + err1.Error())
				}
				return
			} else {
				targetId, data, err1 := ExtractMessageToConnectorIDAndData(buf1[0:r])
				if err1 != nil {
					fmt.Println(InvalidMessageError)
				}
				if conn := c.gConn[targetId]; conn != nil {
					if _, err := conn.Write(data); err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}()
}

func (c *Client) ()  {
	
}

func (c *Client) run(sAddr string) error {
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
	c.sConn = connection
}
