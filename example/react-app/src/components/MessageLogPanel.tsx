import { useEffect, useRef } from 'react';

export interface LogMessage {
  timestamp: Date;
  type: 'lobby' | 'room' | 'error' | 'state' | 'info';
  content: string;
}

interface MessageLogPanelProps {
  messages: LogMessage[];
  onClear: () => void;
}

export function MessageLogPanel({ messages, onClear }: MessageLogPanelProps) {
  const logEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    logEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const getTypeColor = (type: LogMessage['type']) => {
    switch (type) {
      case 'lobby':
        return '#3b82f6';
      case 'room':
        return '#8b5cf6';
      case 'error':
        return '#ef4444';
      case 'state':
        return '#eab308';
      case 'info':
        return '#6b7280';
      default:
        return '#6b7280';
    }
  };

  const getTypeLabel = (type: LogMessage['type']) => {
    const labels = {
      lobby: '大厅',
      room: '房间',
      error: '错误',
      state: '状态',
      info: '信息',
    };
    return labels[type] || type;
  };

  return (
    <div style={{ border: '1px solid #646cff', borderRadius: '8px', padding: '16px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
        <h3 style={{ margin: 0 }}>📋 消息日志</h3>
        <button
          onClick={onClear}
          style={{ padding: '4px 12px', fontSize: '12px' }}
        >
          清空
        </button>
      </div>

      <div style={{
        background: '#1a1a1a',
        borderRadius: '4px',
        padding: '8px',
        maxHeight: '400px',
        overflowY: 'auto',
        fontSize: '12px',
        fontFamily: 'monospace',
      }}>
        {messages.length === 0 ? (
          <div style={{ color: '#6b7280', textAlign: 'center', padding: '20px' }}>
            暂无消息
          </div>
        ) : (
          messages.map((msg, index) => (
            <div
              key={index}
              style={{
                padding: '6px 8px',
                marginBottom: '4px',
                background: '#2a2a2a',
                borderRadius: '4px',
                borderLeft: `3px solid ${getTypeColor(msg.type)}`,
              }}
            >
              <div style={{ display: 'flex', gap: '8px', marginBottom: '4px' }}>
                <span style={{ color: '#6b7280' }}>
                  {msg.timestamp.toLocaleTimeString()}
                </span>
                <span style={{
                  color: getTypeColor(msg.type),
                  fontWeight: 'bold',
                }}>
                  [{getTypeLabel(msg.type)}]
                </span>
              </div>
              <div style={{ color: '#e5e7eb', whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                {msg.content}
              </div>
            </div>
          ))
        )}
        <div ref={logEndRef} />
      </div>
    </div>
  );
}
