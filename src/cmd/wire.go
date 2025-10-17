//go:build wireinject
// +build wireinject

package main

import (
	"lockstep-core/src/config"
	"lockstep-core/src/internal/server"
	"lockstep-core/src/logic"

	"github.com/google/wire"
)

// InitializeApplication 初始化整个应用程序
// Wire 会自动生成这个函数的实现
func InitializeApplication() (*server.HTTPHandlers, error) {
	wire.Build(
		config.ProviderSet,
		logic.ProviderSet,
		server.ProviderSet,
	)
	return nil, nil
}
