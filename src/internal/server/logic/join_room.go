package logic

import (
	"fmt"
	"lockstep-core/src/pkg/lockstep/client"
	"lockstep-core/src/pkg/lockstep/room"
	"lockstep-core/src/pkg/lockstep/session"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// JoinRoomRequest 包含加入房间所需的信息
type JoinRoomRequest struct {
	RoomID uint32
	Key    string // 可选密钥参数
}

// JoinRoomResponse 包含加入房间的结果
type JoinRoomResponse struct {
	PlayerID   int
	UserID     uint32
	RoomID     uint32
	ClientInfo *client.Client
}

// ValidateJoinRoom 验证玩家是否可以加入房间
// 检查房间是否存在、是否满员、鉴权等
func ValidateJoinRoom(
	roomManager room.IRoomManager,
	req *JoinRoomRequest,
) (*room.Room, uint16, error) {
	if req == nil {
		return nil, http.StatusBadRequest, fmt.Errorf("request cannot be nil")
	}

	if req.RoomID == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("roomID cannot be 0")
	}

	// 检查房间是否存在
	r, ok := roomManager.GetRoom(req.RoomID)
	if !ok {
		return nil, http.StatusNotFound, fmt.Errorf("room %d not found", req.RoomID)
	}

	// 检查房间是否满员
	if r.IsRoomFull() {
		return nil, http.StatusConflict, fmt.Errorf("room %d is full", req.RoomID)
	}

	// 鉴权检查
	// 根据 req.Key 验证密钥是否正确
	// 如果 Key 不匹配，返回错误
	if r.HasKey() {
		if !r.CheckKeyCorrect(req.Key) {
			return nil, http.StatusUnauthorized, fmt.Errorf("invalid key for room %d", req.RoomID)
		}
	}

	return r, http.StatusOK, nil
}

// JoinRoom 是一个泛型方法，处理玩家加入房间的通用逻辑
// 支持任何实现了 session.ISession 接口的会话类型
func JoinRoom(
	r *room.Room,
	sessionImpl session.ISession,
) (*JoinRoomResponse, error) {
	// 获取下一个用户 ID
	nextUserId, err := r.GetNextUserID()
	if err != nil {
		return nil, fmt.Errorf("failed to get next user ID: %w", err)
	}

	// 为玩家生成唯一 ID（使用 UUID 的数字部分）
	playerIDNum := int(uuid.New().ID())

	// 创建客户端
	playerClient := client.NewClient(nextUserId, sessionImpl, r.GetIncomingMessagesChan())

	log.Printf("Player %d joining room %d (user ID: %d)", playerIDNum, r.ID, nextUserId)

	// 将玩家添加到房间（这会发送到 register channel）
	r.RegisterPlayer(playerClient)

	// 在独立的 goroutine 中处理此玩家的会话
	// 这会阻塞直到连接关闭
	go r.StartServeClient(playerClient)

	return &JoinRoomResponse{
		PlayerID:   playerIDNum,
		UserID:     nextUserId,
		RoomID:     r.ID,
		ClientInfo: playerClient,
	}, nil
}
