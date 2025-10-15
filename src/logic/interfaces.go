package logic

import (
	"lockstep-core/src/clients"
	"lockstep-core/src/logic/room"
)

// RoomManagerInterface 定义房间管理器的接口
type RoomManagerInterface interface {
	// GetRoom 获取指定 ID 的房间
	GetRoom(roomID string) (*room.Room, bool)

	// CreateRoom 创建一个新房间
	CreateRoom(roomID string) *room.Room

	// GetOrCreateRoom 获取或创建一个房间
	GetOrCreateRoom(roomID string) *room.Room

	// RemoveRoom 删除一个房间
	RemoveRoom(roomID string)

	// ListRooms 列出所有房间 ID
	ListRooms() []string

	// GetRoomCount 获取房间数量
	GetRoomCount() int
}

// PlayerSessionHandler 定义玩家会话处理器的接口（适配新的 Room 结构）
type PlayerSessionHandler interface {
	// HandleSession 处理单个玩家的会话
	HandleSession(r *room.Room, player *clients.Player)
}
