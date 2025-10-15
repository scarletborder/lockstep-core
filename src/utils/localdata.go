package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// 获取本机上本软件的目录

// GetApplicationDataDirectory 获取应用程序数据目录的路径
// 这通常是存放用户特定数据（如数据库、配置、缓存等）的最佳位置。
// 它会自动适配 Windows, macOS, Linux。
func GetApplicationDataDirectory(appName string) (string, error) {
	// 修正：xdg.DataHome 是一个字符串变量，直接引用即可，不需要加 ()
	baseDir := xdg.DataHome

	// appDataDir 会是例如：
	// Windows: C:\Users\<username>\AppData\Local\MyAwesomeWailsApp
	// macOS: /Users/<username>/Library/Application Support/MyAwesomeWailsApp
	// Linux: ~/.local/share/MyAwesomeWailsApp
	appDataDir := filepath.Join(baseDir, appName)

	// 确保目录存在，如果不存在则创建
	// 0755 表示 rwx for owner, rx for group, rx for others
	// 或者 0700 (rwx for owner, no permissions for others) 如果数据非常私密
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create application data directory %q: %w", appDataDir, err)
	}

	return appDataDir, nil
}

// GetDataPath 根据应用名和数据名获取数据文件的完整路径
func GetDataPath(appName, dataFileName string) (string, error) {
	// 获取应用数据目录
	dataDir, err := GetApplicationDataDirectory(appName)
	if err != nil {
		return "", fmt.Errorf("could not get application data %s directory: %w", dataFileName, err)
	}
	return filepath.Join(dataDir, dataFileName), nil
}
