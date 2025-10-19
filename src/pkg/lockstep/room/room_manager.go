package room

import (
	"fmt"
	"lockstep-core/src/config"
	"lockstep-core/src/pkg/lockstep/world"
	"lockstep-core/src/utils"
	"log"
	"sync"
)

// RoomManager 管理所有游戏房间
type RoomManager struct {
	rooms map[uint32]*Room
	mutex sync.RWMutex

	utils.SafeIDAllocator
	// 传入roomid,用于接收房间的停止信号
	stopChan chan uint32

	// function to new game world
	NewGameWorld func(rctx world.IRoomContext) world.IGameWorld

	// cfg
	config.LockstepConfig
	config.ServerConfig
}

// NewRoomManager 创建一个新的 RoomManager 实例
func NewRoomManager(
	newFunc func(rctx world.IRoomContext) world.IGameWorld,
	cfg *config.RuntimeConfig,
) *RoomManager {
	rm := &RoomManager{
		rooms:           make(map[uint32]*Room),
		stopChan:        make(chan uint32, 100), // 缓冲通道
		NewGameWorld:    newFunc,
		LockstepConfig:  cfg.LockstepConfig,
		ServerConfig:    cfg.ServerConfig,
		SafeIDAllocator: *utils.NewSafeIDAllocator(utils.RoundUpTo64(uint32(*cfg.MaxRoomNumber))),
	}

	// 启动监听房间停止信号的 goroutine
	go rm.listenStopSignals()

	return rm
}

// listenStopSignals 监听房间的停止信号
func (rm *RoomManager) listenStopSignals() {
	for roomID := range rm.stopChan {
		log.Printf("🔥 Received stop signal for room %d", roomID)
		rm.RemoveRoom(roomID)
	}
}

// 获取一个可用的roomID

// GetRoom 获取指定 ID 的房间
func (rm *RoomManager) GetRoom(roomID uint32) (*Room, bool) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	room, exists := rm.rooms[roomID]
	return room, exists
}

// CreateRoom 创建一个新房间
func (rm *RoomManager) CreateRoom(name string, key string) (*Room, error) {
	if len(rm.rooms) >= int(*rm.ServerConfig.MaxRoomNumber) {
		return nil, fmt.Errorf("maximum number of rooms reached")
	}

	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	roomID, err := rm.SafeIDAllocator.Allocate()
	if err != nil {
		log.Printf("❌ Failed to allocate room ID: %v", err)
		return nil, err
	}
	// 如果房间已存在，直接返回
	if _, exists := rm.rooms[roomID]; exists {
		return nil, fmt.Errorf("room with ID %d already exists", roomID)
	}
	if name == "" {
		name = fmt.Sprintf("room_%d", roomID)
	}
	room := NewRoom(roomID, rm.stopChan, RoomOptions{
		name:           name,
		key:            key,
		LockstepConfig: rm.LockstepConfig,
	})
	rm.rooms[roomID] = room

	// 启动房间的状态机循环
	room_impl := NewRoomContextImpl(room)
	room.Game = rm.NewGameWorld(room_impl)

	go room.Run()
	log.Printf("🟢 Room %d created and started", roomID)

	return room, nil
}

// RemoveRoom 删除一个房间
func (rm *RoomManager) RemoveRoom(roomID uint32) {
	rm.SafeIDAllocator.Free(roomID)
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if room, exists := rm.rooms[roomID]; exists {
		log.Printf("🔥 Removing room %d from manager", roomID)
		delete(rm.rooms, roomID)
		// 房间已经在 Destroy 中关闭了所有连接
		_ = room
	}
}

// ListRooms 列出所有房间 ID
func (rm *RoomManager) ListRooms() []uint32 {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	roomIDs := make([]uint32, 0, len(rm.rooms))
	for id := range rm.rooms {
		roomIDs = append(roomIDs, id)
	}
	return roomIDs
}

// GetRoomCount 获取房间数量
func (rm *RoomManager) GetRoomCount() int {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	return len(rm.rooms)
}
