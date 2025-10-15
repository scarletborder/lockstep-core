/**
 * Lockstep Core Client SDK
 * 提供 HTTP 和 WebTransport 接口来与 Lockstep 服务器通信
 */

// 导出主客户端类
export { LockstepClient } from './core';

// 导出类型
export type { InitOptions } from './core';
export { ConnectionState } from './requests/stream';
export type { MessageHandlers } from './requests/stream';

// 导出 protobuf 类型
export * from './types/pb/request';
export * from './types/pb/response';

// 导出 HTTP 请求类型
export type {
  ListRoomsResponse,
  CreateRoomRequest,
  CreateRoomResponse,
} from './requests/common';
