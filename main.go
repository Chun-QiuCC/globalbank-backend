package main

import (
	"globalbank-backend/api"
	"globalbank-backend/config"
	"globalbank-backend/db"

	//"globalbank-backend/service" // main 作为入口，可合法依赖 service
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 先初始化 config（符合“基础配置优先”，🔶1-40）
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("加载配置失败：%v", err)
	}

	// 2. 再初始化 db（依赖 config，符合“db 层获取配置”，🔶1-25）
	if err := db.InitMySQL(); err != nil {
		log.Fatalf("初始化 MySQL 失败：%v", err)
	}
	log.Println("MySQL 初始化成功（符合文档“统一数据库管理”需求 🔶1-25）")

	// 3. 最后执行业务初始化（如创建测试账户，依赖 db，符合“service 依赖 db”，🔶1-6） // **移除默认生成初始测试账户**
	// if err := service.CreateTestAccount(); err != nil {
	// 	log.Printf("测试账户已存在：%v（符合文档“分级账户体系”需求 🔶1-15）", err)
	// } else {
	// 	log.Println("测试账户创建成功（管理员/服主/玩家，🔶1-15）")
	// }

	// 后续初始化 API、启动服务...（均符合文档“核心后端调度各模块” 🔶1-6）
	r := gin.Default()
	apiGroup := r.Group("/api")
	api.RegisterWebAPI(apiGroup)
	api.RegisterSpigotAPI(apiGroup)

	port := config.GetServerConfig().Port
	log.Printf("后端服务启动：http://localhost:%s", port)
	log.Fatal(r.Run(":" + port))
}
