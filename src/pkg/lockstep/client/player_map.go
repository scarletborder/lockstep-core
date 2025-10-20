package client

import "sync"

// PlayerMap 封装 sync.Map，强制 key 为 int，value 为 *Player
type PlayerMap struct {
	m sync.Map
}

// Store 存储玩家
func (pm *PlayerMap) Store(uid uint32, value *Client) {
	pm.m.Store(uid, value)
}

// Load 加载玩家
func (pm *PlayerMap) Load(uid uint32) (*Client, bool) {
	v, ok := pm.m.Load(uid)
	if !ok {
		return nil, false
	}
	p, ok := v.(*Client)
	return p, ok
}

// Delete 删除玩家
func (pm *PlayerMap) Delete(uid uint32) {
	pm.m.Delete(uid)
}

// Range 遍历所有玩家
func (pm *PlayerMap) Range(f func(key uint32, value *Client) bool) {
	pm.m.Range(func(k, v interface{}) bool {
		ki, ok1 := k.(uint32)
		vp, ok2 := v.(*Client)
		if !ok1 || !ok2 {
			return true
		}
		return f(ki, vp)
	})
}

// Len 返回玩家数量
func (pm *PlayerMap) Len() int {
	count := 0
	pm.m.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func (pm *PlayerMap) ToSlice() []uint32 {
	ids := make([]uint32, 0)
	pm.m.Range(func(k, _ interface{}) bool {
		ki, ok := k.(uint32)
		if ok {
			ids = append(ids, ki)
		}
		return true
	})
	return ids
}
