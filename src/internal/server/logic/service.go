package logic

import (
	"crypto/sha256"
	"fmt"
	"lockstep-core/src/messages"
	"lockstep-core/src/pkg/lockstep/room"
)

// RoomService 包含所有房间相关的业务逻辑
type RoomService struct {
	roomManager room.IRoomManager
}

// NewRoomService 创建一个新的 RoomService 实例
func NewRoomService(roomManager room.IRoomManager) *RoomService {
	return &RoomService{
		roomManager: roomManager,
	}
}

// ListRooms 获取所有房间列表
// 输入: 无
// 输出: ListRoomsResponse proto 消息
func (s *RoomService) ListRooms() *messages.ListRoomsResponse {
	roomIDs := s.roomManager.ListRooms()
	return &messages.ListRoomsResponse{
		Rooms: roomIDs,
	}
}

// CreateRoom 创建一个新房间
// 输入: CreateRoomRequest proto 消息
// 输出: CreateRoomResponse proto 消息 或 ErrorResponse
func (s *RoomService) CreateRoom(req *messages.CreateRoomRequest) (*messages.CreateRoomResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	createdRoom, err := s.roomManager.CreateRoom(req.Name, req.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	return &messages.CreateRoomResponse{
		RoomId: createdRoom.ID,
	}, nil
}

// HealthCheck 执行健康检查
// 输入: TLS 证书的 DER 编码数据
// 输出: HealthCheckResponse proto 消息
func (s *RoomService) HealthCheck(tlsCertDER []byte) *messages.HealthCheckResponse {
	hash := sha256.Sum256(tlsCertDER)
	// 将字节数组转换为 []uint32
	hashUint32 := make([]uint32, len(hash))
	for i, b := range hash {
		hashUint32[i] = uint32(b)
	}

	return &messages.HealthCheckResponse{
		Status:  "ok",
		Message: "Lockstep server core is running",
		Hash:    hashUint32,
	}
}
