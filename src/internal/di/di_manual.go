package di

import (
	"lockstep-core/src/config"
	"lockstep-core/src/internal/server"
	"lockstep-core/src/pkg/lockstep/room"
	"log"
)

// InitializeApplicationManual 手动构造应用程序依赖（不依赖 wire 生成代码）
func InitializeApplicationManual(newGameWorld room.NewGameWorldFunc) (*server.Serverandlers, error) {
	// 使用默认的数据目录（nil）创建配置
	cfg, err := config.NewDefaultConfig(nil)
	if err != nil {
		return nil, err
	}

	// 创建 room manager
	rm := room.NewRoomManager(newGameWorld, cfg)

	// 创建 server core
	sc := server.NewServerCore(cfg)

	// 创建 handlers
	handlers := server.NewHTTPHandlers(rm, sc)

	log.Printf("Initialized application manually (without wire)")
	return handlers, nil
}
