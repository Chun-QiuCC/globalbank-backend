package service

import (
	"errors"
	"globalbank-backend/db"
	"globalbank-backend/model"
	"globalbank-backend/utils"
	"sync"
	"time"
)

// 本地会话存储（用户≤200人，map足够轻量）
var (
	sessionMap = make(map[string]*model.Account)
	sessionMu  sync.RWMutex
)

// Login 登录：验证账户密码，生成临时会话
func Login(username, password string) (*model.Account, string, error) {
	var account model.Account
	// 从数据库查询账户
	if err := db.DB.Where("username = ?", username).First(&account).Error; err != nil {
		return nil, "", errors.New("账户不存在")
	}

	// 验证密码（utils中实现BCrypt哈希校验）
	if !utils.CheckPasswordHash(password, account.Password) {
		return nil, "", errors.New("密码错误")
	}

	// 生成会话标识（UUID）和过期时间（2小时）
	sessionID := utils.GenerateUUID()
	account.SessionID = sessionID
	account.SessionExp = time.Now().Add(2 * time.Hour).Unix()

	// 更新数据库会话信息，并写入本地会话map
	db.DB.Save(&account)
	sessionMu.Lock()
	sessionMap[sessionID] = &account
	sessionMu.Unlock()

	return &account, sessionID, nil
}

// VerifySession 校验会话：判断会话是否有效，返回账户信息
func VerifySession(sessionID string) (*model.Account, error) {
	sessionMu.RLock()
	account, ok := sessionMap[sessionID]
	sessionMu.RUnlock()

	// 会话不存在或已过期
	if !ok || account.SessionExp < time.Now().Unix() {
		// 清理过期会话
		if ok {
			sessionMu.Lock()
			delete(sessionMap, sessionID)
			sessionMu.Unlock()
		}
		return nil, errors.New("会话无效或已过期")
	}
	return account, nil
}
