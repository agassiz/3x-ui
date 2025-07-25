# 3x-ui 项目概览

## 📋 项目简介

**3x-ui** 是一个基于 Web 的高级开源控制面板，专为管理 Xray-core 服务器而设计。它提供了用户友好的界面，用于配置和监控各种 VPN 和代理协议。

作为原始 X-UI 项目的增强版本，3x-ui 提供了更好的稳定性、更广泛的协议支持和额外的功能。

> **重要提示**: 本项目仅用于个人使用和通信，请勿将其用于非法目的，请勿在生产环境中使用。

## 🎯 核心功能

### 1. 协议支持
- **VMESS** - V2Ray 原生协议
- **VLESS** - 轻量级协议
- **Trojan** - 伪装 HTTPS 流量
- **Shadowsocks** - 经典代理协议
- **WireGuard** - 现代 VPN 协议
- **HTTP/SOCKS** - 传统代理协议
- **Dokodemo-door** - 透明代理

### 2. 管理功能
- 🔐 **用户认证系统** - 支持双因子认证(2FA)
- 📊 **流量统计** - 实时监控和历史数据分析
- 👥 **多用户管理** - 客户端配置和权限控制
- 📱 **Telegram Bot** - 自动化通知和远程管理
- 🔄 **订阅系统** - Clash 配置自动生成
- 🌐 **多语言支持** - 国际化界面

### 3. 系统监控
- 📈 **服务器状态** - CPU、内存、网络监控
- 📋 **日志管理** - 系统和 Xray 日志查看
- 🔧 **配置管理** - 动态配置更新
- 🚀 **性能优化** - 自动重启和故障恢复

## 🛠️ 技术栈

### 后端技术
```
语言: Go 1.24.4
框架: Gin Web Framework v1.10.1
数据库: SQLite + GORM v1.30.0
核心: Xray-core v1.250608.0
```

### 主要依赖
- **Web 框架**: Gin + 中间件生态
- **数据库**: GORM ORM + SQLite 驱动
- **加密**: Go 标准加密库 + TOTP
- **网络**: gRPC + HTTP/2 支持
- **任务调度**: Cron v3 定时任务
- **Telegram**: Telego Bot API
- **国际化**: go-i18n 多语言支持

### 前端技术
- **模板引擎**: Go HTML Template
- **静态资源**: 嵌入式文件系统
- **样式**: CSS + 响应式设计
- **交互**: JavaScript + AJAX
- **图表**: 数据可视化组件

## 🏗️ 项目结构

```
3x-ui/
├── main.go                 # 应用入口点
├── config/                 # 配置管理
├── database/              # 数据库层
│   ├── db.go             # 数据库初始化
│   └── model/            # 数据模型
├── web/                   # Web 层
│   ├── web.go            # Web 服务器
│   ├── controller/       # 控制器
│   ├── service/          # 业务服务
│   ├── middleware/       # 中间件
│   ├── assets/           # 静态资源
│   └── html/             # HTML 模板
├── xray/                  # Xray 集成
├── sub/                   # 订阅系统
├── util/                  # 工具库
└── docs/                  # 技术文档
```

## 🚀 核心特性

### 1. 高性能架构
- **异步处理**: 基于 Goroutine 的并发模型
- **内存优化**: 高效的内存管理和缓存策略
- **网络优化**: HTTP/2 和 gRPC 支持
- **资源嵌入**: 静态资源编译时嵌入

### 2. 安全性设计
- **认证机制**: Session + 双因子认证
- **权限控制**: 基于角色的访问控制
- **数据加密**: 敏感数据加密存储
- **安全传输**: HTTPS 和 TLS 支持

### 3. 可扩展性
- **模块化设计**: 松耦合的组件架构
- **插件系统**: 支持功能扩展
- **配置驱动**: 灵活的配置管理
- **API 友好**: RESTful API 设计

### 4. 运维友好
- **Docker 支持**: 容器化部署
- **健康检查**: 自动故障检测和恢复
- **日志系统**: 结构化日志记录
- **监控集成**: 性能指标收集

## 📊 系统要求

### 最低要求
- **操作系统**: Linux/Windows/macOS
- **内存**: 512MB RAM
- **存储**: 100MB 可用空间
- **网络**: 稳定的网络连接

### 推荐配置
- **操作系统**: Ubuntu 20.04+ / CentOS 8+
- **内存**: 1GB+ RAM
- **存储**: 1GB+ 可用空间
- **CPU**: 1 核心以上

## 🔄 版本信息

- **当前版本**: v2.6.4
- **Go 版本**: 1.24.4
- **Xray 版本**: v1.250608.0
- **更新频率**: 持续集成和部署

## 📚 相关链接

- [系统架构](./02-architecture.md) - 详细的架构设计
- [API 文档](./04-api-documentation.md) - 接口规范
- [部署指南](./08-deployment.md) - 部署和配置
- [开发指南](./09-development-guide.md) - 开发环境搭建

---

*下一步: 查看 [系统架构](./02-architecture.md) 了解详细的技术架构设计*
