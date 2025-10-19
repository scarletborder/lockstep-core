package server

import (
	"encoding/json"
	"fmt"
	"lockstep-core/src/internal/server/logic"
	"lockstep-core/src/messages"
	"lockstep-core/src/pkg/lockstep/room"
	"lockstep-core/src/pkg/lockstep/session"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

// Serverandlers 包含所有请求处理器和其依赖
type Serverandlers struct {
	roomManager room.IRoomManager
	wtServer    *ServerCore
	roomService *logic.RoomService
}

// NewHTTPHandlers 创建一个新的 HTTPHandlers 实例
func NewHTTPHandlers(
	roomManager room.IRoomManager,
	wtServer *ServerCore,
) *Serverandlers {
	roomService := logic.NewRoomService(roomManager)
	return &Serverandlers{
		roomManager: roomManager,
		wtServer:    wtServer,
		roomService: roomService,
	}
}

// ListRoomsHandler 处理获取房间列表的请求 (GET /rooms)
func (h *Serverandlers) ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
	resp := h.roomService.ListRooms()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateRoomHandler 处理创建房间的请求 (POST /rooms)
func (h *Serverandlers) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody messages.CreateRoomRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errResp := &messages.ErrorResponse{
			Error: fmt.Sprintf("Invalid request body: %v", err),
		}
		json.NewEncoder(w).Encode(errResp)
		return
	}

	resp, err := h.roomService.CreateRoom(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errResp := &messages.ErrorResponse{
			Error: fmt.Sprintf("Failed to create room: %v", err),
		}
		json.NewEncoder(w).Encode(errResp)
		return
	}

	log.Printf("Room created: %v", resp.RoomId)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// RoomsHandler 统一处理房间相关请求
func (h *Serverandlers) RoomsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.ListRoomsHandler(w, r)
	case http.MethodPost:
		h.CreateRoomHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// JoinRoomHandler 处理加入房间的请求 (WebTransport /join?roomid={roomID}&key={value}&wt={true|false})
func (h *Serverandlers) JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	log.Printf("JoinRoomHandler called with path: %s", r.URL.Path)
	queryParams := r.URL.Query()
	roomID := queryParams.Get("roomid")
	if roomID == "" {
		errResp := &messages.ErrorResponse{
			Error: "Missing roomID parameter",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	// str to uint32
	roomIDNum, err := strconv.ParseUint(roomID, 10, 32)
	if err != nil {
		errResp := &messages.ErrorResponse{
			Error: "Invalid roomID parameter",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}
	if (roomIDNum > 0xFFFFFFFF) || (roomIDNum == 0) {
		errResp := &messages.ErrorResponse{
			Error: "roomID out of range",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	// is wt
	isWebTransport := queryParams.Get("wt") == "true"

	// 获取可选密钥参数
	key := queryParams.Get("key") // 可选密钥参数

	joinReq := &logic.JoinRoomRequest{
		RoomID: uint32(roomIDNum),
		Key:    key,
	}

	// validate first
	room, status, err := logic.ValidateJoinRoom(h.roomManager, joinReq)
	if err != nil {
		errResp := &messages.ErrorResponse{
			Error: fmt.Sprintf("Failed to validate join room: %v", err),
		}
		w.WriteHeader(int(status))
		json.NewEncoder(w).Encode(errResp)
		return
	}

	var session_impl session.ISession

	if isWebTransport {
		// 升级连接到 WebTransport
		sess, err := h.wtServer.UpgradeToWebTransport(w, r)
		if err != nil {
			errResp := &messages.ErrorResponse{
				Error: fmt.Sprintf("Failed to upgrade to WebTransport: %v", err),
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(errResp)
			return
		}
		// 创建 WebTransport 会话实现
		session_impl = session.NewWtSession(sess)
	} else {
		// 升级连接到 WebSocket
		upgrader := websocket.Upgrader{
			CheckOrigin: h.wtServer.GetWTServer().CheckOrigin,
		}
		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			errResp := &messages.ErrorResponse{
				Error: fmt.Sprintf("Failed to upgrade to WebSocket: %v", err),
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(errResp)
			return
		}
		// 创建 WebSocket 会话实现
		session_impl = session.NewWebsocketSession(wsConn)
	}

	resp, err := logic.JoinRoom(room, session_impl)
	if err != nil {
		session_impl.Close()
		errResp := &messages.ErrorResponse{
			Error: fmt.Sprintf("Failed to join room: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errResp)
		return
	}

	log.Printf("Player %d successfully joined room %d", resp.PlayerID, resp.RoomID)
}

// HealthCheckHandler 处理健康检查和根路径请求 (GET /)
func (h *Serverandlers) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)

	response := h.roomService.HealthCheck(h.wtServer.config.TLSConfig.Certificates[0].Leaf.Raw)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// RegisterHandlers 注册所有 HTTP 处理器
func (h *Serverandlers) RegisterHandlers() {
	h.wtServer.RegisterHandler("/", h.HealthCheckHandler)
	h.wtServer.RegisterHandler("/rooms", h.RoomsHandler)
	h.wtServer.RegisterHandler("/join", h.JoinRoomHandler)
}

// Start 启动服务器
func (h *Serverandlers) Start() error {
	return h.wtServer.Start()
}
