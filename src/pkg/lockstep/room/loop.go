package room

import (
	"lockstep-core/src/constants"
	"lockstep-core/src/pkg/lockstep/client"
	"log"
	"runtime/debug"
	"time"
)

// Run 房间状态机主循环
func (room *Room) Run() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("运行房间 %v 捕获到 Panic: %v\n", room.ID, r)
			log.Printf("堆栈信息:\n%s", string(debug.Stack()))
			log.Println("程序已从 panic 中恢复，将继续运行。")
		}
		// 摧毁本房间
		room.Destroy()
	}()

	// 初始房间状态，大厅中等待玩家
	room.RoomStage.Store(constants.STAGE_InLobby)

	/* 状态机主循环
	根据用户输入和房间的当前状态来进行分支
	*/
	for {
		// 根据当前状态，决定是否需要 ticker
		var tickerChan (<-chan time.Time)

		// 如果不是 InGame 状态，则不需要游戏逻辑定时器
		if room.RoomStage.EqualTo(constants.STAGE_InGame) && room.GameTicker != nil {
			// 只有在 InGame 状态下才有 GameTicker
			tickerChan = room.GameTicker.C
		}

		// 检查是否应该关闭房间
		if room.RoomStage.EqualTo(constants.STAGE_CLOSED) {
			log.Printf("🔴 Room %v is closing, exiting main loop", room.ID)
			return
		}

		select {
		// 处理通用的客户端来源事件

		// 1.1 客户端加入消息来自于 register channel
		case player := <-room.register:
			room.handleRegister(player)

		// 1.2 客户端离开消息来自于 unregister channel
		case player := <-room.unregister:
			room.handleUnregister(player)

		// 2. 处理玩家发来的具体业务消息
		case message := <-room.incomingMessages:
			room.handlePlayerMessage(message)

		// 3. 处理定时器事件，仅在 InGame 状态下有效
		case <-tickerChan:
			room.stepGameTick()
		}
	}
}

// handleRegister 处理玩家注册
func (room *Room) handleRegister(player *client.Client) {
	log.Printf("🔵 Processing registration for player %d", player.GetID())

	// 更新房间活跃时间
	room.UpdateActiveTime()

	// 在 Lobby 状态下才允许新玩家加入
	if room.RoomStage.EqualTo(constants.STAGE_InLobby) {
		log.Printf("🔵 Room is in lobby state, adding player %d", player.GetID())
		// 向 context 中注册用户
		room.ClientsContainer.AddUser(player)

		// TODO: 制作当前 peers 信息 JSON 并广播房间信息
		log.Printf("🔵 Player %d successfully registered", player.GetID())
	} else {
		// 拒绝加入
		log.Printf("🔴 Registration rejected for player %d - room not in lobby state", player.GetID())
	}
}

// handleUnregister 处理玩家注销
func (room *Room) handleUnregister(player *client.Client) {
	if player == nil || player.Session == nil {
		return
	}

	log.Printf("🟡 Unregistering player %d", player.GetID())
	room.ClientsContainer.DelUser(player.GetID())

	// 关闭连接
	if player.Session.IsConnected() {
		player.Session.Close()
	}

	// TODO: 广播人数变化
}

// handlePlayerMessage 处理玩家消息
func (room *Room) handlePlayerMessage(msg *client.ClientMessage) {
	defer func() {
		// 释放到对象池
		msg.Client.ReleasePlayerMessage(msg)

		if r := recover(); r != nil {
			log.Printf("处理用户信息时捕获到 Panic: %v\n", r)
			log.Printf("堆栈信息:\n%s", string(debug.Stack()))
			log.Println("程序已从 panic 中恢复，将继续运行。")
		}
	}()

	// 更新房间活跃时间 - 任何玩家消息都表示房间是活跃的
	room.UpdateActiveTime()

	// TODO: 解析 用户消息并进行分支处理
}

// runGameTick 定时器触发的游戏逻辑帧
// 乐观lockstep, 不等待迟到帧
func (room *Room) stepGameTick() {
	// 仍然没有玩家在线，即全部离开或断开，那么等待，跳过本次
	if room.ClientsContainer.GetPlayerCount() == 0 {
		log.Printf("⚠️ No players online in room %v, skipping game tick", room.ID)
		return
	}
	// 本次frame step行为将有效，更新最后活动时间
	room.LastActiveTime = time.Now()
	// 这一次step行为的目标帧号
	nextRenderFrame := room.SyncData.NextFrameID.Load()

	// TODO: 读取 operation chan 并广播

	defer func() {
		// 步进
		room.SyncData.NextFrameID.Add(1)
		// 删除本帧的操作 ID 记录
		room.SyncData.DeleteOperationID(nextRenderFrame)
		// 更新游戏逻辑
		// room.Logic.Reset()
	}()

	log.Printf("🎮 Game tick: frame %d", nextRenderFrame)
}
