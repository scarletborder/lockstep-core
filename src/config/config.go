package config

import (
	"lockstep-core/src/constants"
	"lockstep-core/src/utils"
	customTLS "lockstep-core/src/utils/tls"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// NewDefaultConfig 从本地目录创建默认配置
func NewDefaultConfig(dataDir *string) (*RuntimeConfig, error) {
	var err error
	if dataDir == nil {
		// 获取 TLS 证书路径
		rdataDir, err := utils.GetApplicationDataDirectory(constants.APPNAME)
		if err != nil {
			return nil, err
		}
		dataDir = &rdataDir
	}

	// general config
	var generalCfg GeneralConfig
	generalCfgPath := filepath.Join(*dataDir, constants.CONFIG)

	if _, err := os.Stat(generalCfgPath); err == nil {
		// 文件存在，尝试读取
		if _, err := toml.DecodeFile(generalCfgPath, &generalCfg); err != nil {
			// 读取失败，使用默认值
			log.Printf("读取配置文件 %s 失败，使用默认配置: %v", generalCfgPath, err)
			generalCfg.ApplyDefaults()
		}
	}

	// 从地址中提取主机部分用于证书生成
	host := *generalCfg.Addr
	// 如果地址包含端口，去掉端口部分
	if colonIndex := len(host) - 1; colonIndex > 0 {
		for i := len(host) - 1; i >= 0; i-- {
			if host[i] == ':' {
				host = host[:i]
				break
			}
		}
	}

	// tls config
	tlsDir := filepath.Join(*dataDir, constants.TLS_DIR)
	tlsConfig, err := customTLS.GetTLSConfigFromPath(tlsDir, host)
	if err != nil {
		return nil, err
	}

	return &RuntimeConfig{
		GeneralConfig:      generalCfg,
		TLSConfig:          tlsConfig,
		CheckOriginEnabled: false, // 生产环境建议启用
	}, nil
}
