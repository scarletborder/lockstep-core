package main

import (
	"flag"
	"fmt"
	"lockstep-core/src/app"
	"lockstep-core/src/constants"
	"lockstep-core/src/internal/defaults"
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

	// 使用对外导出的 app 包启动，内部默认使用 internal/defaults.DefaultNewGameWorld
	if err := app.StartWith(defaults.DefaultNewGameWorld); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
