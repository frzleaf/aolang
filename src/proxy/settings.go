package proxy

// 4 (ID)
// 2 (ServerConnector - ID)
// n - 7 (Data)
var MESSAGE_PREFIX_SIGN = []byte{27, 07, 19, 93}

const ConnectorIDLength = 2
const ServerConnectorID = 0

type GameConfig struct {
	UdpPort          int
	TcpPorts         []int
	TcpPort          int
	LocalIp          string
	InternalRemoteIp string
}

type ConnectionConfig struct {
	serverAddress string
}

type ClientConfig struct {
	gameConfig       *GameConfig
	connectionConfig *ConnectionConfig
}

const CharSplitCommand = " "

const CommandAssignID = "/assign"
const CommandFindOpenTCPPort = "/port"
const CommandFind = "/find"
const CommandSelectTarget = "/to"
const CommandExitGame = "/exit"
const CommandDisconnected = "/disconnected"
