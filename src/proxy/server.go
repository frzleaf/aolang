package proxy

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

type Server struct {
	clients       map[int]*Connector
	clientCounter int
	addr          string
}

func (s *Server) start() {
	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatal(err)
	}
	for {
		newConn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		client := s.onNewClient(newConn)
		s.sendToConnector(ServerConnectorID, client.id, []byte("Your ID: "+strconv.Itoa(client.id)))
		go func() {
			client.listen()
		}()
	}
}

func (s *Server) onNewClient(client net.Conn) *Connector {
	newConnector := Connector{
		s, client, s.clientCounter,
	}
	s.clients[s.clientCounter] = &newConnector
	s.clientCounter++
	return &newConnector
}

func (s *Server) sendToConnector(srcId, targetId int, data []byte) error {
	connector := s.clients[targetId]
	if connector == nil {
		return NotFoundConnectorError
	}
	outMessage := WrapMessage(srcId, data)
	_, err := connector.conn.Write(outMessage)
	return err
}

type Connector struct {
	server *Server
	conn   net.Conn
	id     int
}

func (c *Connector) listen() {
	defer c.conn.Close()
	buf := make([]byte, 512)
	for {
		read, err := c.conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}
		targetID, data, err := ExtractMessageToConnectorIDAndData(buf[0:read])
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = c.server.sendToConnector(c.id, targetID, data)
		if err != nil {
			fmt.Print(err)
		}
	}
}
