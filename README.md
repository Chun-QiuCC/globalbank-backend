## 说明
 - 这是 GlobalBank插件的 后端服务端 代码仓库。
 - GlobalBank是一个为 Minecraft Spigot 服务器设计的跨服经济管理插件套件。  
 - GlobalBank的作用是让多个服务器之间可以互相转账、兑换货币，可以在游戏外的网页上管理自己的存款账户，实现在线转账、兑换货币等功能。  
 - GlobalBank插件分为三个不同的部分，分别为服务器Spigot插件、前端网页和后端服务端，此仓库为 后端服务端。
 - 目前此项目进度较慢，并可能包含大量的BUG，欢迎各位技术大佬加入此项目。  
 - 联系邮箱： herodain@qq.com 或 86herodain@gmail.com
## TODO
- [ ] 添加webAPI的功能（通过Minecraft UUID 绑定web账户，在线转账等功能）
- [ ] 尽快开发Spigot插件以推断Spigot的功能 （游戏内操作）
## 仓库结构
```
globalbank-backend/
├── api/           # 双API接口实现（网页对接API + Spigot插件对接API）
│   ├── web_api.go # 网页前端交互接口（登录、权限、货币管理）
│   └── spigot_api.go # 服务端插件数据同步接口
├── config/        # 配置文件（MySQL连接、服务端口）
│   └── config.go
├── db/            # 数据库操作（MySQL）
│   └── mysql.go
├── model/         # 数据模型（对应文档分级账户、多服货币）
│   ├── account.go # 账户模型（管理员/服主/玩家）
│   └── currency.go # 货币模型（多服分别存储）
├── service/       # 业务逻辑（权限校验、货币处理）
│   ├── auth_service.go # 简化鉴权（账户密码+本地会话）
│   └── currency_service.go # 货币发行、同步、查询
├── utils/         # 工具类（密码哈希、会话管理）
│   └── utils.go
└── main.go        # 服务入口（初始化+路由注册）
```
## API 文档：

### Spigot插件对接API文档
 
1. 同步玩家货币变动
 
**接口描述：** 用于同步玩家货币变动，从服务端插件向核心后端传递玩家货币变化信息。
 
**请求URL：** `POST /api/spigot/currency/sync`
 
**请求参数：**
 
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| server_id | string | 是 | 服务器标识 |
| player_id | string | 是 | 游戏内玩家ID |
| amount | float64 | 是 | 变动金额（+为增加，-为减少） |
 
**请求示例：**
```json
{
  "server_id": "server1",
  "player_id": "player123",
  "amount": 100.50
}
```

**响应示例：**
 - 成功
```json
{
  "status": "success"
}
```
 - 参数错误
```json
{
  "err": "参数错误"
}
```
 2. 查询玩家跨服余额  
**接口描述：** 用于查询玩家跨服余额，从服务端插件向核心后端请求玩家在多个服务器上的余额信息。  
  
**请求URL：** `GET /api/spigot/currency/player`  
  
**请求参数：**  

| 参数名 | 类型 | 必填 | 说明 |
|-------|------|------|------|
|server_id	|string | 是 | 服务器标识 |
|player_id	|string	| 是 | 游戏内玩家ID |
  
**处理逻辑：** 接收服务器ID、玩家ID→查询多服余额→返回结果

### WEB对接API文档  
1. 登录接口  
  
**接口描述：** 简化鉴权入口，用于用户登录获取会话标识和账户角色信息。

**请求URL：** `POST /web/login`
  
**请求参数：**  
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| username | string | 是 | 用户名 |
| password | string | 是 | 密码 |
  
**请求示例：**
```json
{
    "username": "admin",
    "password": "123456"
}
```