package world

import "lockstep-core/src/constants"

// IGameWorld 是需要由具体游戏工程实现的接口
// 需要外部调用时实现游戏世界生命周期
// 核心框架的 Room 将会调用这些方法
type IGameWorld interface {
	// OnCreate 当房间创建时调用，可用于初始化游戏世界
	OnCreate(roomContext IRoomContext)

	OnStageChange(oldStage, newStage constants.Stage)

	// OnPlayerEnter 当有玩家进入或重连房间时，核心框架会调用此方法
	OnPlayerEnter(uid uint32)

	// OnPlayerLeave 当有玩家离开时调用
	OnPlayerLeave(uid uint32)

	// OnReceiveData 核心框架收到客户端的原始数据包后，直接透传给此方法
	// 这是外部游戏世界处理用户输入的核心入口
	//
	// 游戏世界需要自行处理例如延迟补偿等机制
	OnReceiveData(uid uint32, data ClientInputData)

	// OnTick 核心框架的 lockstep 时钟每帧会调用此方法
	// 游戏世界需要在此方法内处理本帧所有输入，并推进游戏状态
	Tick()

	// GetFrameData 核心框架在 Tick() 之后调用多次此方法做到自适应ack冗余发送
	// 外部游戏世界在接受了“必定来自过去”的帧操作后，需要转发用户操作和游戏事件
	// 因此本方法用于从外部游戏世界获取需要 “广播给所有客户端” 的同步数据
	// 即操作序列和权威事件列表(不能仅在客户端预测的事件，如技能释放，伤害计算等)
	//
	// frameId : 为了从(frameId-1)跳到frameId所需的帧数据
	GetFrameData(frameId uint32, o WorldOptions) FrameData

	// GetSnapshot 获取状态快照，以方便拉帧快进
	// frameId : 操作处理完后的已经步进到达的帧号
	GetSnapshot(frameId uint32, o WorldOptions) Snapshot

	// OnDestroy 当房间销毁时调用，用于资源释放
	OnDestroy()
}

type WorldOptions struct {
	// FUTURE: 未来将拓展为多chunk以方便lockstep场景下的大世界
	ChunkID int
}
