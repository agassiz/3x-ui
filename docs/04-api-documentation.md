# 3x-ui API 接口文档

## 📋 API 概览

3x-ui 提供了完整的 RESTful API 接口，支持所有核心功能的程序化操作。API 基于 Gin Web 框架构建，采用 JSON 格式进行数据交换。

### 基础信息
- **基础路径**: `{webBasePath}` (默认: `/`)
- **认证方式**: Session + Cookie
- **数据格式**: JSON
- **字符编码**: UTF-8

## 🔐 认证机制

### 1. 登录认证

**接口**: `POST {basePath}/login`

**请求体**:
```json
{
    "username": "admin",
    "password": "admin",
    "twoFactorCode": "123456"  // 可选，启用2FA时必填
}
```

**响应**:
```json
{
    "success": true,
    "msg": "登录成功",
    "obj": null
}
```

### 2. 会话验证

所有需要认证的接口都会检查会话状态：
- 通过 `checkLogin` 中间件验证
- 未登录时返回 `401 Unauthorized`
- AJAX 请求返回 JSON 错误信息
- 普通请求重定向到登录页

## 🏠 主要 API 分组

### 1. 面板管理 API (`/panel`)

#### 页面路由
```
GET  /panel/           # 主面板页面
GET  /panel/inbounds   # 入站管理页面
GET  /panel/settings   # 系统设置页面
GET  /panel/xray       # Xray 设置页面
```

### 2. 入站管理 API (`/panel/inbound`)

#### 入站配置管理

**获取入站列表**
```
POST /panel/inbound/list
```

**添加入站配置**
```
POST /panel/inbound/add
Content-Type: application/json

{
    "remark": "测试入站",
    "protocol": "vmess",
    "port": 10086,
    "settings": "{\"clients\":[...]}",
    "streamSettings": "{\"network\":\"tcp\"}",
    "enable": true
}
```

**更新入站配置**
```
POST /panel/inbound/update/:id
```

**删除入站配置**
```
POST /panel/inbound/del/:id
```

#### 客户端管理

**添加客户端**
```
POST /panel/inbound/addClient
Content-Type: application/json

{
    "id": 1,
    "settings": "{\"clients\":[{\"id\":\"uuid\",\"email\":\"test@example.com\"}]}"
}
```

**删除客户端**
```
POST /panel/inbound/:id/delClient/:clientId
```

**更新客户端**
```
POST /panel/inbound/updateClient/:clientId
```

#### 流量管理

**重置客户端流量**
```
POST /panel/inbound/:id/resetClientTraffic/:email
```

**重置所有流量**
```
POST /panel/inbound/resetAllTraffics
```

**获取客户端流量**
```
GET /panel/inbound/getClientTraffics/:email
```

### 3. 系统管理 API (`/panel/api/inbounds`)

#### 备份和恢复

**创建备份**
```
GET /panel/api/inbounds/createbackup
```

**获取数据库**
```
GET /server/getDb
```

**导入数据库**
```
POST /server/importDB
Content-Type: multipart/form-data
```

### 4. 服务器管理 API (`/server`)

#### 系统状态

**获取服务器状态**
```
POST /server/status

Response:
{
    "success": true,
    "obj": {
        "cpu": 15.5,
        "mem": {
            "current": 1024000000,
            "total": 8192000000
        },
        "disk": {
            "current": 50000000000,
            "total": 100000000000
        },
        "xray": {
            "state": "running",
            "version": "1.8.0"
        },
        "uptime": 86400,
        "loads": [0.5, 0.6, 0.7],
        "tcpCount": 100,
        "udpCount": 50,
        "netIO": {
            "up": 1000000,
            "down": 2000000
        },
        "netTraffic": {
            "sent": 10000000000,
            "recv": 20000000000
        }
    }
}
```

#### Xray 管理

**获取 Xray 版本**
```
POST /server/getXrayVersion
```

**重启 Xray 服务**
```
POST /server/restartXrayService
```

**停止 Xray 服务**
```
POST /server/stopXrayService
```

**安装 Xray**
```
POST /server/installXray/:version
```

**获取配置 JSON**
```
POST /server/getConfigJson
```

#### 日志管理

**获取日志**
```
POST /server/logs/:count
```

### 5. 设置管理 API (`/panel/setting`)

#### 系统设置

**获取所有设置**
```
POST /panel/setting/all

Response:
{
    "success": true,
    "obj": {
        "webPort": "2053",
        "webBasePath": "/",
        "secret": "randomstring",
        "tgBotEnable": "false",
        "tgBotToken": "",
        "pageSize": "50"
    }
}
```

**更新设置**
```
POST /panel/setting/update
Content-Type: application/json

{
    "webPort": "2054",
    "tgBotEnable": "true",
    "tgBotToken": "your_bot_token"
}
```

**更新用户信息**
```
POST /panel/setting/updateUser
Content-Type: application/json

{
    "username": "newadmin",
    "password": "newpassword"
}
```

### 6. Clash 订阅 API (`/clash`)

#### 订阅配置

**获取 Clash 订阅**
```
GET /clash/subscription/:email

Response: YAML 格式的 Clash 配置文件
```

## 📊 响应格式

### 标准响应结构

```json
{
    "success": boolean,     // 操作是否成功
    "msg": "string",       // 消息内容
    "obj": any             // 返回数据对象
}
```

### 成功响应示例

```json
{
    "success": true,
    "msg": "",
    "obj": {
        "id": 1,
        "remark": "测试入站",
        "protocol": "vmess",
        "port": 10086
    }
}
```

### 错误响应示例

```json
{
    "success": false,
    "msg": "端口已被占用",
    "obj": null
}
```

## 🔒 安全机制

### 1. 认证中间件

```go
func (a *BaseController) checkLogin(c *gin.Context) {
    if !session.IsLogin(c) {
        if isAjax(c) {
            pureJsonMsg(c, http.StatusUnauthorized, false, "请重新登录")
        } else {
            c.Redirect(http.StatusTemporaryRedirect, basePath)
        }
        c.Abort()
    }
}
```

### 2. 域名验证

```go
func DomainValidatorMiddleware(domain string) gin.HandlerFunc {
    return func(c *gin.Context) {
        host := getHostFromRequest(c.Request)
        if host != domain {
            c.AbortWithStatus(http.StatusForbidden)
            return
        }
        c.Next()
    }
}
```

### 3. 输入验证

- 参数类型验证
- 长度限制检查
- 特殊字符过滤
- SQL 注入防护

## 🌐 国际化支持

API 支持多语言错误消息：

```go
func I18nWeb(c *gin.Context, name string, params ...string) string {
    i18nFunc := c.Get("I18n").(func(locale.I18nType, string, ...string) string)
    return i18nFunc(locale.Web, name, params...)
}
```

支持语言：
- 英语 (en)
- 中文 (zh)
- 俄语 (ru)
- 西班牙语 (es)
- 波斯语 (fa)
- 阿拉伯语 (ar)

## 📈 流量统计 API

### 实时流量数据

通过 Xray gRPC API 获取实时流量统计：

```go
type Traffic struct {
    IsInbound bool   `json:"isInbound"`
    Tag       string `json:"tag"`
    Up        int64  `json:"up"`
    Down      int64  `json:"down"`
}

type ClientTraffic struct {
    Email     string `json:"email"`
    Up        int64  `json:"up"`
    Down      int64  `json:"down"`
    Total     int64  `json:"total"`
    Enable    bool   `json:"enable"`
}
```

## 🔄 错误处理

### HTTP 状态码

- `200 OK`: 请求成功
- `400 Bad Request`: 请求参数错误
- `401 Unauthorized`: 未授权访问
- `403 Forbidden`: 禁止访问
- `404 Not Found`: 资源不存在
- `500 Internal Server Error`: 服务器内部错误

### 错误消息格式

```json
{
    "success": false,
    "msg": "具体错误描述",
    "obj": null
}
```

## 📝 使用示例

### cURL 示例

```bash
# 登录
curl -X POST "http://localhost:2053/login" \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"admin"}' \
     -c cookies.txt

# 获取入站列表
curl -X POST "http://localhost:2053/panel/inbound/list" \
     -H "Content-Type: application/json" \
     -b cookies.txt

# 获取服务器状态
curl -X POST "http://localhost:2053/server/status" \
     -H "Content-Type: application/json" \
     -b cookies.txt
```

### JavaScript 示例

```javascript
// 登录
const login = async () => {
    const response = await fetch('/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            username: 'admin',
            password: 'admin'
        })
    });
    return response.json();
};

// 获取入站列表
const getInbounds = async () => {
    const response = await fetch('/panel/inbound/list', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        }
    });
    return response.json();
};
```

---

*下一步: 查看 [前端架构](./05-frontend-architecture.md) 了解前端技术实现*
