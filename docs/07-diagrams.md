# 3x-ui 流程图和时序图

## 📊 业务流程可视化

本文档提供了 3x-ui 系统中关键业务流程的详细时序图和流程图，帮助理解系统的工作机制和组件间的交互关系。

## 🔐 用户认证流程

### 1. 登录认证时序图

```mermaid
sequenceDiagram
    participant U as 用户浏览器
    participant IC as IndexController
    participant US as UserService
    participant SS as SettingService
    participant DB as 数据库
    participant S as Session
    
    U->>IC: POST /login (username, password, 2FA)
    IC->>US: CheckUser(username, password, 2FA)
    US->>DB: 查询用户信息
    DB-->>US: 返回用户数据
    US->>US: 验证密码哈希
    
    alt 启用双因子认证
        US->>SS: GetTwoFactorEnable()
        SS-->>US: true
        US->>SS: GetTwoFactorToken()
        SS-->>US: TOTP密钥
        US->>US: 验证TOTP代码
    end
    
    alt 认证成功
        US-->>IC: 返回用户对象
        IC->>S: 创建用户会话
        S-->>IC: 会话ID
        IC-->>U: 重定向到面板 + Set-Cookie
    else 认证失败
        US-->>IC: 返回null
        IC-->>U: 返回错误信息
    end
```

### 2. 会话验证流程

```mermaid
flowchart TD
    A[请求到达] --> B{检查Cookie}
    B -->|无Cookie| C[重定向登录]
    B -->|有Cookie| D[验证Session]
    D -->|Session无效| C
    D -->|Session有效| E{检查权限}
    E -->|权限不足| F[返回403]
    E -->|权限充足| G[继续处理请求]
    
    style A fill:#e1f5fe
    style G fill:#c8e6c9
    style C fill:#ffcdd2
    style F fill:#ffcdd2
```

## 🔧 Xray 配置管理流程

### 1. 配置生成和应用时序图

```mermaid
sequenceDiagram
    participant C as Controller
    participant XS as XrayService
    participant IS as InboundService
    participant SS as SettingService
    participant DB as 数据库
    participant XC as Xray-core
    participant XA as XrayAPI
    
    C->>XS: RestartXray(force=true)
    XS->>SS: GetXrayConfigTemplate()

    alt 数据库中有配置
        SS->>DB: 查询xrayTemplateConfig
        DB-->>SS: 返回数据库配置
    else 使用默认配置
        SS->>SS: 使用web/service/config.json
    end

    SS-->>XS: 配置模板

    XS->>IS: GetAllInbounds()
    IS->>DB: 查询所有入站配置
    DB-->>IS: 入站配置列表
    IS-->>XS: 入站配置

    XS->>XS: 合并配置和过滤客户端
    
    alt Xray正在运行
        XS->>XC: 停止当前进程
        XC-->>XS: 确认停止
    end
    
    XS->>XC: 启动新配置
    XC-->>XS: 启动成功

    XS->>XA: 初始化API连接
    XA-->>XS: API就绪

    XS->>XS: updateXrayTemplateConfig()
    XS->>SS: SaveXrayTemplateConfig()
    SS->>DB: 更新配置模板
    Note over XS,DB: 确保数据库配置与实际运行配置同步

    XS-->>C: 重启完成
```

### 2. 入站配置更新流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant IC as InboundController
    participant IS as InboundService
    participant DB as 数据库
    participant XA as XrayAPI
    participant XC as Xray-core
    
    U->>IC: 更新入站配置
    IC->>IS: UpdateInbound(inbound)
    IS->>DB: 开始事务
    IS->>DB: 保存配置到数据库
    
    alt 数据库保存成功
        IS->>XA: 连接Xray API
        IS->>XA: DelInbound(oldTag)
        XA->>XC: 删除旧入站
        XC-->>XA: 删除成功
        
        alt 入站启用
            IS->>XA: AddInbound(newConfig)
            XA->>XC: 添加新入站
            XC-->>XA: 添加成功
            XA-->>IS: API操作成功
        else API操作失败
            IS->>IS: 标记需要重启
        end
        
        IS->>DB: 提交事务
        IS-->>IC: 返回成功结果
    else 数据库保存失败
        IS->>DB: 回滚事务
        IS-->>IC: 返回错误
    end
    
    IC-->>U: 返回操作结果
```

## 📊 流量统计流程

### 1. 流量收集和处理时序图

```mermaid
sequenceDiagram
    participant Cron as 定时任务
    participant XTJ as XrayTrafficJob
    participant XS as XrayService
    participant XA as XrayAPI
    participant XC as Xray-core
    participant IS as InboundService
    participant OS as OutboundService
    participant DB as 数据库
    
    Cron->>XTJ: 每10秒触发
    XTJ->>XS: IsXrayRunning()
    XS-->>XTJ: true
    
    XTJ->>XS: GetXrayTraffic()
    XS->>XA: Init(apiPort)
    XS->>XA: GetTraffic(reset=true)
    XA->>XC: gRPC QueryStats
    XC-->>XA: 流量统计数据
    XA-->>XS: [Traffic], [ClientTraffic]
    XS->>XA: Close()
    XS-->>XTJ: 流量数据
    
    par 并行处理
        XTJ->>IS: AddTraffic(traffics, clientTraffics)
        IS->>DB: 开始事务
        IS->>DB: 更新入站流量
        IS->>DB: 更新客户端流量
        IS->>IS: 检查客户端过期
        IS->>DB: 提交事务
        IS-->>XTJ: needRestart1
    and
        XTJ->>OS: AddTraffic(traffics, clientTraffics)
        OS->>DB: 更新出站流量
        OS-->>XTJ: needRestart2
    end
    
    alt 需要重启
        XTJ->>XS: SetToNeedRestart()
    end
```

### 2. 客户端流量限制检查流程

```mermaid
flowchart TD
    A[收到流量数据] --> B[更新客户端流量]
    B --> C{检查流量限制}
    C -->|未超限| D[继续服务]
    C -->|超出限制| E[禁用客户端]
    E --> F[从Xray配置移除]
    F --> G[标记需要重启]
    
    C --> H{检查过期时间}
    H -->|未过期| D
    H -->|已过期| I[禁用客户端]
    I --> F
    
    style A fill:#e1f5fe
    style D fill:#c8e6c9
    style E fill:#ffcdd2
    style I fill:#ffcdd2
    style G fill:#fff3e0
```

## 🤖 Telegram Bot 交互流程

### 1. Bot 命令处理时序图

```mermaid
sequenceDiagram
    participant U as Telegram用户
    participant TG as Telegram服务器
    participant TB as TgbotService
    participant SS as SettingService
    participant XS as XrayService
    participant IS as InboundService
    
    U->>TG: 发送命令 /status
    TG->>TB: Webhook通知
    TB->>TB: 验证用户权限
    
    alt 用户已授权
        TB->>SS: 获取系统设置
        TB->>XS: 获取Xray状态
        TB->>IS: 获取流量统计
        TB->>TB: 格式化状态信息
        TB->>TG: 发送状态消息
        TG->>U: 显示系统状态
    else 用户未授权
        TB->>TG: 发送权限错误
        TG->>U: 显示错误信息
    end
```

### 2. 自动备份通知流程

```mermaid
flowchart TD
    A[定时备份任务] --> B[生成数据库备份]
    B --> C[压缩备份文件]
    C --> D{Telegram Bot启用?}
    D -->|是| E[发送备份到管理员]
    D -->|否| F[仅本地保存]
    E --> G[记录发送日志]
    F --> H[清理旧备份]
    G --> H
    
    style A fill:#e1f5fe
    style E fill:#c8e6c9
    style F fill:#fff3e0
    style H fill:#f3e5f5
```

## 🔄 订阅系统流程

### 1. Clash 订阅生成时序图

```mermaid
sequenceDiagram
    participant C as 客户端
    participant CC as ClashController
    participant CS as ClashService
    participant IS as InboundService
    participant DB as 数据库
    participant API as 外部API
    
    C->>CC: GET /clash/subscription/{email}
    CC->>CS: GetClashSubscription(email)
    CS->>IS: SearchClientTraffic(email)
    IS->>DB: 查询客户端配置
    DB-->>IS: 客户端信息
    IS-->>CS: 客户端配置
    
    CS->>CS: 检查缓存
    
    alt 缓存未命中或已过期
        CS->>CS: 构建订阅URL
        CS->>API: 请求Clash配置生成
        API-->>CS: YAML配置内容
        CS->>DB: 缓存配置内容
    else 缓存命中
        CS->>DB: 读取缓存配置
        DB-->>CS: YAML配置内容
    end
    
    CS-->>CC: YAML配置
    CC-->>C: 返回Clash配置文件
```

### 2. 订阅链接生成流程

```mermaid
flowchart TD
    A[用户请求订阅] --> B[验证客户端邮箱]
    B -->|邮箱无效| C[返回404错误]
    B -->|邮箱有效| D[查询客户端配置]
    D --> E[获取服务器信息]
    E --> F[构建代理配置]
    F --> G{选择订阅格式}
    G -->|Clash| H[生成Clash YAML]
    G -->|V2Ray| I[生成V2Ray JSON]
    G -->|通用| J[生成Base64链接]
    H --> K[返回配置文件]
    I --> K
    J --> K
    
    style A fill:#e1f5fe
    style K fill:#c8e6c9
    style C fill:#ffcdd2
```

## 🔍 系统监控流程

### 1. 健康检查时序图

```mermaid
sequenceDiagram
    participant Cron as 定时任务
    participant CXJ as CheckXrayJob
    participant XS as XrayService
    participant XC as Xray-core
    participant TB as TgbotService
    participant Log as 日志系统
    
    Cron->>CXJ: 每秒触发检查
    CXJ->>XS: IsXrayRunning()
    XS->>XC: 检查进程状态
    XC-->>XS: 进程状态
    XS-->>CXJ: false (未运行)
    
    CXJ->>Log: 记录Xray停止
    CXJ->>XS: RestartXray(force=true)
    
    alt 重启成功
        XS-->>CXJ: 重启成功
        CXJ->>Log: 记录重启成功
    else 重启失败
        XS-->>CXJ: 重启失败 + 错误信息
        CXJ->>Log: 记录重启失败
        CXJ->>TB: SendAlert("Xray重启失败")
        TB->>TB: 发送告警到管理员
    end
```

### 2. 性能监控数据收集流程

```mermaid
flowchart TD
    A[监控任务启动] --> B[收集系统指标]
    B --> C[CPU使用率]
    B --> D[内存使用情况]
    B --> E[磁盘空间]
    B --> F[网络流量]
    B --> G[Xray连接数]
    
    C --> H[数据聚合]
    D --> H
    E --> H
    F --> H
    G --> H
    
    H --> I{超出阈值?}
    I -->|是| J[触发告警]
    I -->|否| K[更新监控面板]
    J --> L[发送通知]
    K --> M[存储历史数据]
    L --> M
    
    style A fill:#e1f5fe
    style J fill:#ffcdd2
    style K fill:#c8e6c9
    style M fill:#f3e5f5
```

## 🔄 配置同步流程

### 1. 配置变更传播时序图

```mermaid
sequenceDiagram
    participant U as 用户
    participant SC as SettingController
    participant SS as SettingService
    participant DB as 数据库
    participant XS as XrayService
    participant WS as WebServer
    
    U->>SC: 更新系统设置
    SC->>SS: UpdateSetting(key, value)
    SS->>DB: 保存设置到数据库
    DB-->>SS: 保存成功
    SS-->>SC: 更新成功
    
    alt 影响Xray配置的设置
        SC->>XS: SetToNeedRestart()
        XS->>XS: 标记需要重启
    end
    
    alt 影响Web服务的设置
        SC->>WS: 通知配置变更
        WS->>WS: 更新运行时配置
    end
    
    SC-->>U: 返回更新结果
    
    Note over XS: 下次定时检查时会重启Xray
```

### 2. 数据库迁移流程

```mermaid
flowchart TD
    A[系统启动] --> B[检查数据库版本]
    B --> C{需要迁移?}
    C -->|否| D[正常启动]
    C -->|是| E[备份当前数据库]
    E --> F[执行迁移脚本]
    F --> G{迁移成功?}
    G -->|是| H[更新版本标记]
    G -->|否| I[恢复备份]
    H --> J[记录迁移日志]
    I --> K[启动失败]
    J --> D
    
    style A fill:#e1f5fe
    style D fill:#c8e6c9
    style K fill:#ffcdd2
    style I fill:#fff3e0
```

---

*下一步: 查看 [部署运维指南](./08-deployment.md) 了解系统部署和运维*
