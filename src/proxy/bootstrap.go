package proxy

import "os"

var LOG *Logger

var Warcraft3Config *GameConfig

func init() {
	LOG = NewLogger(os.Stdout)
	Warcraft3Config = &GameConfig{
		UdpPort: 6112,
		//TcpPorts: []int{6112, 6110, 6111},
		TcpPorts: []int{6110, 6111, 6112},
		LocalIp:  "127.0.0.1",
	}
}
