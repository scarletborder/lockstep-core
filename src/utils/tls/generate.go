package tls

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

func GenerateTLSPEM() (privPEM, certPEM []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
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

func GetTLSFromPath(dir string) (*tls.Config, error) {
	certPath := dir + "/cert.pem"
	keyPath := dir + "/key.pem"

	// 尝试从磁盘加载证书
	tlsCert, err := tls.LoadX509KeyPair(certPath, keyPath)

	if err == nil {
		return GetTLSConfigFromCert(tlsCert), nil
	}

	// 证书不存在或加载失败，生成新证书
	log.Printf("Certificate not found in %s, generating new self-signed certificate...", dir)
	privPEM, certPEM, err := GenerateTLSPEM()
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
