# 单向流功能实现总结

## 改造完成 ✅

已成功为 `ts-app` SDK 和 `react-app` 测试应用添加了 WebTransport 单向流支持，参考了标准 WebTransport 示例服务端的实现。

## 核心改动

### 1. ts-app (SDK)

#### 新增功能

**文件**：`src/requests/stream.ts`

- `UnidirectionalStreamOptions` 接口
- `createUnidirectionalStream()` - 创建单个单向流
- `createMultipleUnidirectionalStreams()` - 批量创建单向流
- `connectToEndpoint()` - 连接到任意端点（支持完整 URL 或相对路径）

**关键改进**：
```typescript
// 智能 URL 处理
if (endpoint.startsWith('http://') || endpoint.startsWith('https://')) {
  url = endpoint;  // 完整 URL
} else {
  url = `${this.serverUrl}${endpoint}`;  // 相对路径
}
```

**文件**：`src/core.ts`

在 `LockstepClient` 类中封装了单向流方法：
- `connectToEndpoint(endpoint: string)`
- `createUnidirectionalStream(data: Uint8Array)`
- `createMultipleUnidirectionalStreams(count: number, data: Uint8Array)`

**文件**：`src/index.ts`

导出新类型：
```typescript
export type { UnidirectionalStreamOptions, SafetyOptions }
```

### 2. react-app (测试应用)

#### 新组件：`UnidirectionalTestPanel.tsx`

**特性**：
- ✅ 独立的客户端实例（不受全局配置影响）
- ✅ 独立的证书哈希配置
- ✅ 完整 URL 或相对路径支持
- ✅ 可配置流数量和数据内容
- ✅ 实时状态显示
- ✅ 一键完整测试
- ✅ 分步操作支持
- ✅ 详细的错误提示和成功反馈

**UI 布局**：
```
🧪 单向流测试
├── 状态指示器（颜色编码）
├── 证书哈希输入框
├── 端点 URL 输入框
├── 流数量配置
├── 数据输入框
├── 操作按钮
│   ├── 连接
│   ├── 发送流
│   ├── 断开连接
│   └── 🚀 完整测试
├── 结果显示
└── 使用说明
```

#### 集成到 App.tsx

面板已添加到中间列，与其他测试面板并列显示。

## 解决的问题

### 问题 1：URL 拼接错误 ✅

**问题**：
```
URL 'https://127.0.0.1:4433https://127.0.0.1:12345/unidirectional' is invalid.
```

**原因**：`connectToEndpoint` 无条件地将 `serverUrl` 和 `endpoint` 拼接

**解决**：添加 URL 类型判断
```typescript
if (endpoint.startsWith('http://') || endpoint.startsWith('https://')) {
  url = endpoint;  // 直接使用完整 URL
} else {
  url = `${this.serverUrl}${endpoint}`;  // 拼接相对路径
}
```

### 问题 2：证书验证失败 ✅

**问题**：
```
ERR_QUIC_PROTOCOL_ERROR.QUIC_TLS_CERTIFICATE_UNKNOWN
```

**原因**：使用了全局配置的证书哈希，但测试的是不同的服务器

**解决**：
1. `UnidirectionalTestPanel` 创建独立的客户端实例
2. 提供独立的证书哈希配置输入框
3. 证书哈希通过 `useEffect` 动态应用到新客户端

```typescript
useEffect(() => {
  const hashes = certHash.split(/[\n,]/).map(h => h.trim()).filter(h => h.length > 0);
  const newClient = new LockstepClient({
    serverUrl: '',
    safety: {
      serverCertificateHashes: hashes.length > 0 ? hashes : undefined,
    },
  });
  setTestClient(newClient);
}, [certHash]);
```

## 参考实现对比

### HTML 示例代码
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

transport.close();
```

### 我们的实现
```typescript
// 1. 连接
const options: WebTransportOptions = {
  serverCertificateHashes: hashes.map(hash => ({
    algorithm: 'sha-256',
    value: this.hexToArrayBuffer(hash)  // 十六进制 → 字节数组
  }))
};
this.transport = new WebTransport(url, options);
await this.transport.ready;

// 2. 发送流
for (let i = 0; i < count; i++) {
  const stream = await this.transport.createUnidirectionalStream();
  const writer = stream.getWriter();
  await writer.write(data);
  await writer.close();
}

// 3. 断开
await this.transport.close();
```

**改进点**：
- ✅ 支持十六进制字符串格式的证书哈希（更易用）
- ✅ 智能处理完整 URL 和相对路径
- ✅ TypeScript 类型安全
- ✅ 友好的 UI 界面
- ✅ 完整的错误处理
- ✅ 状态管理

## 使用方法

### 代码方式

```typescript
import { LockstepClient } from 'lockstep-core-client';

const client = new LockstepClient({
  serverUrl: '',  // 不需要 serverUrl（使用完整 URL）
  safety: {
    serverCertificateHashes: ['your-sha256-hash-here']
  }
});

// 连接并发送
await client.connectToEndpoint('https://127.0.0.1:12345/unidirectional');
const data = new TextEncoder().encode('Hello!');
await client.createMultipleUnidirectionalStreams(5, data);
await client.disconnect();
```

### UI 方式

1. 在 React 应用中找到 "🧪 单向流测试" 面板
2. 输入证书哈希（如果需要）
3. 输入完整 URL：`https://127.0.0.1:12345/unidirectional`
4. 设置流数量和数据
5. 点击 "🚀 完整测试"

## 文档

已创建以下文档：

1. **UNIDIRECTIONAL_STREAMS.md** - 详细的实现说明和 API 文档
2. **UNIDIRECTIONAL_STREAMS_QUICKSTART.md** - 快速开始指南
3. **TESTING_UNIDIRECTIONAL_STREAMS.md** - 测试指南和故障排除

## 测试清单

- [x] SDK 编译成功
- [x] React 应用编译成功
- [x] TypeScript 类型检查通过
- [x] UI 组件正常渲染
- [x] 支持完整 URL
- [x] 支持相对路径
- [x] 证书哈希处理正确
- [x] 错误提示友好
- [x] 文档完整

## 下一步建议

1. **测试实际服务器**：
   - 使用 quiche、aioquic 或其他 WebTransport 服务器
   - 验证与不同实现的兼容性

2. **性能测试**：
   - 测试大量并发流
   - 测试大数据包
   - 测试网络延迟影响

3. **功能扩展**：
   - 添加双向流支持
   - 支持接收单向流数据
   - 添加流的超时和重试机制
   - 提供性能统计（延迟、吞吐量等）

4. **文档完善**：
   - 添加更多示例
   - 视频教程
   - 集成到主文档

## 总结

✅ 成功参考 HTML 示例实现了 WebTransport 单向流功能
✅ 解决了 URL 拼接和证书验证的问题
✅ 提供了友好的 UI 界面和完整的文档
✅ 代码质量高，类型安全，错误处理完善

现在可以使用 react-app 的测试面板来测试任何支持 WebTransport 单向流的服务器！
