package proxy

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strconv"
)

type Server struct {
	clients       map[int]*Connector
	clientCounter int
}

func NewServer() *Server {
	return &Server{
		clientCounter: 0,
		clients:       make(map[int]*Connector),
	}
}

func (s *Server) Start(addr string) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server is running at: " + addr)
	for {
		newConn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		client := s.onNewClient(newConn)
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(client.id))
		s.sendToConnector(ServerConnectorID, client.id, buf)
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
	outMessage := NewInformPacket(srcId, targetId, data).ToBytes()
	_, err := connector.conn.Write(outMessage)
	return err
}

type Connector struct {
	server *Server
	conn   net.Conn
	id     int
}

func (c *Connector) exit() {
	defer c.conn.Close()
	delete(c.server.clients, c.id)
}

func (c *Connector) sendToConnector(pkgType, dst int, data []byte) error {
	dstConn := c.server.clients[dst]
	if dstConn == nil {
		return InvalidTarget
	}
	_, err := dstConn.conn.Write(NewPacket(pkgType, c.id, dst, data).ToBytes())
	return err
}

func (c *Connector) sendInformToConnector(dst int, data []byte) error {
	return c.sendToConnector(PackageTypeInform, dst, data)
}

func (c *Connector) listen() {
	buf := make([]byte, 1000)
	defer c.exit()
	for {
		read, err := c.conn.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				continue
			}
			fmt.Println("Close connection: " + strconv.Itoa(c.id))
			return
		}
		for _, packet := range PacketFromBytes(buf[0:read]) {
			fmt.Printf("\nMsg from %v: %v", packet.src, string(packet.data))
			err = c.sendToConnector(packet.pkgType, packet.dst, packet.data)
			if err != nil {
				fmt.Println("error while sendInform", err)
			}
		}
	}
}
