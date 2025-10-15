package main

import (
	"flag"
	"lockstep-core/src/constants"
	"lockstep-core/src/utils"
	"log"
)

func main() {
	isVersion := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	if *isVersion {
		log.Printf("%s version: %s\n", constants.APPNAME, constants.VERSION)
		dataDir, _ := utils.GetApplicationDataDirectory(constants.APPNAME)
		log.Printf("Data directory: %s\n", dataDir)
		return
	}

	// 使用 Wire 生成的初始化函数
	handlers, err := InitializeApplication()
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
