# Lockstep WebTransport 服务器快速入门

## 项目结构

```
lockstep-core/
├── src/
│   ├── constants/      # 游戏常量定义
│   │   ├── stages.go   # 游戏阶段
│   │   ├── frame.go    # 帧同步配置
│   │   ├── player.go   # 玩家状态
│   │   └── app.go      # 应用信息
│   ├── logic/          # 核心游戏逻辑
│   │   ├── player_context.go  # 玩家连接上下文 (WebTransport)
│   │   ├── player.go           # 玩家结构
│   │   ├── player_map.go       # 并发安全的玩家映射
│   │   ├── room.go             # 房间核心逻辑
│   │   ├── room_context.go     # 房间上下文
│   │   ├── room_loop.go        # 房间状态机循环
│   │   ├── room_manager.go     # 房间管理器
│   │   ├── session_handler.go  # 会话处理
│   │   └── interfaces.go       # 接口定义
│   ├── server/         # WebTransport 服务器
│   │   ├── webtransport_server.go  # 服务器封装
│   │   └── handlers.go             # HTTP/WebTransport 处理器
│   ├── config/         # 配置管理
│   └── types/          # 类型定义
├── old/                # 旧的 WebSocket 实现（保留作为参考）
└── proto/              # Protobuf 消息定义
```

## 核心概念

### 1. 游戏状态机

房间在整个生命周期中经历以下状态：

```
InLobby (0x20)    → 玩家加入，房主设置游戏
    ↓
Preparing (0x21)  → 玩家选择装备并准备
    ↓
Loading (0x22)    → 所有玩家加载游戏资源
    ↓
InGame (0x23)     → 游戏进行中（50ms 定时器驱动）
    ↓
PostGame (0x24)   → 游戏结束，显示结算
    ↓
InLobby           → 重置，开始下一局
```

### 2. 帧同步机制

- **帧率**: 20 FPS (50ms 间隔)
- **延迟容忍**: 最多 10 帧
- **同步检查**: 所有玩家必须在容忍范围内
- **操作序号**: 每帧的操作有唯一序号，防止重复

### 3. 并发模型

```
              ┌─────────────┐
              │   Client    │
              └──────┬──────┘
                     │ WebTransport
                     ↓
            ┌────────────────┐
            │  JoinRoomHandler│
            └────────┬───────┘
                     │
              ┌──────▼──────┐
              │   Room      │
              │  (goroutine)│
              └─────┬───────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
   register    incomingMsg   unregister
     chan         chan          chan
        │           │           │
        └───────────┴───────────┘
                    │
            ┌───────▼────────┐
            │ Room.Run()     │
            │ State Machine  │
            └────────────────┘
```

## 快速开始

### 1. 安装依赖

```bash
cd /Users/songrujia/codes/lockstep-core
go mod tidy
```

### 2. 生成 TLS 证书（WebTransport 需要）

```bash
go run src/utils/tls/generate.go
```

### 3. 运行服务器

```bash
go run main.go
```

服务器将在 `https://localhost:4433` 启动。

### 4. API 端点

#### 获取房间列表
```http
GET /rooms
```

#### 创建房间
```http
POST /rooms
Content-Type: application/json

{
  "room_id": "room123"
}
```

#### 加入房间（WebTransport）
```
WebTransport URL: https://localhost:4433/join/{roomID}
```

## 客户端连接示例

### JavaScript (浏览器)

```javascript
// 连接到房间
const url = 'https://localhost:4433/join/room123';
const transport = new WebTransport(url);

await transport.ready;
console.log('Connected to room!');

// 发送消息（使用 Datagram）
const writer = transport.datagrams.writable.getWriter();
const data = new TextEncoder().encode('Hello, server!');
await writer.write(data);

// 接收消息
const reader = transport.datagrams.readable.getReader();
while (true) {
  const { value, done } = await reader.read();
  if (done) break;
  const message = new TextDecoder().decode(value);
  console.log('Received:', message);
}
```

## 关键类和方法

### Room
```go
// 主要方法
room.Run()                    // 启动状态机循环
room.AddPlayer(player)        // 添加玩家
room.UpdateActiveTime()       // 更新活跃时间
room.BroadcastMessage(msg)    // 广播消息
room.Destroy()                // 销毁房间
```

### RoomManager
```go
// 主要方法
rm.GetOrCreateRoom(roomID)    // 获取或创建房间
rm.ListRooms()                // 列出所有房间
rm.RemoveRoom(roomID)         // 移除房间
```

### Player
```go
// 主要属性
player.Ctx                    // PlayerContext (连接)
player.IsReady                // 是否准备
player.IsLoaded               // 是否加载完毕
player.Write(data)            // 发送消息
```

## 消息流程

### 1. 玩家加入房间

```
Client                Server
  │                     │
  │  WebTransport       │
  │  Upgrade Request    │
  ├──────────────────>  │
  │                     │
  │  Connection OK      │
  │  <───────────────── │
  │                     │
  │                     ├─> CreatePlayer
  │                     │
  │                     ├─> room.register <- player
  │                     │
  │                     ├─> Room.Run() handles register
  │                     │
  │  JoinSuccess        │
  │  <───────────────── │
  │                     │
  │                     ├─> Start goroutine
  │                     │   room.StartServeClient(player)
```

### 2. 游戏循环 (InGame 状态)

```
                  Timer (50ms)
                       │
                       ↓
            Room.runGameTick()
                       │
        ┌──────────────┼──────────────┐
        │              │              │
    Check Sync    Read Operations  Broadcast
        │              │              │
        ↓              ↓              ↓
  All players   Game logic     Send to all
  synced?       processes      players
```

## 配置

### config/config.go
```go
type ServerConfig struct {
    Addr              string        // ":4433"
    TLSConfig         *tls.Config
    CheckOriginEnabled bool
}
```

## 调试

### 日志标记
- 🟢 成功操作
- 🔵 信息日志
- 🟡 警告/待处理
- 🔴 错误
- 🔥 房间销毁/重要事件
- 🕒 时间更新
- 🎮 游戏逻辑

### 常见问题

1. **证书错误**: 确保生成了有效的 TLS 证书
2. **连接拒绝**: 检查防火墙设置
3. **玩家无法同步**: 检查网络延迟和 MaxDelayFrames 设置

## 性能调优

### 帧率调整
```go
// src/constants/frame.go
const FrameIntervalMs = 50  // 降低以提高帧率
```

### 延迟容忍
```go
// src/constants/frame.go
const MaxDelayFrames uint32 = 10  // 增加以容忍更高延迟
```

### Channel 缓冲
```go
// src/logic/room.go
register:         make(chan *Player, 8),      // 增加缓冲
incomingMessages: make(chan *PlayerMessage, 128),
```

## 下一步开发

1. **实现 Protobuf 消息**
   - 定义完整的游戏消息协议
   - 实现序列化/反序列化

2. **完善游戏逻辑**
   - 植物种植
   - 资源管理
   - 战斗逻辑

3. **添加持久化**
   - 玩家数据存储
   - 游戏记录
   - 排行榜

4. **监控和日志**
   - Prometheus 指标
   - 结构化日志
   - 性能追踪

## 参考资源

- [WebTransport 规范](https://www.w3.org/TR/webtransport/)
- [QUIC 协议](https://www.chromium.org/quic)
- [quic-go 文档](https://github.com/quic-go/quic-go)
- [帧同步游戏原理](https://www.gabrielgambetta.com/client-server-game-architecture.html)
