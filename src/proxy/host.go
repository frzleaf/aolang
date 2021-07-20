package proxy

import (
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

		if packet.src == ServerConnectorID {
			h.resolveCommandFromServer(packet)
			return
		}

		switch packet.pkgType {
		case PackageTypeInform:
			h.resolveCommandFromServer(packet)
		case PackageTypeToHost:
			if err := h.forwardPackage(packet); err != nil {
				LOG.Error("can not forward to app", err)
			}
		case PackageTypeBroadCast:
			if err := h.ResponseToBroadCast(packet, func(responseData []byte) {
				if _, err := h.serverConnector.sendData(PackageTypeBroadCastResponse, packet.src, responseData); err != nil {
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
				LOG.Infof("Connection ID assigned: %v", connectionId)
			}
		}
	case CommandExitGame:
		if err := h.Close(); err != nil {
			LOG.Error("error on closing", err)
		}
	default:
		LOG.Warn("Invalid command: ", serverCmd)
	}
}

func (h *Host) ResponseToBroadCast(packet *Packet, onResponse func([]byte)) error {
	conn, err := net.Dial("udp", "localhost:6112")
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
			LOG.Error("can not read")
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
					delete(h.localConnectors, p.src)
				}
				conn.Close()
			}()
			buffer := make([]byte, 1000)
			for {
				if read, err := conn.Read(buffer); err != nil {
					if err == io.EOF {
						continue
					} else {
						LOG.Error("error on read response", err)
					}
				} else {
					h.serverConnector.sendData(PackageTypeToGuest, p.src, buffer[0:read])
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
