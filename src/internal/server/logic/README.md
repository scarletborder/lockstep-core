# Logic Package

此包包含所有 HTTP 路由处理函数的业务逻辑实现。

## 设计原则

- **接收 Proto 消息作为输入**: 所有逻辑函数接收 protobuf 定义的请求消息
- **返回 Proto 消息作为输出**: 所有逻辑函数返回 protobuf 定义的响应消息
- **关注业务逻辑**: 与 HTTP/WebTransport 协议细节分离
- **易于测试**: 由于使用标准化的数据结构，逻辑函数易于单元测试
- **泛型支持多协议**: 使用 Go 泛型处理不同的 Session 实现

## 文件说明

### service.go

`RoomService` 结构体包含所有房间相关的业务逻辑方法：

- `ListRooms()` - 获取所有房间列表
  - 输入: 无
  - 输出: `*messages.ListRoomsResponse`

- `CreateRoom(req *messages.CreateRoomRequest)` - 创建新房间
  - 输入: `*messages.CreateRoomRequest`
  - 输出: `*messages.CreateRoomResponse` 或 error

- `HealthCheck(tlsCertDER []byte)` - 执行健康检查
  - 输入: TLS 证书的 DER 编码字节
  - 输出: `*messages.HealthCheckResponse`

### join_room.go

包含加入房间的通用泛型逻辑，支持任何实现 `session.ISession` 接口的会话类型。

#### 类型定义

```go
type JoinRoomRequest struct {
    RoomID uint32
    Key    string // 可选密钥参数
}

type JoinRoomResponse struct {
    PlayerID   int
    UserID     uint32
    RoomID     uint32
    ClientInfo *client.Client
}
```

#### 泛型函数

```go
// JoinRoom 是一个泛型方法，处理玩家加入房间的通用逻辑
// 支持任何实现了 session.ISession 接口的会话类型
func JoinRoom[S session.ISession](
    roomManager room.IRoomManager,
    req *JoinRoomRequest,
    sessionImpl S,
) (*JoinRoomResponse, error)
```

**特点**：
- 使用 Go 1.18+ 的泛型约束
- 支持 `*session.WtSession`（WebTransport）
- 支持 `*session.WebsocketSession`（WebSocket）
- 零开销抽象，编译时类型检查

## 使用方式

在 `handlers.go` 中，HTTP 处理器调用相应的方法：

### RoomService 方法（非泛型）

```go
func (h *Serverandlers) ListRoomsHandler(w http.ResponseWriter, r *http.Request) {
    resp := h.roomService.ListRooms()
    // 处理 HTTP 细节（header、状态码等）
    json.NewEncoder(w).Encode(resp)
}
```

### JoinRoom 泛型函数（WebTransport）

```go
func (h *Serverandlers) JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
    // 升级到 WebTransport
    sess, err := h.wtServer.UpgradeToWebTransport(w, r)
    
    // 创建 WtSession
    session_impl := session.NewWtSession(sess)
    
    // 调用泛型加入房间逻辑
    resp, err := logic.JoinRoom(h.roomManager, joinReq, session_impl)
}
```

### JoinRoom 泛型函数（WebSocket）

```go
func (h *Serverandlers) LegacyJoinRoomHandler(w http.ResponseWriter, r *http.Request) {
    // 升级到 WebSocket
    wsConn, err := upgrader.Upgrade(w, r, nil)
    
    // 创建 WebsocketSession
    session_impl := session.NewWebsocketSession(wsConn)
    
    // 调用泛型加入房间逻辑
    resp, err := logic.JoinRoom(h.roomManager, joinReq, session_impl)
}
```

## 架构优势

1. **关注点分离**
   - handlers.go: 处理 HTTP 协议细节、CORS、错误处理等
   - service.go: 房间管理业务逻辑
   - join_room.go: 加入房间业务逻辑（支持多协议）

2. **代码复用**
   - WebTransport 和 WebSocket 共用完全相同的加入逻辑
   - 无需重复编写验证、错误处理代码

3. **易于测试**
   - 业务逻辑与 HTTP 完全分离
   - 可以直接测试 `logic.JoinRoom()` 而无需 HTTP

4. **易于扩展**
   - 添加新的连接协议时，只需创建新的 Session 实现
   - 无需修改 join_room.go 代码

## 泛型类型约束

所有用作泛型参数 `S` 的类型必须实现 `session.ISession` 接口：

```go
type ISession interface {
    Close() error
    CloseWithError(code uint32, reason string) error
    IsConnected() bool
    SendDatagram(data []byte) error
    ReceiveDatagram() ([]byte, error)
    GetRemoteAddr() net.Addr
}
```

编译时会自动检查类型是否满足约束。

