package client

import (
	"lockstep-core/src/messages"
	"lockstep-core/src/pkg/lockstep/session"
	lockstep_sync "lockstep-core/src/pkg/lockstep/sync"
	"log"
	"sync"
)

// ClientMessage
// client session转发到room的消息
type ClientMessage struct {
	Client *Client // 发送消息的玩家

	// 二进制消息内容
	// 实际是 本框架的基础类型 + 拓展bytes
	SessionRequest *messages.SessionRequest
}

type ClientMessagePool struct {
	pool sync.Pool
}

func NewClientMessagePool() *ClientMessagePool {
	return &ClientMessagePool{
		pool: sync.Pool{
			New: func() interface{} {
				return new(ClientMessage)
			},
		},
	}
}

// GetPlayerMessage
// 从对象池获取并赋值一个 PlayerMessage
func (p *ClientMessagePool) GetPlayerMessage(client *Client, data *messages.SessionRequest) *ClientMessage {
	msg := p.pool.Get().(*ClientMessage)
	msg.Client = client
	msg.SessionRequest = data
	return msg
}

// ReleasePlayerMessage
// 消费结束某消息后将 PlayerMessage 释放回对象池
func (p *ClientMessagePool) ReleasePlayerMessage(msg *ClientMessage) {
	msg.Client = nil
	if msg.SessionRequest != nil {
		msg.SessionRequest.Reset()
	}
	msg.SessionRequest = nil
	p.pool.Put(msg)
}

// Client 代表一个客户端
// 已和游戏世界逻辑解耦
type Client struct {
	// 客户端会话
	Session session.ISession
	// 持有对 "发送消息到服务器的通道" 的引用
	SendChan chan<- *ClientMessage

	// 共享data池以节省开销
	*ClientMessagePool
	// lockstep
	lockstep_sync.ClientSyncData

	// 房间Life Cycle生命周期相关状态
	IsReady  bool // 是否准备好
	IsLoaded bool // 是否加载完毕

	// 游戏数据 (用于防作弊验证)
	// Deprecated, 在游戏世界中做验证
	// LastEnergySum  int32 // 上一次用户的能量总和
	// LastStarShards int32 // 上一次用户的星之碎片
}

// NewClient 创建一个新的玩家实例
func NewClient(uid uint32, sess session.ISession, sendChan chan<- *ClientMessage) *Client {
	return &Client{
		Session:           sess,
		IsReady:           false,
		IsLoaded:          false,
		SendChan:          sendChan,
		ClientMessagePool: NewClientMessagePool(),
		ClientSyncData:    *lockstep_sync.NewClientSyncData(uid),
	}
}

// ResetData 重置玩家的游戏数据
func (p *Client) ResetData() {
	p.IsReady = false
	p.IsLoaded = false
	p.ClientSyncData.Reset()
}

// GetID 获取玩家 ID
func (p *Client) GetID() uint32 {
	return p.ClientSyncData.ID
}

// Write 写入要发送给客户端的消息
func (p *Client) Write(data []byte) {
	if p == nil || p.Session == nil {
		log.Printf("🔴 Cannot write message: player or context is nil")
		return
	}

	err := p.Session.SendDatagram(data)
	if err != nil {
		log.Printf("🔴 Failed to write message to player %d: %v", p.GetID(), err)
	} else {
		log.Printf("🟢 Message written to player %d, length: %d", p.GetID(), len(data))
	}
}
