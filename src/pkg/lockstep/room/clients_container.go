package room

import (
	"lockstep-core/src/config"
	"lockstep-core/src/pkg/lockstep/client"
	"lockstep-core/src/utils"
	"log"

	mapset "github.com/deckarep/golang-set/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ClientsContainer 房间上下文，维护玩家相关的状态和信息
type ClientsContainer struct {
	Clients client.PlayerMap // 玩家映射
	*utils.SafeIDAllocator
}

// NewClientsContainer 创建一个新的 RoomContext 实例
func NewClientsContainer(cfg config.LockstepConfig) *ClientsContainer {
	return &ClientsContainer{
		SafeIDAllocator: utils.NewSafeIDAllocator(utils.RoundUpTo64(uint32(*cfg.MaxClientsPerRoom))),
	}
}

func (rc *ClientsContainer) GetNextUserID() (uint32, error) {
	return rc.SafeIDAllocator.Allocate()
}

func (rc *ClientsContainer) HasUser(uid uint32) bool {
	_, ok := rc.Clients.Load(uid)
	return ok
}

// 内置方法

// Reset 重置状态以允许下场游戏
func (rc *ClientsContainer) Reset() {
	// 重置玩家数据
	rc.Clients.Range(func(uid uint32, player *client.Client) bool {
		if player != nil {
			player.ResetData()
		}
		return true
	})
}

// AddUser 添加用户
// 这里是最终处理逻辑，不要直接调用
func (rc *ClientsContainer) AddUser(p *client.Client) {
	rc.Clients.Store(p.GetID(), p)
	log.Printf("🔵 Player %d added to room context", p.GetID())
}

// DelUser 删除指定用户
// 这里是最终处理逻辑，不要直接调用
func (rc *ClientsContainer) DelUser(uid uint32) {
	rc.Clients.Delete(uid)
	rc.SafeIDAllocator.Free(uid)
}

// CloseAll 关闭所有用户连接
func (rc *ClientsContainer) CloseAll() {
	rc.Clients.Range(func(key uint32, player *client.Client) bool {
		if player != nil && player.Session != nil {
			player.Session.Close()
		}
		return true
	})
}

// GetPlayerCount 获取玩家数量
func (rc *ClientsContainer) GetPlayerCount() int {
	return int(rc.Clients.Len())
}

// GetActivePlayerCount 获取活跃玩家数量
func (rc *ClientsContainer) GetActivePlayerCount() int {
	count := 0
	rc.Clients.Range(func(key uint32, player *client.Client) bool {
		if player != nil && player.Session != nil && player.Session.IsConnected() {
			count++
		}
		return true
	})
	return count
}

// HasAllPlayerReady 检查是否所有玩家都准备就绪
func (rc *ClientsContainer) HasAllPlayerReady() bool {
	allReady := true
	rc.Clients.Range(func(key uint32, player *client.Client) bool {
		if player != nil && !player.IsReady {
			allReady = false
		}
		return true
	})
	return allReady
}

// PlayerReadyCount 统计准备好的玩家数量
func (rc *ClientsContainer) PlayerReadyCount() int {
	count := 0
	rc.Clients.Range(func(key uint32, player *client.Client) bool {
		if player != nil && player.IsReady {
			count++
		}
		return true
	})
	return count
}

// HasAllPlayerLoaded 检查是否所有玩家都加载完毕
func (rc *ClientsContainer) HasAllPlayerLoaded() bool {
	allLoaded := true

	rc.Clients.Range(func(key uint32, player *client.Client) bool {
		if player != nil && !player.IsLoaded {
			allLoaded = false
		}
		return true
	})
	return allLoaded
}

// BroadcastMessage 广播 protobuf 消息
func (rc *ClientsContainer) BroadcastMessage(msg protoreflect.ProtoMessage, excludeIDs []uint32) {
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("🔴 BroadcastMessage marshal error: %v", err)
		return
	}

	// 创建排除集合
	excludeSet := mapset.NewSet[uint32]()
	for _, id := range excludeIDs {
		excludeSet.Add(id)
	}

	rc.Clients.Range(func(key uint32, client *client.Client) bool {
		if client == nil || client.Session == nil || !client.Session.IsConnected() {
			return true
		}

		// 检查是否在排除列表中
		if excludeSet.Contains(client.ID) {
			return true
		}

		client.Write(data)
		return true
	})
}

// SendMessageToUser 单播消息给指定用户
func (rc *ClientsContainer) SendMessageToUser(data []byte, userID uint32) {
	if client, ok := rc.Clients.Load(userID); ok {
		rc.SendMessageToUserByPlayer(data, client)
	}
}

// SendMessageToUserByPlayer 通过 Player 实例发送消息
func (rc *ClientsContainer) SendMessageToUserByPlayer(data []byte, client *client.Client) {
	if client != nil && client.Session.IsConnected() {
		client.Write(data)
	}
}
