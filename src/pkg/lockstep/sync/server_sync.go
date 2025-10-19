package lockstep_sync

import (
	"lockstep-core/src/pkg/lockstep/world"
	"lockstep-core/src/utils"
	"sync/atomic"
)

type ServerSyncData struct {
	// info

	// 发送给客户端渲染的数据是为了步进到达的帧号
	NextFrameID *atomic.Uint32

	// stored data

	// frameDatas
	FrameDatas utils.ShardedSlice[world.FrameData]

	// snapshots
	Snapshots utils.ShardedSlice[world.Snapshot]

	// 默认全局chunkID=0
	// FUTURE: 未来改为 map shardedslice 以支持多chunk
}

func NewServerSyncData() *ServerSyncData {
	nextRenderFrame := &atomic.Uint32{}
	nextRenderFrame.Store(1) // 下一帧渲染为 1，当前都在 0
	return &ServerSyncData{
		NextFrameID: nextRenderFrame,
	}
}

func (ssd *ServerSyncData) Reset() {
	ssd.NextFrameID.Store(1) // 重置帧 ID 为 1
}
