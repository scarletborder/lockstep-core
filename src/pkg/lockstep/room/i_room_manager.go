package room

// IRoomManager 定义房间管理器的接口
type IRoomManager interface {
	// GetRoom 获取指定 ID 的房间
	GetRoom(roomID uint32) (*Room, bool)

	// CreateRoom 创建一个新房间
	CreateRoom(name string, key string) (*Room, error)

	// RemoveRoom 删除一个房间
	RemoveRoom(roomID uint32)

	// ListRooms 列出所有房间 ID
	ListRooms() []uint32

	// GetRoomCount 获取房间数量
	GetRoomCount() int
}
