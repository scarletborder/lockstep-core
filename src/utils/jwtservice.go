package utils

import (
	"crypto/rand"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// GameTokenClaims 定义了JWT中存储的自定义信息。
type GameTokenClaims struct {
	UserID uint32 `json:"userID"`
	RoomID uint32 `json:"roomID"`
	jwt.RegisteredClaims
}

// JWTService 封装了与JWT相关的操作，每个实例都拥有自己独立的密钥。
type JWTService struct {
	secretKey []byte
}

// NewJWTService 创建一个新的JWTService实例。
// 它会生成一个32字节的加密级随机密钥，确保每个实例的密钥都是唯一的。
func NewJWTService() *JWTService {
	// 生成一个足够安全的密钥 (256 bits for HS256)
	key := make([]byte, 32)
	rand.Read(key)
	return &JWTService{secretKey: key}
}

// GenerateToken 为指定用户和房间生成一个永久有效的Token。
func (s *JWTService) GenerateToken(userID, roomID uint32) (string, error) {
	// 创建自定义的Claims
	// 注意：我们没有设置 ExpiresAt，所以这个Token是永久有效的。
	claims := &GameTokenClaims{
		UserID: userID,
		RoomID: roomID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "game-server-room", // 可以指定签发人
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用实例自身的私钥进行签名
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("为用户 %d 签名Token失败: %w", userID, err)
	}

	return tokenString, nil
}

// ParseToken 解析并验证Token。
// 返回解析出的Claims和布尔值，表示Token是否有效且被信任。
func (s *JWTService) ParseToken(tokenString string) (*GameTokenClaims, bool) {
	claims := &GameTokenClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法是否为HS256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名算法: %v", token.Header["alg"])
		}
		// 返回实例自身的私钥用于验证
		return s.secretKey, nil
	})

	// 如果在解析过程中出现任何错误（如签名不匹配、格式错误等），则认为Token不被信任。
	if err != nil {
		return nil, false
	}

	// 确保token.Valid为true，并且可以成功提取出Claims
	if claims, ok := token.Claims.(*GameTokenClaims); ok && token.Valid {
		return claims, true
	}

	return nil, false
}
