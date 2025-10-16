package tls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

// GenerateTLSPEM 生成自签名 TLS 证书
// host 参数可以是 IP 地址或域名，用于设置证书的 SAN (Subject Alternative Name)
func GenerateTLSPEM(host string) (privPEM, certPEM []byte, err error) {
	// 生成随机序列号
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Failed to generate random serial number: %v", err)
	}
	serial := int64(binary.BigEndian.Uint64(b))
	if serial < 0 {
		serial = -serial
	}

	// 使用 ECDSA P-256 生成密钥对
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		Subject:               pkix.Name{},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		log.Fatalf("Failed to marshal private key: %v", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	privPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	return privPEM, certPEM, nil
}

func GetTLSConfigFromCert(tlsCert tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"h3"},
	}
}

func GetTLSConfigFromPEM(certPEM, privPEM []byte) (*tls.Config, error) {
	tlsCert, err := tls.X509KeyPair(certPEM, privPEM)
	if err != nil {
		return nil, err
	}
	return GetTLSConfigFromCert(tlsCert), nil
}

// GetTLSFromPath 从指定路径加载或生成 TLS 配置
// host 参数用于在生成新证书时设置 SAN
func GetTLSFromPath(dir string, host string) (*tls.Config, error) {
	certPath := dir + "/cert.pem"
	keyPath := dir + "/key.pem"

	// 尝试从磁盘加载证书
	tlsCert, err := tls.LoadX509KeyPair(certPath, keyPath)

	if err == nil {
		log.Printf("Loaded existing TLS certificate from %s", dir)
		return GetTLSConfigFromCert(tlsCert), nil
	}

	// 证书不存在或加载失败，生成新证书
	log.Printf("Certificate not found in %s, generating new self-signed certificate...", dir)
	privPEM, certPEM, err := GenerateTLSPEM(host)
	if err != nil {
		return nil, err
	}

	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	// 保存证书到磁盘
	if err := os.WriteFile(certPath, certPEM, 0644); err != nil {
		log.Fatalf("Failed to write certificate to %s: %v", certPath, err)
	}
	if err := os.WriteFile(keyPath, privPEM, 0600); err != nil {
		log.Fatalf("Failed to write private key to %s: %v", keyPath, err)
	}

	log.Printf("Successfully generated and saved TLS certificate to %s", dir)
	tlsConfig, err := GetTLSConfigFromPEM(certPEM, privPEM)
	if err != nil {
		return nil, err
	}
	return tlsConfig, nil
}
