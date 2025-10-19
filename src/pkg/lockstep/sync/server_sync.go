package lockstep_sync

import (
	"lockstep-core/src/pkg/lockstep/world"
	"sync/atomic"

	"github.com/alphadose/haxmap"
)

type ServerSyncData struct {
	// info

	// 发送给客户端渲染的数据是为了步进到达的帧号
	NextFrameID *atomic.Uint32

	// stored data

	// frameDatas
	FrameDatas *haxmap.Map[uint32, *world.FrameData]

	// snapshots
	Snapshots *haxmap.Map[uint32, world.Snapshot]

	// 默认全局chunkID=0
	// FUTURE: 未来改为 map shardedslice 以支持多chunk
}

func NewServerSyncData() *ServerSyncData {
	nextRenderFrame := &atomic.Uint32{}
	nextRenderFrame.Store(1)      // 下一帧渲染为 1，当前都在 0
	const initFrameNum = 30 * 128 // 初始化30s的缓冲
	return &ServerSyncData{
		NextFrameID: nextRenderFrame,
		FrameDatas:  haxmap.New[uint32, *world.FrameData](initFrameNum),
		Snapshots:   haxmap.New[uint32, world.Snapshot](initFrameNum),
	}
}

func (ssd *ServerSyncData) Reset() {
	ssd.NextFrameID.Store(1) // 重置帧 ID 为 1
}

func (ssd *ServerSyncData) StoreFrame(frameID uint32, frameData *world.FrameData) {
	ssd.FrameDatas.Set(frameID, frameData)
}

func (ssd *ServerSyncData) StoreSnapshot(frameID uint32, snapshot world.Snapshot) {
	ssd.Snapshots.Set(frameID, snapshot)
}

func (ssd *ServerSyncData) GetFrame(frameID uint32) (*world.FrameData, bool) {
	return ssd.FrameDatas.Get(frameID)
}

func (ssd *ServerSyncData) GetSnapshot(frameID uint32) (world.Snapshot, bool) {
	return ssd.Snapshots.Get(frameID)
}
