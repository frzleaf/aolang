package proxy

import (
	"os"
	"strings"
)

var LOG *Logger

var Warcraft3Config *GameConfig

func init() {
	logLevel := InfoLevel
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "panic":
		logLevel = PanicLevel
	case "fatal":
		logLevel = FatalLevel
	case "error":
		logLevel = ErrorLevel
	case "warn":
		logLevel = WarnLevel
	case "info":
		logLevel = InfoLevel
	case "debug":
		logLevel = DebugLevel
	case "trace":
		logLevel = TraceLevel
	}
	LOG = NewLoggerWithLevel(os.Stdout, logLevel)
	Warcraft3Config = &GameConfig{
		UdpPort: 6112,
		//TcpPorts: []int{6112, 6110, 6111},
		TcpPorts: []int{6110, 6111, 6112},
		TcpPort:  6112,
		LocalIp:  "127.0.0.1",
	}
}
