import { useState } from 'react';
import { LockstepClient, ConnectionState, type Request } from 'lockstep-core-client';

interface GameRequestPanelProps {
  client: LockstepClient | null;
  connectionState: ConnectionState;
}

export function GameRequestPanel({ client, connectionState }: GameRequestPanelProps) {
  const [frameId, setFrameId] = useState(0);
  const [ackFrameId, setAckFrameId] = useState(0);
  const [isReady, setIsReady] = useState(false);
  const [isLoaded, setIsLoaded] = useState(false);
  const [chapterId, setChapterId] = useState(1);
  const [stageId, setStageId] = useState(1);
  const [plantRow, setPlantRow] = useState(0);
  const [plantCol, setPlantCol] = useState(0);
  const [plantCardId, setPlantCardId] = useState(1);
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);

  const isConnected = connectionState === ConnectionState.CONNECTED;

  const sendRequest = async (request: Request, description: string) => {
    if (!client) return;
    setLoading(true);
    setMessage('');
    try {
      await client.sendRequest(request);
      setMessage(`✓ ${description}`);
    } catch (error) {
      setMessage(`✗ 错误: ${(error as Error).message}`);
    } finally {
      setLoading(false);
    }
  };

  // 发送心跳/空白消息
  const handleBlank = () => {
    const request: Request = {
      payload: {
        oneofKind: 'blank',
        blank: { frameId, ackFrameId },
      },
    };
    sendRequest(request, '已发送心跳消息');
  };

  // 发送准备消息
  const handleReady = () => {
    const request: Request = {
      payload: {
        oneofKind: 'ready',
        ready: { isReady },
      },
    };
    sendRequest(request, `已发送准备状态: ${isReady ? '准备' : '取消准备'}`);
  };

  // 发送加载完成消息
  const handleLoaded = () => {
    const request: Request = {
      payload: {
        oneofKind: 'loaded',
        loaded: { isLoaded },
      },
    };
    sendRequest(request, `已发送加载状态: ${isLoaded ? '已加载' : '未加载'}`);
  };

  // 发送选择地图消息
  const handleChooseMap = () => {
    const request: Request = {
      payload: {
        oneofKind: 'chooseMap',
        chooseMap: { chapterId, stageId },
      },
    };
    sendRequest(request, `已选择地图: 章节${chapterId}-关卡${stageId}`);
  };

  // 发送离开选择地图消息
  const handleLeaveChooseMap = () => {
    const request: Request = {
      payload: {
        oneofKind: 'leaveChooseMap',
        leaveChooseMap: {},
      },
    };
    sendRequest(request, '已离开地图选择');
  };

  // 发送种植消息
  const handlePlant = () => {
    const request: Request = {
      payload: {
        oneofKind: 'plant',
        plant: {
          base: {
            base: { frameId, ackFrameId },
            row: plantRow,
            col: plantCol,
            processFrameId: frameId + 1,
          },
          pid: plantCardId,
          level: 1,
          cost: 50,
          energySum: 1000,
          starShardsSum: 100,
        },
      },
    };
    sendRequest(request, `已种植卡片${plantCardId}在(${plantRow}, ${plantCol})`);
  };

  // 发送移除植物消息
  const handleRemovePlant = () => {
    const request: Request = {
      payload: {
        oneofKind: 'removePlant',
        removePlant: {
          base: {
            base: { frameId, ackFrameId },
            row: plantRow,
            col: plantCol,
            processFrameId: frameId + 1,
          },
          pid: plantCardId,
        },
      },
    };
    sendRequest(request, `已移除植物在(${plantRow}, ${plantCol})`);
  };

  // 发送星星碎片消息
  const handleStarShards = () => {
    const request: Request = {
      payload: {
        oneofKind: 'starShards',
        starShards: {
          base: {
            base: { frameId, ackFrameId },
            row: plantRow,
            col: plantCol,
            processFrameId: frameId + 1,
          },
          pid: plantCardId,
          cost: 100,
          energySum: 1000,
          starShardsSum: 100,
        },
      },
    };
    sendRequest(request, '已发送星星碎片消息');
  };

  // 发送结束游戏消息
  const handleEndGame = () => {
    const request: Request = {
      payload: {
        oneofKind: 'endGame',
        endGame: {
          gameResult: 1, // 1 表示胜利
        },
      },
    };
    sendRequest(request, '已发送结束游戏消息');
  };

  return (
    <div style={{ border: '1px solid #646cff', borderRadius: '8px', padding: '16px', marginBottom: '16px' }}>
      <h3 style={{ marginTop: 0 }}>🎮 游戏请求测试</h3>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
        {/* 基础消息 */}
        <div>
          <h4 style={{ margin: '0 0 8px 0', fontSize: '14px' }}>基础消息</h4>
          <table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: '8px' }}>
            <tbody>
              <tr>
                <td style={{ padding: '4px 12px 4px 0', fontSize: '12px', verticalAlign: 'middle', width: '100px' }}>
                  帧 ID
                </td>
                <td style={{ padding: '4px 0' }}>
                  <input
                    type="number"
                    value={frameId}
                    onChange={(e) => setFrameId(Number(e.target.value))}
                    disabled={!isConnected}
                    style={{
                      width: '100%',
                      padding: '4px 8px',
                      background: '#2a2a2a',
                      border: '1px solid #555',
                      borderRadius: '4px',
                      color: 'white',
                      fontSize: '12px',
                    }}
                  />
                </td>
              </tr>
              <tr>
                <td style={{ padding: '4px 12px 4px 0', fontSize: '12px', verticalAlign: 'middle' }}>
                  确认帧 ID
                </td>
                <td style={{ padding: '4px 0' }}>
                  <input
                    type="number"
                    value={ackFrameId}
                    onChange={(e) => setAckFrameId(Number(e.target.value))}
                    disabled={!isConnected}
                    style={{
                      width: '100%',
                      padding: '4px 8px',
                      background: '#2a2a2a',
                      border: '1px solid #555',
                      borderRadius: '4px',
                      color: 'white',
                      fontSize: '12px',
                    }}
                  />
                </td>
              </tr>
            </tbody>
          </table>
          <button
            onClick={handleBlank}
            disabled={loading || !isConnected}
            style={{ width: '100%', padding: '4px 12px', fontSize: '12px' }}
          >
            发送心跳
          </button>
        </div>

        {/* 房间准备 */}
        <div>
          <h4 style={{ margin: '0 0 8px 0', fontSize: '14px' }}>房间准备</h4>
          <div style={{ display: 'flex', gap: '8px' }}>
            <label style={{ flex: 1, display: 'flex', alignItems: 'center', fontSize: '12px', cursor: 'pointer' }}>
              <input
                type="checkbox"
                checked={isReady}
                onChange={(e) => setIsReady(e.target.checked)}
                disabled={!isConnected}
                style={{ marginRight: '6px' }}
              />
              准备状态
            </label>
            <button
              onClick={handleReady}
              disabled={loading || !isConnected}
              style={{ padding: '4px 12px', fontSize: '12px' }}
            >
              发送准备
            </button>
          </div>
        </div>

        {/* 地图选择 */}
        <div>
          <h4 style={{ margin: '0 0 8px 0', fontSize: '14px' }}>地图选择</h4>
          <table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: '8px' }}>
            <tbody>
              <tr>
                <td style={{ padding: '4px 12px 4px 0', fontSize: '12px', verticalAlign: 'middle', width: '100px' }}>
                  章节 ID
                </td>
                <td style={{ padding: '4px 0' }}>
                  <input
                    type="number"
                    value={chapterId}
                    onChange={(e) => setChapterId(Number(e.target.value))}
                    disabled={!isConnected}
                    style={{
                      width: '100%',
                      padding: '4px 8px',
                      background: '#2a2a2a',
                      border: '1px solid #555',
                      borderRadius: '4px',
                      color: 'white',
                      fontSize: '12px',
                    }}
                  />
                </td>
              </tr>
              <tr>
                <td style={{ padding: '4px 12px 4px 0', fontSize: '12px', verticalAlign: 'middle' }}>
                  关卡 ID
                </td>
                <td style={{ padding: '4px 0' }}>
                  <input
                    type="number"
                    value={stageId}
                    onChange={(e) => setStageId(Number(e.target.value))}
                    disabled={!isConnected}
                    style={{
                      width: '100%',
                      padding: '4px 8px',
                      background: '#2a2a2a',
                      border: '1px solid #555',
                      borderRadius: '4px',
                      color: 'white',
                      fontSize: '12px',
                    }}
                  />
                </td>
              </tr>
            </tbody>
          </table>
          <div style={{ display: 'flex', gap: '8px' }}>
            <button
              onClick={handleChooseMap}
              disabled={loading || !isConnected}
              style={{ flex: 1, padding: '4px 12px', fontSize: '12px' }}
            >
              选择地图
            </button>
            <button
              onClick={handleLeaveChooseMap}
              disabled={loading || !isConnected}
              style={{ flex: 1, padding: '4px 12px', fontSize: '12px' }}
            >
              离开选择
            </button>
          </div>
        </div>

        {/* 加载状态 */}
        <div>
          <h4 style={{ margin: '0 0 8px 0', fontSize: '14px' }}>加载状态</h4>
          <div style={{ display: 'flex', gap: '8px' }}>
            <label style={{ flex: 1, display: 'flex', alignItems: 'center', fontSize: '12px', cursor: 'pointer' }}>
              <input
                type="checkbox"
                checked={isLoaded}
                onChange={(e) => setIsLoaded(e.target.checked)}
                disabled={!isConnected}
                style={{ marginRight: '6px' }}
              />
              加载完成
            </label>
            <button
              onClick={handleLoaded}
              disabled={loading || !isConnected}
              style={{ padding: '4px 12px', fontSize: '12px' }}
            >
              发送加载
            </button>
          </div>
        </div>

        {/* 游戏操作 */}
        <div>
          <h4 style={{ margin: '0 0 8px 0', fontSize: '14px' }}>游戏操作</h4>
          <table style={{ width: '100%', borderCollapse: 'collapse', marginBottom: '8px' }}>
            <tbody>
              <tr>
                <td style={{ padding: '4px 12px 4px 0', fontSize: '12px', verticalAlign: 'middle', width: '100px' }}>
                  行
                </td>
                <td style={{ padding: '4px 0' }}>
                  <input
                    type="number"
                    value={plantRow}
                    onChange={(e) => setPlantRow(Number(e.target.value))}
                    disabled={!isConnected}
                    style={{
                      width: '100%',
                      padding: '4px 8px',
                      background: '#2a2a2a',
                      border: '1px solid #555',
                      borderRadius: '4px',
                      color: 'white',
                      fontSize: '12px',
                    }}
                  />
                </td>
              </tr>
              <tr>
                <td style={{ padding: '4px 12px 4px 0', fontSize: '12px', verticalAlign: 'middle' }}>
                  列
                </td>
                <td style={{ padding: '4px 0' }}>
                  <input
                    type="number"
                    value={plantCol}
                    onChange={(e) => setPlantCol(Number(e.target.value))}
                    disabled={!isConnected}
                    style={{
                      width: '100%',
                      padding: '4px 8px',
                      background: '#2a2a2a',
                      border: '1px solid #555',
                      borderRadius: '4px',
                      color: 'white',
                      fontSize: '12px',
                    }}
                  />
                </td>
              </tr>
              <tr>
                <td style={{ padding: '4px 12px 4px 0', fontSize: '12px', verticalAlign: 'middle' }}>
                  卡片 ID
                </td>
                <td style={{ padding: '4px 0' }}>
                  <input
                    type="number"
                    value={plantCardId}
                    onChange={(e) => setPlantCardId(Number(e.target.value))}
                    disabled={!isConnected}
                    style={{
                      width: '100%',
                      padding: '4px 8px',
                      background: '#2a2a2a',
                      border: '1px solid #555',
                      borderRadius: '4px',
                      color: 'white',
                      fontSize: '12px',
                    }}
                  />
                </td>
              </tr>
            </tbody>
          </table>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            <div style={{ display: 'flex', gap: '8px' }}>
              <button
                onClick={handlePlant}
                disabled={loading || !isConnected}
                style={{ flex: 1, padding: '4px 12px', fontSize: '12px' }}
              >
                种植
              </button>
              <button
                onClick={handleRemovePlant}
                disabled={loading || !isConnected}
                style={{ flex: 1, padding: '4px 12px', fontSize: '12px' }}
              >
                移除植物
              </button>
            </div>
            <div style={{ display: 'flex', gap: '8px' }}>
              <button
                onClick={handleStarShards}
                disabled={loading || !isConnected}
                style={{ flex: 1, padding: '4px 12px', fontSize: '12px' }}
              >
                星星碎片
              </button>
              <button
                onClick={handleEndGame}
                disabled={loading || !isConnected}
                style={{ flex: 1, padding: '4px 12px', fontSize: '12px' }}
              >
                结束游戏
              </button>
            </div>
          </div>
        </div>

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
