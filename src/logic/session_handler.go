package logic

import (
	"lockstep-core/src/logic/clients"
	"lockstep-core/src/logic/room"
	"log"
)

// DefaultPlayerSessionHandler 默认的玩家会话处理器（适配新的 Room 结构）
type DefaultPlayerSessionHandler struct{}

// NewDefaultPlayerSessionHandler 创建默认的玩家会话处理器
func NewDefaultPlayerSessionHandler() *DefaultPlayerSessionHandler {
	return &DefaultPlayerSessionHandler{}
}

// HandleSession 处理单个玩家的会话
// 注意：这个方法现在由 Room.StartServeClient 代替
// 但保留接口以保持兼容性
func (h *DefaultPlayerSessionHandler) HandleSession(r *room.Room, player *clients.Player) {
	log.Printf("🟢 HandleSession called for player %d in room %s", player.GetID(), r.ID)

	// 调用房间的 StartServeClient 方法
	r.StartServeClient(player)
}
