# Lockstep Client SDK - 快速参考

## 安装与初始化

```typescript
import { LockstepClient } from './src';

const client = new LockstepClient({
  serverUrl: 'https://localhost:8080'
});
```

## HTTP API

```typescript
// 获取房间列表
const rooms: string[] = await client.listRooms();

// 创建房间
const result = await client.createRoom('room-001');

// 便捷方法：创建并加入
await client.createAndJoinRoom('room-001');
```

## WebTransport - 消息处理

```typescript
client.setMessageHandlers({
  // 加入房间响应
  onLobbyResponse: (response) => {
    if (response.payload.oneofKind === 'joinRoomSuccess') {
      const { roomId, myId, key, message } = response.payload.joinRoomSuccess;
      // 保存 key 用于重连
    }
  },
  
  // 房间内消息
  onRoomResponse: (response) => {
    switch (response.payload.oneofKind) {
      case 'roomInfo': /* ... */ break;
      case 'chooseMap': /* ... */ break;
      case 'updateReadyCount': /* ... */ break;
      case 'allReady': /* ... */ break;
      case 'allLoaded': /* ... */ break;
      case 'error': /* ... */ break;
    }
  },
  
  // 错误处理
  onError: (error) => console.error(error),
  
  // 状态变化
  onStateChange: (state) => console.log(state),
});
```

## WebTransport - 连接管理

```typescript
// 加入房间
await client.joinRoom('room-001');

// 重连房间
const roomId = client.getCurrentRoomId();
const key = client.getReconnectKey();
await client.reconnectRoom(roomId, key);

// 断开连接
await client.disconnect();
```

## 发送游戏消息

```typescript
import { Request } from './src';

// 准备消息
await client.sendRequest({
  payload: {
    oneofKind: 'ready',
    ready: { isReady: true }
  }
});

// 选择地图
await client.sendRequest({
  payload: {
    oneofKind: 'chooseMap',
    chooseMap: { chapterId: 1, stageId: 1 }
  }
});

// 心跳消息
await client.sendRequest({
  payload: {
    oneofKind: 'blank',
    blank: { frameId: 1, ackFrameId: 0 }
  }
});

// 加载完成
await client.sendRequest({
  payload: {
    oneofKind: 'loaded',
    loaded: { isLoaded: true }
  }
});

// 种植卡片
await client.sendRequest({
  payload: {
    oneofKind: 'plant',
    plant: {
      base: { frameId: 10, ackFrameId: 9 },
      cardId: 1,
      row: 2,
      col: 3
    }
  }
});

// 移除植物
await client.sendRequest({
  payload: {
    oneofKind: 'removePlant',
    removePlant: {
      base: { frameId: 20, ackFrameId: 19 },
      plantId: 5
    }
  }
});

// 使用星星碎片
await client.sendRequest({
  payload: {
    oneofKind: 'starShards',
    starShards: {
      base: { frameId: 30, ackFrameId: 29 },
      count: 10
    }
  }
});

// 结束游戏
await client.sendRequest({
  payload: {
    oneofKind: 'endGame',
    endGame: {
      win: true,
      score: 1000
    }
  }
});

// 离开选择地图
await client.sendRequest({
  payload: {
    oneofKind: 'leaveChooseMap',
    leaveChooseMap: {}
  }
});
```

## 状态查询

```typescript
// 连接状态
const state = client.getConnectionState();
// 'disconnected' | 'connecting' | 'lobby' | 'connected' | 'reconnecting' | 'error'

// 是否已连接
const connected = client.isConnected();

// 玩家 ID
const playerId = client.getMyPlayerId();

// 当前房间 ID
const roomId = client.getCurrentRoomId();

// 重连密钥
const key = client.getReconnectKey();
```

## 连接状态

```typescript
enum ConnectionState {
  DISCONNECTED = 'disconnected',  // 未连接
  CONNECTING = 'connecting',      // 正在连接
  LOBBY = 'lobby',                // 等待加入响应
  CONNECTED = 'connected',        // 已连接到房间
  RECONNECTING = 'reconnecting',  // 正在重连
  ERROR = 'error'                 // 错误状态
}
```

## 完整流程示例

```typescript
// 1. 初始化
const client = new LockstepClient({ serverUrl: 'https://localhost:8080' });

// 2. 设置处理器
client.setMessageHandlers({
  onLobbyResponse: (resp) => {
    if (resp.payload.oneofKind === 'joinRoomSuccess') {
      console.log('加入成功！');
      // 开始游戏逻辑
    }
  },
  onRoomResponse: (resp) => {
    // 处理游戏消息
  },
});

// 3. 创建房间
await client.createRoom('my-room');

// 4. 加入房间
await client.joinRoom('my-room');

// 5. 等待连接建立（通过 onLobbyResponse 回调）

// 6. 发送游戏消息
if (client.isConnected()) {
  await client.sendRequest({ /* ... */ });
}

// 7. 断开连接
await client.disconnect();
```

## 运行测试

```bash
npm test           # 所有测试
npm test http      # HTTP 测试
npm test wt        # WebTransport 测试
npm test reconnect # 重连测试
```

## 重要提醒

- ⚠️ 需要 HTTPS/TLS 连接
- ⚠️ 使用 datagram（不可靠传输）
- ⚠️ 保存重连密钥
- ⚠️ 监听 onError 和 onStateChange
- ⚠️ 确保浏览器支持 WebTransport

## 文档链接

- [README.md](./README.md) - 项目概览
- [USAGE.md](./USAGE.md) - 详细使用指南
- [IMPLEMENTATION.md](./IMPLEMENTATION.md) - 实现细节
