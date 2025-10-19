package di

import (
	"lockstep-core/src/internal/server"
	"lockstep-core/src/pkg/lockstep/room"
)

// initializeApp 是包级可替换的初始化器。默认实现为 InitializeApplicationManual。
// 如果使用 wire 生成的初始化函数，可以在生成的文件中将此变量替换为生成的实现以获得更优的构造。
var initializeApp func(room.NewGameWorldFunc) (*server.Serverandlers, error) = InitializeApplicationManual

// InitializeWithGameWorld 提供给外部使用的初始化函数
// newGameWorld 为外部实现的游戏世界构造函数
func InitializeWithGameWorld(newGameWorld room.NewGameWorldFunc) (*server.Serverandlers, error) {
	return initializeApp(newGameWorld)
}
