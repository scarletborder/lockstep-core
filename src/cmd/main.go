package main

import (
	"flag"
	"fmt"
	"lockstep-core/src/constants"
	"lockstep-core/src/internal/defaults"
	"lockstep-core/src/internal/di"
	"lockstep-core/src/utils"
	"lockstep-core/src/utils/tls"
	"log"
	"path/filepath"
)

func Verbose() {
	fmt.Printf("%s version: %s\n", constants.APPNAME, constants.VERSION)
	dataDir, _ := utils.GetApplicationDataDirectory(constants.APPNAME)
	fmt.Printf("Data directory: %s\n", dataDir)
	// other verbose info can be added here
	// hash cert
	dir, _ := utils.GetApplicationDataDirectory(constants.APPNAME)
	certPath := filepath.Join(dir, constants.TLS_DIR, "cert.pem")
	fmt.Println(tls.CertToHash(certPath))

}

func main() {
	isVerbose := flag.Bool("v", false, "Print verbose output")
	flag.Parse()

	if *isVerbose {
		Verbose()
		return
	}

	// 使用对外暴露的初始化函数，并注入默认的 NewGameWorld 实现（CLI 保持向后兼容）
	// 生产中，外部调用方可以通过 di.InitializeWithGameWorld 注入自己的 NewGameWorld
	handlers, err := di.InitializeWithGameWorld(defaults.DefaultNewGameWorld)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// 注册路由
	handlers.RegisterHandlers()

	// 启动服务器
	if err := handlers.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
