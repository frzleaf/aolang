package test

import (
	"fmt"
	"net"
	"proxy"
	"strconv"
)

type MockGuestApp struct {
	AppConfig                 *proxy.GameConfig
	host                      net.Conn
	broadcastResponseListener net.Listener
	onHostReplyFunc           func([]byte)
	stopped                   bool
}

func NewMockGuestApp(config *proxy.GameConfig) *MockGuestApp {
	return &MockGuestApp{
		AppConfig: config,
	}
}

func (h *MockGuestApp) broadCast(data []byte) error {
	dial, err := net.Dial("udp", "255.255.255.255:"+strconv.Itoa(h.AppConfig.UdpPort))
	if err != nil {
		return err
	}
	if _, err := dial.Write(data); err != nil {
		return err
	}
	return nil
}

func (h *MockGuestApp) connectToHost(hostAddr string) error {
	dial, err := net.Dial("tcp", hostAddr+":"+strconv.Itoa(h.AppConfig.TcpPort))
	h.host = dial
	go func() {
		for !h.stopped {
			buf := make([]byte, 1000)
			if read, err2 := h.host.Read(buf); err2 != nil {
				return
			} else {
				if h.onHostReplyFunc != nil {
					h.onHostReplyFunc(buf[0:read])
				}
			}
		}
	}()
	return err
}

func (h *MockGuestApp) OnHostReply(replyFunc func([]byte)) {
	h.onHostReplyFunc = replyFunc
}

func (h *MockGuestApp) sendDataToHost(data []byte) error {
	_, err := h.host.Write(data)
	return err
}

func (h *MockGuestApp) close() {
	h.host.Close()
	fmt.Println("disconnect from host")
}
