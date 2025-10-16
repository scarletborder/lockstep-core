/**
 * WebTransport 长连接相关的类型定义和函数
 * 使用不可靠的 datagram 进行通信
 */

import { Request } from '../types/pb/request';
import { LobbyResponse, RoomResponse } from '../types/pb/response';

export interface SafetyOptions {
  allowSelfSigned?: boolean;
  allowInsecureTransport?: boolean;
  allowAnyCert?: boolean;
  /** 服务器证书哈希值（用于浏览器中信任自签名证书） */
  serverCertificateHashes?: string[];
}

// ============== 类型定义 ==============

/**
 * WebTransport 连接状态
 */
export enum ConnectionState {
  DISCONNECTED = 'disconnected',
  CONNECTING = 'connecting',
  LOBBY = 'lobby', // 等待加入房间响应
  CONNECTED = 'connected', // 已加入房间
  RECONNECTING = 'reconnecting',
  ERROR = 'error',
}

/**
 * 消息事件处理器
 */
export interface MessageHandlers {
  onLobbyResponse?: (response: LobbyResponse) => void;
  onRoomResponse?: (response: RoomResponse) => void;
  onError?: (error: Error) => void;
  onStateChange?: (state: ConnectionState) => void;
}

/**
 * 单向流选项
 */
export interface UnidirectionalStreamOptions {
  /** 要发送的数据 */
  data: Uint8Array;
  /** 是否等待流关闭 */
  waitForClose?: boolean;
}

// ============== WebTransport 客户端 ==============

/**
 * WebTransport 流客户端类
 */
export class StreamClient {
  private transport: WebTransport | null = null;
  private state: ConnectionState = ConnectionState.DISCONNECTED;
  private handlers: MessageHandlers = {};
  private encoder = new TextEncoder();
  private decoder = new TextDecoder();
  private readLoopRunning = false;
  private roomId: string | null = null;
  private reconnectKey: string | null = null;
  private myPlayerId: number | null = null;

  constructor(private serverUrl: string, private safety?: SafetyOptions) {
    // 如果在 Node 环境中允许自签名/任意证书，设置环境变量来放宽 TLS 校验
    if (typeof process !== 'undefined' && this.safety && (this.safety.allowSelfSigned || this.safety.allowAnyCert)) {
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';
    }
  }

  /**
   * 在运行时安全地获取 WebTransport 构造函数。
   * 在非浏览器环境（如 Node）中，这将抛出一个明确的错误，便于上层处理或做条件回退。
   */
  private getWebTransportConstructor(): typeof WebTransport {
    // 在浏览器中，WebTransport 通常存在于 globalThis 或 window 上
    // 使用 any 避免 TS 在非浏览器环境下抱怨类型不存在
    const g = globalThis as any;
    const WT = g.WebTransport ?? (typeof window !== 'undefined' ? (window as any).WebTransport : undefined);
    if (!WT) {
      throw new Error('WebTransport is not available in this environment. Ensure code runs in a browser that supports WebTransport.');
    }
    return WT as typeof WebTransport;
  }

  /**
   * 设置消息处理器
   */
  setHandlers(handlers: MessageHandlers): void {
    this.handlers = { ...this.handlers, ...handlers };
  }

  /**
   * 获取当前连接状态
   */
  getState(): ConnectionState {
    return this.state;
  }

  /**
   * 获取玩家 ID
   */
  getMyPlayerId(): number | null {
    return this.myPlayerId;
  }

  /**
   * 改变状态
   */
  private setState(newState: ConnectionState): void {
    if (this.state !== newState) {
      this.state = newState;
      this.handlers.onStateChange?.(newState);
    }
  }

  /**
   * 将十六进制字符串转换为 ArrayBuffer
   */
  private hexToArrayBuffer(hex: string): ArrayBuffer {
    // 移除可能的空格和冒号分隔符
    hex = hex.replace(/[\s:]/g, '');

    // 确保是偶数长度
    if (hex.length % 2 !== 0) {
      throw new Error('Invalid hex string');
    }

    const bytes = new Uint8Array(hex.length / 2);
    for (let i = 0; i < hex.length; i += 2) {
      bytes[i / 2] = parseInt(hex.substr(i, 2), 16);
    }
    return bytes.buffer;
  }

  /**
   * 加入房间
   */
  async joinRoom(roomId: string): Promise<void> {
    if (this.state !== ConnectionState.DISCONNECTED) {
      throw new Error('Already connected or connecting');
    }

    this.roomId = roomId;
    this.setState(ConnectionState.CONNECTING);

    try {
      // 建立 WebTransport 连接
      const url = `${this.serverUrl}/join/${roomId}`;
      const WT = this.getWebTransportConstructor();

      // 如果提供了证书哈希，使用它来信任自签名证书
      const options: WebTransportOptions = {};
      if (this.safety?.serverCertificateHashes && this.safety.serverCertificateHashes.length > 0) {
        options.serverCertificateHashes = this.safety.serverCertificateHashes.map(hash => ({
          algorithm: 'sha-256',
          value: this.hexToArrayBuffer(hash)
        }));
      }

      this.transport = new WT(url, options);

      // 等待连接建立
      await this.transport.ready;

      this.setState(ConnectionState.LOBBY);

      // 启动读取循环
      this.startReadLoop();

    } catch (error) {
      this.setState(ConnectionState.ERROR);
      const err = error instanceof Error ? error : new Error(String(error));
      this.handlers.onError?.(err);
      throw err;
    }
  }

  /**
   * 重连房间
   */
  async reconnectRoom(roomId: string, key: string): Promise<void> {
    if (this.state !== ConnectionState.DISCONNECTED) {
      throw new Error('Already connected or connecting');
    }

    this.roomId = roomId;
    this.reconnectKey = key;
    this.setState(ConnectionState.RECONNECTING);

    try {
      // 建立 WebTransport 连接
      // 注意：这里使用同样的 /join/ 端点，但需要在首个消息中带上重连信息
      // 根据你的服务端实现,可能需要调整
      const url = `${this.serverUrl}/join/${roomId}`;
      const WT = this.getWebTransportConstructor();

      // 如果提供了证书哈希，使用它来信任自签名证书
      const options: any = {};
      if (this.safety?.serverCertificateHashes && this.safety.serverCertificateHashes.length > 0) {
        options.serverCertificateHashes = this.safety.serverCertificateHashes.map(hash => ({
          algorithm: 'sha-256',
          value: this.hexToArrayBuffer(hash)
        }));
      }

      this.transport = new WT(url, options);

      // 等待连接建立
      await this.transport.ready;

      // TODO: 发送重连消息（需要根据服务端协议确定格式）
      // 这里假设需要发送一个包含 key 的特殊消息

      this.setState(ConnectionState.LOBBY);

      // 启动读取循环
      this.startReadLoop();

    } catch (error) {
      this.setState(ConnectionState.ERROR);
      const err = error instanceof Error ? error : new Error(String(error));
      this.handlers.onError?.(err);
      throw err;
    }
  }

  /**
   * 发送请求消息（通过 datagram）
   */
  async sendRequest(request: Request): Promise<void> {
    if (!this.transport) {
      throw new Error('Not connected');
    }

    if (this.state !== ConnectionState.CONNECTED) {
      throw new Error('Not in connected state');
    }

    try {
      // 序列化 protobuf 消息
      const bytes = Request.toBinary(request);

      // 使用 datagram 发送
      const writer = this.transport.datagrams.writable.getWriter();
      await writer.write(bytes);
      writer.releaseLock(); // 发送后立即释放锁，以便下次发送
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      this.handlers.onError?.(err);
      throw err;
    }
  }

  /**
   * 启动读取循环（从 datagram 接收消息）
   */
  private async startReadLoop(): Promise<void> {
    if (!this.transport || this.readLoopRunning) {
      return;
    }

    this.readLoopRunning = true;

    try {
      const reader = this.transport.datagrams.readable.getReader();

      while (this.readLoopRunning) {
        const { value, done } = await reader.read();

        if (done) {
          break;
        }

        // 处理接收到的数据
        this.handleIncomingMessage(value);
      }

      reader.releaseLock();
    } catch (error) {
      console.error('Error in read loop:', error);
      const err = error instanceof Error ? error : new Error(String(error));
      this.handlers.onError?.(err);
    } finally {
      this.readLoopRunning = false;
    }
  }

  /**
   * 处理接收到的消息
   */
  private handleIncomingMessage(data: Uint8Array): void {
    try {
      // 根据当前状态判断消息类型
      if (this.state === ConnectionState.LOBBY || this.state === ConnectionState.RECONNECTING) {
        // 在大厅状态，期待 LobbyResponse
        const lobbyResponse = LobbyResponse.fromBinary(data);

        // 处理加入房间的响应
        if (lobbyResponse.payload.oneofKind === 'joinRoomSuccess') {
          const success = lobbyResponse.payload.joinRoomSuccess;
          this.myPlayerId = success.myId;
          this.reconnectKey = success.key;
          this.setState(ConnectionState.CONNECTED);
          this.handlers.onLobbyResponse?.(lobbyResponse);
        } else if (lobbyResponse.payload.oneofKind === 'joinRoomFailed') {
          this.setState(ConnectionState.ERROR);
          this.handlers.onLobbyResponse?.(lobbyResponse);
        }
      } else if (this.state === ConnectionState.CONNECTED) {
        // 在房间状态，期待 RoomResponse
        const roomResponse = RoomResponse.fromBinary(data);
        this.handlers.onRoomResponse?.(roomResponse);
      }
    } catch (error) {
      console.error('Error parsing message:', error);
      const err = error instanceof Error ? error : new Error(String(error));
      this.handlers.onError?.(err);
    }
  }

  /**
   * 断开连接
   */
  async disconnect(): Promise<void> {
    this.readLoopRunning = false;

    if (this.transport) {
      try {
        await this.transport.close();
      } catch (error) {
        console.error('Error closing transport:', error);
      }
      this.transport = null;
    }

    this.setState(ConnectionState.DISCONNECTED);
    this.roomId = null;
    this.myPlayerId = null;
  }

  /**
   * 获取房间 ID
   */
  getRoomId(): string | null {
    return this.roomId;
  }

  /**
   * 获取重连密钥
   */
  getReconnectKey(): string | null {
    return this.reconnectKey;
  }

  /**
   * 创建单向流并发送数据
   * 参考示例服务端的 /unidirectional 端点
   */
  async createUnidirectionalStream(options: UnidirectionalStreamOptions): Promise<void> {
    if (!this.transport) {
      throw new Error('Not connected');
    }

    try {
      const stream = await this.transport.createUnidirectionalStream();
      const writer = stream.getWriter();
      await writer.write(options.data);

      if (options.waitForClose !== false) {
        await writer.close();
      }

      console.log('Unidirectional stream data sent successfully');
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      this.handlers.onError?.(err);
      throw err;
    }
  }

  /**
   * 批量创建单向流并发送数据
   * 用于测试多个并发流
   */
  async createMultipleUnidirectionalStreams(count: number, data: Uint8Array): Promise<void> {
    if (!this.transport) {
      throw new Error('Not connected');
    }

    const promises: Promise<void>[] = [];

    for (let i = 0; i < count; i++) {
      const promise = this.createUnidirectionalStream({ data, waitForClose: true })
        .then(() => {
          console.log(`Opened and closed stream ${i}.`);
        });
      promises.push(promise);
    }

    await Promise.all(promises);
  }

  /**
   * 建立到指定URL的原始 WebTransport 连接
   * 用于测试非标准端点（如示例服务端的 /unidirectional 端点）
   * @param endpoint 端点路径（相对路径如 '/unidirectional'）或完整 URL（如 'https://127.0.0.1:12345/unidirectional'）
   */
  async connectToEndpoint(endpoint: string): Promise<void> {
    if (this.transport) {
      throw new Error('Already connected. Please disconnect first.');
    }

    this.setState(ConnectionState.CONNECTING);

    try {
      // 判断是完整 URL 还是相对路径
      let url: string;
      if (endpoint.startsWith('http://') || endpoint.startsWith('https://')) {
        // 完整 URL，直接使用
        url = endpoint;
      } else {
        // 相对路径，拼接 serverUrl
        url = `${this.serverUrl}${endpoint}`;
      }

      const WT = this.getWebTransportConstructor();

      // 如果提供了证书哈希，使用它来信任自签名证书
      const options: WebTransportOptions = {};
      if (this.safety?.serverCertificateHashes && this.safety.serverCertificateHashes.length > 0) {
        options.serverCertificateHashes = this.safety.serverCertificateHashes.map(hash => ({
          algorithm: 'sha-256',
          value: this.hexToArrayBuffer(hash)
        }));
      }

      this.transport = new WT(url, options);

      // 设置连接关闭处理
      this.transport.closed
        .then(() => {
          console.log(`The HTTP/3 connection to ${url} closed gracefully.`);
        })
        .catch((error) => {
          console.error(`The HTTP/3 connection to ${url} closed due to ${error}.`);
        });

      // 等待连接建立
      await this.transport.ready;

      this.setState(ConnectionState.CONNECTED);
      console.log(`Connected to ${url}`);
    } catch (error) {
      this.setState(ConnectionState.ERROR);
      const err = error instanceof Error ? error : new Error(String(error));
      this.handlers.onError?.(err);
      throw err;
    }
  }
}
