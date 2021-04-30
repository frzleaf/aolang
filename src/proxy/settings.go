package proxy

// 4 (ID)
// 2 (Connector - ID)
// n - 7 (Data)
var MESSAGE_PREFIX_SIGN = []byte{27, 07, 19, 93}

const ConnectorIDLength = 2
const ServerConnectorID = 0

type GameConfig struct {
	udpPort  int
	tcpPorts []int
	localIp  string
}

type ConnectionConfig struct {
	serverAddress string
}

type ClientConfig struct {
	gameConfig       *GameConfig
	connectionConfig *ConnectionConfig
}
