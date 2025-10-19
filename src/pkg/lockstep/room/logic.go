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
	}
}

// 玩家加载完成
func (room *Room) handleLoaded(from *client.Client, payload *messages.SessionRequest_Loaded) {
	if room == nil || from == nil || payload == nil || payload.Loaded == nil {
		return
	}
	from.IsLoaded = true
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
	room.Game.OnReceiveClientInput(uid, world.ClientInputData{
		Uid:     uid,
		FrameID: payload.InGameFrames.FrameId,
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
	}
}

// GameClientInput 辅助方法：把 []byte 包装为 world.ClientInputData
func (room *Room) GameClientInput(b []byte) world.ClientInputData {
	return world.ClientInputData{Uid: 0, FrameID: 0, Data: b}
}
