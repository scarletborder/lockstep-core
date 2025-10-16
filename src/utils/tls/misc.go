package tls

// cert to hash

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strings"
)

func CertToHash(certPath string) string {
	// 1. 读取证书文件
	pemBytes, err := os.ReadFile(certPath)
	if err != nil {
		log.Fatalf("无法读取证书文件 %s: %v\n请确保服务器已经运行过一次以生成证书，并且路径正确。", certPath, err)
	}
	// 2. 解码 PEM 数据块
	// WebTransport 要求对证书的 DER 编码（即原始二进制内容）进行哈希
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		log.Fatalf("无法从 PEM 数据块解码。文件可能已损坏或格式不正确。")
	}

	if block.Type != "CERTIFICATE" {
		log.Fatalf("PEM 文件中的数据块类型不是 'CERTIFICATE'，而是 '%s'", block.Type)
	}

	// block.Bytes 就是证书的 DER 编码内容
	derBytes := block.Bytes

	// 3. 计算 SHA-256 哈希
	hash := sha256.Sum256(derBytes)

	// 4. 将哈希格式化为十六进制字符串
	// ToUpper 用于匹配 openssl 命令的输出风格，更易读
	hashStringWithoutColons := hex.EncodeToString(hash[:])
	return hashStringWithoutColons

	// hashStringWithColons := toHexWithColons(hash[:])
	// 打印结果
	// fmt.Println("==============================================================================")
	// fmt.Printf("成功计算证书的 SHA-256 哈希值:\n\n")
	// fmt.Printf(" -> 带冒号格式 (用于 openssl 对比):\n    %s\n\n", strings.ToUpper(hashStringWithColons))
	// fmt.Printf(" -> 无冒号格式 (用于前端JS代码):\n    %s\n\n", strings.ToUpper(hashStringWithoutColons))
	// fmt.Println("请将【无冒号格式】的哈希值复制到你的前端 StreamClient 配置中。")
	// fmt.Println("==============================================================================")
}

// toHexWithColons 是一个辅助函数，用于将字节切片格式化为带冒号的十六进制字符串
func toHexWithColons(data []byte) string {
	var hexParts []string
	for _, b := range data {
		hexParts = append(hexParts, fmt.Sprintf("%02X", b))
	}
	return strings.Join(hexParts, ":")
}
