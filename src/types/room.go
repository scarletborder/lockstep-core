package types

import "lockstep-core/src/constants"

// RoomInfo 房间信息
type RoomInfo struct {
	RoomID      string          `json:"room_id"`
	NeedKey     bool            `json:"need_key"`
	PlayerCount int             `json:"player_count"`
	GameState   constants.Stage `json:"game_state"` // 游戏状态
}
