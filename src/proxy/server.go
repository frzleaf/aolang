package proxy

import (
	"bytes"
	"log"
	"net"
	"strconv"
)

type Server struct {
	clients        map[int]net.Conn
	clientStatuses map[int]string
	clientCounter  int
	stopped        bool
}

func NewServer() *Server {
	return &Server{
		clientCounter:  1,
		clients:        make(map[int]net.Conn),
		clientStatuses: make(map[int]string),
		stopped:        false,
	}
}

func (s *Server) Start(addr string) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	LOG.Info("Server is running at:", addr)
	for !s.stopped {
		newConn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		clientId := s.onNewClient(newConn)
		s.sendToConnector(
			PackageTypeInform,
			ServerConnectorID,
			clientId,
			[]byte(CommandToString(CommandAssignID, clientId)),
		)
	}
}

func (s *Server) onNewClient(client net.Conn) int {
	s.clients[s.clientCounter] = client
	go s.watchClient(s.clientCounter)
	clientId := s.clientCounter
	s.clientCounter++
	return clientId
}

func (s *Server) sendToConnector(packetType, srcId, dstId int, data []byte) error {
	conn := s.clients[dstId]
	if conn == nil {
		return NotFoundConnectorError
	}
	_, err := conn.Write(
		NewPacket(packetType, srcId, dstId, data).ToBytes(),
	)
	return err
}

func (s *Server) broadCast(packet *Packet) {
	for id, _ := range s.clients {
		if id != packet.src {
			s.sendToConnector(
				PackageTypeBroadCast,
				packet.src,
				id,
				packet.Data(),
			)
		}
	}
}

func (s *Server) exit(id int) {
	conn := s.clients[id]
	if conn != nil {
		conn.Close()
	}
	delete(s.clients, id)
	delete(s.clientStatuses, id)
	for i := range s.clients {
		s.informClientDisconnected(i, id)
	}
}

func (s *Server) Close() {
	s.stopped = true
	for _, conn := range s.clients {
		conn.Close()
	}
}

func (s *Server) formClientStatuesAsBytes() []byte {
	lineBytes := []byte(" - ")
	newLineBytes := []byte("\n")
	moreInforBytes := []byte("...")

	buffer := bytes.NewBuffer(nil)
	for id, status := range s.clientStatuses {
		buffer.Write([]byte(strconv.Itoa(id)))

		buffer.Write(lineBytes)

		buffer.Write([]byte(status))

		buffer.Write(newLineBytes)

		if buffer.Len() >= 900 {
			buffer.Write(moreInforBytes)
			break
		}
	}
	return buffer.Bytes()
}

func (s *Server) informClientDisconnected(srcId int, disconnectedId int) {
	s.sendToConnector(
		PackageTypeInform,
		ServerConnectorID,
		srcId,
		[]byte(CommandToString(CommandDisconnected, disconnectedId)),
	)
}

func (s *Server) watchClient(id int) {
	buf := make([]byte, 1000)
	conn := s.clients[id]
	defer s.exit(id)
	for {
		read, err := conn.Read(buf)
		if err != nil {
			LOG.Info("Close connection:", id)
			return
		}
		for _, packet := range PacketFromBytes(buf[0:read]) {
			if packet.src == packet.dst {
				continue
			}
			switch packet.pkgType {
			case PackageTypeClientStatus:
				if packet.DstAddr() == ServerConnectorID {
					if len(packet.Data()) > 0 {
						s2 := string(packet.Data())
						// Max status len
						if len(s2) > 128 {
							s2 = s2[0:128]
						}
						s.clientStatuses[packet.src] = s2
					} else {
						s.sendToConnector(PackageTypeClientStatus, ServerConnectorID, packet.src, s.formClientStatuesAsBytes())
					}
				}
			default:
				err = s.sendToConnector(packet.pkgType, packet.src, packet.dst, packet.data)
				if err != nil {
					if err == NotFoundConnectorError {
						s.informClientDisconnected(packet.src, packet.dst)
						break
					}
					LOG.Error("error while sendInform", err)
				}
			}
		}
	}
}
