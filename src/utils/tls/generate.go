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
	"os"
	"path/filepath"
	"time"
)

func generateCert() (*x509.Certificate, *ecdsa.PrivateKey, error) {
	// default valid for 10 days
	start := time.Now()
	end := start.Add(10 * 24 * time.Hour)

	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return nil, nil, err
	}
	serial := int64(binary.BigEndian.Uint64(b))
	if serial < 0 {
		serial = -serial
	}
	certTempl := &x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		Subject:               pkix.Name{},
		NotBefore:             start,
		NotAfter:              end,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, certTempl, certTempl, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, nil, err
	}
	ca, err := x509.ParseCertificate(caBytes)
	if err != nil {
		return nil, nil, err
	}
	return ca, caPrivateKey, nil
}

func saveCertAndKey(cert *x509.Certificate, priv *ecdsa.PrivateKey, certPath string, keyPath string) error {
	// makedir -p
	certDir := filepath.Dir(certPath)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return err
	}
	keyDir := filepath.Dir(keyPath)
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		return err
	}

	// --- 保存证书文件 ---
	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	if err := pem.Encode(certOut, certBlock); err != nil {
		return err
	}
	log.Printf("Certificate saved to %s", certPath)

	// --- 保存私钥文件 ---
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	// 将ECDSA私钥序列化为DER格式
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return err
	}

	keyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY", // 使用 "EC PRIVATE KEY" 更具体
		Bytes: privBytes,
	}

	if err := pem.Encode(keyOut, keyBlock); err != nil {
		return err
	}
	log.Printf("Private key saved to %s", keyPath)

	return nil
}

// GetTLSConfigFromPath 从指定路径加载或生成 TLS 配置
// host 参数用于在生成新证书时设置 SAN
func GetTLSConfigFromPath(dir string, host string) (*tls.Config, error) {
	certPath := dir + "/cert.pem"
	keyPath := dir + "/key.pem"
	needGenerate := false
	// 测试文件是否存在
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		log.Printf("Certificate file %s does not exist.", certPath)
		needGenerate = true
	} else if err != nil {
		log.Fatalf("can not access cert file %s, please remove it manually", err.Error())
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		log.Printf("Key file %s does not exist.", keyPath)
		needGenerate = true
	} else if err != nil {
		log.Fatalf("can not access key file %s, please remove it manually", err.Error())
	}

	if needGenerate {
		cert, priv, err := generateCert()
		if err != nil {
			return nil, err
		}
		// save to disk
		if err := saveCertAndKey(cert, priv, certPath, keyPath); err != nil {
			log.Fatalf("can not save cert and key file to disk:%s", err.Error())
			return nil, err
		}

		return &tls.Config{
			Certificates: []tls.Certificate{{
				Certificate: [][]byte{cert.Raw},
				PrivateKey:  priv,
				Leaf:        cert,
			}},
		}, nil
	} else {
		// load from disk
		log.Printf("Loading certificate and key from %s and %s", certPath, keyPath)

		// 使用 tls.LoadX509KeyPair 直接从文件加载并构建 tls.Certificate
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}

		log.Println("Successfully loaded certificate and key from disk.")
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
		}, nil
	}
}
