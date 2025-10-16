import { useState } from 'react';
import { LockstepClient, ConnectionState } from 'lockstep-core-client';

interface ConnectionPanelProps {
  client: LockstepClient | null;
  connectionState: ConnectionState;
  onStateChange: (state: ConnectionState) => void;
}

export function ConnectionPanel({ client, connectionState }: ConnectionPanelProps) {
  const [roomId, setRoomId] = useState('test-room-001');
  const [reconnectKey, setReconnectKey] = useState('');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);

  const handleJoinRoom = async () => {
    if (!client || !roomId) return;
    setLoading(true);
    setMessage('');
    try {
      await client.joinRoom(roomId);
      setMessage(`✓ 正在连接到房间 ${roomId}...`);
    } catch (error) {
      setMessage(`✗ 错误: ${(error as Error).message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleReconnect = async () => {
    if (!client || !roomId || !reconnectKey) return;
    setLoading(true);
    setMessage('');
    try {
      await client.reconnectRoom(roomId, reconnectKey);
      setMessage(`✓ 正在重连到房间 ${roomId}...`);
    } catch (error) {
      setMessage(`✗ 错误: ${(error as Error).message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleDisconnect = async () => {
    if (!client) return;
    setLoading(true);
    setMessage('');
    try {
      // 保存重连密钥
      const key = client.getReconnectKey();
      if (key) {
        setReconnectKey(key);
      }
      await client.disconnect();
      setMessage('✓ 已断开连接');
    } catch (error) {
      setMessage(`✗ 错误: ${(error as Error).message}`);
    } finally {
      setLoading(false);
    }
  };

  const getStateColor = (state: ConnectionState) => {
    switch (state) {
      case ConnectionState.CONNECTED:
        return '#22c55e';
      case ConnectionState.CONNECTING:
      case ConnectionState.RECONNECTING:
      case ConnectionState.LOBBY:
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

  const currentPlayerId = client?.getMyPlayerId();
  const currentRoomId = client?.getCurrentRoomId();

  return (
    <div style={{ border: '1px solid #646cff', borderRadius: '8px', padding: '16px', marginBottom: '16px' }}>
      <h3 style={{ marginTop: 0 }}>🔌 WebTransport 连接</h3>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        {/* 连接状态 */}
        <div style={{
          background: '#2a2a2a',
          padding: '12px',
          borderRadius: '4px',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}>
          <span style={{ fontSize: '12px' }}>连接状态</span>
          <span style={{
            padding: '4px 12px',
            borderRadius: '12px',
            background: getStateColor(connectionState) + '22',
            color: getStateColor(connectionState),
            fontSize: '12px',
            fontWeight: 'bold',
          }}>
            {getStateText(connectionState)}
          </span>
        </div>

        {/* 当前信息 */}
        {(currentPlayerId !== null || currentRoomId !== null) && (
          <div style={{
            background: '#2a2a2a',
            padding: '8px',
            borderRadius: '4px',
            fontSize: '12px',
          }}>
            {currentRoomId && <div>房间: {currentRoomId}</div>}
            {currentPlayerId !== null && <div>玩家 ID: {currentPlayerId}</div>}
          </div>
        )}

        {/* 房间 ID 输入 */}
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <tbody>
            <tr>
              <td style={{ padding: '8px 12px 8px 0', fontSize: '12px', verticalAlign: 'middle', width: '80px' }}>
                房间 ID
              </td>
              <td style={{ padding: '4px 0' }}>
                <input
                  type="text"
                  value={roomId}
                  onChange={(e) => setRoomId(e.target.value)}
                  disabled={loading || connectionState !== ConnectionState.DISCONNECTED}
                  style={{
                    width: '100%',
                    padding: '8px',
                    background: '#2a2a2a',
                    border: '1px solid #555',
                    borderRadius: '4px',
                    color: 'white',
                  }}
                />
              </td>
            </tr>
          </tbody>
        </table>

        {/* 连接按钮 */}
        <div style={{ display: 'flex', gap: '8px' }}>
          <button
            onClick={handleJoinRoom}
            disabled={loading || !client || !roomId || connectionState !== ConnectionState.DISCONNECTED}
            style={{ flex: 1, padding: '8px 16px' }}
          >
            加入房间
          </button>
          <button
            onClick={handleDisconnect}
            disabled={loading || !client || connectionState === ConnectionState.DISCONNECTED}
            style={{ flex: 1, padding: '8px 16px' }}
          >
            断开连接
          </button>
        </div>

        {/* 重连功能 */}
        {reconnectKey && (
          <div style={{
            background: '#2a2a2a',
            padding: '12px',
            borderRadius: '4px',
          }}>
            <div style={{ fontSize: '12px', marginBottom: '8px' }}>
              重连密钥: <code style={{ fontSize: '10px' }}>{reconnectKey}</code>
            </div>
            <button
              onClick={handleReconnect}
              disabled={loading || !client || connectionState !== ConnectionState.DISCONNECTED}
              style={{ width: '100%', padding: '8px 16px' }}
            >
              使用密钥重连
            </button>
          </div>
        )}

        {/* 消息提示 */}
        {message && (
          <div style={{
            padding: '8px',
            background: message.startsWith('✓') ? '#1a4d1a' : '#4d1a1a',
            borderRadius: '4px',
            fontSize: '12px',
          }}>
            {message}
          </div>
        )}
      </div>
    </div>
  );
}
