package proxy

import (
	"net"
	"strconv"
	"strings"
	"time"
)

type Host struct {
	gConn      map[int]net.Conn // game connections
	sConn      net.Conn         // server connection
	id         int              // connection id
	monitor    *Monitor
	gameConfig *GameConfig
	online     bool
}

func NewHost() *Host {
	return &Host{
		gConn:      make(map[int]net.Conn),
		gameConfig: Warcraft3Config,
	}
}

func (h *Host) BroadCast(packet *Packet) error {
	returnData, err := h.OnBroadCast(packet.Data(), true)
	if err != nil {
		return err
	}
	_, err = h.sConn.Write(
		NewPacket(PackageTypeBroadCastResponse, h.id, packet.SrcAddr(), returnData).ToBytes(),
	)
	return err
}

func (h *Host) BroadCastAndResponse(packet *Packet) error {
	_, err := h.OnBroadCast(packet.Data(), false)
	return err
}

func (h *Host) OpenBroadCastServer() error {
	lAddr := &net.UDPAddr{
		Port: h.gameConfig.udpPort,
		IP:   net.ParseIP(h.gameConfig.internalRemoteIp),
	}
	bcListener, err := net.ListenUDP("udp", lAddr)
	if err != nil {
		return err
	}
	defer bcListener.Close()
	buf := make([]byte, 1000)
	for {
		if read, err := bcListener.Read(buf); err != nil {
			LOG.Warn("Can not read broadcast message", err)
		} else {
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

func (h *Host) OnBroadCastResponse(packet *Packet) error {
	rAddr := net.UDPAddr{
		Port: h.gameConfig.udpPort,
		IP:   net.ParseIP(h.gameConfig.localIp),
	}

	scannerConn, err := net.DialUDP("udp4", nil, &rAddr)
	if err != nil {
		return err
	}

	defer scannerConn.Close()
	_, err = scannerConn.Write(packet.Data())
	return err
}

func (h *Host) OnGameData(packet *Packet) (err error) {
	gConn := h.gConn[packet.SrcAddr()]
	if gConn == nil {
		gConn, err = h.PrepareNewGameConnection(packet.SrcAddr())
		if err != nil {
			return err
		}
	}
	_, err = gConn.Write(packet.data)
	return
}

func (h *Host) PrepareNewGameConnection(srcId int) (gConn net.Conn, err error) {
	gConn, err = net.Dial("tcp", h.getGameBind())
	if err != nil {
		return
	}
	h.gConn[srcId] = gConn
	go h.gameToServer(srcId)
	return
}

func (h *Host) SendDataToServer(packet *Packet) (err error) {
	_, err = h.sConn.Write(packet.ToBytes())
	return
}

func (h *Host) gameToServer(srcId int) {
	gConn := h.gConn[srcId]
	buf := make([]byte, 1000)
	for {
		read, err := gConn.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				continue
			}
			h.monitor.onError(err)
			gConn.Close()
			delete(h.gConn, srcId)
			break
		} else {
			if _, err = h.sConn.Write(NewInformPacket(h.id, srcId, buf[:read]).ToBytes()); err != nil {
				h.monitor.onError(err)
			}
		}
	}
}

func (h *Host) start(serverAddr string) (err error) {
	dial, err := net.Dial("tcp4", serverAddr)
	if err != nil {
		return err
	}
	// Set up broadcast watchClient address
	h.gameConfig.internalRemoteIp = strings.Split(dial.LocalAddr().String(), ":")[0]
	h.sConn = dial
	go h.OpenBroadCastServer()
	return h.watchGameData()
}

func (h *Host) watchGameData() (err error) {
	buf := make([]byte, 1000)
	for {
		if read, err := h.sConn.Read(buf); err != nil {
			return err
		} else {
			h.serverToGame(buf[0:read])
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
	LOG.Infof("Receive msg from %v len %v\n", packet.SrcAddr(), len(packet.Data()))

	if packet.src == ServerConnectorID {
		h.resolveCommandFromServer(packet)
	}

	switch packet.pkgType {
	case PackageTypeInform:
		h.OnGameData(packet)
	case PackageTypeBroadCast:
		h.BroadCast(packet)
	case PackageTypeBroadCastResponse:
		h.BroadCastAndResponse(packet)
	}
	return nil
}

func (h *Host) resolveCommandFromServer(packet *Packet) {
	serverCmd := string(packet.Data())
	split := strings.Split(serverCmd, " ")
	if len(split) <= 1 {
		LOG.Infof("Server - %v", string(packet.Data()))
		return
	}
	switch split[0] {
	case CommandAssignID:
		atoi, err := strconv.Atoi(split[1])
		if err != nil {
			LOG.Error("Invalid connection ID: ", split[1])
		} else {
			h.id = atoi
			LOG.Infof("Connection ID assigned: ", h.id)
		}
	default:
		LOG.Warn("Invalid command: ", serverCmd)
	}
}

func (h *Host) getGameBind() string {
	return h.gameConfig.localIp + strconv.Itoa(h.gameConfig.tcpPorts[0])
}

func (h *Host) Close() {
	for _, conn := range h.gConn {
		conn.Close()
	}
	h.gConn = make(map[int]net.Conn)
	h.sConn.Close()
}
