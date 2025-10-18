package world

// IRoomContext 定义了房间上下文接口。
//
// 游戏逻辑需要一种方式来向客户端发送数据，但它不应该直接接触网络会话 (ISession)。roomContext 提供了这种能力。
type IRoomContext interface {
	// # 通信服务

	// Broadcast 向房间内所有客户端广播消息
	// 游戏逻辑层负责将数据序列化为 []byte
	Broadcast(data []byte)

	// SendTo 向指定的玩家发送消息
	// 游戏逻辑层通过 IPlayer 接口来指定目标
	SendTo(uid uint32, data []byte)

	// SendToMultiple 向一组玩家发送消息，比循环调用 SendTo 更高效
	SendToMultiple(uids []uint32, data []byte)

	// # 状态查询

	// GetRoomID 获取房间的唯一标识符，主要用于日志记录和调试
	GetRoomID() uint32

	// GetAllPlayers 获取当前在房间内的所有玩家列表
	// 返回 IPlayer 接口切片，只暴露核心信息（如UID）
	GetAllPlayers() []uint32

	// GetCurrentFrame 获取当前 lockstep 的帧号
	// 这是与 SyncData 交互的最关键部分
	GetCurrentFrame() uint32

	// # 动作请求

	// KickPlayer 请求核心框架踢掉一个玩家
	// 游戏逻辑判断“为什么”踢，核心框架执行“如何”踢（关闭连接、清理资源等）
	KickPlayer(uid uint32, reason string)

	// DestroyRoom 请求核心框架销毁当前房间
	// 例如，游戏逻辑在 Tick() 中判断出胜负已分，可以调用此方法来结束游戏
	DestroyRoom()
}
