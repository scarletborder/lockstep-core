package session

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/gorilla/websocket"
)

// WebsocketSession 实现了与 WtSession 兼容的接口，用于 WebSocket 连接。
type WebsocketSession struct {
	ctx    context.Context
	cancel context.CancelFunc
	// gorilla/websocket 连接
	conn  *websocket.Conn
	mutex sync.Mutex // 用于保护对 conn 的并发访问
}

// NewWebsocketSession 创建一个新的 WebsocketSession 实例。
func NewWebsocketSession(conn *websocket.Conn) *WebsocketSession {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebsocketSession{
		conn:   conn,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Close 关闭 WebSocket 会话和连接。
func (ws *WebsocketSession) Close() error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if ws.cancel != nil {
		ws.cancel()
		ws.cancel = nil
	}

	if ws.conn != nil {
		// 根据 WebSocket 协议发送关闭帧
		// 1000 表示正常关闭
		err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, "player disconnected"))
		if err != nil {
			// 如果写入关闭消息失败，直接关闭底层连接
			return ws.conn.Close()
		}
		return ws.conn.Close()
	}
	return nil
}

// IsConnected 检查 WebSocket 连接是否仍然活跃。
func (ws *WebsocketSession) IsConnected() bool {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	// 检查 conn 是否为 nil，并检查会话的 context 是否已被取消。
	return ws.conn != nil && ws.ctx.Err() == nil
}

// SendDatagram 通过 WebSocket 连接发送数据。
// WebSocket 没有原生的 "datagram"，所以我们使用二进制消息 (BinaryMessage)。
func (ws *WebsocketSession) SendDatagram(data []byte) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if !ws.IsConnected() {
		return fmt.Errorf("session is closed or nil")
	}

	// 使用 WriteMessage 发送二进制数据
	err := ws.conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		log.Printf("🔴 SendDatagram (WebSocket) error for player %v: %v", ws.conn.RemoteAddr(), err)
	}
	return err
}

// ReceiveDatagram 从 WebSocket 连接接收数据。
// 同样，我们读取二进制消息作为 "datagram"。
func (ws *WebsocketSession) ReceiveDatagram() ([]byte, error) {
	// 注意：gorilla/websocket 的 ReadMessage 会阻塞，直到有消息、发生错误或连接关闭。
	// 它的行为受到我们创建的 context 的控制。
	// 当 context 被取消时，并发的读取操作应该会出错。

	if !ws.IsConnected() {
		return nil, fmt.Errorf("session is closed or nil")
	}

	// ReadMessage 是线程安全的，所以这里不需要锁
	_, data, err := ws.conn.ReadMessage()
	if err != nil {
		// 检查错误类型，如果是预期的关闭，则不打印为错误日志
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
			log.Printf("🔴 ReceiveDatagram (WebSocket) error: %v", err)
		}
		return nil, err
	}

	return data, nil
}

// RemoteAddr 返回客户端的网络地址
func (ws *WebsocketSession) GetRemoteAddr() net.Addr {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	if ws.conn == nil {
		return nil
	}
	return ws.conn.RemoteAddr()
}

func (ws *WebsocketSession) CloseWithError(code uint32, reason string) error {
	if !ws.IsConnected() {
		return fmt.Errorf("session is nil")
	}
	return ws.conn.CloseHandler()(int(code), reason)
}
