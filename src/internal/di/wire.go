//go:build wireinject
// +build wireinject

package di

import (
	"lockstep-core/src/config"
	"lockstep-core/src/internal/server"

	"github.com/google/wire"
)

// InitializeApplication 初始化整个应用程序
// Wire 会自动生成这个函数的实现
func InitializeApplication() (*server.Serverandlers, error) {
	wire.Build(
		config.ProviderSet,
		server.ProviderSet,
	)
	return nil, nil
}
