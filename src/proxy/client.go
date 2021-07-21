package proxy

type Client interface {
	ServerConnector() *ServerConnector
	SelectTargetId(targetId int)
	TargetId() int
	GameConfig() *GameConfig
	OnMatch() bool
}
