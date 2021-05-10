package proxy

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Host struct {
	hostConn          map[int]net.Conn // game connections
	connLock          sync.Mutex       // support concurrent map
	hostPort          int              // local host port
	guestCon          net.Conn         // guest player connection
	sConn             net.Conn         // server connection
	virtualHost       net.Listener     // local virtual host
	id                int              // connection id
	monitor           *Monitor
	gameConfig        *GameConfig
	connectedHost     int
	connectedHostPort int
	isHost            bool
}

func NewHost() *Host {
	return &Host{
		hostConn:          make(map[int]net.Conn),
		gameConfig:        Warcraft3Config,
		connectedHost:     -1,
		connectedHostPort: -1,
		hostPort:          -1,
		monitor:           &Monitor{},
		connLock:          sync.Mutex{},
	}
}

func (h *Host) BroadCast(packet *Packet, responsePacketType int) error {
	returnData, err := h.OnBroadCast(packet.Data(), true)
	if err != nil || returnData == nil {
		return err
	}

	if len(returnData) > 100 {
		h.isHost = true
	}

	if h.isHost && h.hostPort < 0 {
		h.hostPort = FindPortOpen(h.gameConfig.localIp, h.gameConfig.tcpPorts)
	}
	portBinary := make([]byte, 4)
	binary.BigEndian.PutUint32(portBinary, uint32(h.hostPort))
	sendData := make([]byte, 4+len(returnData))
	copy(sendData, portBinary)
	copy(sendData[4:], returnData)
	_, err = h.sConn.Write(
		NewPacket(responsePacketType, h.id, packet.SrcAddr(), sendData).ToBytes(),
	)
	return err
}

func (h *Host) BroadCastResponse(packet *Packet) error {
	gameData := packet.Data()
	if len(gameData) <= 4 {
		return errors.New("invalid broadcast response")
	}

	h.connectedHost = packet.src
	dedicatePort := int(binary.BigEndian.Uint32(gameData[0:4]))
	if 0 < dedicatePort && dedicatePort < (2<<16) {
		h.connectedHostPort = dedicatePort
	}

	_, err := h.OnBroadCast(gameData[4:], false)
	if h.virtualHost == nil && h.connectedHost != h.id {
		go h.OpenProxyHost()
	}
	return err
}

func (h *Host) OpenProxyHost() (err error) {
	if h.isHost {
		return
	}
	h.virtualHost, err = net.Listen("tcp", h.getVirtualHostGameBind())
	if err != nil {
		return
	}
	LOG.Info("Virtual host open at:", h.virtualHost.Addr().String())
	defer func() {
		if h.virtualHost != nil {
			h.virtualHost.Close()
		}
		h.virtualHost = nil
	}()
	for {
		if gConn, err := h.virtualHost.Accept(); err == nil {
			h.guestCon = gConn
			go func() {
				defer func() {
					if gConn != nil {
						gConn.Close()
					}
					gConn = nil
				}()
				if h.connectedHost < 0 {
					return
				}
				targetHost := h.connectedHost
				_, _ = h.sConn.Write(NewInformPacket(h.id, targetHost, []byte(CommandExitGame)).ToBytes())
				time.Sleep(time.Millisecond * 100)
				LOG.Infof("Game joined to: %v (%v)", targetHost, gConn.LocalAddr())
				buf := make([]byte, 1000)
				for {
					if read, err2 := gConn.Read(buf); err2 != nil {
						if err2.Error() == "EOF" {
							continue
						}
						LOG.Error(err2)
						break
					} else {
						// TODO need parse the packet to specify the target host
						_, err2 := h.sConn.Write(NewPacket(PackageTypeToHost, h.id, targetHost, buf[0:read]).ToBytes())
						if err2 != nil {
							LOG.Warn("Error send host:", err2)
						}
					}
				}
			}()
		}
	}
	LOG.Info("Exit proxy")
	return
}

func (h *Host) OpenBroadCastListener() error {
	lAddr := &net.UDPAddr{
		Port: h.gameConfig.udpPort,
		IP:   net.ParseIP(h.gameConfig.internalRemoteIp),
	}
	bcListener, err := net.ListenUDP("udp", lAddr)
	if err != nil {
		return err
	}
	LOG.Infof("Open broadcast listener at %v", bcListener.LocalAddr().String())
	defer bcListener.Close()
	buf := make([]byte, 1000)
	for {
		if read, err := bcListener.Read(buf); err != nil {
			LOG.Warn("Can not read broadcast message", err)
		} else {
			if h.isHost {
				continue
			}
			LOG.Debugf("Receive broadcast message, len = %v", len(buf[0:read]))
			h.sConn.Write(NewPacket(
				PackageTypeBroadCast,
				h.id,
				ServerConnectorID,
				buf[0:read],
			).ToBytes())
		}
	}
}

func (h *Host) OnBroadCast(receive []byte, getResponse bool) ([]byte, error) {
	rAddr := net.UDPAddr{
		Port: h.gameConfig.udpPort,
		IP:   net.ParseIP(h.gameConfig.localIp),
	}

	scannerConn, err := net.DialUDP("udp4", nil, &rAddr)
	if err != nil {
		return nil, err
	}

	defer scannerConn.Close()
	buf := make([]byte, 1024)

	_, err = scannerConn.Write(receive)

	if getResponse {
		scannerConn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
		read, err := scannerConn.Read(buf)
		if err == nil {
			return buf[0:read], nil
		}
	}
	return nil, err
}

func (h *Host) DataToHost(packet *Packet) (err error) {
	hConn := h.hostConn[packet.SrcAddr()]
	if hConn == nil {
		hConn, err = h.PrepareNewGameConnection(packet.SrcAddr())
		if err != nil {
			return err
		}
	}
	_, err = hConn.Write(packet.data)
	return
}

func (h *Host) DataToGuest(packet *Packet) (err error) {
	if h.connectedHost == packet.src && h.guestCon != nil {
		_, err = h.guestCon.Write(packet.data)
	}
	return
}

func (h *Host) PrepareNewGameConnection(srcId int) (gConn net.Conn, err error) {
	gConn, err = net.Dial("tcp", h.getGameBind())
	if err != nil {
		return
	}
	h.addConnection(srcId, gConn)
	go h.DataHostToServer(srcId)
	return
}

func (h *Host) SendDataToServer(packet *Packet) (err error) {
	_, err = h.sConn.Write(packet.ToBytes())
	return
}

func (h *Host) DataHostToServer(srcId int) {
	hConn := h.hostConn[srcId]
	buf := make([]byte, 1000)
	for {
		read, err := hConn.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				continue
			}
			h.monitor.onError(err)
			h.removeConnection(srcId)
			break
		} else {
			if _, err = h.sConn.Write(NewPacket(PackageTypeToGuest, h.id, srcId, buf[:read]).ToBytes()); err != nil {
				h.monitor.onError(err)
			}
		}
	}
}

func (h *Host) addConnection(srcId int, newConn net.Conn) {
	h.connLock.Lock()
	h.hostConn[srcId] = newConn
	h.connLock.Unlock()
}

func (h *Host) removeConnection(srcId int) {
	h.connLock.Lock()
	if h.hostConn[srcId] != nil {
		h.hostConn[srcId].Close()
		delete(h.hostConn, srcId)
	}
	h.connLock.Unlock()
}

func (h *Host) start(serverAddr string) (err error) {
	h.sConn, err = net.Dial("tcp4", serverAddr)
	if err != nil {
		return
	}
	// Set up broadcast watchClient address
	h.gameConfig.internalRemoteIp = strings.Split(h.sConn.LocalAddr().String(), ":")[0]
	go func() {
		for {
			go h.getLocalhostGameAndSendToServer()
			if !h.isHost {
				go h.getServerGame()
			}
			time.Sleep(time.Second)
		}
	}()
	return h.watchGameData()
}

func (h *Host) watchGameData() (err error) {
	buf := make([]byte, 1000)
	for {
		if read, err := h.sConn.Read(buf); err != nil {
			return err
		} else {
			err := h.serverToGame(buf[0:read])
			if err != nil {
				LOG.Error(err)
			}
		}
	}
}

func (h *Host) serverToGame(data []byte) (err error) {
	packets := PacketFromBytes(data)
	for _, packet := range packets {
		if err = h.gameReceiveMessage(packet); err != nil {
			return err
		}
	}
	return
}

// TODO consider using event emitter
func (h *Host) gameReceiveMessage(packet *Packet) (err error) {
	LOG.Debugf("Receive msg from %v len %v\n", packet.SrcAddr(), len(packet.Data()))

	if packet.src == ServerConnectorID {
		h.resolveCommandFromServer(packet)
		return nil
	}

	switch packet.pkgType {
	case PackageTypeInform:
		h.resolveCommandFromServer(packet)
		return nil
	//case PackageTypeBroadCast:
	//err = h.BroadCast(packet)
	case PackageTypeToHost:
		err = h.DataToHost(packet)
	case PackageTypeToGuest:
		h.isHost = false
		err = h.DataToGuest(packet)
	case PackageTypeGameListGet:
		h.isHost = false
		h.sendServerGameToLocal(packet)
		//case PackageTypeBroadCastResponse:
		//	err = h.BroadCastResponse(packet)
	}
	return err
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
				h.id = connectionId
				LOG.Infof("Connection ID assigned: %v", h.id)
			}
		}
	case CommandExitGame:
		h.removeConnection(packet.SrcAddr())
	default:
		LOG.Warn("Invalid command: ", serverCmd)
	}
}

func (h *Host) getGameBind() string {
	return h.gameConfig.localIp + ":" + strconv.Itoa(h.hostPort)
}

func (h *Host) getVirtualHostGameBind() string {
	return h.gameConfig.localIp + ":" + strconv.Itoa(h.connectedHostPort)
}

func (h *Host) getLocalhostGameAndSendToServer() {
	i := byte(time.Now().Second() % 10)
	bytes := []byte{0xf7, 0x2f, 0x10, 0x00, 0x50, 0x58, 0x33, 0x57, 0x18, 0x00, 0x00, 0x00, i, 0x00, 0x00, 0x00}
	if gameListData, err := h.OnBroadCast(bytes, true); err == nil && len(gameListData) > 0 {
		h.BroadCast(NewPacket(PackageTypeGameListPush, ServerConnectorID, h.id, bytes), PackageTypeGameListPush)
	}
}

func (h *Host) getServerGame() {
	h.sConn.Write(
		NewPacket(PackageTypeGameListGet, h.id, ServerConnectorID, nil).ToBytes(),
	)
}

func (h *Host) sendServerGameToLocal(packet *Packet) {
	h.BroadCastResponse(packet)
}

func (h *Host) Close() {
	for _, conn := range h.hostConn {
		conn.Close()
	}
	h.hostConn = make(map[int]net.Conn)
	h.sConn.Close()
}
