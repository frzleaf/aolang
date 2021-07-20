package test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"proxy"
	"strconv"
)

type MockHostApp struct {
	AppConfig     *proxy.GameConfig
	guests        map[int]net.Conn
	guestCount    int
	stopped       bool
	onDataFunc    func([]byte, int)
	onUdpDataFunc func([]byte, *net.UDPConn)
}

func NewMockHostApp(config *proxy.GameConfig) *MockHostApp {
	return &MockHostApp{
		AppConfig: config,
		guests:    make(map[int]net.Conn),
	}
}

func (h *MockHostApp) listenUdp() error {
	udpAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: h.AppConfig.UdpPort,
	}

	listener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	go func() {
		defer listener.Close()
		buf := make([]byte, 1000)
		for !h.stopped {
			if read, err2 := listener.Read(buf); err2 != nil {
				return
			} else {
				if h.onUdpDataFunc != nil {
					h.onUdpDataFunc(buf[0:read], listener)
				}
			}
		}
	}()
	return nil
}

func (h *MockHostApp) onNewClient(conn net.Conn) {
	h.guestCount += 1
	currentCounter := h.guestCount
	h.guests[currentCounter] = conn
	go func() {
		defer conn.Close()
		buf := make([]byte, 1000)
		for !h.stopped {
			if read, err := conn.Read(buf); err != nil {
				if err == io.EOF {
					continue
				} else {
					return
				}
			} else {
				readData := buf[0:read]
				if h.onDataFunc != nil {
					h.onDataFunc(readData, currentCounter)
				}
			}
		}
	}()
}

func (h *MockHostApp) OnData(onDataFunc func([]byte, int)) {
	h.onDataFunc = onDataFunc
}

func (h *MockHostApp) listen() error {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(h.AppConfig.TcpPort))
	if err != nil {
		return err
	} else {
		fmt.Println("Open connection at:", listener.Addr().String())
	}

	go func() {
		defer listener.Close()
		for !h.stopped {
			if newConn, err1 := listener.Accept(); err1 != nil {
				fmt.Errorf("error on accept new conn %s", err1)
			} else {
				h.onNewClient(newConn)
			}
		}
	}()
	return nil
}

func (h *MockHostApp) sendData(data []byte, guestId int) error {
	conn := h.guests[guestId]
	if conn == nil {
		return errors.New("no guest found with id " + strconv.Itoa(guestId))
	}
	_, err := conn.Write(data)
	return err
}

func (h *MockHostApp) close() {
	for i, guest := range h.guests {
		guest.Close()
		fmt.Println("close connection ", i)
	}
	h.stopped = true
}

func (h *MockHostApp) onUdpData(onUdpData func([]byte, *net.UDPConn)) {
	h.onUdpDataFunc = onUdpData
}
