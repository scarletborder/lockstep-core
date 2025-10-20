package room

import (
	"fmt"
	"lockstep-core/src/constants"
	"lockstep-core/src/messages"
	"lockstep-core/src/pkg/lockstep/client"
	"lockstep-core/src/pkg/lockstep/world"
	"log"
	"runtime/debug"
	"time"

	"google.golang.org/protobuf/proto"
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

	log.Printf("🔵 Room is in lobby state, adding player %d", player.GetID())
	// 向 context 中注册用户
	room.ClientsContainer.AddUser(player)
	reconnKey, err := room.JwtService.GenerateToken(player.GetID(), room.ID)
	if err != nil {
		resp := &messages.ResponseJoin{
			Code: 500,
			Payload: &messages.ResponseJoin_Fail{
				Fail: &messages.ResponseJoinFail{
					Message: fmt.Sprintf("Fail to Generate reconnect token: %s", err.Error()),
				},
			},
		}
		b, err := proto.Marshal(resp)
		if err == nil {
			room.SendMessageToUserByPlayer(b, player)
		}
		log.Printf("🔴 Failed to generate reconnect token for player %d: %v", player.GetID(), err)
		return
	}
	extraData := room.Game.OnPlayerJoin(player.GetID(), player.IsReconnected)
	// 发送欢迎消息
	roomInfo := &messages.RoomInfo{
		RoomKey:        room.key,
		MaxPlayers:     int32(room.MaxClientPerRoom),
		CurrentPlayers: int32(room.GetPlayerCount()),
		PlayerIDs:      room.Clients.ToSlice(),
		Data:           extraData,
	}
	resp := &messages.ResponseJoin{
		Code: 200,
		Payload: &messages.ResponseJoin_Success{
			Success: &messages.ResponseJoinSuccess{
				RoomID:         room.ID,
				MyID:           player.GetID(),
				ReconnectToken: reconnKey,
				RoomInfo:       roomInfo,
			},
		},
	}
	b, err := proto.Marshal(resp)
	if err == nil {
		room.SendMessageToUserByPlayer(b, player)
	}

	// 制作当前 peers 信息并广播房间信息
	room.BroadcastMessage(resp, []uint32{})

	log.Printf("🔵 Player %d successfully registered", player.GetID())

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

	// 解析 用户消息并进行分支处理
	payload := msg.SessionRequest.Payload

	// 对 oneof 字段的具体类型进行 switch 并交由相应的 handler 处理
	switch p := payload.(type) {
	case *messages.SessionRequest_InLobby:
		room.handleInLobby(msg.Client, p)
	case *messages.SessionRequest_ToPreparing:
		room.handleToPreparing(msg.Client, p)
	case *messages.SessionRequest_Ready:
		room.handleReady(msg.Client, p)
	case *messages.SessionRequest_ToInLobby:
		room.handleToInLobby(msg.Client, p)
	case *messages.SessionRequest_Loaded:
		room.handleLoaded(msg.Client, p)
	case *messages.SessionRequest_InGameFrames:
		room.handleInGameFrames(msg.Client, p)
	case *messages.SessionRequest_Other:
		room.handleOther(msg.Client, p)
	case *messages.SessionRequest_EndGame:
		room.handleEndGame(msg.Client, p)
	case *messages.SessionRequest_PostGameData:
		room.handlePostGameData(msg.Client, p)
	default:
		// unknown type - ignore
	}
}

// runGameTick 定时器触发的游戏逻辑帧
// 乐观lockstep, 不等待迟到帧
func (room *Room) stepGameTick() {
	// 仍然没有玩家在线，即全部离开或断开，那么等待，跳过本次
	if room.ClientsContainer.GetPlayerCount() == 0 {
		log.Printf("⚠️ No players online in room %v, skipping game tick", room.ID)
		return
	}

	// 如果没有启用乐观锁，判断是否停止等待
	if *room.LockstepConfig.MaxDelayFrames >= 0 && !room.HasAllPlayerSync() {
		// 跳过本次
		return
	}

	// 本次frame step行为将有效，更新最后活动时间
	room.LastActiveTime = time.Now()
	// 这一次step行为的目标帧号
	nextRenderFrame := room.SyncData.NextFrameID.Load()

	// 步进到下一帧所需的FrameData
	frameData := room.Game.GetFrameData(nextRenderFrame, world.WorldOptions{
		ChunkID: 0,
	})

	room.SyncData.StoreFrame(nextRenderFrame, &frameData)

	// 步进，防止耗时的发送操作阻塞逻辑更新
	room.SyncData.NextFrameID.Add(1)

	// 预组装所有帧数据以优化发送
	var oldestAsk uint32 = 0xFFFFFFFF
	room.ClientsContainer.Clients.Range(func(key uint32, value *client.Client) bool {
		ack := value.LatestAckNextFrameID.Load()
		if ack < nextRenderFrame {
			oldestAsk = ack
		}
		return true
	})

	if oldestAsk == 0xFFFFFFFF {
		// 发送空
		resp := &messages.SessionResponse{
			Payload: &messages.SessionResponse_InGameFrames{
				InGameFrames: &messages.ResponseInGameFrames{
					Frames: []*messages.FrameData{},
				},
			},
		}
		room.ClientsContainer.Clients.Range(func(key uint32, value *client.Client) bool {
			go func(client *client.Client) {
				data, err := proto.Marshal(resp)
				if err != nil {
					log.Printf("Failed to marshal empty frame data for client %d: %v", client.GetID(), err)
					return
				}
				client.Write(data)
			}(value)
			return true
		})
		// 结束发送空
		return
	}

	allFrames := make([]*messages.FrameData, 0, nextRenderFrame-oldestAsk)
	for i := oldestAsk + 1; i <= nextRenderFrame; i++ {
		if frame, ok := room.SyncData.GetFrame(i); ok {
			allFrames = append(allFrames, (*messages.FrameData)(frame))
		}
	}

	// 为每位用户发送ack至目前的帧
	room.ClientsContainer.Clients.Range(func(key uint32, value *client.Client) bool {
		go func(client *client.Client) {
			// 用户已经确认了“步进到ack”所需的帧数据，
			// 需要向他传递 "步进到ack+1", "步进到ack+2" ... "步进到nextRenderFrame" 的所有帧数据
			ack := client.LatestAckNextFrameID.Load()
			if ack >= nextRenderFrame {
				return
			}
			frames := allFrames[ack-oldestAsk : nextRenderFrame-oldestAsk]
			resp := &messages.SessionResponse{
				Payload: &messages.SessionResponse_InGameFrames{
					InGameFrames: &messages.ResponseInGameFrames{
						Frames: frames,
					},
				},
			}
			data, err := proto.Marshal(resp)
			if err != nil {
				log.Printf("Failed to marshal frame data for client %d: %v", client.GetID(), err)
				return
			}
			client.Write(data)
		}(value)
		return true
	})

}
