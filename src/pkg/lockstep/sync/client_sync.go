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
	LatestFrameID    atomic.Uint32 // 最近服务器获知的该用户所在的帧
	LatestAckFrameID atomic.Uint32 // 最近该用户确认(ACK)的帧
}

func NewClientSyncData(id uint32) *ClientSyncData {
	csd := &ClientSyncData{
		ID: id,
	}
	csd.LatestFrameID.Store(0)
	csd.LatestAckFrameID.Store(0)
	return csd
}

func (pc *ClientSyncData) Reset() {
	pc.LatestFrameID.Store(0)
	pc.LatestAckFrameID.Store(0)
}

// UpdatePlayerFrame 更新玩家的帧同步信息
func (pc *ClientSyncData) UpdatePlayerFrame(frameID, ackFrameID uint32) {
	oldFrame := pc.LatestFrameID.Load()
	if frameID > oldFrame {
		pc.LatestFrameID.Store(frameID)
	}

	oldAck := pc.LatestAckFrameID.Load()
	if ackFrameID > oldAck {
		pc.LatestAckFrameID.Store(ackFrameID)
	}
}
