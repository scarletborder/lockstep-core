import { useState, useEffect } from 'react';
import { LockstepClient, ConnectionState, type LobbyResponse, type RoomResponse } from 'lockstep-core-client';
import './App.css';
import { ConfigPanel, type ClientConfig } from './components/ConfigPanel';
import { HTTPTestPanel } from './components/HTTPTestPanel';
import { ConnectionPanel } from './components/ConnectionPanel';
import { GameRequestPanel } from './components/GameRequestPanel';
import { MessageLogPanel, type LogMessage } from './components/MessageLogPanel';
import { QuickTestPanel } from './components/QuickTestPanel';
import { UnidirectionalTestPanel } from './components/UnidirectionalTestPanel';

function App() {
  // 配置状态
  const [config, setConfig] = useState<ClientConfig>({
    serverUrl: 'https://127.0.0.1:4433',
    allowSelfSigned: true,
    allowInsecureTransport: true,
    allowAnyCert: true,
    serverCertificateHashes: '',
  });

  // 客户端实例
  const [client, setClient] = useState<LockstepClient | null>(null);
  const [connectionState, setConnectionState] = useState<ConnectionState>(ConnectionState.DISCONNECTED);

  // 消息日志
  const [messages, setMessages] = useState<LogMessage[]>([]);

  // 初始化客户端
  useEffect(() => {
    // 解析证书哈希（每行一个或空格/冒号分隔）
    const hashes = config.serverCertificateHashes
      .split(/[\n,]/)
      .map(h => h.trim())
      .filter(h => h.length > 0);

    const newClient = new LockstepClient({
      serverUrl: config.serverUrl,
      safety: {
        allowSelfSigned: config.allowSelfSigned,
        allowInsecureTransport: config.allowInsecureTransport,
        allowAnyCert: config.allowAnyCert,
        serverCertificateHashes: hashes.length > 0 ? hashes : undefined,
      },
    });

    // 设置消息处理器
    newClient.setMessageHandlers({
      onLobbyResponse: (response: LobbyResponse) => {
        addLog('lobby', formatLobbyResponse(response));
      },
      onRoomResponse: (response: RoomResponse) => {
        addLog('room', formatRoomResponse(response));
      },
      onError: (error: Error) => {
        addLog('error', error.message);
      },
      onStateChange: (state: ConnectionState) => {
        setConnectionState(state);
        addLog('state', `连接状态变化: ${state}`);
      },
    });

    setClient(newClient);

    return () => {
      newClient.disconnect();
    };
  }, [config]);

  // 添加日志
  const addLog = (type: LogMessage['type'], content: string) => {
    setMessages(prev => [...prev, {
      timestamp: new Date(),
      type,
      content,
    }]);
  };

  // 清空日志
  const clearLogs = () => {
    setMessages([]);
  };

  // 格式化大厅响应
  const formatLobbyResponse = (response: LobbyResponse): string => {
    if (response.payload.oneofKind === 'joinRoomSuccess') {
      const data = response.payload.joinRoomSuccess;
      return `✓ 加入房间成功\n房间ID: ${data.roomId}\n玩家ID: ${data.myId}\n密钥: ${data.key}\n消息: ${data.message}`;
    } else if (response.payload.oneofKind === 'joinRoomFailed') {
      const data = response.payload.joinRoomFailed;
      return `✗ 加入房间失败\n${data.message}`;
    }
    return JSON.stringify(response.payload, null, 2);
  };

  // 格式化房间响应
  const formatRoomResponse = (response: RoomResponse): string => {
    const kind = response.payload.oneofKind;
    if (!kind) return '未知消息类型';

    const typeNames: Record<string, string> = {
      roomInfo: '房间信息',
      chooseMap: '选择地图',
      quitChooseMap: '退出选择地图',
      roomClosed: '房间关闭',
      gameEnd: '游戏结束',
      error: '错误',
      updateReadyCount: '更新准备数',
      allReady: '全部准备',
      allLoaded: '全部加载完成',
      frameData: '帧数据',
    };

    const typeName = typeNames[kind] || kind;
    const data = response.payload[kind as keyof typeof response.payload];

    return `[${typeName}]\n${JSON.stringify(data, null, 2)}`;
  };

  return (
    <div style={{
      minHeight: '100vh',
      padding: '20px',
      background: '#0a0a0a',
    }}>
      {/* 配置面板 */}
      <ConfigPanel config={config} onConfigChange={setConfig} />

      {/* 主标题 */}
      <div style={{ textAlign: 'center', marginBottom: '32px' }}>
        <h1 style={{ margin: '0 0 8px 0' }}>🎮 Lockstep Client 测试面板</h1>
        <p style={{ margin: 0, color: '#888', fontSize: '14px' }}>
          完整测试 WebTransport 和 HTTP 功能
        </p>
      </div>

      {/* 主内容区域 */}
      <div style={{
        maxWidth: '1400px',
        margin: '0 auto',
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))',
        gap: '20px',
      }}>
        {/* 左侧列 */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
          <HTTPTestPanel client={client} />
          <ConnectionPanel
            client={client}
            connectionState={connectionState}
            onStateChange={setConnectionState}
          />
          <QuickTestPanel
            client={client}
            connectionState={connectionState}
            onLog={addLog}
          />
        </div>

        {/* 中间列 */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
          <GameRequestPanel
            client={client}
            connectionState={connectionState}
          />
          <UnidirectionalTestPanel
            client={client}
            connectionState={connectionState}
          />
        </div>

        {/* 右侧列 */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
          <MessageLogPanel messages={messages} onClear={clearLogs} />
        </div>
      </div>

      {/* 页脚 */}
      <div style={{
        textAlign: 'center',
        marginTop: '32px',
        padding: '16px',
        color: '#666',
        fontSize: '12px',
      }}>
        <p>Lockstep Core Testing Suite</p>
        <p>基于 WebTransport 的帧同步游戏服务器测试工具</p>
      </div>
    </div>
  );
}

export default App;
