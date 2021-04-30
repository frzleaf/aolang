package proxy

import (
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
	monitor  *Monitor
	sync.Mutex
}

func (c *Guest) SendListGameAndProxyHost(listGameData []byte) error {
	rUdpAddr := net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP(strings.Split(os.Args[2], ":")[0]),
	}
	gameListConn, err := net.DialUDP("udp4", nil, &rUdpAddr)
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
				LOG.Info("Game joined to: ", c.pConn.RemoteAddr())
				go c.gameToServer()
			}
		}
	}()
	return err
}

func (c *Guest) gameToServer() {
	buf := make([]byte, 1000)
	for {
		read, err := c.pConn.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				continue
			}
			c.monitor.onError(err)
		} else {
			_, err := c.sConn.Write(NewInformPacket(c.id, dstClient, buf[0:read]).ToBytes())
			c.monitor.onError(err)
		}
	}
}

func (c *Guest) serverToGame(data []byte) error {
	packets := PacketFromBytes(data)
	for _, packet := range packets {
		LOG.Debugf("\nReceive msg from #%v len %v\n", packet.SrcAddr(), len(packet.Data()))
		switch packet.pkgType {
		case PackageTypeFindHostResponse:
			err1 := c.SendListGameAndProxyHost(packet.data)
			if err1 != nil {
				LOG.Info(err1)
			}
		case PackageTypeInform:
			if _, err := c.pConn.Write(packet.data); err != nil {
				LOG.Info("Error on write: ", err)
			}
		}
	}
	return nil
}
