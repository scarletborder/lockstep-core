# 单向流测试使用指南

## 快速测试步骤

### 1. 获取服务器证书哈希

如果你的测试服务器使用自签名证书，需要先获取证书的 SHA-256 哈希：

```bash
# 从证书文件获取哈希
openssl x509 -in server.crt -outform DER | openssl dgst -sha256 -binary | xxd -p -c 64

# 或者从运行中的服务器获取
echo | openssl s_client -connect 127.0.0.1:12345 2>/dev/null | \
  openssl x509 -outform DER | openssl dgst -sha256 -binary | xxd -p -c 64
```

输出示例：
```
3c7f87e9d2b8a1f4e6c5d3a2b9f1e8d7c6a5b4f3e2d1c0b9a8f7e6d5c4b3a2f1
```

### 2. 在 React 应用中测试

1. 启动 React 应用（如果还没启动）：
   ```bash
   cd example/react-app
   npm run dev
   ```

2. 在浏览器中打开应用

3. 找到 "🧪 单向流测试" 面板

4. 配置参数：
   - **证书哈希**：粘贴上面获取的哈希值（如果是自签名证书）
   - **端点 URL**：输入完整 URL，例如：
     ```
     https://127.0.0.1:12345/unidirectional
     ```
   - **流数量**：设置要创建的流数量（默认 5）
   - **数据**：输入要发送的数据（默认 "Hello from unidirectional stream!"）

5. 点击 "🚀 完整测试" 按钮

### 3. 观察结果

成功的输出示例：
```
步骤 1/3: 连接到端点...
步骤 2/3: 发送单向流...
步骤 3/3: 断开连接...
✓ 完整测试成功！发送了 5 个单向流
```

## 常见问题

### Q: 提示 "ERR_QUIC_PROTOCOL_ERROR.QUIC_TLS_CERTIFICATE_UNKNOWN"

**原因**：证书验证失败

**解决方案**：
1. 确保已正确输入证书哈希
2. 检查哈希格式（纯十六进制，无空格、无冒号）
3. 确保哈希值与服务器证书匹配

### Q: 连接超时

**原因**：
- 服务器未运行
- URL 或端口错误
- 防火墙阻止连接

**解决方案**：
1. 确认服务器正在运行
2. 检查 URL 和端口是否正确
3. 检查防火墙设置

### Q: 如何验证证书哈希是否正确？

在浏览器开发者工具的控制台中，会看到详细的错误信息。如果哈希不匹配，会有类似的提示：
```
The provided certificate hash does not match the server's certificate.
```

## 与 HTML 示例的对比

### 原始 HTML 示例

```javascript
const transport = new WebTransport(url, {
    "serverCertificateHashes": [{
        "algorithm": "sha-256",
        "value": new Uint8Array([...])  // 证书哈希字节数组
    }]
});

for(let i = 0; i < 5; i++) {
    const stream = await transport.createUnidirectionalStream();
    const writer = stream.getWriter();
    writer.write(data);
    await writer.close();
}
```

### 我们的实现

```typescript
// 在 StreamClient 中
const options: WebTransportOptions = {
  serverCertificateHashes: hashes.map(hash => ({
    algorithm: 'sha-256',
    value: this.hexToArrayBuffer(hash)  // 十六进制转字节数组
  }))
};

this.transport = new WebTransport(url, options);
await this.transport.ready;

// 创建多个单向流
for (let i = 0; i < count; i++) {
  const stream = await this.transport.createUnidirectionalStream();
  const writer = stream.getWriter();
  await writer.write(data);
  await writer.close();
}
```

**主要改进**：
1. 支持十六进制字符串格式的证书哈希（更易用）
2. 自动将完整 URL 或相对路径转换
3. 提供友好的 UI 界面
4. 完整的错误处理和状态管理

## 测试服务器示例

如果你想搭建自己的测试服务器，可以参考以下项目：

1. **quiche HTTP/3 示例**：https://github.com/cloudflare/quiche
2. **aioquic**：https://github.com/aiortc/aioquic
3. **Go 实现**：使用 quic-go 库

## 下一步

- 尝试不同的数据大小
- 测试大量并发流
- 观察网络延迟和吞吐量
- 集成到实际应用中
