# 3x-ui 系统架构

## 🏗️ 整体架构

3x-ui 采用分层架构设计，确保系统的可维护性、可扩展性和高性能。

```mermaid
graph TB
    subgraph "用户层"
        A[Web 浏览器]
        B[Telegram Bot]
        C[API 客户端]
    end
    
    subgraph "接入层"
        D[Gin Web Server]
        E[中间件层]
        F[路由层]
    end
    
    subgraph "控制层"
        G[IndexController]
        H[XUIController]
        I[APIController]
        J[ServerController]
        K[ClashController]
    end
    
    subgraph "服务层"
        L[UserService]
        M[InboundService]
        N[XrayService]
        O[SettingService]
        P[TgbotService]
    end
    
    subgraph "数据层"
        Q[GORM ORM]
        R[SQLite 数据库]
    end
    
    subgraph "核心层"
        S[Xray-core]
        T[配置管理]
        U[流量统计]
    end
    
    A --> D
    B --> D
    C --> D
    D --> E
    E --> F
    F --> G
    F --> H
    F --> I
    F --> J
    F --> K
    G --> L
    H --> M
    I --> N
    J --> O
    K --> P
    L --> Q
    M --> Q
    N --> Q
    O --> Q
    P --> Q
    Q --> R
    N --> S
    N --> T
    N --> U
```

## 📦 核心模块

### 1. Web 服务器层 (web/)

#### 主要组件
- **web.go**: Web 服务器核心，负责启动和配置
- **路由管理**: 基于 Gin 的路由系统
- **中间件**: 认证、日志、CORS 等中间件
- **静态资源**: 嵌入式文件系统

#### 关键特性
```go
type Server struct {
    httpServer *http.Server
    listener   net.Listener
    
    // 控制器
    index  *controller.IndexController
    server *controller.ServerController
    panel  *controller.XUIController
    api    *controller.APIController
    clash  *controller.ClashController
    
    // 服务
    xrayService    service.XrayService
    settingService service.SettingService
    tgbotService   service.Tgbot
    
    // 任务调度
    cron *cron.Cron
}
```

### 2. 控制器层 (web/controller/)

#### 控制器职责分工

```mermaid
graph LR
    A[IndexController] --> A1[登录认证]
    A[IndexController] --> A2[首页展示]
    
    B[XUIController] --> B1[面板管理]
    B[XUIController] --> B2[入站配置]
    B[XUIController] --> B3[系统设置]
    
    C[APIController] --> C1[REST API]
    C[APIController] --> C2[数据操作]
    
    D[ServerController] --> D1[服务器状态]
    D[ServerController] --> D2[系统监控]
    
    E[ClashController] --> E1[订阅生成]
    E[ClashController] --> E2[配置下载]
```

### 3. 服务层 (web/service/)

#### 核心服务架构

```mermaid
classDiagram
    class XrayService {
        +RestartXray()
        +GetXrayConfig()
        +AddInbound()
        +RemoveInbound()
        +GetTraffic()
    }
    
    class InboundService {
        +GetInbounds()
        +AddInbound()
        +UpdateInbound()
        +DeleteInbound()
        +AddTraffic()
    }
    
    class UserService {
        +Login()
        +CheckAuth()
        +UpdatePassword()
        +VerifyTwoFactor()
    }
    
    class SettingService {
        +GetSettings()
        +UpdateSetting()
        +GetBasePath()
        +GetSecret()
    }
    
    class TgbotService {
        +SendMessage()
        +HandleCommand()
        +SendBackup()
    }
    
    XrayService --> InboundService
    InboundService --> UserService
    UserService --> SettingService
    TgbotService --> SettingService
```

### 4. 数据层 (database/)

#### 数据模型关系

```mermaid
erDiagram
    User ||--o{ Inbound : manages
    Inbound ||--o{ ClientTraffic : tracks
    Inbound ||--o{ InboundClientIps : monitors
    User ||--o{ Setting : configures
    ClashSubscription ||--|| User : belongs_to
    
    User {
        int id PK
        string username
        string password
    }
    
    Inbound {
        int id PK
        int user_id FK
        string protocol
        int port
        string settings
        string stream_settings
        bool enable
        int64 expiry_time
    }
    
    ClientTraffic {
        int id PK
        int inbound_id FK
        string email
        int64 up
        int64 down
        int64 total
    }
    
    Setting {
        int id PK
        string key
        string value
    }
```

## 🔄 核心流程

### 1. 用户认证流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant C as IndexController
    participant S as UserService
    participant D as Database
    
    U->>C: 提交登录表单
    C->>S: 验证用户凭据
    S->>D: 查询用户信息
    D-->>S: 返回用户数据
    S->>S: 验证密码和2FA
    S-->>C: 返回认证结果
    C->>C: 创建会话
    C-->>U: 重定向到面板
```

### 2. Xray 配置更新流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant IC as InboundController
    participant IS as InboundService
    participant XS as XrayService
    participant X as Xray-core
    
    U->>IC: 添加/修改入站配置
    IC->>IS: 处理入站数据
    IS->>IS: 验证配置
    IS->>Database: 保存配置
    IS->>XS: 触发配置更新
    XS->>XS: 生成 Xray 配置
    XS->>X: 重启 Xray 服务
    X-->>XS: 确认重启成功
    XS-->>IS: 返回更新结果
    IS-->>IC: 返回操作结果
    IC-->>U: 显示操作结果
```

### 3. 流量统计流程

```mermaid
sequenceDiagram
    participant Cron as 定时任务
    participant XS as XrayService
    participant X as Xray-core
    participant IS as InboundService
    participant D as Database
    
    Cron->>XS: 每10秒触发统计
    XS->>X: 获取流量数据
    X-->>XS: 返回流量统计
    XS->>IS: 处理流量数据
    IS->>D: 更新流量记录
    IS->>IS: 检查流量限制
    IS->>XS: 触发必要的重启
```

## 🔧 技术架构特点

### 1. 分层设计
- **表现层**: Web UI + REST API
- **业务层**: 服务组件 + 业务逻辑
- **数据层**: ORM + 数据库
- **集成层**: Xray-core + 外部服务

### 2. 依赖注入
- 控制器依赖服务层
- 服务层依赖数据层
- 松耦合设计便于测试

### 3. 中间件架构
- 认证中间件
- 日志中间件
- 错误处理中间件
- 域名验证中间件

### 4. 任务调度
- Cron 定时任务
- Xray 状态检查
- 流量统计收集
- 自动重启机制

## 📊 性能优化

### 1. 数据库优化
- 连接池管理
- 索引优化
- 事务控制
- 批量操作

### 2. 内存管理
- 对象池复用
- 缓存策略
- 垃圾回收优化

### 3. 网络优化
- HTTP/2 支持
- Gzip 压缩
- 静态资源缓存
- 连接复用

## 🔒 安全架构

### 1. 认证安全
- Session 管理
- 双因子认证
- 密码加密存储
- 登录失败限制

### 2. 数据安全
- 敏感数据加密
- SQL 注入防护
- XSS 防护
- CSRF 保护

### 3. 网络安全
- HTTPS 强制
- 域名验证
- IP 白名单
- 访问控制

---

*下一步: 查看 [数据库设计](./03-database-design.md) 了解详细的数据模型*
