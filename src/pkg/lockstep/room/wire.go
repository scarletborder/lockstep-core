package room

import "github.com/google/wire"

// room manager
// 现在 RoomManager 的构造函数依赖一个 NewGameWorldFunc，由外部注入
var ProviderSet = wire.NewSet(
	NewRoomManager,
	// 绑定接口到实现
	wire.Bind(new(IRoomManager), new(*RoomManager)),
)
