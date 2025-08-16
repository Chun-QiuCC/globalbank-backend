package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// MySQLConfig/ServerConfig 仅定义配置结构（符合文档“集中管理配置”需求 🔶1-40）
type MySQLConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Addr     string `yaml:"addr"`
	DBName   string `yaml:"dbname"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

var (
	MySQLCfg  MySQLConfig
	ServerCfg ServerConfig
)

// LoadConfig 仅加载配置文件，不调用任何业务函数（符合文档“配置模块仅管配置”定位）
func LoadConfig() error {
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return err
	}
	type Config struct {
		MySQL  MySQLConfig  `yaml:"mysql"`
		Server ServerConfig `yaml:"server"`
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}
	MySQLCfg = cfg.MySQL
	ServerCfg = cfg.Server
	return nil
}

// 仅保留配置获取函数，无任何业务逻辑
func GetMySQLConfig() MySQLConfig   { return MySQLCfg }
func GetServerConfig() ServerConfig { return ServerCfg }
