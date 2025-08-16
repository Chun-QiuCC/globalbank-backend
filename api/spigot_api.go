package api

import (
	"github.com/gin-gonic/gin"
	"globalbank-backend/service"
	"net/http"
)

// RegisterSpigotAPI 注册Spigot插件对接API路由
func RegisterSpigotAPI(router *gin.RouterGroup) {
	spigot := router.Group("/spigot")
	{
		// 1. 同步玩家货币变动（服务端插件→核心后端）
		spigot.POST("/currency/sync", func(c *gin.Context) {
			var req struct {
				ServerID string  `json:"server_id" binding:"required"` // 服务器标识
				PlayerID string  `json:"player_id" binding:"required"` // 游戏内玩家ID
				Amount   float64 `json:"amount" binding:"required"`    // 变动金额（+为增加，-为减少）
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"err": "参数错误"})
				return
			}

			// 调用服务层同步货币数据（对应文档“实时同步玩家货币变动”🔶1-11）
			if err := service.SyncPlayerCurrency(req.ServerID, req.PlayerID, req.Amount); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "success"})
		})

		// 2. 查询玩家跨服余额（服务端插件→核心后端）
		spigot.GET("/currency/player", func(c *gin.Context) {
			// 实现逻辑：接收服务器ID、玩家ID→查询多服余额→返回结果
		})
	}
}
