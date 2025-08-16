package service

import (
	"errors"
	"fmt"
	"globalbank-backend/db"
	"globalbank-backend/model"
	"gorm.io/gorm"
)

// 1. QueryCurrency 货币查询：按服务器ID、账户角色过滤数据（对应 🔶1-25、🔶1-27）
// 参数说明：
// - serverID：目标服务器标识（如"ServerA"，管理员传空可查所有服）
// - accountID：当前登录账户ID（用于权限校验，非玩家ID）
// - accountRole：当前登录账户角色（admin/owner/player）
// 返回：符合权限的货币数据列表
func QueryCurrency(serverID string, accountID int, accountRole string) ([]model.Currency, error) {
	var currencies []model.Currency
	dbConn := db.DB

	// 权限过滤：遵循文档“分级权限管控”规则（🔶1-15、🔶1-30/31/32）
	switch accountRole {
	case model.RoleAdmin:
		// 管理员：可查所有服务器（serverID为空）或指定服务器
		if serverID != "" {
			dbConn = dbConn.Where("server_id = ?", serverID)
		}
	case model.RoleOwner:
		// 服主：仅能查自己所属服务器（需先获取服主绑定的ServerID）
		var ownerAccount model.Account
		if err := db.DB.Where("id = ?", accountID).First(&ownerAccount).Error; err != nil {
			return nil, errors.New("获取服主信息失败")
		}
		// 强制按服主所属服务器过滤，防止越权查询
		dbConn = dbConn.Where("server_id = ?", ownerAccount.ServerID)
	case model.RolePlayer:
		// 玩家：仅能查自己在各服的余额（需关联玩家ID，此处简化用accountID映射，实际可关联游戏内PlayerID）
		dbConn = dbConn.Where("player_id = ?", accountID) // 实际项目可优化为“账户ID-游戏PlayerID”关联表
	default:
		return nil, errors.New("无效账户角色，无查询权限")
	}

	// 执行查询，返回符合条件的货币数据
	if err := dbConn.Find(&currencies).Error; err != nil {
		return nil, errors.New("查询货币数据失败：" + err.Error())
	}
	return currencies, nil
}

// 2. SyncPlayerCurrency 玩家货币同步：接收服务端插件的实时变动（对应 🔶1-11、🔶1-26）
// 参数说明：
// - serverID：变动发生的服务器标识
// - playerID：游戏内玩家ID（如"Steve123"，对应文档“玩家账户”）
// - amount：变动金额（正数=增加，负数=减少，如任务奖励+100、购买道具-50）
// 返回：同步结果（成功/失败）
func SyncPlayerCurrency(serverID, playerID string, amount float64) error {
	dbConn := db.DB
	var currency model.Currency

	// 步骤1：查询玩家在该服务器的货币记录（按“server_id+player_id”唯一匹配，符合 🔶1-25“分别存储”）
	err := dbConn.Where("server_id = ? AND player_id = ?", serverID, playerID).First(&currency).Error
	if err != nil {
		// 若记录不存在（新玩家首次在该服产生货币变动），创建新记录
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 新记录默认总发行量继承该服当前总量（需先查该服总发行量，无则初始化为0）
			var serverTotal model.Currency
			dbConn.Where("server_id = ?", serverID).First(&serverTotal)

			newCurrency := model.Currency{
				ServerID:    serverID,
				PlayerID:    playerID,
				Balance:     amount,                  // 首次变动直接设为amount（如初始奖励+50）
				TotalIssued: serverTotal.TotalIssued, // 继承该服总发行量
			}
			if err := dbConn.Create(&newCurrency).Error; err != nil {
				return errors.New("创建玩家货币记录失败：" + err.Error())
			}
			return nil
		}
		// 其他查询错误（如数据库异常）
		return errors.New("查询玩家货币记录失败：" + err.Error())
	}

	// 步骤2：更新玩家余额（实时同步变动，对应 🔶1-11“玩家货币变动同步”）
	// 校验：余额不能为负（防止货币超支，保障经济稳定 🔶1-42）
	if currency.Balance+amount < 0 {
		return errors.New("玩家余额不足，无法完成扣减")
	}
	currency.Balance += amount

	// 执行更新
	if err := dbConn.Save(&currency).Error; err != nil {
		return errors.New("同步玩家货币变动失败：" + err.Error())
	}
	return nil
}

// 3. IssueCurrency 货币发行：调整服务器货币总量（对应 🔶1-34、🔶1-35）
// 参数说明：
// - serverID：目标服务器标识（必填，服主仅能传自己所属服务器）
// - newTotal：新的货币总发行量（需大于当前玩家总余额，避免总量小于已流通量）
// - accountID：当前登录账户ID（用于校验服主所属服务器）
// - accountRole：当前登录账户角色（admin/owner）
// 返回：发行结果（成功/失败）
func IssueCurrency(serverID string, newTotal float64, accountID int, accountRole string) error {
	// 前置校验：货币总量不能为负（保障经济可控性 🔶1-42）
	if newTotal < 0 {
		return errors.New("货币总发行量不能为负数")
	}

	dbConn := db.DB
	var (
		serverCurrencies []model.Currency
		ownerServerID    string
	)

	// 权限校验：遵循“分级发行管控”规则（🔶1-35）
	switch accountRole {
	case model.RoleAdmin:
		// 管理员：可发行任意服务器货币（无需额外校验）
	case model.RoleOwner:
		// 服主：仅能发行自己所属服务器货币（先获取服主绑定的ServerID）
		var ownerAccount model.Account
		if err := db.DB.Where("id = ?", accountID).First(&ownerAccount).Error; err != nil {
			return errors.New("获取服主信息失败：" + err.Error())
		}
		ownerServerID = ownerAccount.ServerID
		// 校验：服主只能操作自己的服务器
		if ownerServerID != serverID {
			return errors.New("服主无权限发行其他服务器货币（仅可操作：" + ownerServerID + "）")
		}
	default:
		return errors.New("仅管理员、服主可执行货币发行操作（玩家无权限）")
	}

	// 校验：新总量不能小于该服当前玩家总余额（防止“总量<流通量”，避免经济混乱 🔶1-42）
	if err := dbConn.Where("server_id = ?", serverID).Find(&serverCurrencies).Error; err != nil {
		return errors.New("查询服务器货币数据失败：" + err.Error())
	}
	var currentTotalBalance float64
	for _, c := range serverCurrencies {
		currentTotalBalance += c.Balance
	}
	if newTotal < currentTotalBalance {
		return errors.New("新发行量（" + fmt.Sprintf("%.2f", newTotal) + "）小于当前流通总量（" + fmt.Sprintf("%.2f", currentTotalBalance) + "），请调整")
	}

	// 执行发行：更新该服务器所有货币记录的“总发行量”字段（确保全服数据一致 🔶1-25）
	if err := dbConn.Model(&model.Currency{}).Where("server_id = ?", serverID).Update("total_issued", newTotal).Error; err != nil {
		return errors.New("更新货币发行量失败：" + err.Error())
	}
	return nil
}

// 4. GetPlayerBalance 玩家单服余额查询：供玩家查询自己在指定服的余额（对应 🔶1-22）
// 参数说明：
// - serverID：目标服务器标识
// - playerID：游戏内玩家ID
// 返回：玩家在该服的余额
func GetPlayerBalance(serverID, playerID string) (float64, error) {
	var currency model.Currency
	err := db.DB.Where("server_id = ? AND player_id = ?", serverID, playerID).First(&currency).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil // 新玩家无记录，余额默认为0
		}
		return 0, errors.New("查询玩家余额失败：" + err.Error())
	}
	return currency.Balance, nil
}
