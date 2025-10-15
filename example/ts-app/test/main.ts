/**
 * Lockstep Client SDK 测试示例
 */

import { LockstepClient, ConnectionState, Request, LobbyResponse, RoomResponse } from '../src';

// 配置
const SERVER_URL = 'https://127.0.0.1:4433';
const ROOM_ID = 'test-room-001';

async function testHTTPRequests() {
  console.log('\n=== 测试 HTTP 请求 ===\n');

  const client = new LockstepClient({
    serverUrl: SERVER_URL,
    safety: {
      allowAnyCert: true,
      allowInsecureTransport: true,
      allowSelfSigned: true,
    }
  });

  try {
    // 测试获取房间列表
    console.log('1. 获取房间列表...');
    const rooms = await client.listRooms();
    console.log(`   找到 ${rooms.length} 个房间:`, rooms);

    // 测试创建房间
    console.log('\n2. 创建房间...');
    const createResult = await client.createRoom(ROOM_ID);
    console.log(`   ${createResult.message}`);

    // 再次获取房间列表
    console.log('\n3. 再次获取房间列表...');
    const roomsAfter = await client.listRooms();
    console.log(`   找到 ${roomsAfter.length} 个房间:`, roomsAfter);
  } catch (error) {
    console.error('HTTP 请求测试失败:', error);
  }
}

async function testWebTransport() {
  console.log('\n=== 测试 WebTransport 连接 ===\n');

  const client = new LockstepClient({ serverUrl: SERVER_URL });

  // 设置消息处理器
  client.setMessageHandlers({
    onLobbyResponse: (response: LobbyResponse) => {
      console.log('\n收到大厅响应:');

      if (response.payload.oneofKind === 'joinRoomSuccess') {
        const success = response.payload.joinRoomSuccess;
        console.log(`✓ 成功加入房间！`);
        console.log(`  - 房间 ID: ${success.roomId}`);
        console.log(`  - 我的玩家 ID: ${success.myId}`);
        console.log(`  - 重连密钥: ${success.key}`);
        console.log(`  - 消息: ${success.message}`);

        // 加入成功后发送一些测试消息
        setTimeout(() => testGameMessages(client), 1000);
      } else if (response.payload.oneofKind === 'joinRoomFailed') {
        const failed = response.payload.joinRoomFailed;
        console.error(`✗ 加入房间失败: ${failed.message}`);
      }
    },

    onRoomResponse: (response: RoomResponse) => {
      console.log('\n收到房间消息:');
      console.log(`  类型: ${response.payload.oneofKind}`);

      switch (response.payload.oneofKind) {
        case 'roomInfo':
          console.log('  房间信息:', response.payload.roomInfo);
          break;
        case 'chooseMap':
          console.log('  选择地图:', response.payload.chooseMap);
          break;
        case 'quitChooseMap':
          console.log('  退出选择地图:', response.payload.quitChooseMap);
          break;
        case 'roomClosed':
          console.log('  房间关闭:', response.payload.roomClosed);
          break;
        case 'gameEnd':
          console.log('  游戏结束:', response.payload.gameEnd);
          break;
        case 'error':
          console.log('  错误:', response.payload.error);
          break;
        case 'updateReadyCount':
          console.log('  更新准备数:', response.payload.updateReadyCount);
          break;
        case 'allReady':
          console.log('  全部准备:', response.payload.allReady);
          break;
        case 'allLoaded':
          console.log('  全部加载完成:', response.payload.allLoaded);
          break;
        default:
          console.log('  内容:', response.payload);
      }
    },

    onError: (error: Error) => {
      console.error('\n发生错误:', error.message);
    },

    onStateChange: (state: ConnectionState) => {
      console.log(`\n连接状态变化: ${state}`);
    },
  });

  try {
    // 加入房间
    console.log(`正在连接到房间 "${ROOM_ID}"...`);
    await client.joinRoom(ROOM_ID);

    // 保持连接，等待消息
    await new Promise((resolve) => setTimeout(resolve, 10000));

    // 断开连接
    console.log('\n正在断开连接...');
    await client.disconnect();
    console.log('已断开连接');
  } catch (error) {
    console.error('WebTransport 测试失败:', error);
  }
}

async function testGameMessages(client: LockstepClient) {
  console.log('\n=== 测试游戏消息 ===\n');

  try {
    // 1. 发送空白消息（心跳）
    console.log('1. 发送空白消息...');
    const blankRequest: Request = {
      payload: {
        oneofKind: 'blank',
        blank: {
          frameId: 0,
          ackFrameId: 0,
        },
      },
    };
    await client.sendRequest(blankRequest);
    console.log('   ✓ 已发送');

    await sleep(500);

    // 2. 发送准备消息
    console.log('\n2. 发送准备消息...');
    const readyRequest: Request = {
      payload: {
        oneofKind: 'ready',
        ready: {
          isReady: true,
        },
      },
    };
    await client.sendRequest(readyRequest);
    console.log('   ✓ 已发送');

    await sleep(500);

    // 3. 发送选择地图消息
    console.log('\n3. 发送选择地图消息...');
    const chooseMapRequest: Request = {
      payload: {
        oneofKind: 'chooseMap',
        chooseMap: {
          chapterId: 1,
          stageId: 1,
        },
      },
    };
    await client.sendRequest(chooseMapRequest);
    console.log('   ✓ 已发送');

    await sleep(500);

    // 4. 发送加载完成消息
    console.log('\n4. 发送加载完成消息...');
    const loadedRequest: Request = {
      payload: {
        oneofKind: 'loaded',
        loaded: {
          isLoaded: true,
        },
      },
    };
    await client.sendRequest(loadedRequest);
    console.log('   ✓ 已发送');

  } catch (error) {
    console.error('发送游戏消息失败:', error);
  }
}

async function testReconnect() {
  console.log('\n=== 测试重连功能 ===\n');

  const client = new LockstepClient({ serverUrl: SERVER_URL });
  let reconnectKey: string | null = null;

  // 设置消息处理器
  client.setMessageHandlers({
    onLobbyResponse: (response: LobbyResponse) => {
      if (response.payload.oneofKind === 'joinRoomSuccess') {
        const success = response.payload.joinRoomSuccess;
        reconnectKey = success.key;
        console.log(`✓ 加入成功，重连密钥: ${reconnectKey}`);
      }
    },
    onStateChange: (state: ConnectionState) => {
      console.log(`连接状态: ${state}`);
    },
  });

  try {
    // 首次连接
    console.log('1. 首次连接...');
    await client.joinRoom(ROOM_ID);
    await sleep(2000);

    // 断开连接
    console.log('\n2. 断开连接...');
    await client.disconnect();
    await sleep(1000);

    // 重连
    if (reconnectKey) {
      console.log('\n3. 使用密钥重连...');
      await client.reconnectRoom(ROOM_ID, reconnectKey);
      await sleep(2000);

      console.log('\n✓ 重连成功！');
    }

    // 清理
    await client.disconnect();
  } catch (error) {
    console.error('重连测试失败:', error);
  }
}

// 辅助函数
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// 主函数
async function main() {
  console.log('=================================');
  console.log('Lockstep Client SDK 测试程序');
  console.log('=================================');

  const args = process.argv.slice(2);
  const testType = args[args.length - 1] || 'all';

  switch (testType) {
    case 'http':
      await testHTTPRequests();
      break;
    case 'wt':
    case 'webtransport':
      await testWebTransport();
      break;
    case 'reconnect':
      await testReconnect();
      break;
    case 'all':
    default:
      await testHTTPRequests();
      await testWebTransport();
      break;
  }

  console.log('\n=== 测试完成 ===\n');
}

// 运行测试
if (require.main === module) {
  main().catch(console.error);
}

export { testHTTPRequests, testWebTransport, testReconnect };
