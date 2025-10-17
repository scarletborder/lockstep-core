package room

import (
	"lockstep-core/src/logic/clients"
	"lockstep-core/src/constants"
	"log"
	"sync"
	"sync/atomic"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/quic-go/webtransport-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// RoomContext 房间上下文，维护玩家相关的状态和信息
type RoomContext struct {
	OwnerUserID int               // 房主 ID
	Players     clients.PlayerMap // 玩家映射
	nextID      int32             // 用于生成用户 ID，初始值设置为 100

	// 帧同步
	NextFrameID    *atomic.Uint32    // 发送给客户端渲染的下一帧 ID
	OperationID    map[uint32]uint32 // 某帧下一次操作的序号 (FrameID -> OperationID)
	OperationMutex sync.Mutex        // 确保操作 ID 的原子性,仅用作操作Frame ID

	// 游戏逻辑定时器
	GameTicker *time.Ticker // 一直是 nil 直到 InGame 状态切换前赋值
}

// NewRoomContext 创建一个新的 RoomContext 实例
func NewRoomContext() *RoomContext {
	nextRenderFrame := &atomic.Uint32{}
	nextRenderFrame.Store(1) // 下一帧渲染为 1，当前都在 0
	return &RoomContext{
		NextFrameID:    nextRenderFrame,
		OwnerUserID:    0, // unset
		nextID:         100,
		OperationID:    make(map[uint32]uint32),
		OperationMutex: sync.Mutex{},
		GameTicker:     nil,
	}
}

// GetNextOperationID 获得某帧下一次操作的序号并自增
func (rc *RoomContext) GetNextOperationID(frameID uint32) uint32 {
	rc.OperationMutex.Lock()
	defer rc.OperationMutex.Unlock()

	nextID, exists := rc.OperationID[frameID]
	if !exists {
		rc.OperationID[frameID] = 2
		return 1
	}

	rc.OperationID[frameID] = nextID + 1
	return nextID
}

// DeleteOperationID 在广播某帧后，删除其 map 记录
func (rc *RoomContext) DeleteOperationID(frameID uint32) {
	rc.OperationMutex.Lock()
	defer rc.OperationMutex.Unlock()
	delete(rc.OperationID, frameID)
}

// Reset 重置状态以允许下场游戏
func (rc *RoomContext) Reset() {
	rc.NextFrameID.Store(1) // 重置帧 ID 为 1
	if rc.GameTicker != nil {
		rc.GameTicker.Stop()
	}
	rc.GameTicker = nil

	// 重置玩家数据
	rc.Players.Range(func(key int, player *clients.Player) bool {
		if player != nil {
			player.ResetData()
		}
		return true
	})
}

// StartGameTicker 设置游戏逻辑定时器
func (rc *RoomContext) StartGameTicker() {
	if rc.GameTicker != nil {
		rc.GameTicker.Stop()
	}
	rc.GameTicker = time.NewTicker(constants.FrameIntervalMs * time.Millisecond)
}

// CreatePlayerContext 从 WebTransport Session 创建玩家上下文
func (rc *RoomContext) CreatePlayerContext(session interface{}, id int) *clients.PlayerContext {
	// 这里暂时不使用自动生成 ID，因为我们在外部创建
	// newID := int(atomic.AddInt32(&rc.nextID, 1))
	return clients.NewPlayerContext(session.(*webtransport.Session), id)
}

// AddUser 添加用户
func (rc *RoomContext) AddUser(p *clients.Player) {
	rc.Players.Store(p.Ctx.ID, p)
	log.Printf("🔵 Player %d added to room context", p.GetID())

	// 如果 OwnerUserID 还没设置，则设置为第一个加入的用户
	if rc.OwnerUserID == 0 {
		rc.OwnerUserID = p.Ctx.ID
	}
}

// DelUser 删除指定用户
func (rc *RoomContext) DelUser(id int) {
	rc.Players.Delete(id)
}

// CloseAll 关闭所有用户连接
func (rc *RoomContext) CloseAll() {
	rc.Players.Range(func(key int, player *clients.Player) bool {
		if player != nil && player.Ctx != nil {
			player.Ctx.Close()
		}
		return true
	})
}

// GetPeerAddr 返回所有连接的远程地址
func (rc *RoomContext) GetPeerAddr() []string {
	var addrs []string
	rc.Players.Range(func(key int, player *clients.Player) bool {
		if player != nil && player.Ctx != nil {
			if addr := player.Ctx.GetRemoteAddr(); addr != "" {
				addrs = append(addrs, addr)
			}
		}
		return true
	})
	return addrs
}

// GetPlayerCount 获取玩家数量
func (rc *RoomContext) GetPlayerCount() uint32 {
	return uint32(rc.Players.Len())
}

// BroadcastMessage 广播 protobuf 消息
func (rc *RoomContext) BroadcastMessage(msg protoreflect.ProtoMessage, excludeIDs []int) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("🔴 BroadcastMessage marshal error: %v", err)
		return
	}

	// 创建排除集合
	excludeSet := mapset.NewSet[int]()
	for _, id := range excludeIDs {
		excludeSet.Add(id)
	}

	rc.Players.Range(func(key int, player *clients.Player) bool {
		if player == nil || player.Ctx == nil || !player.Ctx.IsConnected() {
			return true
		}

		// 检查是否在排除列表中
		if excludeSet.Contains(player.Ctx.ID) {
			return true
		}

		player.Write(data)
		return true
	})
}

// SendMessageToUser 单播消息给指定用户
func (rc *RoomContext) SendMessageToUser(msg protoreflect.ProtoMessage, userID int) {
	if player, ok := rc.Players.Load(userID); ok {
		rc.SendMessageToUserByPlayer(msg, player)
	}
}

// SendMessageToUserByPlayer 通过 Player 实例发送消息
func (rc *RoomContext) SendMessageToUserByPlayer(msg protoreflect.ProtoMessage, player *clients.Player) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("SendMessageToUserByPlayer marshal error: %v", err)
		return
	}
	if player != nil && player.Ctx != nil && player.Ctx.IsConnected() {
		player.Write(data)
	}
}
