package config

import (
	"github.com/google/wire"
)

// ProviderSet 是 config 模块的 Wire provider set
var ProviderSet = wire.NewSet(
	NewDefaultConfig,
)
