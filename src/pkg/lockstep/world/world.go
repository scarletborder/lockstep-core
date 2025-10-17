package world

type PlayerData struct {
	// 用户发送数据时,其所在的FrameID
	// 游戏世界需要自行处理例如延迟补偿等机制
	FrameID uint32
	// 解析后的框架无关的额外Bytes,不带有帧ID的原始数据
	Data []byte
}

// IGameWorld 是需要由具体游戏工程实现的接口
// 核心框架的 Room 将会调用这些方法
type IGameWorld interface {
	// OnCreate 当房间创建时调用，可用于初始化游戏世界
	OnCreate(roomContext IRoomContext)

	// OnPlayerEnter 当有玩家进入房间时，核心框架会调用此方法
	// player 是一个只包含核心信息（如UID）的接口或结构体
	OnPlayerEnter(player IPlayer)

	// OnPlayerLeave 当有玩家离开时调用
	OnPlayerLeave(player IPlayer)

	// OnReceiveData 核心框架收到客户端的原始数据包后，直接透传给此方法
	// 这是处理游戏逻辑输入的核心入口
	OnReceiveData(player IPlayer, data PlayerData)

	// OnTick 核心框架的 lockstep 时钟每帧会调用此方法
	// 游戏世界需要在此方法内处理本帧所有输入，并推进游戏状态
	Tick()

	// GetSnapshot 核心框架在 Tick() 之后调用此方法，获取需要广播给所有客户端的帧同步数据
	GetSnapshot() []byte

	// OnDestroy 当房间销毁时调用，用于资源释放
	OnDestroy()
}
