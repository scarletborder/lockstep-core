package server

import (
	"encoding/json"
	"fmt"
	"lockstep-core/src/logic"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// HTTPHandlers 包含所有 HTTP 请求处理器
type HTTPHandlers struct {
	roomManager    logic.RoomManagerInterface
	sessionHandler logic.PlayerSessionHandler
	wtServer       *WebTransportServer
}

// NewHTTPHandlers 创建一个新的 HTTPHandlers 实例
func NewHTTPHandlers(
	roomManager logic.RoomManagerInterface,
	sessionHandler logic.PlayerSessionHandler,
	wtServer *WebTransportServer,
) *HTTPHandlers {
	return &HTTPHandlers{
		roomManager:    roomManager,
		sessionHandler: sessionHandler,
		wtServer:       wtServer,
	}
}

// ListRoomsHandler 处理获取房间列表的请求 (GET /rooms)
func (h *HTTPHandlers) ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
	roomIDs := h.roomManager.ListRooms()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(roomIDs); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateRoomHandler 处理创建房间的请求 (POST /rooms)
func (h *HTTPHandlers) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		RoomID string `json:"room_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || reqBody.RoomID == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.roomManager.GetOrCreateRoom(reqBody.RoomID)
	log.Printf("Room created: %s", reqBody.RoomID)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Room %s created successfully", reqBody.RoomID)
}

// RoomsHandler 统一处理房间相关请求
func (h *HTTPHandlers) RoomsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListRoomsHandler(w, r)
	case http.MethodPost:
		h.CreateRoomHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// JoinRoomHandler 处理加入房间的请求 (WebTransport /join/{roomID})
func (h *HTTPHandlers) JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimPrefix(r.URL.Path, "/join/")
	if roomID == "" {
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	room := h.roomManager.GetOrCreateRoom(roomID)

	// 升级连接到 WebTransport
	session, err := h.wtServer.UpgradeToWebTransport(w, r)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebTransport", http.StatusInternalServerError)
		return
	}

	// 为玩家生成唯一 ID（使用数字 ID）
	playerIDNum := int(uuid.New().ID()) // 使用 UUID 的数字部分

	// 创建玩家上下文
	playerCtx := logic.NewPlayerContext(session, playerIDNum)

	// 创建玩家实例
	player := logic.NewPlayer(playerCtx, room.GetIncomingMessagesChan())

	log.Printf("Player %d joining room %s", playerIDNum, roomID)

	// TODO: 发送加入成功消息给客户端
	// joinMsg := []byte(fmt.Sprintf("Player %d joined room %s", playerIDNum, room.ID))

	// 将玩家添加到房间（这会发送到 register channel）
	room.AddPlayer(player)

	// 在独立的 goroutine 中处理此玩家的会话
	// 这会阻塞直到连接关闭
	go h.sessionHandler.HandleSession(room, player)
}

// RegisterHandlers 注册所有 HTTP 处理器
func (h *HTTPHandlers) RegisterHandlers() {
	h.wtServer.RegisterHandler("/rooms", h.RoomsHandler)
	h.wtServer.RegisterHandler("/join/", h.JoinRoomHandler)
}

// Start 启动服务器
func (h *HTTPHandlers) Start() error {
	return h.wtServer.Start()
}
