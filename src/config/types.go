package config

import "crypto/tls"

type ServerConfig struct {
	// 服务器监听地址
	Addr *string `toml:"addr"`
}

const DefaultAddr = "127.0.0.1:4433"

type LockstepConfig struct {
	// 帧间隔
	FrameInterval *uint32 `toml:"frame_interval"`

	// 容忍的最大延迟帧数,接受迟到帧
	// 建议为1s表示的帧数，例如FrameInterval=100ms则设置为10
	// -1 为不限制
	MaxDelayFrames *int32 `toml:"max_delay_frames"`

	// 是否启用确定性锁步，
	// 如果此值为-1，表示乐观锁步
	// 如果此值大于0，表示悲观锁步，且为等待确认的最大帧数
	DeterministicLockstep *int32 `toml:"deterministic_lockstep"`

	// 最大人数
	MaxClientsPerRoom *uint16 `toml:"max_clients_per_room"`
}

const (
	DefaultFrameInterval         = 66       // 默认帧间隔 66ms (~15fps)
	DefaultMaxDelayFrames        = 500 / 66 // 默认最大延迟帧500ms
	DefaultMaxClientsPerRoom     = 8        // 默认每个房间最大人数 8 人
	DefaultDeterministicLockstep = -1       // 默认乐观锁步
)

type GeneralConfig struct {
	ServerConfig   `toml:"server"`
	LockstepConfig `toml:"lockstep"`
}

func Uint32Ptr(v uint32) *uint32 {
	return &v
}

func Int32Ptr(v int32) *int32 {
	return &v
}

func Uint16Ptr(v uint16) *uint16 {
	return &v
}

func (c *GeneralConfig) ApplyDefaults() {
	if c.FrameInterval == nil {
		c.FrameInterval = Uint32Ptr(DefaultFrameInterval)
	}
	if c.MaxDelayFrames == nil {
		c.MaxDelayFrames = Int32Ptr(DefaultMaxDelayFrames)
	}
	if c.MaxClientsPerRoom == nil {
		c.MaxClientsPerRoom = Uint16Ptr(DefaultMaxClientsPerRoom)
	}
	if c.DeterministicLockstep == nil {
		c.DeterministicLockstep = Int32Ptr(DefaultDeterministicLockstep)
	}
}

// RuntimeConfig 包含运行时的配置信息
type RuntimeConfig struct {
	GeneralConfig

	// TLS 配置
	TLSConfig *tls.Config

	// 是否启用 Origin 检查
	CheckOriginEnabled bool
}
