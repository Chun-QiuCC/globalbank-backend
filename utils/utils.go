package utils

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 密码加密（把明文密码转哈希，存数据库）
func HashPassword(password string) (string, error) {
	// GenerateFromPassword 生成哈希，cost=10（复杂度适中，新手不用改）
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash 校验密码（明文密码 vs 数据库哈希）
func CheckPasswordHash(password, hash string) bool {
	// CompareHashAndPassword 对比，返回nil表示匹配
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateUUID 生成UUID（作为会话标识，唯一不重复）
func GenerateUUID() string {
	return uuid.NewString()
}
