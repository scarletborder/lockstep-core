# Lockstep Client SDK 实现总结

## 概述

已完成一个功能完整的 TypeScript 客户端 SDK，用于与 Lockstep Go 服务器通信。该 SDK 支持：

1. **HTTP 请求**：获取房间列表、创建房间
2. **WebTransport 长连接**：使用 datagram API 进行实时游戏通信
3. **加入房间**：建立 WebTransport 连接并加入游戏房间
4. **断线重连**：使用密钥重新连接到房间

## 文件结构

```
example/ts-app/
├── src/
│   ├── core.ts              # 主客户端类 (LockstepClient)
│   ├── index.ts             # 导出入口
│   ├── requests/
│   │   ├── common.ts        # HTTP 请求实现 (HTTPClient)
│   │   └── stream.ts        # WebTransport 实现 (StreamClient)
│   └── types/
│       ├── index.ts         # 类型统一导出
│       └── pb/              # Protobuf 类型定义
│           ├── request.ts   # 请求消息类型
│           └── response.ts  # 响应消息类型
├── test/
│   └── main.ts              # 完整的测试套件
├── example.ts               # 简单使用示例
├── USAGE.md                 # 详细使用文档
├── README.md                # 项目说明
└── package.json             # 包配置（已更新测试脚本）
```

## 核心实现

### 1. HTTPClient (`requests/common.ts`)

实现了两个 HTTP API：

- `listRooms()`: GET /rooms - 获取房间列表
- `createRoom(roomId)`: POST /rooms - 创建房间

```typescript
const httpClient = new HTTPClient('https://localhost:8080');
const rooms = await httpClient.listRooms();
await httpClient.createRoom('room-001');
```

### 2. StreamClient (`requests/stream.ts`)

实现了 WebTransport 长连接功能：

**核心特性：**
- 使用 datagram API（不可靠传输）
- 状态管理：DISCONNECTED → CONNECTING → LOBBY → CONNECTED
- 消息处理：区分 LobbyResponse 和 RoomResponse
- 自动消息读取循环

**关键方法：**
- `joinRoom(roomId)`: 连接到 /join/{roomId}
- `reconnectRoom(roomId, key)`: 使用密钥重连
- `sendRequest(request)`: 发送 protobuf 消息
- `setHandlers(handlers)`: 设置消息回调

**实现细节：**
```typescript
// 发送消息（使用 datagram）
const writer = transport.datagrams.writable.getWriter();
await writer.write(bytes);
writer.releaseLock();

// 接收消息（使用 datagram）
const reader = transport.datagrams.readable.getReader();
const { value, done } = await reader.read();
```

### 3. LockstepClient (`core.ts`)

统一封装 HTTPClient 和 StreamClient，提供一站式接口：

```typescript
const client = new LockstepClient({ serverUrl: 'https://localhost:8080' });

// HTTP 请求
await client.listRooms();
await client.createRoom('room-001');

// WebTransport 连接
client.setMessageHandlers({ ... });
await client.joinRoom('room-001');
await client.sendRequest(request);

// 状态查询
client.isConnected();
client.getMyPlayerId();
client.getReconnectKey();

// 断开连接
await client.disconnect();
```

## 消息流程

### 加入房间流程

```
Client                          Server
  |                               |
  |--- WebTransport Connect ----->|  /join/{roomId}
  |                               |
  |<--- WebTransport Ready -------|
  |                               |
  |    (读取循环开始)              |
  |                               |
  |<---- LobbyResponse ----------|  joinRoomSuccess
  |                               |
  |      (状态: CONNECTED)         |
  |                               |
  |---- Request (datagram) ------>|
  |                               |
  |<--- RoomResponse (datagram) --|
  |                               |
```

### 消息类型处理

SDK 根据连接状态自动区分消息类型：

- **LOBBY / RECONNECTING 状态**：期待 `LobbyResponse`
  - `joinRoomSuccess`: 加入成功，切换到 CONNECTED 状态
  - `joinRoomFailed`: 加入失败
  
- **CONNECTED 状态**：期待 `RoomResponse`
  - `roomInfo`: 房间信息
  - `chooseMap`: 选择地图
  - `updateReadyCount`: 准备人数更新
  - `allReady`: 所有人准备
  - `allLoaded`: 所有人加载完成
  - `error`: 错误消息
  - 等等...

## 使用示例

### 基本使用

```typescript
import { LockstepClient } from './src';

const client = new LockstepClient({ serverUrl: 'https://localhost:8080' });

// 设置回调
client.setMessageHandlers({
  onLobbyResponse: (response) => { /* 处理加入结果 */ },
  onRoomResponse: (response) => { /* 处理游戏消息 */ },
  onError: (error) => { /* 处理错误 */ },
  onStateChange: (state) => { /* 监听状态变化 */ },
});

// 创建并加入房间
await client.createAndJoinRoom('my-room');

// 发送游戏消息
await client.sendRequest({
  payload: {
    oneofKind: 'ready',
    ready: { isReady: true },
  },
});
```

### 断线重连

```typescript
// 保存重连信息
const roomId = client.getCurrentRoomId();
const key = client.getReconnectKey();

// 断线后重连
await client.reconnectRoom(roomId, key);
```

## 运行测试

```bash
# 安装依赖
pnpm install

# 运行所有测试
npm test

# 运行特定测试
npm test http        # HTTP 请求测试
npm test wt          # WebTransport 测试
npm test reconnect   # 重连测试

# 运行示例
tsx example.ts
```

## API 对照表

| 服务端 API | 客户端方法 | 说明 |
|-----------|-----------|------|
| GET /rooms | `client.listRooms()` | 获取房间列表 |
| POST /rooms | `client.createRoom(roomId)` | 创建房间 |
| WebTransport /join/{roomId} | `client.joinRoom(roomId)` | 加入房间 |
| WebTransport /join/{roomId} | `client.reconnectRoom(roomId, key)` | 重连房间 |
| Datagram (send) | `client.sendRequest(request)` | 发送游戏消息 |
| Datagram (receive) | `onRoomResponse` callback | 接收游戏消息 |

## 关键技术点

### 1. Datagram API 使用

所有 WebTransport 通信都使用不可靠的 datagram：

```typescript
// 发送
const writer = transport.datagrams.writable.getWriter();
await writer.write(data);
writer.releaseLock();  // 立即释放锁

// 接收（循环）
const reader = transport.datagrams.readable.getReader();
while (running) {
  const { value, done } = await reader.read();
  if (done) break;
  handleMessage(value);
}
```

### 2. Protobuf 序列化

使用 @protobuf-ts 库：

```typescript
// 序列化
const bytes = Request.toBinary(request);

// 反序列化
const response = LobbyResponse.fromBinary(data);
```

### 3. 状态管理

通过状态机管理连接状态：

```
DISCONNECTED → CONNECTING → LOBBY → CONNECTED
                    ↓           ↓
                  ERROR ← ← ← ← ←
```

### 4. 类型安全

完整的 TypeScript 类型支持：
- Protobuf 消息类型
- HTTP 请求/响应类型
- 连接状态枚举
- 回调函数类型

## 注意事项

1. **HTTPS 要求**：WebTransport 必须使用 HTTPS/TLS
2. **浏览器支持**：需要支持 WebTransport API 的现代浏览器
3. **Datagram 不可靠性**：消息可能丢失或乱序，应用层需要处理
4. **重连密钥**：务必保存 `getReconnectKey()` 返回的密钥
5. **状态同步**：通过 `onStateChange` 监听连接状态变化

## 扩展建议

1. **消息队列**：为 datagram 添加重传机制
2. **心跳机制**：定期发送 blank 消息保持连接
3. **错误重试**：自动重连失败的连接
4. **日志系统**：添加详细的调试日志
5. **性能监控**：统计消息延迟和丢包率

## 总结

这个 SDK 完全实现了你的需求：

✅ HTTP 请求（房间列表、创建房间）  
✅ WebTransport 长连接（datagram API）  
✅ 加入房间功能  
✅ 重连功能  
✅ Protobuf 消息支持  
✅ 完整的类型定义  
✅ 测试套件和使用示例  
✅ 详细的文档  

SDK 结构清晰，易于使用和扩展，完全对应服务端的 API 设计。
