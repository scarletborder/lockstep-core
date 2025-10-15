package room

import (
	"lockstep-core/src/constants"
	"lockstep-core/src/logic/clients"
	"log"
	"runtime/debug"
	"time"
)

// Run 房间状态机主循环
func (room *Room) Run() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("运行房间 %s 捕获到 Panic: %v\n", room.ID, r)
			log.Printf("堆栈信息:\n%s", string(debug.Stack()))
			log.Println("程序已从 panic 中恢复，将继续运行。")
		}
		// 摧毁本房间
		room.Destroy()
	}()

	// 初始房间状态，大厅中等待玩家
	room.GameStage.Store(constants.STAGE_InLobby)

	/* 状态机主循环
	根据用户输入和房间的当前状态来进行分支
	*/
	for {
		// 根据当前状态，决定是否需要 ticker
		var tickerChan (<-chan time.Time)

		// 如果不是 InGame 状态，则不需要游戏逻辑定时器
		if room.GameStage.EqualTo(constants.STAGE_InGame) && room.RoomCtx.GameTicker != nil {
			// 只有在 InGame 状态下才有 GameTicker
			tickerChan = room.RoomCtx.GameTicker.C
		}

		// 检查是否应该关闭房间
		if room.GameStage.EqualTo(constants.STAGE_CLOSED) {
			log.Printf("🔴 Room %s is closing, exiting main loop", room.ID)
			return
		}

		select {
		// 1. 处理通用的客户端管理事件
		case player := <-room.register:
			room.handleRegister(player)

		case player := <-room.unregister:
			room.handleUnregister(player)

		// 2. 处理玩家发来的具体业务消息
		case message := <-room.incomingMessages:
			room.handlePlayerMessage(message)

		// 3. 处理定时器事件，仅在 InGame 状态下有效
		case <-tickerChan:
			room.runGameTick()
		}
	}
}

// handleRegister 处理玩家注册
func (room *Room) handleRegister(player *clients.Player) {
	log.Printf("🔵 Processing registration for player %d", player.GetID())

	// 更新房间活跃时间
	room.UpdateActiveTime()

	// 在 Lobby 状态下才允许新玩家加入
	if room.GameStage.EqualTo(constants.STAGE_InLobby) {
		log.Printf("🔵 Room is in lobby state, adding player %d", player.GetID())
		// 向 context 中注册用户
		room.RoomCtx.AddUser(player)

		// TODO: 制作当前 peers 信息 JSON 并广播房间信息
		log.Printf("🔵 Player %d successfully registered", player.GetID())
	} else {
		// 拒绝加入
		log.Printf("🔴 Registration rejected for player %d - room not in lobby state", player.GetID())
	}
}

// handleUnregister 处理玩家注销
func (room *Room) handleUnregister(player *clients.Player) {
	if player == nil || player.Ctx == nil {
		return
	}

	log.Printf("🟡 Unregistering player %d", player.GetID())
	room.RoomCtx.DelUser(player.Ctx.ID)

	// 关闭连接
	if player.Ctx.IsConnected() {
		player.Ctx.Close()
	}

	// TODO: 广播人数变化
}

// handlePlayerMessage 处理玩家消息
func (room *Room) handlePlayerMessage(msg *clients.PlayerMessage) {
	defer func() {
		// 释放到对象池
		clients.ReleasePlayerMessage(msg)

		if r := recover(); r != nil {
			log.Printf("处理用户信息时捕获到 Panic: %v\n", r)
			log.Printf("堆栈信息:\n%s", string(debug.Stack()))
			log.Println("程序已从 panic 中恢复，将继续运行。")
		}
	}()

	// 更新房间活跃时间 - 任何玩家消息都表示房间是活跃的
	room.UpdateActiveTime()

	// TODO: 解析 protobuf 消息并根据类型分发
	log.Printf("🟡 Received message from player %d, length: %d", msg.Player.GetID(), len(msg.Data))
}

// runGameTick 定时器触发的游戏逻辑帧
func (room *Room) runGameTick() {
	if !room.HasAllPlayerSync() {
		// 如果没有所有玩家同步，则跳过此次逻辑帧
		return
	}

	room.LastActiveTime = time.Now() // 更新最后活动时间
	nextRenderFrame := room.RoomCtx.NextFrameID.Load()

	// TODO: 读取 operation chan 并广播

	defer func() {
		// 步进
		room.RoomCtx.NextFrameID.Add(1)
		// 删除本帧的操作 ID 记录
		room.RoomCtx.DeleteOperationID(nextRenderFrame)
		// 更新游戏逻辑
		// room.Logic.Reset()
	}()

	log.Printf("🎮 Game tick: frame %d", nextRenderFrame)
}

// HasAllPlayerSync 检查是否所有玩家都同步
func (room *Room) HasAllPlayerSync() bool {
	// 延迟等待，最多容忍 maxDelayFrames 帧的延迟
	nextRenderFrame := room.RoomCtx.NextFrameID.Load()
	var minFrameID uint32

	if nextRenderFrame < constants.MaxDelayFrames {
		minFrameID = 0
	} else {
		minFrameID = nextRenderFrame - constants.MaxDelayFrames
	}

	synced := true
	// 遍历每个玩家的 frameID，若有任意玩家低于阈值，则返回 false
	room.RoomCtx.Players.Range(func(key int, value *clients.Player) bool {
		// 检查玩家是否为空或玩家上下文为空
		if value == nil || value.Ctx == nil {
			synced = false
			return false
		}

		// 获取当前玩家实际的帧号
		playerCurrentFrame := value.Ctx.LatestFrameID.Load()
		if playerCurrentFrame < minFrameID {
			synced = false
			return false
		}
		return true
	})
	return synced
}

// StartServeClient 开始为客户端服务（接收消息）
func (room *Room) StartServeClient(player *clients.Player) {
	ctx := player.Ctx
	log.Printf("🟡 StartServeClient for player %d", player.GetID())

	// 检查基本有效性
	if ctx == nil {
		log.Printf("🔴 Player context is nil for player %d at start", player.GetID())
		return
	}

	log.Printf("🟢 Starting client service for player %d", player.GetID())

	defer func() {
		log.Printf("🟡 StartServeClient ending for player %d", player.GetID())

		// 发送 unregister 信号，通知房间移除这个玩家
		select {
		case room.unregister <- player:
			log.Printf("🟡 Sent unregister signal for player %d", player.GetID())
		default:
			log.Printf("🔴 Failed to send unregister signal for player %d (channel full)", player.GetID())
			if player != nil && player.Ctx != nil {
				player.Ctx.Close()
			}
		}

		if r := recover(); r != nil {
			log.Printf("服务用户 %d 时捕获到 Panic: %v\n", player.GetID(), r)
			log.Printf("堆栈信息:\n%s", string(debug.Stack()))
			log.Println("程序已从 panic 中恢复，将继续运行。")
		}
	}()

	// 接收消息循环
	log.Printf("🟡 Starting message loop for player %d", player.GetID())
	for {
		// 使用 WebTransport 接收 datagram
		data, err := ctx.ReceiveDatagram(ctx.Ctx)
		if err != nil {
			log.Printf("🔴 ReceiveDatagram error for player %d: %v", player.GetID(), err)
			return
		}

		log.Printf("🟡 Received datagram from player %d, length: %d", player.GetID(), len(data))

		// 发送到消息管道
		msg := clients.GetPlayerMessage(player, data)
		room.incomingMessages <- msg
	}
}
