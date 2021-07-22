package proxy

import (
	"errors"
	"net"
	"strings"
)

type ServerConnector struct {
	sAddr         string
	sConn         net.Conn
	connectionId  int
	onPackageFunc func(p *Packet)
}

func NewServerConnector(sAddr string) *ServerConnector {
	return &ServerConnector{
		sAddr: sAddr,
	}
}

func (s *ServerConnector) connect() (err error) {
	s.sConn, err = net.Dial("tcp", s.sAddr)
	return
}

func (s *ServerConnector) waitAndForward() error {
	if s.onPackageFunc == nil {
		return errors.New("onPackageFunc not setup")
	}

	buff := make([]byte, 1000)
	for s.sConn != nil {
		if read, err := s.sConn.Read(buff); err != nil {
			if s.sConn != nil {
				LOG.Error("error on reading server data", err)
				s.close()
			}
		} else {
			for _, pkg := range PacketFromBytes(buff[0:read]) {
				s.onPackageFunc(pkg)
			}
		}
	}
	return nil
}

func (s *ServerConnector) sendData(pkgType int, targetId int, data []byte) (int, error) {
	if s.sConn != nil {
		return s.sConn.Write(NewPacket(pkgType, s.connectionId, targetId, data).ToBytes())
	}
	return 0, errors.New("server not connected")
}

func (s *ServerConnector) onPacket(onDataFunc func(p *Packet)) {
	s.onPackageFunc = onDataFunc
}

func (s *ServerConnector) close() (err error) {
	if s.sConn == nil {
		return nil
	}
	conn := s.sConn
	s.sConn = nil
	return conn.Close()
}

func (s *ServerConnector) SetConnectionId(connectionId int) {
	s.connectionId = connectionId
}

func (s *ServerConnector) LocalAddr() string {
	return strings.Split(s.sConn.LocalAddr().String(), ":")[0]
}
