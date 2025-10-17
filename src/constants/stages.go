package constants

import "sync/atomic"

// 游戏的各个阶段
type Stage uint32

const (
	STAGE_InLobby   Stage = 0x20 // InLobby (房间.等待中): 房间刚被创建,所有人还在房间中，等待玩家加入，房主可以设置游戏。
	STAGE_Preparing Stage = 0x21 // Preparing (房间.准备中): 房主已发起游戏，所有玩家选择装备并确认准备。
	STAGE_Loading   Stage = 0x22 // Loading (游戏.加载中): 游戏开始前的加载阶段，所有玩家准备完毕后进入 InGame。
	STAGE_InGame    Stage = 0x23 // InGame (游戏.游戏中): 所有玩家准备就绪，游戏正式开始，由定时器驱动逻辑。
	STAGE_PostGame  Stage = 0x24 // PostGame (游戏.游戏后结算): 游戏结束，显示战绩，等待返回大厅。

	STAGE_CLOSED Stage = 0xEE
	STAGE_Error  Stage = 0xFF
)

// ForwardStage 将当前阶段推进到下一个阶段。
// 这是一个纯函数，不修改原始值。
func (s Stage) ForwardStage() Stage {
	switch s {
	case STAGE_InLobby:
		return STAGE_Preparing
	case STAGE_Preparing:
		return STAGE_Loading
	case STAGE_Loading:
		return STAGE_InGame
	case STAGE_InGame:
		return STAGE_PostGame
	case STAGE_PostGame:
		return STAGE_InLobby
	default:
		return STAGE_Error
	}
}

// IsLaterThanOrEqual 判断当前阶段是否晚于或等于目标阶段。
func (s Stage) IsLaterThanOrEqual(target Stage) bool {
	return s >= target
}

// IsEarlierThan 判断当前阶段是否早于目标阶段。
func (s Stage) IsEarlierThan(target Stage) bool {
	return s < target
}

// EqualTo 判断当前阶段是否等于目标阶段。
func (s Stage) EqualTo(target Stage) bool {
	return s == target
}

// --- 以下是新增的 AtomStage ---

// AtomStage 是 Stage 的并发安全封装。
// 它使用原子操作来保证所有方法的线程安全。
type AtomStage struct {
	value uint32
}

// NewAtomStage 创建一个 AtomStage 实例。
func NewAtomStage(initialStage Stage) *AtomStage {
	return &AtomStage{
		value: uint32(initialStage),
	}
}

// Load 原子地加载并返回当前的 Stage。
func (as *AtomStage) Load() Stage {
	return Stage(atomic.LoadUint32(&as.value))
}

// Store 原子地存储一个 Stage。
func (as *AtomStage) Store(s Stage) {
	atomic.StoreUint32(&as.value, uint32(s))
}

// CompareAndSwap (CAS) 原子地比较当前值与 old，如果相等，则替换为 new。
// 如果操作成功，返回 true。
func (as *AtomStage) CompareAndSwap(old, new Stage) bool {
	return atomic.CompareAndSwapUint32(&as.value, uint32(old), uint32(new))
}

// ForwardStage 原子地将当前阶段推进到下一个阶段。
// 使用 compare-and-swap 循环来确保操作的原子性。
func (as *AtomStage) ForwardStage() Stage {
	for {
		old := as.Load()
		new := old.ForwardStage()

		// 如果当前阶段无法推进，则直接返回，不进行任何修改
		if new == STAGE_Error {
			return old
		}

		// 尝试用新值替换旧值
		if as.CompareAndSwap(old, new) {
			return new
		}
		// 如果 CAS 失败，说明有其他 goroutine 修改了值，循环重试
	}
}

// EqualTo 原子地判断当前阶段是否等于目标阶段。
func (as *AtomStage) EqualTo(target Stage) bool {
	return as.Load() == target
}

// IsLaterThanOrEqual 原子地判断当前阶段是否晚于或等于目标阶段。
func (as *AtomStage) IsLaterThanOrEqual(target Stage) bool {
	return as.Load() >= target
}

// IsEarlierThan 原子地判断当前阶段是否早于目标阶段。
func (as *AtomStage) IsEarlierThan(target Stage) bool {
	return as.Load() < target
}
