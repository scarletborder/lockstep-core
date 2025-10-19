// Package utils 提供了一个高性能、并发安全的、可自动扩容的泛型分片切片。
// 它内部使用 int 作为索引，与 Go 原生切片行为保持一致，支持任意类型的值。
package utils

import (
	"errors"
	"sync"
)

const (
	// shardSizeLog2 定义了每个分片的大小，2^8 = 256
	shardSizeLog2 = 8
	// shardSize 是每个分片的实际大小
	shardSize = 1 << shardSizeLog2
)

// ErrIndexOutOfRange 是一个公共错误，当索引为负数时返回。
var ErrIndexOutOfRange = errors.New("index out of range: cannot be negative")

// item 是存储在分片中的元素，使用指针以支持 nil 值，表示空槽位。
type item[ValueT any] *ValueT

// shard 代表一个数据分片，包含一个读写锁和元素切片。
type shard[ValueT any] struct {
	sync.RWMutex
	items []item[ValueT]
}

// ShardedSlice 是并发安全的分片切片结构。
type ShardedSlice[ValueT any] struct {
	shards []*shard[ValueT]
	mu     sync.RWMutex // 用于保护 shards 切片本身的锁
}

// New 创建并返回一个新的 ShardedSlice 实例。
func New[ValueT any]() *ShardedSlice[ValueT] {
	return &ShardedSlice[ValueT]{
		shards: make([]*shard[ValueT], 0),
	}
}

// getOrCreateShard 根据分片索引获取或创建对应的分片。
func (s *ShardedSlice[ValueT]) getOrCreateShard(shardIndex int) *shard[ValueT] {
	s.mu.RLock()
	if shardIndex < len(s.shards) {
		sh := s.shards[shardIndex]
		s.mu.RUnlock()
		return sh
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	// 双重检查，防止在获取写锁期间其他 goroutine 已经创建了分片
	if shardIndex < len(s.shards) {
		return s.shards[shardIndex]
	}

	// 扩容 shards 切片直到能容纳 shardIndex
	newShards := make([]*shard[ValueT], shardIndex+1)
	copy(newShards, s.shards)
	for i := len(s.shards); i <= shardIndex; i++ {
		newShards[i] = &shard[ValueT]{
			items: make([]item[ValueT], 0, shardSize),
		}
	}
	s.shards = newShards

	return s.shards[shardIndex]
}

// Set 在指定的索引处设置值。如果索引为负，会返回错误。
func (s *ShardedSlice[ValueT]) Set(index int, value ValueT) error {
	// 索引必须是非负数
	if index < 0 {
		return ErrIndexOutOfRange
	}

	shardIndex := index >> shardSizeLog2
	shard := s.getOrCreateShard(shardIndex)

	indexInShard := index & (shardSize - 1)

	shard.Lock()
	defer shard.Unlock()

	// 如果分片内部的 items 切片长度不足，则扩容
	if indexInShard >= len(shard.items) {
		// 创建一个足够大的新切片，并用 nil 填充中间的空位
		newItems := make([]item[ValueT], indexInShard+1)
		copy(newItems, shard.items)
		shard.items = newItems
	}

	shard.items[indexInShard] = &value
	return nil
}

// Get 读取指定索引处的值。如果索引无效或该位置没有值，则返回零值和 false。
func (s *ShardedSlice[ValueT]) Get(index int) (ValueT, bool) {
	var zeroVal ValueT

	// 索引必须是非负数
	if index < 0 {
		return zeroVal, false
	}

	shardIndex := index >> shardSizeLog2

	s.mu.RLock()
	// 如果分片索引超出了当前已存在的分片范围，则该值肯定不存在
	if shardIndex >= len(s.shards) {
		s.mu.RUnlock()
		return zeroVal, false
	}
	shard := s.shards[shardIndex]
	s.mu.RUnlock()

	indexInShard := index & (shardSize - 1)

	shard.RLock()
	defer shard.RUnlock()

	// 如果分片内索引超出了范围，或者该位置的值为 nil，则认为值不存在
	if indexInShard >= len(shard.items) || shard.items[indexInShard] == nil {
		return zeroVal, false
	}

	return *shard.items[indexInShard], true
}
