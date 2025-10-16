package config

import (
	"crypto/tls"
	"lockstep-core/src/constants"
	"lockstep-core/src/utils"
	customTLS "lockstep-core/src/utils/tls"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type GeneralConfig struct {
	// 服务器监听地址
	Addr string
}

// ServerConfig 包含服务器的配置信息
type ServerConfig struct {
	GeneralConfig

	// TLS 配置
	TLSConfig *tls.Config

	// 是否启用 Origin 检查
	CheckOriginEnabled bool
}

// NewDefaultConfig 创建默认配置
func NewDefaultConfig() (*ServerConfig, error) {
	// 获取 TLS 证书路径
	dataDir, err := utils.GetApplicationDataDirectory(constants.APPNAME)
	if err != nil {
		return nil, err
	}

	// general config
	generalCfgPath := filepath.Join(dataDir, constants.CONFIG)
	generalCfg := GeneralConfig{Addr: "127.0.0.1:4433"} // 默认值

	if _, err := os.Stat(generalCfgPath); err == nil {
		// 文件存在，尝试读取
		if _, err := toml.DecodeFile(generalCfgPath, &generalCfg); err != nil {
			// 读取失败，使用默认值
		}
	}

	// 从地址中提取主机部分用于证书生成
	host := generalCfg.Addr
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
	tlsDir := filepath.Join(dataDir, constants.TLS_DIR)
	tlsConfig, err := customTLS.GetTLSConfigFromPath(tlsDir, host)
	if err != nil {
		return nil, err
	}

	return &ServerConfig{
		GeneralConfig:      generalCfg,
		TLSConfig:          tlsConfig,
		CheckOriginEnabled: false, // 生产环境建议启用
	}, nil
}

// WithAddr 设置服务器地址
func (c *ServerConfig) WithAddr(addr string) *ServerConfig {
	c.Addr = addr
	return c
}

// WithTLSConfig 设置 TLS 配置
func (c *ServerConfig) WithTLSConfig(tlsConfig *tls.Config) *ServerConfig {
	c.TLSConfig = tlsConfig
	return c
}

// WithCheckOrigin 设置是否检查 Origin
func (c *ServerConfig) WithCheckOrigin(enabled bool) *ServerConfig {
	c.CheckOriginEnabled = enabled
	return c
}
