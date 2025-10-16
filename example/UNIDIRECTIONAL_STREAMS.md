# 单向流测试功能 (Unidirectional Streams Testing)

## 概述

本次更新为 `lockstep-core-client` SDK 和 React 测试应用添加了对 WebTransport 单向流的支持，参考了标准 WebTransport 示例服务端的实现。

## 修改内容

### 1. ts-app (SDK 核心库)

#### `src/requests/stream.ts`

新增以下功能：

- **`UnidirectionalStreamOptions` 接口**：定义单向流选项
  ```typescript
  export interface UnidirectionalStreamOptions {
    data: Uint8Array;
    waitForClose?: boolean;
  }
  ```

- **`createUnidirectionalStream()`**：创建单个单向流并发送数据
  ```typescript
  async createUnidirectionalStream(options: UnidirectionalStreamOptions): Promise<void>
  ```

- **`createMultipleUnidirectionalStreams()`**：批量创建多个单向流
  ```typescript
  async createMultipleUnidirectionalStreams(count: number, data: Uint8Array): Promise<void>
  ```

- **`connectToEndpoint()`**：连接到指定端点（用于测试非标准端点）
  ```typescript
  async connectToEndpoint(endpoint: string): Promise<void>
  ```

#### `src/core.ts`

在 `LockstepClient` 类中添加了对单向流的封装方法：

```typescript
async connectToEndpoint(endpoint: string): Promise<void>
async createUnidirectionalStream(data: Uint8Array): Promise<void>
async createMultipleUnidirectionalStreams(count: number, data: Uint8Array): Promise<void>
```

#### `src/index.ts`

导出新的类型定义：

```typescript
export type { UnidirectionalStreamOptions, SafetyOptions }
```

### 2. react-app (React 测试应用)

#### 新增组件：`src/components/UnidirectionalTestPanel.tsx`

这是一个全新的测试面板，提供以下功能：

1. **端点配置**：可以设置要连接的端点路径（如 `/unidirectional`）
2. **流数量配置**：设置要创建的单向流数量
3. **数据输入**：自定义要发送的数据内容
4. **连接状态显示**：实时显示连接状态
5. **操作按钮**：
   - **连接**：建立到指定端点的 WebTransport 连接
   - **发送流**：创建并发送单向流数据
   - **断开连接**：关闭连接
   - **完整测试**：自动完成连接→发送→断开的完整流程

#### 修改文件：`src/App.tsx`

- 导入 `UnidirectionalTestPanel` 组件
- 将面板添加到中间列的布局中

## 使用方法

### 在代码中使用

```typescript
import { LockstepClient } from 'lockstep-core-client';

// 初始化客户端
const client = new LockstepClient({
  serverUrl: 'https://127.0.0.1:4433',
  safety: {
    allowSelfSigned: true,
    serverCertificateHashes: ['your-cert-hash-here']
  }
});

// 连接到端点
await client.connectToEndpoint('/unidirectional');

// 发送数据
const encoder = new TextEncoder();
const data = encoder.encode('Hello from unidirectional stream!');
await client.createMultipleUnidirectionalStreams(5, data);

// 断开连接
await client.disconnect();
```

### 在 React 应用中测试

1. 启动 React 应用：
   ```bash
   cd example/react-app
   npm run dev
   ```

2. 在浏览器中打开应用

3. 配置服务器 URL 和证书哈希（在配置面板中）

4. 在 "🧪 单向流测试" 面板中：
   - 设置端点路径（如 `/unidirectional`）
   - 设置流数量（默认 5）
   - 输入要发送的数据
   - 点击 "🚀 完整测试" 按钮

## 参考实现

本实现参考了以下 WebTransport 示例代码：

```javascript
async function establishSession(url) {
    const transport = new WebTransport(url, {
        "serverCertificateHashes": [{
            "algorithm": "sha-256",
            "value": new Uint8Array([...])
        }]
    });

    transport.closed.then(() => {
        console.log(`Connection closed gracefully.`);
    }).catch((error) => {
        console.error(`Connection closed due to ${error}.`);
    });

    await transport.ready;
    return transport;
}

async function runUnidirectionalTest() {
    const transport = await establishSession('https://127.0.0.1:12345/unidirectional');
    const data = new Uint8Array([...]);

    for(let i = 0; i < 5; i++) {
        const stream = await transport.createUnidirectionalStream();
        const writer = stream.getWriter();
        writer.write(data);
        await writer.close();
    }
    
    transport.close();
}
```

## 兼容性说明

- 需要浏览器支持 WebTransport API
- 建议使用 Chrome 97+ 或其他支持 WebTransport 的现代浏览器
- 对于自签名证书，需要配置 `serverCertificateHashes`

## 测试服务端

可以使用以下服务端进行测试：

1. **quiche 示例服务器**：
   ```bash
   # 使用 quiche 提供的示例
   cargo run --manifest-path=tools/http3_test/Cargo.toml
   ```

2. **其他 WebTransport 服务器**：
   确保服务器实现了 `/unidirectional` 端点，接受单向流数据

## 调试提示

- 打开浏览器的开发者工具查看控制台日志
- 检查 "Network" 标签查看 WebTransport 连接
- 使用 "完整测试" 按钮可以快速验证整个流程
- 如果连接失败，检查证书哈希是否正确

## 后续改进

可能的改进方向：

1. 添加双向流（bidirectional streams）支持
2. 支持接收单向流数据
3. 添加流的超时和重试机制
4. 提供更详细的性能统计（延迟、吞吐量等）
