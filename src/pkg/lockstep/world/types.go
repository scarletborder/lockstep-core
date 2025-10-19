package world

import (
	"encoding/binary"
	"lockstep-core/src/messages"

	"github.com/zeebo/xxh3"
)

// 用户的输入序列
// 1. 影响权威游戏世界
// 2. 直接进行帧同步广播
type ClientInputData = messages.ClientInputData

// 游戏事件
type WorldEventData = messages.WorldEventData

type FrameData = messages.FrameData

// CalculateChecksum 使用 github.com/zeebo/xxh3 计算 FrameData 的校验和。
// 这个函数被设计为高性能且零内存分配（除了初始的 hasher）。
func CalculateChecksum(fd *FrameData) uint64 {
	// 1. 创建一个新的 Hasher 实例。
	// New() 是一个非常轻量级的操作。
	h := xxh3.New()

	// 2. 为了避免在循环中重复分配内存，创建一个可复用的缓冲区。
	// uint32 需要4个字节。
	var buf [4]byte

	// 3. 写入根级别的 FrameID。
	// 必须使用固定的字节序（如 BigEndian）来保证所有平台的计算结果一致。
	binary.BigEndian.PutUint32(buf[:], fd.FrameId)
	h.Write(buf[:]) // Hasher 的 Write 方法永不返回错误

	// 4. 依序写入 InputArray 的内容。
	// **重要**: 为了保证确定性，InputArray 在传入此函数前必须是已排序的。
	// 例如，可以根据 Uid 和 FrameID 进行排序。
	for i := range fd.InputArray {
		input := fd.InputArray[i]
		binary.BigEndian.PutUint32(buf[:], input.Uid)
		h.Write(buf[:])
		binary.BigEndian.PutUint32(buf[:], input.FrameId)
		h.Write(buf[:])
		h.Write(input.Data)
	}

	// 5. 依序写入 Events 的内容。
	// **重要**: 同样，Events 数组也必须预先排序以保证确定性。
	// 例如，可以根据 FrameID 和 Data 的某些内容排序。
	for i := range fd.Events {
		event := fd.Events[i]
		binary.BigEndian.PutUint32(buf[:], event.FrameId)
		h.Write(buf[:])
		h.Write(event.Data)
	}

	// 6. 计算并返回最终的 64位 哈希值。
	return h.Sum64()
}

// SetChecksum 是一个辅助方法，用于计算并设置 FrameData 的 Checksum 字段。
// 必须在发送之前调用
func SetChecksum(fd *FrameData) {
	fd.Checksum = CalculateChecksum(fd)
}

// Snapshot 表示游戏世界的状态快照
type Snapshot interface {
	GetData() []byte
	// 是否有效，因为有些帧暂时无快照
	IsAvailable() bool
}
