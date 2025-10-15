package constants

// PlayerState 玩家连接状态
type PlayerState uint8

const (
	PlayerStateConnected    PlayerState = iota + 0x10 // 在线
	PlayerStateDisconnected                           // 断线（等待重连中）
)
