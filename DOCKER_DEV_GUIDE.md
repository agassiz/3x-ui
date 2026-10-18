# 3x-ui Docker 开发环境指南

本指南介绍如何使用优化的 Docker 开发环境进行 3x-ui 开发，支持**分层缓存**和**快速重建**，大幅提升开发效率。

## 🚀 快速开始

### 一键启动开发环境

```bash
./docker-dev.sh
```

访问地址：http://localhost:54321

需要 Docker 和支持 `docker compose` 命令的 Compose 插件。脚本始终使用项目目录作为构建上下文，可以从其他目录调用。

每次启动都会先构建镜像，构建成功后才更新容器。构建失败不会停止原容器；容器启动失败或等待超时会返回非零退出状态，不会提示启动成功。

### Clash 订阅

在面板设置 → 订阅中开启订阅服务及 Clash。客户端列表中的 Clash 按钮使用设置返回的 `subClashURI` 和客户端 `subId`，请求经过标准订阅入口的 HWID 校验。

开发 Compose 默认映射订阅端口 `2096:2096`。已有数据库配置其他订阅端口时，请同步调整映射。通过反向代理或其他域名、端口访问订阅时，将 Clash 订阅地址设置为可从客户端访问的完整前缀，例如 `https://sub.example.com/clash/`；通过本机开发端口访问可使用 `http://localhost:2096/clash/`。

原面板路径 `/clash/subscription/<email>` 已移除，旧链接需重新导入。若反向代理或 CDN 缓存过旧路径，请清理相关缓存。启用 HWID 限制时，客户端必须支持发送 `X-HWID`。

## 🔥 前端开发

前端是 `frontend/` 下的 React + Vite 工程，构建产物会嵌入 Go 二进制，因此容器内没有文件挂载式的热重载：

- 🔄 **在容器里看效果** - 修改 `frontend/src` 后运行 `./docker-dev.sh`，镜像的前端阶段会重新构建
- ⚡ **实时热更新** - 本地运行 `go run main.go`（面板监听 2053），再在 `frontend/` 下运行 `npm run dev`，访问 http://localhost:5173 ；Vite 把接口代理到 `localhost:2053`，所以连不到映射在 54321 的容器

### Go 源代码 - 快速重建

- ⚡ **分层缓存** - 复用未变化的依赖层，实际耗时取决于缓存及网络
- 🔄 **自动构建** - 每次启动都包含最新修改
- 📦 **增量编译** - 只重新编译变更部分

## ⚡ 性能优化

### 智能缓存策略

1. **Go 依赖缓存**: `go.mod`/`go.sum` 不变时复用
2. **外部资源层**: 当前位于源码复制之后，源码变化也会重新下载 Xray 和 geo 数据文件
3. **目标架构一致**: Compose 不覆盖架构，由 BuildKit 的 `TARGETARCH` 选择对应的内核文件
4. **npm 依赖缓存**: `package.json`/`package-lock.json` 不变时复用

## 🛠️ 可用脚本

### docker-build-fast.sh

快速构建脚本，支持分层缓存

该脚本调用 `docker compose -f docker-compose.dev.yml build`，与启动脚本共用构建配置和镜像。默认使用 Docker 的目标平台，不写死 amd64 或 arm64。

```bash
./docker-build-fast.sh              # 标准构建
./docker-build-fast.sh --no-cache   # 不使用缓存
./docker-build-fast.sh --verbose    # 详细输出
```

### docker-dev.sh

开发环境管理脚本

```bash
./docker-dev.sh                     # 启动开发环境（自动构建最新代码）
./docker-dev.sh --rebuild     # 强制重新构建（不使用缓存）
```

## 📁 文件说明

### 生产环境文件（保持原样）

- `Dockerfile`: 原始生产环境配置
- `docker-compose.yml`: 原始生产环境配置

### 开发环境文件

- `Dockerfile.dev`: 开发专用 Dockerfile，优化了缓存层级
- `docker-compose.dev.yml`: 开发环境配置（数据目录挂载 + 端口映射）
- `docker-build-fast.sh`: 快速构建脚本
- `docker-dev.sh`: 开发环境管理脚本
- `.dockerignore`: 优化构建上下文，排除不必要文件

## 🔧 开发工作流

### 🚀 **首次设置**

```bash
./docker-dev.sh  # 一键启动，自动构建
```

### 🔥 **日常开发 - 前端**

```bash
# 本地启动面板 + Vite 开发服务器（热更新）
go run main.go
cd frontend && npm run dev
# 或者修改后重建容器
./docker-dev.sh
```

### ⚡ **Go 代码开发 - 快速重建**

```bash
# 修改Go源代码
vim main.go
# 重新构建并更新容器
./docker-dev.sh
```

### 🔄 **依赖变更**

```bash
# 修改go.mod后强制重建
./docker-dev.sh --rebuild
```

## 🎯 性能优化要点

### Dockerfile 层级优化

```dockerfile
# 0. 前端（package-lock.json 变化时才重新安装依赖）
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

# 1. 系统依赖（很少变化）
RUN apk add build-base gcc curl unzip

# 2. Go依赖（go.mod变化时才重建）
COPY go.mod go.sum ./
RUN go mod download

# 3. 源代码与前端构建产物
COPY . .
COPY --from=frontend /src/internal/web/dist ./internal/web/dist

# 4. 外部资源与应用程序，使用相同目标架构
ARG TARGETARCH
RUN ./DockerInit.sh "$TARGETARCH"
RUN go build ...
```

## 🐛 故障排除

### 构建失败

```bash
# 清理缓存重新构建
./docker-build-fast.sh --no-cache
```

### 容器启动失败

```bash
# 查看日志
docker logs 3xui_app_dev

# 重启容器
docker restart 3xui_app_dev
```

### 权限问题

```bash
# 确保脚本有执行权限
chmod +x docker-build-fast.sh docker-dev.sh
```

## 📊 效果验证

### 快速重建测试

```bash
# 修改Go代码
touch main.go

# 测试重建速度
time ./docker-dev.sh
```

观察构建输出中的缓存命中情况和脚本报告的实际耗时。首次构建、缓存失效或外部资源下载可能需要更长时间。

## 🎯 开发体验提升

### 前端开发

- ✅ **即时反馈** - 本地 Vite 开发服务器热更新
- ✅ **整体验证** - `./docker-dev.sh` 重建后在容器中检查

### 后端开发

- ✅ **失败可见** - 构建或启动失败返回非零状态
- ✅ **智能缓存** - 只重新编译变更部分
- ✅ **一键启动** - 自动构建最新代码

现在您可以享受现代化的开发体验！🚀
