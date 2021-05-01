package proxy

import "net"

type Agent struct {
	sConn      net.Conn // server connection
	id         int      // connection id
	monitor    *Monitor
	gameConfig *GameConfig
	online     bool
}

func (h *Agent) IsOn() bool {
	return h.online
}

func (h *Agent) TurnOn() {
	h.online = true
}

func (h *Agent) TurnOff() {
	h.online = false
}
