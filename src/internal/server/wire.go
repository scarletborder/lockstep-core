package server

import (
	"github.com/google/wire"
)

// ProviderSet 是 server 模块的 Wire provider set
var ProviderSet = wire.NewSet(
	NewServerCore,
	NewHTTPHandlers,
)
