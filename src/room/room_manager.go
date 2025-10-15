package room

import (
	"log"
	"sync"
)

// RoomManager 管理所有游戏房间
type RoomManager struct {
	rooms    map[string]*Room
	mutex    sync.RWMutex
	stopChan chan string // 用于接收房间的停止信号
}

// NewRoomManager 创建一个新的 RoomManager 实例
func NewRoomManager() *RoomManager {
	rm := &RoomManager{
		rooms:    make(map[string]*Room),
		stopChan: make(chan string, 100), // 缓冲通道
	}

	// 启动监听房间停止信号的 goroutine
	go rm.listenStopSignals()

	return rm
}

// listenStopSignals 监听房间的停止信号
func (rm *RoomManager) listenStopSignals() {
	for roomID := range rm.stopChan {
		log.Printf("🔥 Received stop signal for room %s", roomID)
		rm.RemoveRoom(roomID)
	}
}

// GetRoom 获取指定 ID 的房间
func (rm *RoomManager) GetRoom(roomID string) (*Room, bool) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	room, exists := rm.rooms[roomID]
	return room, exists
}

// CreateRoom 创建一个新房间
func (rm *RoomManager) CreateRoom(roomID string) *Room {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// 如果房间已存在，直接返回
	if room, exists := rm.rooms[roomID]; exists {
		return room
	}

	room := NewRoom(roomID, rm.stopChan)
	rm.rooms[roomID] = room

	// 启动房间的状态机循环
	go room.Run()
	log.Printf("🟢 Room %s created and started", roomID)

	return room
}

// GetOrCreateRoom 获取或创建一个房间
func (rm *RoomManager) GetOrCreateRoom(roomID string) *Room {
	// 先尝试读锁获取
	rm.mutex.RLock()
	room, exists := rm.rooms[roomID]
	rm.mutex.RUnlock()

	if exists {
		return room
	}

	// 不存在则创建
	return rm.CreateRoom(roomID)
}

// RemoveRoom 删除一个房间
func (rm *RoomManager) RemoveRoom(roomID string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if room, exists := rm.rooms[roomID]; exists {
		log.Printf("🔥 Removing room %s from manager", roomID)
		delete(rm.rooms, roomID)
		// 房间已经在 Destroy 中关闭了所有连接
		_ = room
	}
}

// ListRooms 列出所有房间 ID
func (rm *RoomManager) ListRooms() []string {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	roomIDs := make([]string, 0, len(rm.rooms))
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
