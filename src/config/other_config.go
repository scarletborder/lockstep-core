package config

type GeneralConfig struct {
	// 服务器监听地址
	Addr string
}

type LockstepConfig struct {
	// 帧间隔
	FrameInterval uint32

	// 最大延迟帧数,接受迟到帧
	MaxDelayFrames uint32
}
