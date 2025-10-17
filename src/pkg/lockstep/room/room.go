package room

import (
	"crypto/subtle"
	"lockstep-core/src/config"
	"lockstep-core/src/constants"
	"lockstep-core/src/logic/clients"
	lockstep_sync "lockstep-core/src/pkg/lockstep/sync"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type DataChannel struct {
	// 客户端session用于发送注册信息
	register chan *clients.Player
	// 客户端session用于发送解除注册信息
	unregister chan *clients.Player
	// 客户端用于发送bytes消息
	incomingMessages chan *clients.PlayerMessage
	// ingameOperations chan *messages.InGameOperation // TODO: 整合 protobuf 消息
}

func (dc *DataChannel) Reset() {
	// 重置通道（关闭旧通道，创建新通道）
	dc.register = make(chan *clients.Player, 8)
	dc.unregister = make(chan *clients.Player, 8)
	dc.incomingMessages = make(chan *clients.PlayerMessage, 128)
	// dc.ingameOperations = make(chan *messages.InGameOperation, 128)
}

type Room struct {
	// 基础属性
	ID   uint32
	Name string
	// 安全
	key string // 房间密钥

	// Logic   *GameLogic // TODO: 后续整合游戏逻辑

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
		register:         make(chan *clients.Player, 8),
		unregister:       make(chan *clients.Player, 8),
		incomingMessages: make(chan *clients.PlayerMessage, 128),
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

		room.RoomCtx.CloseAll()

		// 通知房间管理器移除引用
		room.StopChan <- room.ID
	})
}

// GetPlayerCount 获取玩家数量
func (room *Room) GetPlayerCount() uint32 {
	return room.RoomCtx.GetPlayerCount()
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

// HasAllPlayerReady 检查是否所有玩家都准备就绪
func (room *Room) HasAllPlayerReady() bool {
	allReady := true
	room.RoomCtx.Players.Range(func(key int, value *clients.Player) bool {
		if !value.IsReady {
			allReady = false
			return false // 立刻停止遍历
		}
		return true
	})
	return allReady
}

// HasAllPlayerLoaded 检查是否所有玩家都加载完毕
func (room *Room) HasAllPlayerLoaded() bool {
	allLoaded := true
	room.RoomCtx.Players.Range(func(key int, value *clients.Player) bool {
		if !value.IsLoaded {
			allLoaded = false
			return false
		}
		return true
	})
	return allLoaded
}

// PlayerReadyCount 统计准备好的玩家数量
func (room *Room) PlayerReadyCount() uint32 {
	count := uint32(0)
	room.RoomCtx.Players.Range(func(key int, value *clients.Player) bool {
		if value.IsReady {
			count++
		}
		return true
	})
	return count
}

// UpdateActiveTime 更新房间的最后活跃时间
func (room *Room) UpdateActiveTime() {
	room.LastActiveTime = time.Now()
	log.Printf("🕒 Updated active time for room %s", room.ID)
}

// AddPlayer 添加一个玩家到房间（发送注册信号）
func (room *Room) AddPlayer(player *clients.Player) {
	select {
	case room.register <- player:
		log.Printf("🟢 Player %d registered to room %s", player.GetID(), room.ID)
	default:
		log.Printf("🔴 Failed to register player %d - channel full", player.GetID())
	}
}

// RemovePlayer 移除一个玩家（发送注销信号）
func (room *Room) RemovePlayer(player *clients.Player) {
	select {
	case room.unregister <- player:
		log.Printf("🟡 Player %d unregistered from room %s", player.GetID(), room.ID)
	default:
		log.Printf("🔴 Failed to unregister player %d - channel full", player.GetID())
	}
}

// BroadcastMessage 广播 protobuf 消息到所有玩家（除了排除列表中的）
func (room *Room) BroadcastMessage(msg protoreflect.ProtoMessage, excludeIDs []int) {
	room.RoomCtx.BroadcastMessage(msg, excludeIDs)
}

// SendMessageToPlayer 发送消息给指定玩家
func (room *Room) SendMessageToPlayer(msg protoreflect.ProtoMessage, playerID int) {
	room.RoomCtx.SendMessageToUser(msg, playerID)
}

// GetIncomingMessagesChan 获取接收消息的通道
func (room *Room) GetIncomingMessagesChan() chan<- *clients.PlayerMessage {
	return room.incomingMessages
}
