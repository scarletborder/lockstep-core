import { useState } from 'react';
import { LockstepClient, ConnectionState } from 'lockstep-core-client';

interface QuickTestPanelProps {
  client: LockstepClient | null;
  connectionState: ConnectionState;
  onLog: (type: 'lobby' | 'room' | 'error' | 'state' | 'info', content: string) => void;
}

export function QuickTestPanel({ client, connectionState, onLog }: QuickTestPanelProps) {
  const [loading, setLoading] = useState(false);
  const [testResult, setTestResult] = useState('');

  const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

  // 快速测试：完整流程
  const handleFullFlowTest = async () => {
    if (!client) return;
    setLoading(true);
    setTestResult('');

    try {
      const roomId = `test-${Date.now()}`;

      onLog('info', '=== 开始完整流程测试 ===');

      // 1. 创建房间
      onLog('info', '步骤 1/6: 创建房间...');
      await client.createRoom(roomId);
      await sleep(500);

      // 2. 加入房间
      onLog('info', '步骤 2/6: 加入房间...');
      await client.joinRoom(roomId);
      await sleep(2000);

      if (connectionState !== ConnectionState.CONNECTED) {
        throw new Error('未能成功连接');
      }

      // 3. 发送准备消息
      onLog('info', '步骤 3/6: 发送准备消息...');
      await client.sendRequest({
        payload: { oneofKind: 'ready', ready: { isReady: true } },
      });
      await sleep(500);

      // 4. 选择地图
      onLog('info', '步骤 4/6: 选择地图...');
      await client.sendRequest({
        payload: { oneofKind: 'chooseMap', chooseMap: { chapterId: 1, stageId: 1 } },
      });
      await sleep(500);

      // 5. 发送加载完成
      onLog('info', '步骤 5/6: 发送加载完成...');
      await client.sendRequest({
        payload: { oneofKind: 'loaded', loaded: { isLoaded: true } },
      });
      await sleep(500);

      // 6. 断开连接
      onLog('info', '步骤 6/6: 断开连接...');
      await client.disconnect();

      setTestResult('✓ 完整流程测试成功！');
      onLog('info', '=== 完整流程测试完成 ===');
    } catch (error) {
      const errMsg = `✗ 测试失败: ${(error as Error).message}`;
      setTestResult(errMsg);
      onLog('error', errMsg);
    } finally {
      setLoading(false);
    }
  };

  // 快速测试：重连流程
  const handleReconnectTest = async () => {
    if (!client) return;
    setLoading(true);
    setTestResult('');

    try {
      const roomId = `reconnect-test-${Date.now()}`;

      onLog('info', '=== 开始重连测试 ===');

      // 1. 创建并加入房间
      onLog('info', '步骤 1/4: 创建并加入房间...');
      await client.createRoom(roomId);
      await client.joinRoom(roomId);
      await sleep(2000);

      // 2. 保存重连密钥
      const key = client.getReconnectKey();
      if (!key) {
        throw new Error('未能获取重连密钥');
      }
      onLog('info', `步骤 2/4: 获取重连密钥: ${key}`);

      // 3. 断开连接
      onLog('info', '步骤 3/4: 断开连接...');
      await client.disconnect();
      await sleep(1000);

      // 4. 重连
      onLog('info', '步骤 4/4: 使用密钥重连...');
      await client.reconnectRoom(roomId, key);
      await sleep(2000);

      if (connectionState === ConnectionState.CONNECTED) {
        setTestResult('✓ 重连测试成功！');
        onLog('info', '=== 重连测试完成 ===');
      } else {
        throw new Error('重连后未能成功连接');
      }

      // 清理
      await client.disconnect();
    } catch (error) {
      const errMsg = `✗ 测试失败: ${(error as Error).message}`;
      setTestResult(errMsg);
      onLog('error', errMsg);
    } finally {
      setLoading(false);
    }
  };

  // 快速测试：消息压力测试
  const handleStressTest = async () => {
    if (!client || connectionState !== ConnectionState.CONNECTED) {
      setTestResult('✗ 请先连接到房间');
      return;
    }

    setLoading(true);
    setTestResult('');

    try {
      onLog('info', '=== 开始消息压力测试 ===');

      const messageCount = 50;
      onLog('info', `发送 ${messageCount} 条心跳消息...`);

      for (let i = 0; i < messageCount; i++) {
        await client.sendRequest({
          payload: {
            oneofKind: 'blank',
            blank: { frameId: i, ackFrameId: i - 1 },
          },
        });

        if (i % 10 === 0) {
          onLog('info', `已发送 ${i + 1}/${messageCount} 条消息`);
        }

        // 稍微延迟以避免过载
        await sleep(50);
      }

      setTestResult(`✓ 压力测试完成！成功发送 ${messageCount} 条消息`);
      onLog('info', '=== 消息压力测试完成 ===');
    } catch (error) {
      const errMsg = `✗ 测试失败: ${(error as Error).message}`;
      setTestResult(errMsg);
      onLog('error', errMsg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ border: '1px solid #646cff', borderRadius: '8px', padding: '16px', marginBottom: '16px' }}>
      <h3 style={{ marginTop: 0 }}>⚡ 快速测试</h3>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
        <button
          onClick={handleFullFlowTest}
          disabled={loading || !client || connectionState !== ConnectionState.DISCONNECTED}
          style={{ padding: '12px', fontSize: '14px' }}
        >
          {loading ? '测试中...' : '完整流程测试'}
        </button>

        <button
          onClick={handleReconnectTest}
          disabled={loading || !client || connectionState !== ConnectionState.DISCONNECTED}
          style={{ padding: '12px', fontSize: '14px' }}
        >
          {loading ? '测试中...' : '重连流程测试'}
        </button>

        <button
          onClick={handleStressTest}
          disabled={loading || !client || connectionState !== ConnectionState.CONNECTED}
          style={{ padding: '12px', fontSize: '14px' }}
        >
          {loading ? '测试中...' : '消息压力测试 (需已连接)'}
        </button>

        {testResult && (
          <div style={{
            padding: '12px',
            background: testResult.startsWith('✓') ? '#1a4d1a' : '#4d1a1a',
            borderRadius: '4px',
            fontSize: '14px',
            fontWeight: 'bold',
          }}>
            {testResult}
          </div>
        )}
      </div>
    </div>
  );
}
