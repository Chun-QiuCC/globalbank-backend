package api

import (
	"globalbank-backend/model"
	"globalbank-backend/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterWebAPI 注册网页对接API路由
func RegisterWebAPI(router *gin.RouterGroup) {
	web := router.Group("/web")
	{
		// 1. 登录接口（简化鉴权入口）
		web.POST("/login", func(c *gin.Context) {
			var req struct {
				Username string `json:"username" binding:"required"`
				Password string `json:"password" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"err": "参数错误"})
				return
			}

			account, sessionID, err := service.Login(req.Username, req.Password)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"err": err.Error()})
				return
			}

			// 返回会话标识和账户角色（供网页前端加载对应模块）
			c.JSON(http.StatusOK, gin.H{
				"session_id": sessionID,
				"role":       account.Role,
				"server_id":  account.ServerID,
			})
		})

		// 2. 货币查询接口（按角色权限过滤）
		web.GET("/currency/query", authMiddleware, func(c *gin.Context) {
			account := c.MustGet("account").(*model.Account)
			serverID := c.Query("server_id")

			// 权限校验：管理员可查任意服，服主仅查本服，玩家查自己所属服
			if account.Role == model.RoleOwner && account.ServerID != serverID {
				c.JSON(http.StatusForbidden, gin.H{"err": "无权限查询其他服务器"})
				return
			}

			// 调用服务层查询货币数据
			currencies, err := service.QueryCurrency(serverID, account.ID, account.Role)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": currencies})
		})

		// 3. 货币发行接口（管理员全服/服主单服）
		web.POST("/currency/issue", authMiddleware, func(c *gin.Context) {
			// 实现逻辑：校验角色权限→修改对应服务器货币总量→返回结果
			// （具体代码参考QueryCurrency，核心是调用service.IssueCurrency）
		})
	}
}

// authMiddleware 会话鉴权中间件（网页API通用）
func authMiddleware(c *gin.Context) {

	sessionID := c.GetHeader("X-Session-ID") // c.GetHeader() 不是 c.Header() SB豆包，查半天也不对

	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"err":  "未携带会话标识",
			"msg":  "请先通过登录接口获取 X-Session-ID，再携带该请求头访问",
			"docs": "对应文档“网页前端单入口登录+服务端鉴权”规则（🔶1-28）",
		})
		c.Abort() // 终止请求，不执行后续逻辑
		return
	}

	// 3. 校验会话有效性（调用 service 层鉴权逻辑，符合文档“核心后端权限校验”定位 🔶1-6）
	validAccount, err := service.VerifySession(sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"err":  "会话无效或已过期：" + err.Error(),
			"msg":  "请重新登录获取新的会话标识",
			"docs": "对应文档“服务端鉴权验证账户权限”需求（🔶1-28）",
		})
		c.Abort()
		return
	}

	// 4. 会话有效：将合法账户信息存入上下文，供后续接口使用（符合文档“模块协同”逻辑 🔶1-5）
	c.Set("account", validAccount)
	c.Next() // 继续执行后续接口
}
