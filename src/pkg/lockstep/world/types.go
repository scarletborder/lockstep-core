package world

import (
	"lockstep-core/src/messages"
)

// 用户的输入序列
// 1. 影响权威游戏世界
// 2. 直接进行帧同步广播
type ClientInputData = messages.ClientInputData

// 游戏事件
// 目前只主要用于传递chunk数据
type WorldEventData = messages.WorldEventData

type FrameData = messages.FrameData

type Snapshot = []byte
