package room

import (
	"lockstep-core/src/constants"
	"lockstep-core/src/messages"
	"lockstep-core/src/pkg/lockstep/client"
	"lockstep-core/src/pkg/lockstep/world"
)

// 以下为 Room 的各个消息处理方法实现，均为非导出方法（首字母小写）
// 入参均为原始 protobuf 消息对应的具体 oneof 类型与来源客户端

// 处理大厅阶段的透传消息到游戏世界
func (room *Room) handleInLobby(from *client.Client, payload *messages.SessionRequest_InLobby) {
	if room == nil || payload == nil || payload.InLobby == nil || from == nil {
		return
	}
	// 直接透传给游戏世界
	if room.Game == nil {
		return
	}
	// 防御性地检查 from.GetID() 可用性
	uid := from.GetID()
	room.Game.OnHandleInLobby(uid, payload.InLobby.GetData())

}

// 请求进入 Preparing 阶段（房主／客户端发起）
func (room *Room) handleToPreparing(from *client.Client, payload *messages.SessionRequest_ToPreparing) {
	if room == nil || from == nil || payload == nil || payload.ToPreparing == nil {
		return
	}
	if room.Game == nil {
		return
	}
	if room.Game.OnHandleToPreparingStage(from.GetID(), payload.ToPreparing.GetData()) {
		// 允许进入 Preparing 阶段
		room.RoomStage.Store(constants.STAGE_Preparing)
	}
}

// 玩家准备就绪
func (room *Room) handleReady(from *client.Client, payload *messages.SessionRequest_Ready) {
	if room == nil || from == nil || payload == nil || payload.Ready == nil {
		return
	}
	from.IsReady = true
	if room.Game == nil {
		return
	}
	room.Game.OnHandleReady(from.GetID(), payload.Ready.IsReady, payload.Ready.GetData())
	var readyPlayerIds []uint32 = make([]uint32, 0)
	var playerCount uint32 = 0
	room.ClientsContainer.Clients.Range(func(key uint32, value *client.Client) bool {
		if value != nil && value.IsReady {
			readyPlayerIds = append(readyPlayerIds, key)
		}
		playerCount++
		return true
	})
	// 广播准备状态变更
	innerReadyResp := &messages.ResponseReadyCountUpdate{
		ReadyPlayerIds: readyPlayerIds,
		TotalCount:     uint32(playerCount),
	}
	sresp := &messages.SessionResponse{Payload: &messages.SessionResponse_ReadyCountUpdate{ReadyCountUpdate: innerReadyResp}}
	room.BroadcastMessage(sresp, []uint32{})

	if playerCount == uint32(len(readyPlayerIds)) {
		// 所有玩家均已准备好
		data := room.Game.OnHandleAllReady()
		room.RoomStage.ForwardStage()
		innerStage := &messages.ResponseStageChange{
			NewStage: uint32(constants.STAGE_Loading),
			Data:     data,
		}
		sresp := &messages.SessionResponse{Payload: &messages.SessionResponse_StageChange{StageChange: innerStage}}
		room.BroadcastMessage(sresp, []uint32{})
	}
}

// 返回大厅
func (room *Room) handleToInLobby(from *client.Client, payload *messages.SessionRequest_ToInLobby) {
	if room == nil || from == nil || payload == nil || payload.ToInLobby == nil {
		return
	}
	if room.Game == nil {
		return
	}
	if room.Game.OnHandleToLobbyStage(from.GetID(), payload.ToInLobby.GetData()) {
		// 允许返回大厅
		room.RoomStage.Store(constants.STAGE_InLobby)
		innerStage := &messages.ResponseStageChange{NewStage: uint32(constants.STAGE_InLobby)}
		sresp := &messages.SessionResponse{Payload: &messages.SessionResponse_StageChange{StageChange: innerStage}}
		room.BroadcastMessage(sresp, []uint32{})
	}
}

// 玩家加载完成
func (room *Room) handleLoaded(from *client.Client, payload *messages.SessionRequest_Loaded) {
	var playerCount uint32 = 0
	if room == nil || from == nil || payload == nil || payload.Loaded == nil {
		return
	}
	room.Game.OnHandleLoaded(from.GetID())
	from.IsLoaded = true
	var loadedPlayerIds []uint32 = make([]uint32, 0)
	room.ClientsContainer.Clients.Range(func(key uint32, value *client.Client) bool {
		if value != nil && value.IsLoaded {
			loadedPlayerIds = append(loadedPlayerIds, key)
		}
		playerCount++
		return true
	})
	// 广播加载状态变更
	innerLoadedResp := &messages.ResponseLoadedCountUpdate{
		LoadedPlayerIds: loadedPlayerIds,
		TotalCount:      uint32(playerCount),
	}
	srespLoaded := &messages.SessionResponse{Payload: &messages.SessionResponse_LoadedCountUpdate{LoadedCountUpdate: innerLoadedResp}}
	room.BroadcastMessage(srespLoaded, []uint32{})

	if playerCount == uint32(len(loadedPlayerIds)) {
		// 所有玩家均已加载完毕，进入游戏阶段
		// world 无方法需要调用，因为通过step来进行游戏开始
		room.RoomStage.ForwardStage()
		innerStage := &messages.ResponseStageChange{NewStage: uint32(constants.STAGE_InGame)}
		sresp := &messages.SessionResponse{Payload: &messages.SessionResponse_StageChange{StageChange: innerStage}}
		room.BroadcastMessage(sresp, []uint32{})
	}
}

// 处理游戏中帧数据
func (room *Room) handleInGameFrames(from *client.Client, payload *messages.SessionRequest_InGameFrames) {
	if room == nil || payload == nil || payload.InGameFrames == nil || from == nil {
		return
	}
	if room.Game == nil {
		return
	}
	uid := from.GetID()
	// 更新ack
	from.LatestAckNextFrameID.Store(payload.InGameFrames.GetAckFrameId())
	from.LatestNextFrameID.Store(payload.InGameFrames.GetFrameId())

	room.Game.OnReceiveClientInput(uid, world.ClientInputData{
		Uid:     uid,
		FrameId: payload.InGameFrames.FrameId,
		Data:    payload.InGameFrames.GetData(),
	})
}

// 处理其他自定义消息
func (room *Room) handleOther(from *client.Client, payload *messages.SessionRequest_Other) {
	if room == nil || payload == nil || payload.Other == nil || from == nil {
		return
	}
	// 默认透传给游戏世界
	if room.Game == nil {
		return
	}
	room.Game.OnReceiveOtherData(from.GetID(), payload.Other.GetData())
}

// 处理结束游戏请求
func (room *Room) handleEndGame(from *client.Client, payload *messages.SessionRequest_EndGame) {
	if room == nil || from == nil || payload == nil || payload.EndGame == nil {
		return
	}
	if room.Game == nil {
		return
	}
	if room.Game.OnHandleEndGame(from.GetID(), payload.EndGame.GetStatusCode(), payload.EndGame.GetData()) {
		room.RoomStage.Store(constants.STAGE_PostGame)
		innerStage := &messages.ResponseStageChange{NewStage: uint32(constants.STAGE_PostGame)}
		sresp := &messages.SessionResponse{Payload: &messages.SessionResponse_StageChange{StageChange: innerStage}}
		room.BroadcastMessage(sresp, []uint32{})
	}
}

// 处理进入 PostGame
func (room *Room) handlePostGameData(from *client.Client, payload *messages.SessionRequest_PostGameData) {
	if room == nil || from == nil || payload == nil || payload.PostGameData == nil {
		return
	}
	if room.Game == nil {
		return
	}
	backToLobby := room.Game.OnHandlePostGameData(from.GetID(), payload.PostGameData.GetData())
	if backToLobby {
		room.RoomStage.Store(constants.STAGE_InLobby)
		innerStage := &messages.ResponseStageChange{NewStage: uint32(constants.STAGE_InLobby)}
		sresp := &messages.SessionResponse{Payload: &messages.SessionResponse_StageChange{StageChange: innerStage}}
		room.BroadcastMessage(sresp, []uint32{})
	}
}

// GameClientInput 辅助方法：把 []byte 包装为 world.ClientInputData
func (room *Room) GameClientInput(b []byte) world.ClientInputData {
	return world.ClientInputData{Uid: 0, FrameId: 0, Data: b}
}
