package main

import (
	"log"
)

func main() {
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
