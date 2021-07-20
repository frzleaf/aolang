package proxy

import (
	"errors"
	"strconv"
	"strings"
)

type Guest struct {
	serveConnector *ServerConnector // guestId player connection
	localConnector *LocalConnector  // connection id
	hostId         int              // connection id
	monitor        *Monitor
	gameConfig     *GameConfig
}

func NewGuest(sAddr string, config *GameConfig) (g *Guest) {
	g = &Guest{
		serveConnector: NewServerConnector(sAddr),
		gameConfig:     config,
	}
	g.init()
	return
}

func (g *Guest) init() {
	g.localConnector = NewLocalConnector(g.gameConfig.LocalIp + ":" + strconv.Itoa(g.gameConfig.TcpPort))
}

func (g *Guest) selectHost(hostId int) {
	g.hostId = hostId
}

func (g *Guest) openProxy() (err error) {
	g.localConnector.onData(func(data []byte) {
		_, err = g.serveConnector.sendData(PackageTypeToHost, g.hostId, data)
		if err != nil {
			LOG.Error("can not send package ", err)
		}
	})
	if err = g.localConnector.startListen(); err != nil {
		return errors.New("can not open local proxy")
	}
	err = g.localConnector.startListenUdp(":"+strconv.Itoa(g.gameConfig.UdpPort), func(data []byte) {
		_, err := g.serveConnector.sendData(PackageTypeBroadCast, g.hostId, data)
		if err != nil {
			LOG.Error("can not send package ", err)
		}
	})
	return
}

func (g *Guest) ConnectServer() error {
	if err := g.serveConnector.connect(); err != nil {
		return err
	}
	g.serveConnector.onPacket(func(packet *Packet) {
		LOG.Debugf("Receive msg from %v len %v\n", packet.SrcAddr(), len(packet.Data()))

		if packet.src == ServerConnectorID {
			g.resolveCommandFromServer(packet)
			return
		}

		switch packet.pkgType {
		case PackageTypeInform:
			g.resolveCommandFromServer(packet)
		case PackageTypeToGuest:
			if _, err := g.localConnector.sentBytes(packet.Data()); err != nil {
				LOG.Error("error on write data to local")
			}
		case PackageTypeBroadCastResponse:
			g.BroadCastResponse(packet)
		}
	})
	go g.serveConnector.waitAndForward()
	return g.openProxy()
}

func (g *Guest) resolveCommandFromServer(packet *Packet) {
	serverCmd := string(packet.Data())
	split := strings.Split(serverCmd, " ")
	if len(split) < 1 {
		LOG.Infof("Server - %v", string(packet.Data()))
		return
	}
	switch split[0] {
	case CommandAssignID:
		if len(split) < 2 {
			LOG.Warnf("Invalid argument - %v", string(packet.Data()))
		} else {
			connectionId, err := strconv.Atoi(split[1])
			if err != nil {
				LOG.Error("Invalid connection ID: ", split[1])
			} else {
				g.serveConnector.SetConnectionId(connectionId)
				LOG.Infof("Connection ID assigned: %v", connectionId)
			}
		}
	case CommandExitGame:
		if err := g.Close(); err != nil {
			LOG.Error("error on closing", err)
		}
	default:
		LOG.Warn("Invalid command: ", serverCmd)
	}
}

func (g *Guest) BroadCastResponse(packet *Packet) {
	if _, err := g.localConnector.sentBytesTo(packet.Data(), ":"+strconv.Itoa(g.gameConfig.TcpPort)); err != nil {
		LOG.Error("can not receive broadcast response", err)
	}
}

func (g *Guest) Close() (err error) {
	err = g.serveConnector.close()
	err = g.localConnector.close()
	return
}

func (g *Guest) ConnectionId() int {
	return g.serveConnector.connectionId
}

func (g *Guest) SelectHost(hostId int) {
	g.hostId = hostId
}
