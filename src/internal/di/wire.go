//go:build wireinject
// +build wireinject

package di

import (
	"lockstep-core/src/config"
	"lockstep-core/src/internal/server"
	"lockstep-core/src/pkg/lockstep/room"

	"github.com/google/wire"
)

// InitializeApplication 初始化整个应用程序
// Wire 会自动生成这个函数的实现
// 增加一个参数 newGameWorld 用于注入外部实现的游戏世界构造函数
func InitializeApplication(newGameWorld room.NewGameWorldFunc) (*server.Serverandlers, error) {
	wire.Build(
		config.ProviderSet,
		server.ProviderSet,
	)
	return nil, nil
}
