package proxy

import "os"

var LOG *Logger

var Warcraft3Config *GameConfig

func init() {
	LOG = NewLogger(os.Stdout)
	Warcraft3Config = &GameConfig{
		udpPort:  6112,
		//tcpPorts: []int{6112, 6110, 6111},
		tcpPorts: []int{6110, 6111, 6112},
		localIp:  "127.0.0.1",
	}
}
