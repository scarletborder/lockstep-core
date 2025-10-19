package room

import (
	"lockstep-core/src/constants"
	"lockstep-core/src/messages"
	"lockstep-core/src/pkg/lockstep/client"
	"lockstep-core/src/pkg/lockstep/world"
	"log"
)

// 以下为 Room 的各个消息处理方法实现，均为非导出方法（首字母小写）
// 入参均为原始 protobuf 消息对应的具体 oneof 类型与来源客户端

// 处理大厅阶段的透传消息到游戏世界
func (room *Room) handleInLobby(from *client.Client, payload *messages.SessionRequest_InLobby) {
	if payload == nil || payload.InLobby == nil {
		return
	}
	// 直接透传给游戏世界
	if room.Game != nil {
		room.Game.OnReceiveData(from.GetID(), room.GameClientInput(payload.InLobby.Data))
	}
}

// 请求进入 Preparing 阶段（房主／客户端发起）
func (room *Room) handleToPreparing(from *client.Client, payload *messages.SessionRequest_ToPreparing) {
	// 简单允许推进到 Preparing（以后可加权限判断）
	old := room.RoomStage.Load()
	new := old.ForwardStage()
	room.RoomStage.Store(new)
	log.Printf("Room %d stage changed %d -> %d by player %d", room.ID, old, room.RoomStage.Load(), from.GetID())
	if room.Game != nil {
		room.Game.OnStageChange(old, room.RoomStage.Load())
	}
}

// 玩家准备就绪
func (room *Room) handleReady(from *client.Client, payload *messages.SessionRequest_Ready) {
	if from == nil {
		return
	}
	from.IsReady = true
	log.Printf("Player %d set ready in room %d", from.GetID(), room.ID)
}

// 返回大厅
func (room *Room) handleToInLobby(from *client.Client, payload *messages.SessionRequest_ToInLobby) {
	// 将房间状态回到大厅
	old := room.RoomStage.Load()
	room.RoomStage.Store(constants.STAGE_InLobby)
	if room.Game != nil {
		room.Game.OnStageChange(old, room.RoomStage.Load())
	}
}

// 玩家加载完成
func (room *Room) handleLoaded(from *client.Client, payload *messages.SessionRequest_Loaded) {
	if from == nil {
		return
	}
	from.IsLoaded = true
	log.Printf("Player %d loaded in room %d", from.GetID(), room.ID)
}

// 处理游戏中帧数据
func (room *Room) handleInGameFrames(from *client.Client, payload *messages.SessionRequest_InGameFrames) {
	if payload == nil || payload.InGameFrames == nil || from == nil {
		return
	}
	// 将原始数据转为游戏世界客户端输入并交由 Game 处理
	if room.Game != nil {
		room.Game.OnReceiveData(from.GetID(), room.GameClientInput(payload.InGameFrames.Data))
	}
}

// 处理其他自定义消息
func (room *Room) handleOther(from *client.Client, payload *messages.SessionRequest_Other) {
	if payload == nil || payload.Other == nil {
		return
	}
	// 默认透传给游戏世界
	if room.Game != nil {
		room.Game.OnReceiveData(from.GetID(), room.GameClientInput(payload.Other.Data))
	}
}

// 处理结束游戏请求
func (room *Room) handleEndGame(from *client.Client, payload *messages.SessionRequest_EndGame) {
	// 将房间推进到 PostGame
	old := room.RoomStage.Load()
	room.RoomStage.Store(constants.STAGE_PostGame)
	if room.Game != nil {
		room.Game.OnStageChange(old, constants.STAGE_PostGame)
	}
}

// 处理进入 PostGame
func (room *Room) handlePostGameData(from *client.Client, payload *messages.SessionRequest_ToPostGame) {
	old := room.RoomStage.Load()
	room.RoomStage.Store(constants.STAGE_PostGame)
	if room.Game != nil {
		room.Game.OnStageChange(old, constants.STAGE_PostGame)
	}
}

// GameClientInput 辅助方法：把 []byte 包装为 world.ClientInputData
func (room *Room) GameClientInput(b []byte) world.ClientInputData {
	return world.ClientInputData{Uid: 0, FrameID: 0, Data: b}
}
