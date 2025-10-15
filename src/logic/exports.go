// Package logic 提供游戏逻辑的统一入口点
// 这个包重新导出了分散在 clients、room、game 等包中的核心类型
package logic

import (
	"lockstep-core/src/logic/clients"
	"lockstep-core/src/logic/room"
)

// 类型别名，用于向后兼容
type (
	// Player 是 clients.Player 的别名
	Player = clients.Player

	// PlayerContext 是 clients.PlayerContext 的别名
	PlayerContext = clients.PlayerContext

	// PlayerMessage 是 clients.PlayerMessage 的别名
	PlayerMessage = clients.PlayerMessage

	// PlayerMap 是 clients.PlayerMap 的别名
	PlayerMap = clients.PlayerMap

	// RoomContext 是 room.RoomContext 的别名
	RoomContext = room.RoomContext

	// RoomManager 是 room.RoomManager 的别名
	RoomManager = room.RoomManager
)

// 导出的构造函数
var (
	// NewPlayer 创建新的玩家实例
	NewPlayer = clients.NewPlayer

	// NewPlayerContext 创建新的玩家上下文
	NewPlayerContext = clients.NewPlayerContext

	// GetPlayerMessage 从对象池获取 PlayerMessage
	GetPlayerMessage = clients.GetPlayerMessage

	// ReleasePlayerMessage 将 PlayerMessage 放回对象池
	ReleasePlayerMessage = clients.ReleasePlayerMessage

	// NewRoomContext 创建新的房间上下文
	NewRoomContext = room.NewRoomContext

	// NewRoomManager 创建新的房间管理器
	NewRoomManager = room.NewRoomManager
)
