package config

import (
	"crypto/tls"
	"lockstep-core/src/constants"
	"lockstep-core/src/utils"
	customTLS "lockstep-core/src/utils/tls"
	"path/filepath"
)

// ServerConfig 包含服务器的配置信息
type ServerConfig struct {
	// 服务器监听地址
	Addr string

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

	tlsDir := filepath.Join(dataDir, constants.TLS_DIR)
	tlsConfig, err := customTLS.GetTLSFromPath(tlsDir)
	if err != nil {
		return nil, err
	}

	return &ServerConfig{
		Addr:               ":4433",
		TLSConfig:          tlsConfig,
		CheckOriginEnabled: true, // 生产环境建议启用
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
