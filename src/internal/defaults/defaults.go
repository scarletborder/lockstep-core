package defaults

import "lockstep-core/src/pkg/lockstep/world"

// DefaultGameWorld 是一个空的、最小的游戏世界实现，仅用于默认占位。
// 该类型放在 internal 包中，外部模块无法直接引用或依赖。
type DefaultGameWorld struct{}

// OnCreateRoom 当房间创建时调用（空实现）
func (d *DefaultGameWorld) OnCreateRoom(roomContext world.IRoomContext) {}

func (d *DefaultGameWorld) CouldJoinRoom(isReconnect bool) bool { return true }
func (d *DefaultGameWorld) OnPlayerJoin(uid uint32, isReconnect bool) []byte { return nil }
func (d *DefaultGameWorld) OnPlayerLeave(uid uint32)                       {}
func (d *DefaultGameWorld) OnHandleInLobby(uid uint32, data []byte)        {}
func (d *DefaultGameWorld) OnHandleToPreparingStage(uid uint32, data []byte) bool {
    return true
}
func (d *DefaultGameWorld) OnHandleReady(uid uint32, isReady bool, extraData []byte) {}
func (d *DefaultGameWorld) OnHandleToLobbyStage(uid uint32, extraData []byte) bool { return true }
func (d *DefaultGameWorld) OnHandleLoaded(uid uint32)                             {}
func (d *DefaultGameWorld) OnReceiveClientInput(uid uint32, data world.ClientInputData) {
}
func (d *DefaultGameWorld) OnReceiveOtherData(uid uint32, data []byte)             {}
func (d *DefaultGameWorld) OnHandleEndGame(uid uint32, statusCode uint32, data []byte) bool {
    return true
}
func (d *DefaultGameWorld) OnHandlePostGameData(uid uint32, data []byte) bool { return true }
func (d *DefaultGameWorld) Tick()                                            {}
func (d *DefaultGameWorld) GetFrameData(frameId uint32, o world.WorldOptions) world.FrameData {
    return world.FrameData{}
}
func (d *DefaultGameWorld) GetSnapshot(frameId uint32, o world.WorldOptions) world.Snapshot {
    return nil
}
func (d *DefaultGameWorld) OnDestroy() {}

// DefaultNewGameWorld 是 DefaultGameWorld 的工厂函数，供内部默认使用
func DefaultNewGameWorld(rctx world.IRoomContext) world.IGameWorld {
    return &DefaultGameWorld{}
}
