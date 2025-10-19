package app

import (
	"lockstep-core/src/internal/di"
	"lockstep-core/src/internal/server"
	"lockstep-core/src/pkg/lockstep/room"
)

// NewHandlers 使用外部提供的 newGameWorld 构造函数初始化并返回 handlers
// 这是对外可见的入口，隐藏了 internal/di 的实现细节
func NewHandlers(newGameWorld room.NewGameWorldFunc) (*server.Serverandlers, error) {
	return di.InitializeWithGameWorld(newGameWorld)
}

// StartWith 直接初始化并启动服务器（阻塞直到服务器返回或出错）
func StartWith(newGameWorld room.NewGameWorldFunc) error {
	handlers, err := NewHandlers(newGameWorld)
	if err != nil {
		return err
	}
	handlers.RegisterHandlers()
	return handlers.Start()
}
