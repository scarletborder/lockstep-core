package logic

import (
	"github.com/google/wire"
)

// ProviderSet 是 logic 模块的 Wire provider set
var ProviderSet = wire.NewSet(
	NewRoomManager,
	NewDefaultPlayerSessionHandler,
	// 绑定接口到实现
	wire.Bind(new(RoomManagerInterface), new(*RoomManager)),
	wire.Bind(new(PlayerSessionHandler), new(*DefaultPlayerSessionHandler)),
)
