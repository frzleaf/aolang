package proxy

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

type Host struct {
	gConn      map[int]net.Conn // game connections
	sConn      net.Conn         // server connection
	id         int
	config     *ClientConfig
	monitor    *Monitor
	gameConfig *GameConfig
}

func NewHost() *Host {
	return &Host{
		gConn: make(map[int]net.Conn),
		id:    -1,
	}
}

func (h *Host) GetGameList() (data []byte, err error) {
	raddr := net.UDPAddr{
		Port: 6112,
		IP:   net.ParseIP("localhost"),
	}

	scannerConn, err := net.DialUDP("udp4", nil, &raddr)
	if err != nil {
		log.Fatal(err)
	}

	defer scannerConn.Close()
	buf := make([]byte, 1000)
	var scanCounter byte = 0

	for {
		_, err = scannerConn.Write(
			[]byte{0xf7, 0x2f, 0x10, 0x00, 0x50, 0x58, 0x33, 0x57, 0x18, 0x00, 0x00, 0x00, scanCounter, 0x00, 0x00, 0x00},
		)
		scanCounter = scanCounter + 1
		scannerConn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
		read, _, err1 := scannerConn.ReadFromUDP(buf)
		if err1 != nil {
			return nil, err1

		} else {
			return buf[0:read], nil
		}
	}
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

func (h *Host) gameToServer(srcId int) {
	gConn := h.gConn[srcId]
	buf := make([]byte, 1000)
	for {
		read, err := gConn.Read(buf)
		if err != nil {
			if err.Error() == "EOF" {
				continue
			}
			fmt.Println("Error on game send ", err)
			gConn.Close()
			delete(h.gConn, srcId)
			break
		} else {
			if _, err = h.sConn.Write(NewInformPacket(h.id, srcId, buf[:read]).ToBytes()); err != nil {
				fmt.Println("Host response error: ", err)
			}
		}
	}
}

func (h *Host) serverToGame(data []byte) (err error) {
	packets := PacketFromBytes(data)
	for _, packet := range packets {
		LOG.Infof("\nReceive msg from #%v len %v\n", packet.SrcAddr(), len(packet.Data()))
		switch packet.pkgType {
		case PackageTypeInform:
			conn := h.gConn[packet.SrcAddr()]
			if conn == nil {
				conn, err = h.PrepareNewGameConnection(packet.src)
				if err != nil {
					return err
				}
				LOG.Info("New client join host: ", packet.SrcAddr())
			}
			if _, err = conn.Write(packet.data); err != nil {
				return err
			}
		case PackageTypeConnectHost:
			gConn := h.gConn[packet.SrcAddr()]
			if gConn == nil {
				gConn, err = h.PrepareNewGameConnection(packet.src)
				if err != nil {
					return err
				}
			}
			if _, err := gConn.Write(packet.data); err == nil {
				LOG.Info("New client join host: ", packet.SrcAddr())
			}
		case PackageTypeFindHost:
			if gameData, err := h.GetGameList(); err == nil {
				h.sConn.Write(NewPacket(PackageTypeFindHostResponse, h.id, packet.SrcAddr(), gameData).ToBytes())
			} else {
				LOG.Info("No game found: ", err)
			}
		}
	}
	return
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
