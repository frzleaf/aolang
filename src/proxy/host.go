package proxy

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type Host struct {
	serverConnector *ServerConnector // guestId player connection
	localConnectors map[int]net.Conn // connection id
	monitor         *Monitor
	gameConfig      *GameConfig
	targetId        int
}

func NewHost(sAddr string, gameConfig *GameConfig) *Host {
	return &Host{
		serverConnector: NewServerConnector(sAddr),
		localConnectors: make(map[int]net.Conn),
		gameConfig:      gameConfig,
	}
}

func (h *Host) ConnectServer() error {
	if err := h.serverConnector.connect(); err != nil {
		return err
	}
	h.serverConnector.onPacket(func(packet *Packet) {
		LOG.Debugf("Receive msg from %v len %v\n", packet.SrcAddr(), len(packet.Data()))

		switch packet.pkgType {
		case PackageTypeConverse:
			LOG.Infof("#%v: %v\n", packet.src, string(packet.Data()))
		case PackageTypeInform:
			if packet.src == ServerConnectorID {
				h.resolveCommandFromServer(packet)
				return
			}
		case PackageTypeClientStatus:
			fmt.Println(string(packet.Data()))
		case PackageTypeAppData:
			if err := h.forwardPackage(packet); err != nil {
				LOG.Error("can not forward to app", err)
			}
		case PackageTypeBroadCast:
			if err := h.ResponseToBroadCast(packet, func(responseData []byte) {
				if _, err := h.serverConnector.sendData(PackageTypeBroadCast, packet.src, responseData); err != nil {
					LOG.Error("can not send response broadcast", err)
				}
			}); err != nil {
				LOG.Error("error on response broadcast", err)
			}
		}
	})
	return h.serverConnector.waitAndForward()
}

func (h *Host) resolveCommandFromServer(packet *Packet) {
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
				h.serverConnector.SetConnectionId(connectionId)
				h.SelectTargetId(connectionId)
				h.serverConnector.sendData(PackageTypeClientStatus, ServerConnectorID, []byte(ClientModeHost))
				LOG.Infof("Connection ID assigned: %v", connectionId)
			}
		}
	case CommandExitGame:
		if err := h.Close(); err != nil {
			LOG.Error("error on closing", err)
		}
	case CommandDisconnected:
		if len(split) > 1 {
			if guestId, err := strconv.Atoi(split[1]); err == nil {
				if err := h.closeLocalConnection(guestId); err != nil {
					LOG.Error("error on close local connection", err)
				}
			}
		}
	default:
		LOG.Warn("Invalid command: ", serverCmd)
	}
}

func (h *Host) ResponseToBroadCast(packet *Packet, onResponse func([]byte)) error {
	addr := &net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP(h.serverConnector.LocalAddr()),
	}
	net.DialUDP("udp", nil, addr)
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	if _, err = conn.Write(packet.Data()); err != nil {
		return err
	}
	go func() {
		defer conn.Close()
		buffer := make([]byte, 1000)
		if read, err := conn.Read(buffer); err != nil {
			LOG.Error("can not read", err)
		} else {
			onResponse(buffer[0:read])
		}
	}()
	return nil
}

func (h *Host) Close() (err error) {
	err = h.serverConnector.close()
	for _, c := range h.localConnectors {
		if err1 := c.Close(); err1 != nil {
			err = err1
		}
	}
	return
}

func (h *Host) closeLocalConnection(connectionId int) error {
	if conn := h.localConnectors[connectionId]; conn != nil {
		delete(h.localConnectors, connectionId)
		return conn.Close()
	}
	return nil
}

func (h *Host) forwardPackage(p *Packet) error {
	conn := h.localConnectors[p.src]
	if conn == nil {
		if dial, err := net.Dial("tcp", h.gameConfig.LocalIp+":"+strconv.Itoa(h.gameConfig.TcpPort)); err != nil {
			return err
		} else {
			conn = dial
		}
		go func() {
			defer func() {
				if h.localConnectors[p.src] == conn {
					if err := h.closeLocalConnection(p.src); err != nil {
						LOG.Error("error on closing local connection ", "#"+strconv.Itoa(p.src), err)
					}
				}
			}()
			buffer := make([]byte, 1000)
			for {
				if read, err := conn.Read(buffer); err != nil {
					if err != io.EOF {
						LOG.Error("error on read response", err)
					}
					return
				} else {
					h.serverConnector.sendData(PackageTypeAppData, p.src, buffer[0:read])
				}
			}
		}()
		// Create connection to local host
		h.localConnectors[p.src] = conn
	}
	_, err := conn.Write(p.Data())
	return err
}

func (h *Host) ConnectionId() int {
	return h.serverConnector.connectionId
}

func (h *Host) SelectTargetId(targetId int) {
	h.targetId = targetId
}

func (h *Host) TargetId() int {
	return h.targetId
}

func (h *Host) ServerConnector() *ServerConnector {
	return h.serverConnector
}

func (h *Host) GameConfig() *GameConfig {
	return h.gameConfig
}

func (g *Host) OnMatch() bool {
	return true
}
