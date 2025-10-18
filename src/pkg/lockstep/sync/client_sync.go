package lockstep_sync

import (
	"sync/atomic"
)

// EnumPlayerState 玩家连接状态
type EnumPlayerState uint8

const (
	PlayerStateConnected    EnumPlayerState = iota + 0x10 // 在线
	PlayerStateDisconnected                               // 断线（等待重连中）
)

type ClientSyncData struct {
	ID uint32 // 用户 ID
	// 帧同步信息
	LatestNextFrameID    atomic.Uint32 // 最近服务器获知的该用户所在的下一帧
	LatestAckNextFrameID atomic.Uint32 // 最近该用户确认(ACK)的帧
}

func NewClientSyncData(id uint32) *ClientSyncData {
	csd := &ClientSyncData{
		ID: id,
	}
	csd.LatestNextFrameID.Store(1)
	csd.LatestAckNextFrameID.Store(0)
	return csd
}

func (pc *ClientSyncData) Reset() {
	pc.LatestNextFrameID.Store(1)
	pc.LatestAckNextFrameID.Store(0)
}

// UpdatePlayerFrame 更新玩家的帧同步信息
func (pc *ClientSyncData) UpdatePlayerFrame(nextFrameID, ackNextFrameID uint32) {
	oldFrame := pc.LatestNextFrameID.Load()
	if nextFrameID > oldFrame {
		pc.LatestNextFrameID.Store(nextFrameID)
	}

	oldAck := pc.LatestAckNextFrameID.Load()
	if ackNextFrameID > oldAck {
		pc.LatestAckNextFrameID.Store(ackNextFrameID)
	}
}

// GetCurrentFrame 获取玩家当前所在的帧号
func (pc *ClientSyncData) GetCurrentNextFrame() (nextFrameID, ackNextFrameID uint32) {
	nextFrameID = pc.LatestNextFrameID.Load()
	ackNextFrameID = pc.LatestAckNextFrameID.Load()
	return
}
