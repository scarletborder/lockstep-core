import { useState } from 'react';
import { LockstepClient } from 'lockstep-core-client';

interface HTTPTestPanelProps {
  client: LockstepClient | null;
}

export function HTTPTestPanel({ client }: HTTPTestPanelProps) {
  const [roomId, setRoomId] = useState('test-room-001');
  const [rooms, setRooms] = useState<string[]>([]);
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);

  const handleListRooms = async () => {
    if (!client) return;
    setLoading(true);
    setMessage('');
    try {
      const roomList = await client.listRooms();
      setRooms(roomList);
      setMessage(`✓ 找到 ${roomList.length} 个房间`);
    } catch (error) {
      setMessage(`✗ 错误: ${(error as Error).message}`);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateRoom = async () => {
    if (!client || !roomId) return;
    setLoading(true);
    setMessage('');
    try {
      const result = await client.createRoom(roomId);
      setMessage(result.success ? `✓ ${result.message}` : `✗ ${result.message}`);
      // 创建成功后刷新列表
      if (result.success) {
        await handleListRooms();
      }
    } catch (error) {
      setMessage(`✗ 错误: ${(error as Error).message}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ border: '1px solid #646cff', borderRadius: '8px', padding: '16px', marginBottom: '16px' }}>
      <h3 style={{ marginTop: 0 }}>🌐 HTTP 请求测试</h3>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
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
                  disabled={loading}
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

        <div style={{ display: 'flex', gap: '8px' }}>
          <button
            onClick={handleListRooms}
            disabled={loading || !client}
            style={{ flex: 1, padding: '8px 16px' }}
          >
            获取房间列表
          </button>
          <button
            onClick={handleCreateRoom}
            disabled={loading || !client || !roomId}
            style={{ flex: 1, padding: '8px 16px' }}
          >
            创建房间
          </button>
        </div>

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

        {rooms.length > 0 && (
          <div style={{
            background: '#2a2a2a',
            borderRadius: '4px',
            padding: '8px',
            maxHeight: '150px',
            overflowY: 'auto',
          }}>
            <div style={{ fontSize: '12px', fontWeight: 'bold', marginBottom: '8px' }}>
              房间列表 ({rooms.length})
            </div>
            {rooms.map((room, index) => (
              <div key={index} style={{ fontSize: '12px', padding: '4px 0' }}>
                • {room}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
