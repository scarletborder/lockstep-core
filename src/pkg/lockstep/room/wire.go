package room

import "github.com/google/wire"

// room manager
var ProviderSet = wire.NewSet(
	NewRoomManager,
	// 绑定接口到实现
	wire.Bind(new(IRoomManager), new(*RoomManager)),
)
