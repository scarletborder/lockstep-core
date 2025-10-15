package room

import (
	"crypto/subtle"
	"lockstep-core/src/clients"
	"lockstep-core/src/constants"
	"log"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Room struct {
	// 基础属性
	ID      string
	RoomCtx *RoomContext
	// Logic   *GameLogic // TODO: 后续整合游戏逻辑

	// 网络通道
	register         chan *clients.Player
	unregister       chan *clients.Player
	incomingMessages chan *clients.PlayerMessage
	// ingameOperations chan *messages.InGameOperation // TODO: 整合 protobuf 消息

	// 安全
	key string // 房间密钥

	// 游戏状态
	GameStage *constants.AtomStage
	ChapterID uint32
	StageID   uint32
	Seed      int32

	// 生命周期管理
	destroyOnce    sync.Once
	LastActiveTime time.Time     // 上次活动时间
	StopChan       chan<- string // 通知房间管理器的停止信号通道
}

// NewRoom 创建一个新的游戏房间
func NewRoom(id string, stopChan chan string) *Room {
	// gameOperationChan := make(chan *messages.InGameOperation, 128)
	// logic := NewGameLogic(gameOperationChan)
	seed := rand.Int31n(40960000) // 随机种子

	return &Room{
		ID:      id,
		RoomCtx: NewRoomContext(),
		// Logic:   logic,

		ChapterID: 0,
		StageID:   0,
		Seed:      seed,

		LastActiveTime: time.Now(),
		key:            "",
		GameStage:      constants.NewAtomStage(constants.STAGE_InLobby),
		StopChan:       stopChan,
		destroyOnce:    sync.Once{},

		// 网络
		register:         make(chan *clients.Player, 8),
		unregister:       make(chan *clients.Player, 8),
		incomingMessages: make(chan *clients.PlayerMessage, 128),
		// ingameOperations: gameOperationChan,
	}
}

// Reset 重置房间为大厅状态，以允许下一场游戏
func (room *Room) Reset() {
	room.RoomCtx.Reset()
	// room.Logic.Reset()

	// room 本身 reset
	room.ChapterID = 0
	room.StageID = 0
	room.Seed = rand.Int31n(40960000)             // 重置随机种子
	room.LastActiveTime = time.Now()              // 重置最后活动时间
	room.GameStage.Store(constants.STAGE_InLobby) // 重置游戏状态为大厅

	// 清空 ingameOperations
	// for len(room.ingameOperations) > 0 {
	// 	<-room.ingameOperations
	// }
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
		if room.RoomCtx.GameTicker != nil {
			room.RoomCtx.GameTicker.Stop()
			room.RoomCtx.GameTicker = nil
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
