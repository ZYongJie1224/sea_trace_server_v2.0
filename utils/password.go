package utils

import (
	"github.com/beego/beego/v2/core/logs"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 对密码进行哈希处理
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 验证密码是否匹配哈希
func CheckPasswordHash(password, hash string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	logs.Emergency(err)
	return err == nil
}
