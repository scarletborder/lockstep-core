import { HTTPClient } from './requests/common';
import { StreamClient, MessageHandlers, ConnectionState } from './requests/stream';
import { Request } from './types/pb/request';
import { LobbyResponse, RoomResponse } from './types/pb/response';

/**
 * 初始化选项
 */
export interface InitOptions {
  serverUrl: string;
}

/**
 * 非安全选项（用于测试/本地开发，允许自签名证书等）
 */
export interface SafetyOptions {
  /** 允许自签名证书（在 Node 环境下会设置 NODE_TLS_REJECT_UNAUTHORIZED=0） */
  allowSelfSigned?: boolean;
  /** 允许不安全的传输（例如在某些实现中允许 ws:// 或 http://，仅作占位） */
  allowInsecureTransport?: boolean;
  /** 更广泛地放宽证书校验（等同于 allowSelfSigned，但命名更笼统一些） */
  allowAnyCert?: boolean;
}

/**
 * Lockstep 客户端主类
 * 提供统一的接口来管理普通 HTTP 请求和 WebTransport 长连接
 */
export class LockstepClient {
  private httpClient: HTTPClient;
  private streamClient: StreamClient;

  constructor(private options: InitOptions & { safety?: SafetyOptions }) {
    // 初始化 HTTP 客户端
    this.httpClient = new HTTPClient(options.serverUrl, options.safety);

    // 初始化 WebTransport 客户端
    this.streamClient = new StreamClient(options.serverUrl, options.safety);
  }

  // ============== 普通 HTTP 请求方法 ==============

  /**
   * 获取房间列表
   * @returns 房间 ID 列表
   */
  async listRooms(): Promise<string[]> {
    return this.httpClient.listRooms();
  }

  /**
   * 创建房间
   * @param roomId 房间 ID
   * @returns 创建结果
   */
  async createRoom(roomId: string): Promise<{ success: boolean; message: string }> {
    return this.httpClient.createRoom(roomId);
  }

  // ============== WebTransport 长连接方法 ==============

  /**
   * 加入房间（建立 WebTransport 连接）
   * @param roomId 房间 ID
   */
  async joinRoom(roomId: string): Promise<void> {
    await this.streamClient.joinRoom(roomId);
  }

  /**
   * 重连房间（使用之前的密钥）
   * @param roomId 房间 ID
   * @param key 重连密钥
   */
  async reconnectRoom(roomId: string, key: string): Promise<void> {
    await this.streamClient.reconnectRoom(roomId, key);
  }

  /**
   * 发送游戏请求消息
   * @param request protobuf 请求消息
   */
  async sendRequest(request: Request): Promise<void> {
    await this.streamClient.sendRequest(request);
  }

  /**
   * 设置消息处理器
   * @param handlers 消息处理器对象
   */
  setMessageHandlers(handlers: MessageHandlers): void {
    this.streamClient.setHandlers(handlers);
  }

  /**
   * 断开连接
   */
  async disconnect(): Promise<void> {
    await this.streamClient.disconnect();
  }

  // ============== 状态查询方法 ==============

  /**
   * 获取当前连接状态
   */
  getConnectionState(): ConnectionState {
    return this.streamClient.getState();
  }

  /**
   * 获取当前玩家 ID
   */
  getMyPlayerId(): number | null {
    return this.streamClient.getMyPlayerId();
  }

  /**
   * 获取当前房间 ID
   */
  getCurrentRoomId(): string | null {
    return this.streamClient.getRoomId();
  }

  /**
   * 获取重连密钥（用于断线重连）
   */
  getReconnectKey(): string | null {
    return this.streamClient.getReconnectKey();
  }

  /**
   * 检查是否已连接
   */
  isConnected(): boolean {
    return this.streamClient.getState() === ConnectionState.CONNECTED;
  }

  // ============== 便捷方法 ==============

  /**
   * 创建并加入房间的便捷方法
   * @param roomId 房间 ID
   */
  async createAndJoinRoom(roomId: string): Promise<void> {
    await this.createRoom(roomId);
    await this.joinRoom(roomId);
  }
}

// 导出相关类型
export { ConnectionState, MessageHandlers };
export type { Request, LobbyResponse, RoomResponse };