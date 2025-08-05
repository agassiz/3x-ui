# Clash 订阅 API 接口文档

## 概述

本接口提供了基于客户端email生成Clash订阅配置的功能。接口会自动缓存生成的配置，并在客户端配置发生变化时自动更新。

## 接口详情

### 获取Clash订阅配置

**接口地址：** `GET {webBasePath}/clash/subscription/{email}`

**请求方式：** GET

**是否需要登录：** 否（公开接口）

**参数说明：**
- `email`: 客户端的唯一标识邮箱地址（路径参数）
- `{webBasePath}`: 面板配置的URL根路径（如 `/admin/`、`/panel/` 等，默认为 `/`）

**请求示例：**
```bash
# 默认根路径 (webBasePath = "/")
curl -X GET "http://your-domain.com/clash/subscription/icuwuugf"

# 自定义根路径 (webBasePath = "/admin/")
curl -X GET "http://your-domain.com/admin/clash/subscription/icuwuugf"

# 自定义根路径 (webBasePath = "/panel/")
curl -X GET "http://your-domain.com/panel/clash/subscription/icuwuugf"
```

**响应说明：**
- 成功时返回YAML格式的Clash配置文件
- 失败时返回JSON格式的错误信息

**响应头：**
- `Content-Type: application/x-yaml; charset=utf-8`
- `Content-Disposition: attachment; filename=clash-config.yaml`
- `Cache-Control: public, max-age=300`

**成功响应示例：**
```yaml
# Clash配置内容
proxies:
  - name: "proxy-name"
    type: vmess
    server: 154.83.95.153
    port: 23121
    uuid: b50565c9-88ec-4709-9dc9-798773ba608b
    # ... 其他配置
```

**错误响应示例：**
```json
{
  "success": false,
  "message": "Client not found for the specified email"
}
```

**使用说明：**
- 前端会自动生成完整的订阅URL
- 用户直接复制订阅地址到Clash客户端即可
- 系统会自动缓存配置，提高响应速度

## 工作原理

1. **订阅地址生成：** 前端生成完整的订阅URL（如：`https://yourdomain.com/clash/subscription/email`）
2. **服务器地址获取：** 后端自动从请求头中获取真实的服务器地址（X-Forwarded-Host或Host），并正确分离端口号
3. **客户端查找：** 根据提供的email在数据库中查找对应的客户端配置
4. **安全链接生成：** 生成使用固定端口45556的代理链接作为缓存键（隐藏真实端口）
5. **缓存检查：** 计算缓存键的MD5签名，检查数据库中是否有缓存
6. **配置生成：** 如果缓存不存在或配置发生变化，调用外部API生成Clash配置
7. **端口替换：** 将生成的配置中的所有45556直接替换为真实端口
8. **缓存更新：** 将最终配置保存到数据库中

## 安全特性

- **端口隐藏：** 调用外部API时使用固定端口45556，避免暴露真实端口信息
- **简单替换：** 生成配置后直接将所有45556替换为真实端口，确保配置正确
- **缓存保护：** 缓存的配置已经是替换后的真实配置，提高安全性

## 缓存机制

- 系统会自动缓存生成的Clash配置
- 当客户端配置（如端口、密码、传输方式等）发生变化时，会自动重新生成配置
- 缓存基于客户端链接的MD5签名进行判断

## 支持的协议

- **VMESS：** 支持TCP、WebSocket、gRPC等传输方式
- **VLESS：** 支持各种传输方式和流控
- **Trojan：** 支持TLS加密
- **Shadowsocks：** 支持各种加密方法，包括2022版协议

## 数据库表结构

系统会自动创建`clash_subscriptions`表来存储缓存数据：

```sql
CREATE TABLE clash_subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR UNIQUE NOT NULL,
    url_md5 VARCHAR NOT NULL,
    yaml_content TEXT,
    created_at INTEGER,
    updated_at INTEGER
);
```

## 错误码说明

- **400 Bad Request：** 缺少email参数
- **404 Not Found：** 指定email的客户端不存在
- **500 Internal Server Error：** 服务器内部错误或外部API调用失败

## 使用注意事项

1. 确保客户端的email在系统中存在且已启用
2. 客户端必须属于支持的协议类型
3. 外部API服务需要可访问（https://sub.datapipe.top）
4. 建议定期清理过期的缓存数据
5. 服务器地址会自动从请求头中获取，支持反向代理环境

## 安全考虑

- 接口不需要登录验证，但建议在生产环境中添加适当的访问控制
- email参数会被清理和验证
- 生成的配置文件包含敏感信息，请妥善保管

## 部署说明

1. 确保数据库已正确初始化
2. 重启3X-UI服务以加载新的API接口
3. 接口将在主服务端口上提供服务

## 示例用法

## API调用示例

```bash
# 获取客户端icuwuugf的Clash配置（默认根路径）
curl -X GET "http://localhost:2053/clash/subscription/icuwuugf" \
     -H "Accept: application/x-yaml" \
     -o clash-config.yaml

# 获取客户端icuwuugf的Clash配置（自定义根路径 /admin/）
curl -X GET "http://localhost:2053/admin/clash/subscription/icuwuugf" \
     -H "Accept: application/x-yaml" \
     -o clash-config.yaml

# 注意：删除缓存功能暂未实现，配置会自动根据客户端变更更新
```

## Web界面使用

### 1. 在客户端列表中使用Clash订阅功能

1. 进入入站规则管理页面
2. 展开任意入站规则的客户端列表
3. 在客户端操作菜单中点击"Clash订阅"选项
4. 在弹出的模态框中可以：
   - 扫描二维码获取订阅链接
   - 复制订阅链接到剪贴板
   - 下载Clash配置文件
   - 刷新配置（清除缓存重新生成）

### 2. 功能特性

- **智能缓存**：配置会自动缓存，只有在客户端配置变化时才重新生成
- **实时更新**：支持手动刷新配置，确保获取最新设置
- **多种获取方式**：支持二维码扫描、链接复制、文件下载
- **错误处理**：提供详细的错误信息和重试机制

### 3. 支持的客户端

- 支持所有具有代理链接的协议（VMESS、VLESS、Trojan、Shadowsocks）
- 自动检测客户端配置变化
- 支持各种传输方式和安全设置

## 🚀 使用效果

### 订阅链接示例
- 默认根路径：`https://yourdomain.com/clash/subscription/icuwuugf`
- 自定义根路径：`https://yourdomain.com/admin/clash/subscription/icuwuugf`
- IP访问（默认）：`http://1.2.3.4:2053/clash/subscription/icuwuugf`
- IP访问（自定义）：`http://1.2.3.4:2053/panel/clash/subscription/icuwuugf`
- 反向代理：自动适配真实的访问地址和根路径

### 生成的代理链接
订阅配置中的代理链接会包含正确的服务器地址：
- VMESS：`vmess://xxx@yourdomain.com:443`
- VLESS：`vless://xxx@yourdomain.com:443`
- 支持反向代理：自动识别真实的服务器地址
