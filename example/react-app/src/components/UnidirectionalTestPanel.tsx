import { useState, useEffect } from 'react';
import { LockstepClient, ConnectionState } from 'lockstep-core-client';

interface UnidirectionalTestPanelProps {
  // 不使用传入的 client，而是自己创建独立的客户端
  client?: LockstepClient | null;
  connectionState?: ConnectionState;
}

export function UnidirectionalTestPanel(_props: UnidirectionalTestPanelProps) {
  const [endpoint, setEndpoint] = useState('https://127.0.0.1:12345/unidirectional');
  const [streamCount, setStreamCount] = useState(5);
  const [dataInput, setDataInput] = useState('Hello from unidirectional stream!');
  const [certHash, setCertHash] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);

  // 独立的客户端实例和连接状态
  const [testClient, setTestClient] = useState<LockstepClient | null>(null);
  const [testConnectionState, setTestConnectionState] = useState<ConnectionState>(ConnectionState.DISCONNECTED);

  const [testResult, setTestResult] = useState<{
    success: boolean;
    message: string;
    timestamp: Date;
  } | null>(null);

  // 创建独立的客户端实例
  useEffect(() => {
    const hashes = certHash
      .split(/[\n,]/)
      .map(h => h.trim())
      .filter(h => h.length > 0);

    const newClient = new LockstepClient({
      serverUrl: '', // 不使用 serverUrl，因为我们会传入完整 URL
      safety: {
        allowSelfSigned: true,
        allowInsecureTransport: true,
        allowAnyCert: true,
        serverCertificateHashes: hashes.length > 0 ? hashes : undefined,
      },
    });

    newClient.setMessageHandlers({
      onStateChange: (state: ConnectionState) => {
        setTestConnectionState(state);
      },
      onError: (error: Error) => {
        console.error('Test client error:', error);
      },
    });

    setTestClient(newClient);

    return () => {
      newClient.disconnect();
    };
  }, [certHash]);

  const handleConnect = async () => {
    if (!testClient) return;
    setLoading(true);
    setMessage('');
    setTestResult(null);

    try {
      await testClient.connectToEndpoint(endpoint);
      setMessage(`✓ 已连接到端点: ${endpoint}`);
      setTestResult({
        success: true,
        message: `成功连接到 ${endpoint}`,
        timestamp: new Date()
      });
    } catch (error) {
      const errorMsg = `✗ 连接失败: ${(error as Error).message}`;
      setMessage(errorMsg);
      setTestResult({
        success: false,
        message: errorMsg,
        timestamp: new Date()
      });
    } finally {
      setLoading(false);
    }
  };

  const handleRunTest = async () => {
    if (!testClient) return;

    // 检查是否已连接
    if (testConnectionState !== ConnectionState.CONNECTED) {
      setMessage('✗ 请先连接到端点');
      return;
    }

    setLoading(true);
    setMessage('');
    setTestResult(null);

    try {
      // 将输入的字符串转换为 Uint8Array
      const encoder = new TextEncoder();
      const data = encoder.encode(dataInput);

      setMessage(`正在创建 ${streamCount} 个单向流并发送数据...`);

      await testClient.createMultipleUnidirectionalStreams(streamCount, data);

      const successMsg = `✓ 成功发送 ${streamCount} 个单向流`;
      setMessage(successMsg);
      setTestResult({
        success: true,
        message: successMsg,
        timestamp: new Date()
      });
    } catch (error) {
      const errorMsg = `✗ 测试失败: ${(error as Error).message}`;
      setMessage(errorMsg);
      setTestResult({
        success: false,
        message: errorMsg,
        timestamp: new Date()
      });
    } finally {
      setLoading(false);
    }
  };

  const handleDisconnect = async () => {
    if (!testClient) return;
    setLoading(true);
    setMessage('');

    try {
      await testClient.disconnect();
      setMessage('✓ 已断开连接');
      setTestResult(null);
    } catch (error) {
      setMessage(`✗ 断开连接失败: ${(error as Error).message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleRunFullTest = async () => {
    if (!testClient) return;

    setLoading(true);
    setMessage('');
    setTestResult(null);

    try {
      // 步骤 1: 连接到端点
      setMessage('步骤 1/3: 连接到端点...');
      await testClient.connectToEndpoint(endpoint);
      await new Promise(resolve => setTimeout(resolve, 500));

      // 步骤 2: 发送流数据
      setMessage('步骤 2/3: 发送单向流...');
      const encoder = new TextEncoder();
      const data = encoder.encode(dataInput);
      await testClient.createMultipleUnidirectionalStreams(streamCount, data);
      await new Promise(resolve => setTimeout(resolve, 500));

      // 步骤 3: 断开连接
      setMessage('步骤 3/3: 断开连接...');
      await testClient.disconnect();

      const successMsg = `✓ 完整测试成功！发送了 ${streamCount} 个单向流`;
      setMessage(successMsg);
      setTestResult({
        success: true,
        message: successMsg,
        timestamp: new Date()
      });
    } catch (error) {
      const errorMsg = `✗ 完整测试失败: ${(error as Error).message}`;
      setMessage(errorMsg);
      setTestResult({
        success: false,
        message: errorMsg,
        timestamp: new Date()
      });

      // 尝试清理连接
      try {
        await testClient.disconnect();
      } catch (e) {
        console.error('Failed to disconnect:', e);
      }
    } finally {
      setLoading(false);
    }
  };

  const getStateColor = (state: ConnectionState) => {
    switch (state) {
      case ConnectionState.CONNECTED:
        return '#22c55e';
      case ConnectionState.CONNECTING:
        return '#eab308';
      case ConnectionState.ERROR:
        return '#ef4444';
      default:
        return '#6b7280';
    }
  };

  const getStateText = (state: ConnectionState) => {
    const stateMap: Record<ConnectionState, string> = {
      [ConnectionState.DISCONNECTED]: '未连接',
      [ConnectionState.CONNECTING]: '连接中',
      [ConnectionState.LOBBY]: '大厅中',
      [ConnectionState.CONNECTED]: '已连接',
      [ConnectionState.RECONNECTING]: '重连中',
      [ConnectionState.ERROR]: '错误',
    };
    return stateMap[state] || state;
  };

  return (
    <div style={{ border: '1px solid #646cff', borderRadius: '8px', padding: '16px', marginBottom: '16px' }}>
      <h3 style={{ marginTop: 0 }}>🧪 单向流测试 (Unidirectional Streams)</h3>

      {/* 连接状态指示器 */}
      <div style={{
        background: '#2a2a2a',
        padding: '12px',
        borderRadius: '4px',
        marginBottom: '12px',
        display: 'flex',
        alignItems: 'center',
        gap: '8px'
      }}>
        <div style={{
          width: '12px',
          height: '12px',
          borderRadius: '50%',
          background: getStateColor(testConnectionState),
          boxShadow: `0 0 8px ${getStateColor(testConnectionState)}`,
        }} />
        <span style={{ color: '#e5e7eb' }}>
          状态: <strong>{getStateText(testConnectionState)}</strong>
        </span>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {/* 证书哈希配置 */}
        <div>
          <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', color: '#9ca3af' }}>
            证书哈希 (SHA-256):
          </label>
          <textarea
            value={certHash}
            onChange={(e) => setCertHash(e.target.value)}
            disabled={loading || testConnectionState !== ConnectionState.DISCONNECTED}
            rows={2}
            placeholder="输入证书 SHA-256 哈希值（十六进制）"
            style={{
              width: '100%',
              padding: '8px',
              borderRadius: '4px',
              border: '1px solid #4b5563',
              background: '#1f2937',
              color: '#fff',
              fontFamily: 'monospace',
              fontSize: '12px',
              resize: 'vertical',
            }}
          />
          <div style={{ fontSize: '12px', color: '#6b7280', marginTop: '4px' }}>
            自签名证书需要提供 SHA-256 哈希值
          </div>
        </div>

        {/* 端点配置 */}
        <div>
          <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', color: '#9ca3af' }}>
            端点 URL:
          </label>
          <input
            type="text"
            value={endpoint}
            onChange={(e) => setEndpoint(e.target.value)}
            disabled={loading || testConnectionState !== ConnectionState.DISCONNECTED}
            placeholder="/unidirectional"
            style={{
              width: '100%',
              padding: '8px',
              borderRadius: '4px',
              border: '1px solid #4b5563',
              background: '#1f2937',
              color: '#fff',
            }}
          />
          <div style={{ fontSize: '12px', color: '#6b7280', marginTop: '4px' }}>
            完整 URL（如 https://127.0.0.1:12345/unidirectional）或相对路径（如 /unidirectional）
          </div>
        </div>

        {/* 流数量配置 */}
        <div>
          <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', color: '#9ca3af' }}>
            流数量:
          </label>
          <input
            type="number"
            value={streamCount}
            onChange={(e) => setStreamCount(parseInt(e.target.value) || 1)}
            min="1"
            max="100"
            disabled={loading}
            style={{
              width: '100%',
              padding: '8px',
              borderRadius: '4px',
              border: '1px solid #4b5563',
              background: '#1f2937',
              color: '#fff',
            }}
          />
        </div>

        {/* 数据输入 */}
        <div>
          <label style={{ display: 'block', marginBottom: '4px', fontSize: '14px', color: '#9ca3af' }}>
            要发送的数据:
          </label>
          <textarea
            value={dataInput}
            onChange={(e) => setDataInput(e.target.value)}
            disabled={loading}
            rows={3}
            style={{
              width: '100%',
              padding: '8px',
              borderRadius: '4px',
              border: '1px solid #4b5563',
              background: '#1f2937',
              color: '#fff',
              fontFamily: 'monospace',
              resize: 'vertical',
            }}
          />
          <div style={{ fontSize: '12px', color: '#6b7280', marginTop: '4px' }}>
            字节数: {new TextEncoder().encode(dataInput).length}
          </div>
        </div>

        {/* 按钮组 */}
        <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
          <button
            onClick={handleConnect}
            disabled={loading || testConnectionState !== ConnectionState.DISCONNECTED}
            style={{
              padding: '8px 16px',
              borderRadius: '4px',
              border: 'none',
              background: testConnectionState !== ConnectionState.DISCONNECTED ? '#4b5563' : '#3b82f6',
              color: '#fff',
              cursor: testConnectionState !== ConnectionState.DISCONNECTED ? 'not-allowed' : 'pointer',
              fontSize: '14px',
            }}
          >
            {loading ? '连接中...' : '连接'}
          </button>

          <button
            onClick={handleRunTest}
            disabled={loading || testConnectionState !== ConnectionState.CONNECTED}
            style={{
              padding: '8px 16px',
              borderRadius: '4px',
              border: 'none',
              background: testConnectionState !== ConnectionState.CONNECTED ? '#4b5563' : '#10b981',
              color: '#fff',
              cursor: testConnectionState !== ConnectionState.CONNECTED ? 'not-allowed' : 'pointer',
              fontSize: '14px',
            }}
          >
            {loading ? '发送中...' : '发送流'}
          </button>

          <button
            onClick={handleDisconnect}
            disabled={loading || testConnectionState === ConnectionState.DISCONNECTED}
            style={{
              padding: '8px 16px',
              borderRadius: '4px',
              border: 'none',
              background: testConnectionState === ConnectionState.DISCONNECTED ? '#4b5563' : '#ef4444',
              color: '#fff',
              cursor: testConnectionState === ConnectionState.DISCONNECTED ? 'not-allowed' : 'pointer',
              fontSize: '14px',
            }}
          >
            断开连接
          </button>

          <button
            onClick={handleRunFullTest}
            disabled={loading || testConnectionState !== ConnectionState.DISCONNECTED}
            style={{
              padding: '8px 16px',
              borderRadius: '4px',
              border: 'none',
              background: testConnectionState !== ConnectionState.DISCONNECTED ? '#4b5563' : '#8b5cf6',
              color: '#fff',
              cursor: testConnectionState !== ConnectionState.DISCONNECTED ? 'not-allowed' : 'pointer',
              fontSize: '14px',
              fontWeight: 'bold',
            }}
          >
            {loading ? '测试中...' : '🚀 完整测试'}
          </button>
        </div>

        {/* 消息显示 */}
        {message && (
          <div style={{
            padding: '12px',
            borderRadius: '4px',
            background: message.startsWith('✓') ? '#064e3b' : message.startsWith('✗') ? '#7f1d1d' : '#1e3a8a',
            color: '#fff',
            fontSize: '14px',
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
          }}>
            {message}
          </div>
        )}

        {/* 测试结果 */}
        {testResult && (
          <div style={{
            padding: '12px',
            borderRadius: '4px',
            background: testResult.success ? '#064e3b' : '#7f1d1d',
            border: testResult.success ? '1px solid #10b981' : '1px solid #ef4444',
          }}>
            <div style={{ fontWeight: 'bold', marginBottom: '8px', color: '#fff' }}>
              {testResult.success ? '✓ 测试成功' : '✗ 测试失败'}
            </div>
            <div style={{ fontSize: '14px', color: '#e5e7eb', marginBottom: '4px' }}>
              {testResult.message}
            </div>
            <div style={{ fontSize: '12px', color: '#9ca3af' }}>
              时间: {testResult.timestamp.toLocaleTimeString()}
            </div>
          </div>
        )}

        {/* 说明 */}
        <div style={{
          padding: '12px',
          background: '#1e293b',
          borderRadius: '4px',
          fontSize: '13px',
          color: '#94a3b8',
        }}>
          <div style={{ fontWeight: 'bold', marginBottom: '8px', color: '#e2e8f0' }}>
            📖 使用说明
          </div>
          <ul style={{ margin: 0, paddingLeft: '20px' }}>
            <li>此面板独立于全局配置，可以连接到不同的 WebTransport 服务器</li>
            <li>支持完整 URL 或相对路径（相对路径将使用全局配置的服务器地址）</li>
            <li>对于自签名证书，需要在上方输入证书的 SHA-256 哈希值（十六进制格式）</li>
            <li>点击 "🚀 完整测试" 按钮可自动完成连接→发送→断开的完整流程</li>
            <li>或手动分步操作：先连接，再发送流，最后断开</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
