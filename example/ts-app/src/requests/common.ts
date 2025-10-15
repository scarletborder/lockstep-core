/**
 * 普通 HTTP 请求相关的类型定义和函数
 */

// ============== 类型定义 ==============

/**
 * 房间列表响应
 */
export interface ListRoomsResponse {
  rooms: string[];
}

/**
 * 创建房间请求
 */
export interface CreateRoomRequest {
  room_id: string;
}

/**
 * 创建房间响应
 */
export interface CreateRoomResponse {
  success: boolean;
  message: string;
}

// ============== HTTP 请求封装 ==============

/**
 * HTTP 客户端类
 */
export interface SafetyOptions {
  allowSelfSigned?: boolean;
  allowInsecureTransport?: boolean;
  allowAnyCert?: boolean;
}

export class HTTPClient {
  constructor(private baseUrl: string, private safety?: SafetyOptions) {
    // 如果在 Node 环境中允许自签名/任意证书，设置环境变量来放宽 TLS 校验
    if (typeof process !== 'undefined' && this.safety && (this.safety.allowSelfSigned || this.safety.allowAnyCert)) {
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore
      process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';
    }
  }

  /**
   * 获取房间列表
   */
  async listRooms(): Promise<string[]> {
    const response = await fetch(`${this.baseUrl}/rooms`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to list rooms: ${response.statusText}`);
    }

    const rooms: string[] = await response.json();
    return rooms;
  }

  /**
   * 创建房间
   */
  async createRoom(roomId: string): Promise<CreateRoomResponse> {
    const response = await fetch(`${this.baseUrl}/rooms`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ room_id: roomId }),
    });

    if (!response.ok) {
      throw new Error(`Failed to create room: ${response.statusText}`);
    }

    const text = await response.text();
    return {
      success: true,
      message: text,
    };
  }
}
