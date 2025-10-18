package room

import (
	"crypto/subtle"
	"lockstep-core/src/config"
	"lockstep-core/src/constants"

	"lockstep-core/src/pkg/lockstep/client"
	lockstep_sync "lockstep-core/src/pkg/lockstep/sync"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

type DataChannel struct {
	// 客户端session用于发送注册信息
	register chan *client.Client
	// 客户端session用于发送解除注册信息
	unregister chan *client.Client
	// 客户端用于发送bytes消息
	incomingMessages chan *client.ClientMessage
	// ingameOperations chan *messages.InGameOperation // TODO: 整合 protobuf 消息
}

func (dc *DataChannel) Reset() {
	// 重置通道（关闭旧通道，创建新通道）
	dc.register = make(chan *client.Client, 8)
	dc.unregister = make(chan *client.Client, 8)
	dc.incomingMessages = make(chan *client.ClientMessage, 128)
	// dc.ingameOperations = make(chan *messages.InGameOperation, 128)
}

type Room struct {
	// 基础属性
	ID   uint32
	Name string
	// 安全
	key string // 房间密钥

	// Logic   *GameLogic // TODO: 后续整合游戏逻辑

	// clients
	ClientsContainer

	// 共享数据通道
	DataChannel

	// lockstep sync
	// ticker
	GameTicker *time.Ticker
	// data
	SyncData *lockstep_sync.ServerSyncData
	// config
	config.LockstepConfig

	// 房间Life Cycle生命周期管理
	GameStage constants.AtomStage // 房间当前状态

	// 是否已经摧毁本房间
	destroyOnce sync.Once
	// 房间上次活动时间
	LastActiveTime time.Time
	// 传入本房间id,通知房间管理器的停止信号通道
	StopChan chan<- uint32
}

type RoomOptions struct {
	key  string
	name string // 房间密钥
	config.LockstepConfig
}

// NewRoom 创建一个新的游戏房间
func NewRoom(id uint32, stopChan chan uint32, o RoomOptions) *Room {
	// gameOperationChan := make(chan *messages.InGameOperation, 128)
	// logic := NewGameLogic(gameOperationChan)

	var channel = DataChannel{
		register:         make(chan *client.Client, *o.LockstepConfig.MaxClientsPerRoom),
		unregister:       make(chan *client.Client, *o.LockstepConfig.MaxClientsPerRoom),
		incomingMessages: make(chan *client.ClientMessage, 16*(*o.LockstepConfig.MaxClientsPerRoom)),
		// ingameOperations: gameOperationChan,
	}

	return &Room{
		ID:   id,
		Name: o.name,
		key:  o.key,

		// lockstep
		GameTicker:     nil,
		SyncData:       lockstep_sync.NewServerSyncData(),
		LockstepConfig: o.LockstepConfig,
		// 网络
		DataChannel: channel,
		// ingameOperations: gameOperationChan,

		GameStage:      *constants.NewAtomStage(constants.STAGE_InLobby),
		LastActiveTime: time.Now(),
		StopChan:       stopChan,
		destroyOnce:    sync.Once{},
	}
}

func (r *Room) IsRoomFull() bool {
	return r.GetPlayerCount() >= int(*r.LockstepConfig.MaxClientsPerRoom)
}

// Reset 重置房间为大厅状态，以允许下一场游戏
func (room *Room) Reset() {
	// lockstep sync reset
	room.SyncData.Reset()
	if room.GameTicker != nil {
		room.GameTicker.Stop()
		room.GameTicker = nil
	}
	// room.Logic.Reset()

	// room 本身 reset
	room.LastActiveTime = time.Now()              // 重置最后活动时间
	room.GameStage.Store(constants.STAGE_InLobby) // 重置游戏状态为大厅
	// 清空共享数据,ingameOperations
	room.DataChannel.Reset()
}

// Destroy 摧毁房间
func (room *Room) Destroy() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("摧毁房间出错:捕获到 Panic: %v\n", r)
			log.Printf("堆栈信息:\n%s", string(debug.Stack()))
			log.Println("程序已从 panic 中恢复，将继续运行。")
		}
	}()

	room.destroyOnce.Do(func() {
		log.Printf("🔥 Destroying room %s (stage: %d, players: %d, idle time: %v)",
			room.ID, room.GameStage.Load(), room.GetPlayerCount(), time.Since(room.LastActiveTime))

		// TODO: 发送房间关闭消息
		// room.RoomCtx.BroadcastMessage(...)

		// 通知房间关闭
		room.GameStage.Store(constants.STAGE_CLOSED)

		// 停止定时器
		if room.GameTicker != nil {
			room.GameTicker.Stop()
			room.GameTicker = nil
		}

		room.ClientsContainer.CloseAll()

		// 通知房间管理器移除引用
		room.StopChan <- room.ID
	})
}

// CheckKeyCorrect 检查密钥是否正确（时长无关的检查）
func (room *Room) CheckKeyCorrect(key string) bool {
	return subtle.ConstantTimeCompare([]byte(room.key), []byte(key)) == 1
}

// SetKey 设置房间密钥
func (room *Room) SetKey(key string) {
	room.key = key
}

// HasKey 检查是否有密钥
func (room *Room) HasKey() bool {
	return room.key != ""
}

// UpdateActiveTime 更新房间的最后活跃时间
func (room *Room) UpdateActiveTime() {
	room.LastActiveTime = time.Now()
	log.Printf("🕒 Updated active time for room %s", room.ID)
}

// RegisterPlayer 添加一个玩家到房间（发送注册信号）
func (room *Room) RegisterPlayer(player *client.Client) {
	select {
	case room.register <- player:
		log.Printf("🟢 Player %d registered to room %s", player.GetID(), room.ID)
	default:
		log.Printf("🔴 Failed to register player %d - channel full", player.GetID())
	}
}

// UnregisterPlayer 移除一个玩家（发送注销信号）
func (room *Room) UnregisterPlayer(player *client.Client) {
	select {
	case room.unregister <- player:
		log.Printf("🟡 Player %d unregistered from room %s", player.GetID(), room.ID)
	default:
		log.Printf("🔴 Failed to unregister player %d - channel full", player.GetID())
	}
}

// GetIncomingMessagesChan 获取接收消息的通道
func (room *Room) GetIncomingMessagesChan() chan<- *client.ClientMessage {
	return room.incomingMessages
}

// HasAllPlayerSync 检查是否所有玩家都同步(差帧小于容忍量)
func (room *Room) HasAllPlayerSync() bool {
	if *room.LockstepConfig.MaxDelayFrames < 0 {
		// 不进行延迟检查，直接返回 true
		return true
	}

	// 延迟等待，最多容忍 maxDelayFrames 帧的延迟
	nextRenderFrame := room.SyncData.NextFrameID.Load()
	MaxDelayFrames := uint32(*room.LockstepConfig.MaxDelayFrames)
	var minFrameID uint32

	if nextRenderFrame < MaxDelayFrames {
		minFrameID = 0
	} else {
		minFrameID = nextRenderFrame - MaxDelayFrames
	}

	synced := true
	// 遍历每个玩家的 frameID，若有任意玩家低于阈值，则返回 false
	room.ClientsContainer.Clients.Range(func(key uint32, value *client.Client) bool {
		// 检查玩家是否为空或玩家上下文为空
		if value == nil || value.Session == nil {
			synced = false
			return false
		}

		// 获取当前玩家实际的帧号
		playerCurrentFrame := value.ClientSyncData.LatestNextFrameID.Load()
		if playerCurrentFrame < minFrameID {
			synced = false
			return false
		}
		return true
	})
	return synced
}

// StartServeClient 开始为客户端服务（接收消息）
func (room *Room) StartServeClient(client *client.Client) {
	log.Printf("🟡 StartServeClient for player %d", client.GetID())

	// 检查基本有效性
	if client.Session == nil {
		log.Printf("🔴 Player session is nil for player %d at start", client.GetID())
		return
	}

	log.Printf("🟢 Starting client service for player %d", client.GetID())

	defer func() {
		log.Printf("🟡 StartServeClient ending for player %d", client.GetID())

		// 发送 unregister 信号，通知房间移除这个玩家
		select {
		case room.unregister <- client:
			log.Printf("🟡 Sent unregister signal for player %d", client.GetID())
		default:
			log.Printf("🔴 Failed to send unregister signal for player %d (channel full)", client.GetID())
			if client != nil && client.Session != nil {
				client.Session.Close()
			}
		}

		if r := recover(); r != nil {
			log.Printf("服务用户 %d 时捕获到 Panic: %v\n", client.GetID(), r)
			log.Printf("堆栈信息:\n%s", string(debug.Stack()))
			log.Println("程序已从 panic 中恢复，将继续运行。")
		}
	}()

	// 接收消息循环
	log.Printf("🟡 Starting message loop for player %d", client.GetID())
	for {
		// 使用 WebTransport 接收 datagram
		data, err := client.Session.ReceiveDatagram()
		if err != nil {
			log.Printf("🔴 ReceiveDatagram error for player %d: %v", client.GetID(), err)
			return
		}

		log.Printf("🟡 Received datagram from player %d, length: %d", client.GetID(), len(data))

		// 发送到消息管道
		msg := client.GetPlayerMessage(client, data)
		room.incomingMessages <- msg
	}
}
