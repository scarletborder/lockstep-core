import { useState } from 'react';

export interface ClientConfig {
  serverUrl: string;
  allowSelfSigned: boolean;
  allowInsecureTransport: boolean;
  allowAnyCert: boolean;
  serverCertificateHashes: string;
}

interface ConfigPanelProps {
  config: ClientConfig;
  onConfigChange: (config: ClientConfig) => void;
}

export function ConfigPanel({ config, onConfigChange }: ConfigPanelProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  const handleChange = (key: keyof ClientConfig, value: string | boolean) => {
    onConfigChange({ ...config, [key]: value });
  };

  return (
    <div style={{
      position: 'fixed',
      top: '10px',
      right: '10px',
      background: '#1a1a1a',
      border: '1px solid #646cff',
      borderRadius: '8px',
      padding: '10px',
      minWidth: '250px',
      zIndex: 1000,
    }}>
      <div
        onClick={() => setIsExpanded(!isExpanded)}
        style={{
          cursor: 'pointer',
          fontWeight: 'bold',
          marginBottom: isExpanded ? '10px' : '0',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <span>⚙️ 配置</span>
        <span>{isExpanded ? '▼' : '▶'}</span>
      </div>

      {isExpanded && (
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <tbody>
            <tr>
              <td style={{ padding: '8px 12px 8px 0', fontSize: '12px', verticalAlign: 'middle', width: '100px' }}>
                服务器 URL
              </td>
              <td style={{ padding: '4px 0' }}>
                <input
                  type="text"
                  value={config.serverUrl}
                  onChange={(e) => handleChange('serverUrl', e.target.value)}
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
              <td style={{ padding: '8px 12px 8px 0', fontSize: '12px', verticalAlign: 'middle' }}>
                自签名证书
              </td>
              <td style={{ padding: '4px 0' }}>
                <label style={{ display: 'flex', alignItems: 'center', fontSize: '12px', cursor: 'pointer' }}>
                  <input
                    type="checkbox"
                    checked={config.allowSelfSigned}
                    onChange={(e) => handleChange('allowSelfSigned', e.target.checked)}
                    style={{ marginRight: '6px' }}
                  />
                  允许
                </label>
              </td>
            </tr>
            <tr>
              <td style={{ padding: '8px 12px 8px 0', fontSize: '12px', verticalAlign: 'middle' }}>
                不安全传输
              </td>
              <td style={{ padding: '4px 0' }}>
                <label style={{ display: 'flex', alignItems: 'center', fontSize: '12px', cursor: 'pointer' }}>
                  <input
                    type="checkbox"
                    checked={config.allowInsecureTransport}
                    onChange={(e) => handleChange('allowInsecureTransport', e.target.checked)}
                    style={{ marginRight: '6px' }}
                  />
                  允许
                </label>
              </td>
            </tr>
            <tr>
              <td style={{ padding: '8px 12px 8px 0', fontSize: '12px', verticalAlign: 'middle' }}>
                任何证书
              </td>
              <td style={{ padding: '4px 0' }}>
                <label style={{ display: 'flex', alignItems: 'center', fontSize: '12px', cursor: 'pointer' }}>
                  <input
                    type="checkbox"
                    checked={config.allowAnyCert}
                    onChange={(e) => handleChange('allowAnyCert', e.target.checked)}
                    style={{ marginRight: '6px' }}
                  />
                  允许
                </label>
              </td>
            </tr>
            <tr>
              <td colSpan={2} style={{ padding: '8px 0 4px 0', fontSize: '12px', color: '#aaa' }}>
                证书哈希 (SHA-256)
              </td>
            </tr>
            <tr>
              <td colSpan={2} style={{ padding: '4px 0' }}>
                <textarea
                  value={config.serverCertificateHashes}
                  onChange={(e) => handleChange('serverCertificateHashes', e.target.value)}
                  placeholder="每行一个哈希值&#10;例如: ABC123...&#10;或空格/冒号分隔"
                  style={{
                    width: '100%',
                    padding: '6px 8px',
                    background: '#2a2a2a',
                    border: '1px solid #555',
                    borderRadius: '4px',
                    color: 'white',
                    fontSize: '11px',
                    fontFamily: 'monospace',
                    minHeight: '60px',
                    resize: 'vertical',
                  }}
                />
              </td>
            </tr>
          </tbody>
        </table>
      )}
    </div>
  );
}
