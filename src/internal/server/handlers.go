package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"lockstep-core/src/pkg/lockstep/client"
	"lockstep-core/src/pkg/lockstep/room"
	"lockstep-core/src/pkg/lockstep/session"
	customTLS "lockstep-core/src/utils/tls"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

// HTTPHandlers 包含所有 HTTP 请求处理器
type HTTPHandlers struct {
	roomManager room.IRoomManager
	wtServer    *WebTransportServer
}

// NewHTTPHandlers 创建一个新的 HTTPHandlers 实例
func NewHTTPHandlers(
	roomManager room.IRoomManager,
	wtServer *WebTransportServer,
) *HTTPHandlers {
	return &HTTPHandlers{
		roomManager: roomManager,
		wtServer:    wtServer,
	}
}

// ListRoomsHandler 处理获取房间列表的请求 (GET /rooms)
func (h *HTTPHandlers) ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
	roomIDs := h.roomManager.ListRooms()
	resp := struct {
		Rooms []uint32 `json:"rooms"`
	}{
		Rooms: roomIDs,
	}

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
func (h *HTTPHandlers) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Name string `json:"name"`
		Key  string `json:"key"`
	}

	room, err := h.roomManager.CreateRoom(reqBody.Name, reqBody.Key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp := struct {
			Error string `json:"error"`
		}{
			Error: fmt.Sprintf("Failed to create room: %v", err),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Printf("Room created: %s", room.ID)

	resp := struct {
		RoomID uint32 `json:"room_id"`
	}{
		RoomID: room.ID,
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// RoomsHandler 统一处理房间相关请求
func (h *HTTPHandlers) RoomsHandler(w http.ResponseWriter, r *http.Request) {
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

// JoinRoomHandler 处理加入房间的请求 (WebTransport /join?roomid={roomID}&name={name}&&key={value})
func (h *HTTPHandlers) JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("JoinRoomHandler called with path: %s", r.URL.Path)
	queryParams := r.URL.Query()
	roomID := queryParams.Get("roomid")
	if roomID == "" {
		http.Error(w, "Missing roomID parameter", http.StatusBadRequest)
		return
	}

	// str to uint32
	roomIDNum, err := strconv.ParseUint(roomID, 10, 32)
	if err != nil {
		http.Error(w, "Invalid roomID parameter", http.StatusBadRequest)
		return
	}
	if (roomIDNum > 0xFFFFFFFF) || (roomIDNum == 0) {
		http.Error(w, "roomID out of range", http.StatusBadRequest)
		return
	}
	//other params
	name := queryParams.Get("name")
	if name == "" {
		name = fmt.Sprintf("room_%d", roomIDNum)
	}
	key := queryParams.Get("key") // 可选密钥参数
	_ = key                       // TODO:目前未使用密钥

	room, ok := h.roomManager.GetRoom(uint32(roomIDNum))
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		resp := struct {
			Error string `json:"error"`
		}{
			Error: fmt.Sprintf("Room %s not found", roomID),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if room.IsRoomFull() {
		w.WriteHeader(http.StatusForbidden)
		resp := struct {
			Error string `json:"error"`
		}{
			Error: fmt.Sprintf("Room %s is full", roomID),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	nextUserId, err := room.GetNextUserID()
	if err != nil {
		resp := struct {
			Error string `json:"error"`
		}{
			Error: fmt.Sprintf("Failed to get next user ID: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 升级连接到 WebTransport
	sess, err := h.wtServer.UpgradeToWebTransport(w, r)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebTransport", http.StatusInternalServerError)
		return
	}

	// 为玩家生成唯一 ID（使用数字 ID）
	playerIDNum := int(uuid.New().ID()) // 使用 UUID 的数字部分

	// 创建玩家上下文
	session_impl := session.NewWtSession(sess)

	client := client.NewClient(nextUserId, session_impl, room.GetIncomingMessagesChan())

	// playerCtx := logic.NewPlayerContext(session, playerIDNum)

	// 创建玩家实例
	// player := logic.NewPlayer(playerCtx, room.GetIncomingMessagesChan())
	//
	log.Printf("Player %d joining room %s", playerIDNum, roomID)

	// TODO: 发送加入成功消息给客户端
	// joinMsg := []byte(fmt.Sprintf("Player %d joined room %s", playerIDNum, room.ID))

	// 将玩家添加到房间（这会发送到 register channel）
	room.RegisterPlayer(client)

	// 在独立的 goroutine 中处理此玩家的会话
	// 这会阻塞直到连接关闭
	go room.StartServeClient(client)
}

// HealthCheckHandler 处理健康检查和根路径请求 (GET /)
func (h *HTTPHandlers) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.WriteHeader(http.StatusOK)

	hash := sha256.Sum256((h.wtServer.config.TLSConfig.Certificates[0].Leaf.Raw))

	response := map[string]interface{}{
		"status":  "ok",
		"message": "WebTransport server is running",
		"hash":    customTLS.FormatByteSlice(hash[:]),
		"endpoints": map[string]string{
			"health":      "GET /",
			"list_rooms":  "GET /rooms",
			"create_room": "POST /rooms",
			"join_room":   "WebTransport /join?roomid={roomID}&key={value}",
		},
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// RegisterHandlers 注册所有 HTTP 处理器
func (h *HTTPHandlers) RegisterHandlers() {
	h.wtServer.RegisterHandler("/", h.HealthCheckHandler)
	h.wtServer.RegisterHandler("/rooms", h.RoomsHandler)
	h.wtServer.RegisterHandler("/join", h.JoinRoomHandler)
}

// Start 启动服务器
func (h *HTTPHandlers) Start() error {
	return h.wtServer.Start()
}
