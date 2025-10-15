package clients

import "sync"

// PlayerMap 封装 sync.Map，强制 key 为 int，value 为 *Player
type PlayerMap struct {
	m sync.Map
}

// Store 存储玩家
func (pm *PlayerMap) Store(key int, value *Player) {
	pm.m.Store(key, value)
}

// Load 加载玩家
func (pm *PlayerMap) Load(key int) (*Player, bool) {
	v, ok := pm.m.Load(key)
	if !ok {
		return nil, false
	}
	p, ok := v.(*Player)
	return p, ok
}

// Delete 删除玩家
func (pm *PlayerMap) Delete(key int) {
	pm.m.Delete(key)
}

// Range 遍历所有玩家
func (pm *PlayerMap) Range(f func(key int, value *Player) bool) {
	pm.m.Range(func(k, v interface{}) bool {
		ki, ok1 := k.(int)
		vp, ok2 := v.(*Player)
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
