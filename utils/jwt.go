package utils

import (
	"errors"
	"time"

	"github.com/beego/beego/v2/server/web"
	"github.com/golang-jwt/jwt"
)

// JWTClaims 自定义JWT声明
type JWTClaims struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CompanyID int    `json:"company_id"`
	jwt.StandardClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID int, username, role string, companyID int) (string, error) {
	// 从配置获取密钥
	secretKey, err := web.AppConfig.String("JWTSecretKey")
	if err != nil {
		return "", err
	}

	// 创建Token
	claims := JWTClaims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		CompanyID: companyID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "sea_trace_system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*JWTClaims, error) {
	// 从配置获取密钥
	secretKey, err := web.AppConfig.String("JWTSecretKey")
	if err != nil {
		return nil, err
	}

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌有效性
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}
