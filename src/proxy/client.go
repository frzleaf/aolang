package proxy

const ClientModeGuest = "guest"
const ClientModeHost = "host"

type Client interface {
	ServerConnector() *ServerConnector
	SelectTargetId(targetId int)
	TargetId() int
	GameConfig() *GameConfig
	OnMatch() bool
	ConnectServer() error
	Close() error
	OnConnectSuccess(func(client Client))
}
