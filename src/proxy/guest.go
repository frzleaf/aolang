package proxy

import (
	"strconv"
	"strings"
)

type Guest struct {
	serverConnector *ServerConnector // guestId player connection
	localConnector  *LocalConnector  // connection id
	hostId          int              // connection id
	monitor         *Monitor
	gameConfig      *GameConfig
}

func NewGuest(sAddr string, config *GameConfig) (g *Guest) {
	g = &Guest{
		serverConnector: NewServerConnector(sAddr),
		gameConfig:      config,
		hostId:          -1,
	}
	g.init()
	return
}

func (g *Guest) init() {
	g.localConnector = NewLocalConnector(g.gameConfig.LocalIp + ":" + strconv.Itoa(g.gameConfig.TcpPort))
}

func (g *Guest) openProxy() (err error) {
	g.localConnector.onData(func(data []byte) {
		_, err = g.serverConnector.sendData(PackageTypeAppData, g.hostId, data)
		if err != nil {
			LOG.Error("can not send package ", err)
		}
	})
	go func() {
		if err = g.localConnector.startListen(); err != nil {
			LOG.Error("can not open local proxy", err)
		}
	}()
	err = g.localConnector.startListenUdp(g.serverConnector.LocalAddr(), g.gameConfig.UdpPort, func(data []byte) {
		_, err := g.serverConnector.sendData(PackageTypeBroadCast, g.hostId, data)
		if err != nil {
			LOG.Error("can not send package ", err)
		}
	})
	return
}

func (g *Guest) ConnectServer() error {
	if err := g.serverConnector.connect(); err != nil {
		return err
	}
	g.serverConnector.onPacket(func(packet *Packet) {
		LOG.Debugf("Receive msg from %v len %v\n", packet.SrcAddr(), len(packet.Data()))

		if packet.src == ServerConnectorID {
			g.resolveCommandFromServer(packet)
			return
		}

		switch packet.pkgType {
		case PackageTypeConverse:
			LOG.Infof("#%v: %v\n", packet.src, string(packet.Data()))
		case PackageTypeInform:
			g.resolveCommandFromServer(packet)
		case PackageTypeAppData:
			if _, err := g.localConnector.sentBytes(packet.Data()); err != nil {
				LOG.Error("error on write data to local")
			}
		case PackageTypeBroadCast:
			g.BroadCastResponse(packet)
		}
	})
	go func() {
		if err := g.openProxy(); err != nil {
			LOG.Error("error on openProxy", err)
		}
	}()
	return g.serverConnector.waitAndForward()
}

func (g *Guest) resolveCommandFromServer(packet *Packet) {
	serverCmd := string(packet.Data())
	split := strings.Split(serverCmd, CharSplitCommand)
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
				g.serverConnector.SetConnectionId(connectionId)
				g.SelectTargetId(connectionId)
				LOG.Infof("Connection ID assigned: %v", connectionId)
			}
		}
	case CommandExitGame:
		if err := g.Close(); err != nil {
			LOG.Error("error on closing", err)
		}
	case CommandDisconnected:
		if len(split) > 1 {
			if strconv.Itoa(g.hostId) == split[1] {
				g.localConnector.Release()
				LOG.Infof("Disconnect to host #%v", g.hostId)
			}
		}
	default:
		LOG.Warn("Invalid command: ", serverCmd)
	}
}

func (g *Guest) BroadCastResponse(packet *Packet) {
	if _, err := g.localConnector.sentUdpBytesTo(packet.Data(), g.gameConfig.LocalIp, g.gameConfig.TcpPort); err != nil {
		LOG.Error("can not receive broadcast response", err)
	}
}

func (g *Guest) Close() (err error) {
	err = g.serverConnector.close()
	err = g.localConnector.close()
	return
}

func (g *Guest) ConnectionId() int {
	return g.serverConnector.connectionId
}

func (g *Guest) SelectHost(hostId int) {
	g.hostId = hostId
}

func (g *Guest) ServerConnector() *ServerConnector {
	return g.serverConnector
}

func (g *Guest) SelectTargetId(targetId int) {
	g.hostId = targetId
}

func (g *Guest) TargetId() int {
	return g.hostId
}

func (g *Guest) GameConfig() *GameConfig {
	return g.gameConfig
}

func (g *Guest) OnMatch() bool {
	return g.localConnector.isOnline()
}
