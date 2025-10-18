package world

// 用户的输入序列
// 1. 影响权威游戏世界
// 2. 直接进行帧同步广播
type ClientInputData struct {
	Uid uint32
	// 用户发送数据时,其所在的FrameID
	// 游戏世界需要自行处理例如延迟补偿等机制
	//
	// e.g. Lag Compensation, Display of Targets
	FrameID uint32
	// 解析后的框架无关的额外Bytes,不带有帧ID的原始输入队列
	// e.g. [{"method" : "move", "direction": "up"}]
	Data []byte
}

// 游戏事件
type WorldEventData struct {
	// 事件对哪个帧号造成影响
	// 即事件后的下一帧
	FrameID uint32
	// 事件的原始数据队列
	// e.g. [{"eventType": "explosion", "position": [100,200]}]
	//
	// 如果是超大地图需要分chunk的游戏，此data也推荐用于顺便携带某next帧的某chunk状态同步数据
	Data []byte
}

// IGameWorld 是需要由具体游戏工程实现的接口
// 核心框架的 Room 将会调用这些方法
type IGameWorld interface {
	// OnCreate 当房间创建时调用，可用于初始化游戏世界
	OnCreate(roomContext IRoomContext)

	// OnPlayerEnter 当有玩家进入房间时，核心框架会调用此方法
	// player 是一个只包含核心信息（如UID）的接口或结构体
	OnPlayerEnter(uid uint32)

	// OnPlayerLeave 当有玩家离开时调用
	OnPlayerLeave(uid uint32)

	// OnReceiveData 核心框架收到客户端的原始数据包后，直接透传给此方法
	// 这是外部游戏世界处理用户输入的核心入口
	OnReceiveData(uid uint32, data ClientInputData)

	// OnTick 核心框架的 lockstep 时钟每帧会调用此方法
	// 游戏世界需要在此方法内处理本帧所有输入，并推进游戏状态
	Tick()

	// EnableDifferentSnapshot 如果启用，那么GetSnapshot的参数将含有意义，并且将为每个玩家独立请求一份数据
	// 小游戏建议DISABLE此功能以提升性能
	// IMPORTANT: 如果禁用，在InGame阶段将转发所有的input data，如果input data中存在例如”私聊“等事件也会被转发
	// FUTURE ： 考虑优化为分group请求数据
	EnableDifferentSnapshot() bool

	// GetSnapshot 核心框架在 Tick() 之后调用此方法
	// 外部游戏世界在接受了“必定来自过去”的帧操作后，需要转发用户操作和游戏事件
	// 因此本方法用于从外部游戏世界获取需要 “广播给所有客户端” 的同步数据
	// 即权威事件列表(不能仅在客户端预测的事件，如技能释放，伤害计算等)
	GetSnapshot(uid uint32) []WorldEventData

	// OnDestroy 当房间销毁时调用，用于资源释放
	OnDestroy()
}
