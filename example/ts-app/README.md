# Lockstep Core - TypeScript Client SDK

TypeScript 客户端 SDK，用于与 Lockstep 服务器进行通信。支持普通 HTTP 请求和 WebTransport 长连接（使用不可靠的 datagram）。

## 功能特性

- ✅ **HTTP 请求**: 获取房间列表、创建房间
- ✅ **WebTransport 长连接**: 使用 datagram 进行实时通信
- ✅ **加入房间**: 建立 WebTransport 连接并加入游戏房间
- ✅ **断线重连**: 支持使用密钥重新连接到房间
- ✅ **Protobuf 消息**: 完整的 protobuf 类型定义
- ✅ **TypeScript**: 完整的类型支持

## 安装

```bash
pnpm install
```

## 快速开始

### 基本使用

```typescript
import { LockstepClient, ConnectionState } from './src';

// 1. 初始化客户端
const client = new LockstepClient({
  serverUrl: 'https://localhost:8080',
});

// 2. 设置消息处理器
client.setMessageHandlers({
  onLobbyResponse: (response) => {
    if (response.payload.oneofKind === 'joinRoomSuccess') {
      console.log('成功加入房间！');
    }
  },
  onRoomResponse: (response) => {
    console.log('收到房间消息:', response);
  },
  onError: (error) => {
    console.error('错误:', error);
  },
  onStateChange: (state) => {
    console.log('连接状态:', state);
  },
});

// 3. 创建并加入房间
await client.createAndJoinRoom('my-room');

// 4. 发送游戏消息
if (client.isConnected()) {
  await client.sendRequest({
    payload: {
      oneofKind: 'ready',
      ready: { isReady: true },
    },
  });
}
```

## API 文档

### LockstepClient

主客户端类，提供所有功能的统一接口。

#### 构造函数

```typescript
new LockstepClient(options: InitOptions)
```

#### HTTP 方法

- `listRooms(): Promise<string[]>` - 获取房间列表
- `createRoom(roomId: string): Promise<CreateRoomResponse>` - 创建房间

#### WebTransport 方法

- `joinRoom(roomId: string): Promise<void>` - 加入房间
- `reconnectRoom(roomId: string, key: string): Promise<void>` - 重连房间
- `sendRequest(request: Request): Promise<void>` - 发送游戏请求
- `disconnect(): Promise<void>` - 断开连接

#### 状态方法

- `getConnectionState(): ConnectionState` - 获取连接状态
- `getMyPlayerId(): number | null` - 获取玩家 ID
- `getCurrentRoomId(): string | null` - 获取当前房间 ID
- `getReconnectKey(): string | null` - 获取重连密钥
- `isConnected(): boolean` - 检查是否已连接

#### 事件处理

```typescript
client.setMessageHandlers({
  onLobbyResponse?: (response: LobbyResponse) => void;
  onRoomResponse?: (response: RoomResponse) => void;
  onError?: (error: Error) => void;
  onStateChange?: (state: ConnectionState) => void;
});
```

## 运行测试

```bash
# 运行所有测试
npm test

# 只测试 HTTP 请求
npm test http

# 只测试 WebTransport
npm test wt

# 测试重连功能
npm test reconnect
```

## 项目结构

```
src/
├── core.ts              # 主客户端类
├── index.ts             # 导出入口
├── requests/
│   ├── common.ts        # HTTP 请求实现
│   └── stream.ts        # WebTransport 实现
└── types/
    ├── index.ts         # 类型导出
    └── pb/              # Protobuf 类型定义
        ├── request.ts   # 请求消息类型
        └── response.ts  # 响应消息类型
```

## 详细文档

查看 [USAGE.md](./USAGE.md) 获取更详细的使用说明和示例。

## 注意事项

1. **HTTPS 要求**: WebTransport 需要 HTTPS/TLS 连接
2. **Datagram API**: 所有消息通过不可靠的 datagram 发送
3. **浏览器支持**: 需要支持 WebTransport 的现代浏览器
4. **重连密钥**: 务必保存 `getReconnectKey()` 返回的密钥用于重连

## 许可证

MIT
