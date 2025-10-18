package lockstep_sync

import (
	"sync"
	"sync/atomic"
)

type ServerSyncData struct {
	//data

	// 发送给客户端渲染的下一帧 ID
	NextFrameID *atomic.Uint32
	// 某帧下一次操作的序号 (FrameID -> OperationID)
	OperationIDMaps sync.Map // map[uint32]uint32
}

func NewServerSyncData() *ServerSyncData {
	nextRenderFrame := &atomic.Uint32{}
	nextRenderFrame.Store(1) // 下一帧渲染为 1，当前都在 0
	return &ServerSyncData{
		NextFrameID:          nextRenderFrame,
		OperationIDMaps:      sync.Map{},
	}
}

func (ssd *ServerSyncData) Reset() {
	ssd.NextFrameID.Store(1) // 重置帧 ID 为 1

	// 清空 OperationIDMaps
	ssd.OperationIDMaps = sync.Map{}
}

// GetNextOperationID 获得某帧下一次操作的序号并自增
func (ssd *ServerSyncData) GetNextOperationID(frameID uint32) uint32 {
	v, _ := ssd.OperationIDMaps.LoadOrStore(frameID, uint32(1))
	nextID := v.(uint32)
	ssd.OperationIDMaps.Store(frameID, nextID+1)
	return nextID
}

// DeleteOperationID 在广播某帧后，删除其 map 记录
func (ssd *ServerSyncData) DeleteOperationID(frameID uint32) {
	ssd.OperationIDMaps.Delete(frameID)
}