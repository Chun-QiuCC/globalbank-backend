package db

import (
	"globalbank-backend/config"
	"globalbank-backend/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitMySQL 初始化MySQL连接
func InitMySQL() error {
	// 读取配置（config中定义MySQL地址、账号、密码）
	cfg := config.GetMySQLConfig()
	dsn := cfg.User + ":" + cfg.Password + "@tcp(" + cfg.Addr + ")/" + cfg.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动创建表（账户表、货币表）
	err = db.AutoMigrate(&model.Account{}, &model.Currency{})
	if err != nil {
		return err
	}

	DB = db
	return nil
}
