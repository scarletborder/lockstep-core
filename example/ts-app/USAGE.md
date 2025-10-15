# Lockstep Client SDK 使用示例

这是一个完整的客户端 SDK，用于与 Lockstep 服务器进行通信。支持普通 HTTP 请求和 WebTransport 长连接。

## 安装

```bash
npm install
# 或
pnpm install
```

## 基本使用

### 1. 初始化客户端

```typescript
import { LockstepClient, ConnectionState } from './src';

const client = new LockstepClient({
  serverUrl: 'https://localhost:8080', // 你的服务器地址
});
```

### 2. 普通 HTTP 请求

#### 获取房间列表

```typescript
try {
  const rooms = await client.listRooms();
  console.log('可用房间:', rooms);
} catch (error) {
  console.error('获取房间列表失败:', error);
}
```

#### 创建房间

```typescript
try {
  const result = await client.createRoom('room-001');
  console.log('创建房间成功:', result.message);
} catch (error) {
  console.error('创建房间失败:', error);
}
```

### 3. WebTransport 长连接

#### 设置消息处理器

在连接之前，先设置消息处理器来接收服务器的消息：

```typescript
import { LobbyResponse, RoomResponse } from './src';

client.setMessageHandlers({
  // 处理大厅响应（加入房间的结果）
  onLobbyResponse: (response: LobbyResponse) => {
    if (response.payload.oneofKind === 'joinRoomSuccess') {
      const success = response.payload.joinRoomSuccess;
      console.log(`成功加入房间！`);
      console.log(`房间 ID: ${success.roomId}`);
      console.log(`我的玩家 ID: ${success.myId}`);
      console.log(`重连密钥: ${success.key}`);
      console.log(`消息: ${success.message}`);
    } else if (response.payload.oneofKind === 'joinRoomFailed') {
      const failed = response.payload.joinRoomFailed;
      console.error(`加入房间失败: ${failed.message}`);
    }
  },

  // 处理房间内的游戏消息
  onRoomResponse: (response: RoomResponse) => {
    console.log('收到房间消息:', response);
    
    // 根据不同的消息类型进行处理
    switch (response.payload.oneofKind) {
      case 'roomInfo':
        console.log('房间信息:', response.payload.roomInfo);
        break;
      case 'playerJoined':
        console.log('玩家加入:', response.payload.playerJoined);
        break;
      case 'playerLeft':
        console.log('玩家离开:', response.payload.playerLeft);
        break;
      // ... 处理其他消息类型
    }
  },

  // 处理错误
  onError: (error: Error) => {
    console.error('发生错误:', error);
  },

  // 处理连接状态变化
  onStateChange: (state: ConnectionState) => {
    console.log('连接状态变化:', state);
  },
});
```

#### 加入房间

```typescript
try {
  await client.joinRoom('room-001');
  console.log('正在连接到房间...');
  
  // 等待连接建立（通过 onStateChange 或 onLobbyResponse 回调获知结果）
} catch (error) {
  console.error('加入房间失败:', error);
}
```

#### 发送游戏消息

加入房间成功后，可以发送游戏消息：

```typescript
import { Request } from './src';

// 等待连接建立
if (client.isConnected()) {
  // 发送准备消息
  const readyRequest: Request = {
    payload: {
      oneofKind: 'ready',
      ready: {
        isReady: true,
      },
    },
  };
  
  await client.sendRequest(readyRequest);
  
  // 发送选择地图消息
  const chooseMapRequest: Request = {
    payload: {
      oneofKind: 'chooseMap',
      chooseMap: {
        chapterId: 1,
        stageId: 1,
      },
    },
  };
  
  await client.sendRequest(chooseMapRequest);
}
```

#### 断线重连

如果连接断开，可以使用之前保存的重连密钥进行重连：

```typescript
const roomId = client.getCurrentRoomId();
const reconnectKey = client.getReconnectKey();

if (roomId && reconnectKey) {
  try {
    await client.reconnectRoom(roomId, reconnectKey);
    console.log('正在重连...');
  } catch (error) {
    console.error('重连失败:', error);
  }
}
```

#### 断开连接

```typescript
await client.disconnect();
console.log('已断开连接');
```

## 完整示例

```typescript
import { LockstepClient, ConnectionState, Request } from './src';

async function main() {
  // 1. 初始化客户端
  const client = new LockstepClient({
    serverUrl: 'https://localhost:8080',
  });

  // 2. 设置消息处理器
  client.setMessageHandlers({
    onLobbyResponse: (response) => {
      if (response.payload.oneofKind === 'joinRoomSuccess') {
        console.log('成功加入房间！');
        // 加入成功后可以开始发送游戏消息
        sendGameMessages(client);
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

  try {
    // 3. 创建房间（可选）
    await client.createRoom('my-room');
    
    // 4. 加入房间
    await client.joinRoom('my-room');
    
    // 连接建立后的操作在 onLobbyResponse 回调中处理
  } catch (error) {
    console.error('操作失败:', error);
  }
}

async function sendGameMessages(client: LockstepClient) {
  // 发送准备消息
  const readyRequest: Request = {
    payload: {
      oneofKind: 'ready',
      ready: { isReady: true },
    },
  };
  
  await client.sendRequest(readyRequest);
  console.log('已发送准备消息');
}

main();
```

## 便捷方法

SDK 还提供了一些便捷方法：

```typescript
// 创建并加入房间
await client.createAndJoinRoom('my-room');

// 查询状态
const isConnected = client.isConnected();
const state = client.getConnectionState();
const playerId = client.getMyPlayerId();
const roomId = client.getCurrentRoomId();
const key = client.getReconnectKey();
```

## 注意事项

1. **使用 Datagram API**: 所有的 WebTransport 通信都使用不可靠的 datagram 进行，这适用于实时游戏场景。

2. **连接状态管理**: 通过 `onStateChange` 回调可以监听连接状态的变化，包括：
   - `DISCONNECTED`: 未连接
   - `CONNECTING`: 正在连接
   - `LOBBY`: 等待加入房间响应
   - `CONNECTED`: 已连接到房间
   - `RECONNECTING`: 正在重连
   - `ERROR`: 错误状态

3. **消息处理**: `onLobbyResponse` 用于处理加入房间的响应，`onRoomResponse` 用于处理房间内的游戏消息。

4. **错误处理**: 始终设置 `onError` 回调来处理可能的错误。

5. **重连机制**: 保存 `getReconnectKey()` 返回的密钥，用于断线后重连。

6. **HTTPS 要求**: WebTransport 需要 HTTPS/TLS 连接。

## 打包与发布 (用于浏览器 / npm)

下面说明如何把 `ts-app` 打包并发布为一个在浏览器和 npm 上都能使用的包。

1. 安装依赖（在 `example/ts-app` 目录下）:

```bash
pnpm install
```

2. 本地构建:

```bash
pnpm run build
```

构建完成后会在 `dist/` 目录得到以下产物：
- `index.esm.js` — ESM 模块
- `index.cjs.js` — CommonJS 模块
- `index.umd.js` — 浏览器 UMD 构建（可通过 `<script>` 引入）
- `index.d.ts` — TypeScript 类型声明

3. 发布到 npm:

```bash
npm publish --access public
```

package.json 中的 `exports` 与 `unpkg` 字段会让像 unpkg / jsDelivr 这样的 CDN 能够直接为浏览器提供 `index.umd.js`。

4. 从 CDN 使用（示例）:

```html
<script src="https://unpkg.com/lockstep-core-client@1.0.0/dist/index.umd.js"></script>
<script>
  const client = new LockstepCoreClient.LockstepClient({ serverUrl: 'https://example.com' });
  // 使用 client
</script>
```

注: 如果你想使用更小的浏览器构建 (minified)，可以在 `tsup.config.ts` 中开启 `minify: true` 或额外传入 `--minify`。
