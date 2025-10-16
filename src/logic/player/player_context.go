package player

import (
	"context"
	"fmt"
	"lockstep-core/src/constants"
	"log"
	"sync"
	"sync/atomic"

	"github.com/quic-go/webtransport-go"
)

// PlayerContext 封装玩家的 WebTransport 连接和上下文信息
type PlayerContext struct {
	mu      sync.RWMutex          // 保护会话的读写锁
	Session *webtransport.Session // WebTransport 会话
	Ctx     context.Context       // 上下文，用于控制生命周期
	Cancel  context.CancelFunc    // 取消函数

	// 基础信息
	ID int // 用户 ID

	// 帧同步信息
	LatestFrameID    atomic.Uint32 // 最近服务器获知的该用户所在的帧
	LatestAckFrameID atomic.Uint32 // 最近该用户确认(ACK)的帧

	// 连接状态
	ReconnectionToken string                // 重连令牌
	State             constants.PlayerState // 当前状态（在线/断线）
}

// NewPlayerContext 创建一个新的玩家上下文
func NewPlayerContext(session *webtransport.Session, id int) *PlayerContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &PlayerContext{
		Session: session,
		Ctx:     ctx,
		Cancel:  cancel,
		ID:      id,
		State:   constants.PlayerStateConnected,
	}
}

// Close 关闭玩家连接
func (pc *PlayerContext) Close() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.Cancel != nil {
		log.Printf("🔴 Closing connection for player %d", pc.ID)
		pc.Cancel()
		pc.Cancel = nil
	}

	if pc.Session != nil {
		pc.Session.CloseWithError(0, "player disconnected")
		pc.Session = nil
	}
}

// IsConnected 安全地检查连接是否有效
func (pc *PlayerContext) IsConnected() bool {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.Session != nil && pc.Ctx.Err() == nil
}

// SendDatagram 安全地发送 datagram 消息
func (pc *PlayerContext) SendDatagram(data []byte) error {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if pc.Session == nil {
		return fmt.Errorf("session is nil")
	}

	err := pc.Session.SendDatagram(data)
	if err != nil {
		log.Printf("🔴 SendDatagram error for player %d: %v", pc.ID, err)
	}
	return err
}

// ReceiveDatagram 安全地接收 datagram 消息
func (pc *PlayerContext) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	pc.mu.RLock()
	session := pc.Session
	pc.mu.RUnlock()

	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}

	data, err := session.ReceiveDatagram(ctx)
	if err != nil {
		log.Printf("🔴 ReceiveDatagram error for player %d: %v", pc.ID, err)
	}
	return data, err
}

// UpdatePlayerFrame 更新玩家的帧同步信息
func (pc *PlayerContext) UpdatePlayerFrame(frameID, ackFrameID uint32) {
	oldFrame := pc.LatestFrameID.Load()
	if frameID > oldFrame {
		pc.LatestFrameID.Store(frameID)
	}

	oldAck := pc.LatestAckFrameID.Load()
	if ackFrameID > oldAck {
		pc.LatestAckFrameID.Store(ackFrameID)
	}
}

// GetRemoteAddr 获取远程地址（WebTransport 版本）
func (pc *PlayerContext) GetRemoteAddr() string {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	if pc.Session == nil {
		return ""
	}
	return pc.Session.RemoteAddr().String()
}
