package model

// 账户角色枚举（对应文档管理员/服主/玩家）
const (
	RoleAdmin  = "admin"  // 管理员（最高权限）
	RoleOwner  = "owner"  // 服主（单服权限）
	RolePlayer = "player" // 玩家（基础权限）
)

// Account 账户模型
type Account struct {
	ID         int    `gorm:"primaryKey"` // 账户ID
	Username   string `gorm:"unique"`     // 用户名
	Password   string // 密码哈希（不存明文）
	Role       string // 角色（admin/owner/player）
	ServerID   string // 所属服务器（仅服主有值，管理员/玩家为""）
	SessionID  string // 临时会话标识（简化鉴权用）
	SessionExp int64  // 会话过期时间（时间戳）
}

// 表名
func (Account) TableName() string {
	return "globalbank_accounts"
}
