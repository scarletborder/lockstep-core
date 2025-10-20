package logic

import (
	"fmt"
	"lockstep-core/src/pkg/lockstep/client"
	"lockstep-core/src/pkg/lockstep/room"
	"lockstep-core/src/pkg/lockstep/session"
	"log"
	"net/http"
)

// JoinRoomRequest 包含加入房间所需的信息
type JoinRoomRequest struct {
	RoomID         uint32
	Key            string // 可选密钥参数
	ReconnectToken string // 可选重连令牌
}

// JoinRoomResponse 包含加入房间的结果
type JoinRoomResponse struct {
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
	var isReconnect bool = false

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

	// 重连请求
	if req.ReconnectToken != "" {
		parsed, ok := r.JwtService.ParseToken(req.ReconnectToken)
		if !ok {
			return nil, http.StatusUnauthorized, fmt.Errorf("invalid reconnect token for room %d", req.RoomID)
		} else if parsed.RoomID != req.RoomID {
			return nil, http.StatusUnauthorized, fmt.Errorf("reconnect token roomID mismatch for room %d", req.RoomID)
		} else {
			user, _ := r.ClientsContainer.Clients.Load(parsed.UserID)
			// 如果没有找到用户，说明unregister了，不需要在意
			if user.Session.IsConnected() {
				return nil, http.StatusConflict, fmt.Errorf("user %d already connected in room %d", parsed.UserID, req.RoomID)
			} else {
				isReconnect = true
			}
		}
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

	// 游戏世界特殊设置
	if r.Game != nil {
		if !r.Game.CouldJoinRoom(isReconnect) {
			return nil, http.StatusForbidden, fmt.Errorf("cannot join room %d due to game rules", req.RoomID)
		}
	}

	return r, http.StatusOK, nil
}

// JoinRoom 是一个泛型方法，处理玩家加入房间的通用逻辑
// 支持任何实现了 session.ISession 接口的会话类型
func JoinRoom(
	r *room.Room,
	sessionImpl session.ISession,
	reconnectToken string,
) (*JoinRoomResponse, error) {
	var nextUserId uint32
	var err error
	var isReconnect bool = false
	if reconnectToken != "" {
		parse, _ := r.JwtService.ParseToken(reconnectToken)
		nextUserId = parse.UserID
		isReconnect = true
	} else {
		// 获取下一个用户 ID
		nextUserId, err = r.GetNextUserID()
		if err != nil {
			return nil, fmt.Errorf("failed to get next user ID: %w", err)
		}
	}

	// 创建客户端
	playerClient := client.NewClient(nextUserId, sessionImpl, r.GetIncomingMessagesChan())
	playerClient.IsReconnected = isReconnect

	log.Printf("Player joining room %d (user ID: %d)", r.ID, nextUserId)

	// 将玩家添加到房间（这会发送到 register channel）
	r.RegisterPlayer(playerClient)

	// 在独立的 goroutine 中处理此玩家的会话
	// 这会阻塞直到连接关闭
	go r.StartServeClient(playerClient)

	return &JoinRoomResponse{
		UserID:     nextUserId,
		RoomID:     r.ID,
		ClientInfo: playerClient,
	}, nil
}
