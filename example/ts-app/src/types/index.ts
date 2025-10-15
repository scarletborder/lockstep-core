/**
 * 统一导出所有类型定义
 */

// 从 protobuf 导出
export * from './pb/request';
export * from './pb/response';

// HTTP 请求类型
export interface ListRoomsResponse {
  rooms: string[];
}

export interface CreateRoomRequest {
  room_id: string;
}

export interface CreateRoomResponse {
  success: boolean;
  message: string;
}

// WebTransport 相关类型
export enum ConnectionState {
  DISCONNECTED = 'disconnected',
  CONNECTING = 'connecting',
  LOBBY = 'lobby',
  CONNECTED = 'connected',
  RECONNECTING = 'reconnecting',
  ERROR = 'error',
}

export interface MessageHandlers {
  onLobbyResponse?: (response: any) => void;
  onRoomResponse?: (response: any) => void;
  onError?: (error: Error) => void;
  onStateChange?: (state: ConnectionState) => void;
}
