# 快速使用指南：单向流测试

## 快速开始

### 1. 构建 SDK

```bash
cd example/ts-app
pnpm install
pnpm run build
```

### 2. 启动 React 应用

```bash
cd example/react-app
npm install
npm run dev
```

### 3. 测试单向流

在浏览器中：

1. 打开 `http://localhost:5173`
2. 在配置面板中设置：
   - 服务器 URL: `https://127.0.0.1:12345`（根据你的测试服务器调整）
   - 证书哈希: 输入你的服务器证书 SHA-256 哈希（可选，用于自签名证书）
   - 勾选 "允许自签名证书" 和 "允许任意证书"
3. 滚动到 "🧪 单向流测试" 面板
4. 设置：
   - 端点路径: `/unidirectional`
   - 流数量: `5`
   - 数据: 输入任意文本
5. 点击 "🚀 完整测试" 按钮

## 代码示例

### 基础用法

```typescript
import { LockstepClient } from 'lockstep-core-client';

const client = new LockstepClient({
  serverUrl: 'https://127.0.0.1:12345',
  safety: {
    allowSelfSigned: true,
    serverCertificateHashes: ['your-sha256-hash']
  }
});

// 方式 1: 使用便捷方法
await client.connectToEndpoint('/unidirectional');
const data = new TextEncoder().encode('Hello!');
await client.createMultipleUnidirectionalStreams(5, data);
await client.disconnect();
```

### 进阶用法：单独控制

```typescript
import { StreamClient } from 'lockstep-core-client/requests/stream';

const streamClient = new StreamClient('https://127.0.0.1:12345', {
  allowSelfSigned: true
});

// 连接
await streamClient.connectToEndpoint('/unidirectional');

// 发送单个流
const data = new TextEncoder().encode('Test data');
await streamClient.createUnidirectionalStream({ 
  data, 
  waitForClose: true 
});

// 断开
await streamClient.disconnect();
```

## 与参考 HTML 的对应关系

参考 HTML 代码：

```javascript
const transport = new WebTransport(url, {
    "serverCertificateHashes": [{
        "algorithm": "sha-256",
        "value": new Uint8Array([...])
    }]
});

await transport.ready;

for(let i = 0; i < 5; i++) {
    const stream = await transport.createUnidirectionalStream();
    const writer = stream.getWriter();
    writer.write(data);
    await writer.close();
}
```

我们的实现：

```typescript
// 在 StreamClient.connectToEndpoint() 中
const options: WebTransportOptions = {};
if (this.safety?.serverCertificateHashes) {
  options.serverCertificateHashes = 
    this.safety.serverCertificateHashes.map(hash => ({
      algorithm: 'sha-256',
      value: this.hexToArrayBuffer(hash)
    }));
}
this.transport = new WebTransport(url, options);
await this.transport.ready;

// 在 StreamClient.createMultipleUnidirectionalStreams() 中
for (let i = 0; i < count; i++) {
  const stream = await this.transport.createUnidirectionalStream();
  const writer = stream.getWriter();
  await writer.write(data);
  await writer.close();
}
```

## 测试服务器

如果你有 quiche 的示例服务器：

```bash
# 编译运行
cargo run --example http3-server -- --listen 127.0.0.1:12345

# 获取证书哈希
openssl x509 -in cert.pem -outform DER | openssl dgst -sha256 -binary | xxd -p -c 64
```

## 常见问题

### Q: 连接失败，显示证书错误

A: 确保：
1. 在配置面板中启用 "允许自签名证书"
2. 如果是浏览器环境，正确配置 `serverCertificateHashes`
3. 证书哈希格式正确（十六进制字符串，无空格或冒号）

### Q: 如何获取服务器证书哈希？

A: 
```bash
# 方法 1: 从证书文件
openssl x509 -in server.crt -outform DER | openssl dgst -sha256 -binary | xxd -p -c 64

# 方法 2: 从运行中的服务器
echo | openssl s_client -connect 127.0.0.1:12345 2>/dev/null | \
  openssl x509 -outform DER | openssl dgst -sha256 -binary | xxd -p -c 64
```

### Q: 在 Node.js 环境中使用

A: Node.js 默认不支持 WebTransport。可以：
1. 使用实验性的 Node.js WebTransport 支持
2. 或在浏览器环境中使用此功能

## 下一步

- 查看 `UNIDIRECTIONAL_STREAMS.md` 了解详细的实现说明
- 查看 React 应用中的实时示例
- 尝试修改流数量和数据大小进行性能测试
