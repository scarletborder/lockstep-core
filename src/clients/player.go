package clients

import (
	"log"
	"sync"
)

// PlayerMessage 直接转发客户端传来的数据
type PlayerMessage struct {
	Player *Player // 发送消息的玩家
	Data   []byte  // 二进制消息内容
}

var playerMessagePool = sync.Pool{
	New: func() interface{} {
		return new(PlayerMessage)
	},
}

// GetPlayerMessage 从对象池获取 PlayerMessage
func GetPlayerMessage(player *Player, data []byte) *PlayerMessage {
	msg := playerMessagePool.Get().(*PlayerMessage)
	msg.Player = player
	msg.Data = data
	return msg
}

// ReleasePlayerMessage 将 PlayerMessage 放回对象池
func ReleasePlayerMessage(msg *PlayerMessage) {
	msg.Player = nil
	msg.Data = nil
	playerMessagePool.Put(msg)
}

// Player 代表一个游戏玩家 (适配 WebTransport)
type Player struct {
	Ctx      *PlayerContext        // 玩家上下文
	SendChan chan<- *PlayerMessage // 发送消息到服务器的通道

	// 游戏相关状态
	IsReady  bool // 是否准备好
	IsLoaded bool // 是否加载完毕

	// 游戏数据 (用于防作弊验证)
	LastEnergySum  int32 // 上一次用户的能量总和
	LastStarShards int32 // 上一次用户的星之碎片
}

// NewPlayer 创建一个新的玩家实例
func NewPlayer(ctx *PlayerContext, sendChan chan<- *PlayerMessage) *Player {
	return &Player{
		Ctx:            ctx,
		IsReady:        false,
		IsLoaded:       false,
		SendChan:       sendChan,
		LastEnergySum:  0,
		LastStarShards: 0,
	}
}

// ResetData 重置玩家的游戏数据
func (p *Player) ResetData() {
	p.IsReady = false
	p.IsLoaded = false
	p.LastEnergySum = 0
	p.LastStarShards = 0
	if p.Ctx != nil {
		p.Ctx.LatestFrameID.Store(0)
		p.Ctx.LatestAckFrameID.Store(0)
	}
}

// GetID 获取玩家 ID
func (p *Player) GetID() int {
	if p == nil || p.Ctx == nil {
		return -1
	}
	return p.Ctx.ID
}

// Write 写入要发送给客户端的消息
func (p *Player) Write(data []byte) {
	if p == nil || p.Ctx == nil {
		log.Printf("🔴 Cannot write message: player or context is nil")
		return
	}

	err := p.Ctx.SendDatagram(data)
	if err != nil {
		log.Printf("🔴 Failed to write message to player %d: %v", p.GetID(), err)
	} else {
		log.Printf("🟢 Message written to player %d, length: %d", p.GetID(), len(data))
	}
}
