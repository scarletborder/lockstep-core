package main

import (
	"flag"
	"fmt"
	"lockstep-core/src/constants"
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
