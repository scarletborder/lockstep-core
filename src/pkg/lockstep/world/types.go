package world

import (
	"encoding/binary"

	"github.com/zeebo/xxh3"
)

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

type FrameData struct {
	// 本次服务端将要跳转到的帧 ID
	// FrameData和input, events都带有FrameID字段是为了延迟补偿
	// 告知客户端，之前的帧中有遗漏的操作和时间发生
	FrameID uint32
	// 帧数据内容

	InputArray []ClientInputData
	Events     []WorldEventData
	// Checksum 的类型应为 uint64 以接收 xxh3.Sum64() 的结果
	Checksum uint64
}

// CalculateChecksum 使用 github.com/zeebo/xxh3 计算 FrameData 的校验和。
// 这个函数被设计为高性能且零内存分配（除了初始的 hasher）。
func (fd *FrameData) CalculateChecksum() uint64 {
	// 1. 创建一个新的 Hasher 实例。
	// New() 是一个非常轻量级的操作。
	h := xxh3.New()

	// 2. 为了避免在循环中重复分配内存，创建一个可复用的缓冲区。
	// uint32 需要4个字节。
	var buf [4]byte

	// 3. 写入根级别的 FrameID。
	// 必须使用固定的字节序（如 BigEndian）来保证所有平台的计算结果一致。
	binary.BigEndian.PutUint32(buf[:], fd.FrameID)
	h.Write(buf[:]) // Hasher 的 Write 方法永不返回错误

	// 4. 依序写入 InputArray 的内容。
	// **重要**: 为了保证确定性，InputArray 在传入此函数前必须是已排序的。
	// 例如，可以根据 Uid 和 FrameID 进行排序。
	for i := range fd.InputArray {
		input := &fd.InputArray[i] // 使用指针避免复制
		binary.BigEndian.PutUint32(buf[:], input.Uid)
		h.Write(buf[:])
		binary.BigEndian.PutUint32(buf[:], input.FrameID)
		h.Write(buf[:])
		h.Write(input.Data)
	}

	// 5. 依序写入 Events 的内容。
	// **重要**: 同样，Events 数组也必须预先排序以保证确定性。
	// 例如，可以根据 FrameID 和 Data 的某些内容排序。
	for i := range fd.Events {
		event := &fd.Events[i]
		binary.BigEndian.PutUint32(buf[:], event.FrameID)
		h.Write(buf[:])
		h.Write(event.Data)
	}

	// 6. 计算并返回最终的 64位 哈希值。
	return h.Sum64()
}

// SetChecksum 是一个辅助方法，用于计算并设置 FrameData 的 Checksum 字段。
// 必须在发送之前调用
func (fd *FrameData) SetChecksum() {
	fd.Checksum = fd.CalculateChecksum()
}

// Snapshot 表示游戏世界的状态快照
type Snapshot interface {
	GetData() []byte
}
