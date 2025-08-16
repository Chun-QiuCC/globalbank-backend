package model

// Currency 多服货币模型（按服务器独立存储）
type Currency struct {
	ID          int     `gorm:"primaryKey"` // ID
	ServerID    string  `gorm:"index"`      // 服务器标识（如ServerA/ServerB）
	PlayerID    string  `gorm:"index"`      // 玩家ID（对应游戏内ID）
	Balance     float64 // 货币余额
	TotalIssued float64 // 本服总发行量（仅管理员/服主可修改）
}

// 表名
func (Currency) TableName() string {
	return "globalbank_currencies"
}
