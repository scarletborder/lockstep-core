# 🚀 单向流测试 - 快速参考

## 立即开始

```bash
# 1. 构建 SDK
cd example/ts-app && pnpm run build

# 2. 安装依赖
cd ../react-app && npm install

# 3. 启动应用
npm run dev

# 4. 打开浏览器访问显示的地址（通常是 http://localhost:5173 或 5174）
```

## 获取证书哈希

```bash
# 方法 1：从证书文件
openssl x509 -in server.crt -outform DER | openssl dgst -sha256 -binary | xxd -p -c 64

# 方法 2：从运行的服务器
echo | openssl s_client -connect 127.0.0.1:12345 2>/dev/null | \
  openssl x509 -outform DER | openssl dgst -sha256 -binary | xxd -p -c 64
```

## 在 UI 中测试

1. 找到 "🧪 单向流测试" 面板
2. **证书哈希**：粘贴上面的输出（如果是自签名证书）
3. **端点 URL**：`https://127.0.0.1:12345/unidirectional`
4. **流数量**：`5`
5. **数据**：任意文本
6. 点击 **🚀 完整测试**

## 代码示例

```typescript
import { LockstepClient } from 'lockstep-core-client';

const client = new LockstepClient({
  serverUrl: '',
  safety: {
    serverCertificateHashes: ['your-cert-hash']
  }
});

await client.connectToEndpoint('https://127.0.0.1:12345/unidirectional');
const data = new TextEncoder().encode('Hello!');
await client.createMultipleUnidirectionalStreams(5, data);
await client.disconnect();
```

## 常见错误

| 错误 | 原因 | 解决方法 |
|------|------|----------|
| `QUIC_TLS_CERTIFICATE_UNKNOWN` | 证书哈希错误或缺失 | 输入正确的证书哈希 |
| `URL is invalid` | URL 格式错误 | 确保使用 `https://` 开头的完整 URL |
| 连接超时 | 服务器未运行 | 确认服务器地址和端口 |

## 关键改进

✅ 支持完整 URL（`https://127.0.0.1:12345/...`）  
✅ 支持相对路径（`/unidirectional`）  
✅ 独立证书哈希配置  
✅ 一键完整测试  
✅ 实时状态显示  

## 文档

- 📖 [详细实现说明](./UNIDIRECTIONAL_STREAMS.md)
- 🚀 [快速开始指南](./UNIDIRECTIONAL_STREAMS_QUICKSTART.md)
- 🧪 [测试指南](./TESTING_UNIDIRECTIONAL_STREAMS.md)
- 📝 [实现总结](./UNIDIRECTIONAL_STREAMS_SUMMARY.md)

## 问题反馈

遇到问题？检查：
1. 浏览器是否支持 WebTransport
2. 证书哈希格式是否正确（纯十六进制，无空格）
3. 服务器是否正在运行
4. 查看浏览器开发者工具的控制台和网络标签
