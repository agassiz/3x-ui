#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
DEV_COMPOSE_FILE="$SCRIPT_DIR/docker-compose.dev.yml"
cd "$SCRIPT_DIR"

# 解析命令行参数
FORCE_REBUILD=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --rebuild)
      FORCE_REBUILD=true
      shift
      ;;
    --help|-h)
      echo "用法: $0 [选项]"
      echo "选项:"
      echo "  --rebuild  强制重新构建镜像（不使用缓存）"
      echo "  --help, -h       显示此帮助信息"
      echo ""
      echo "先构建最新代码，成功后再更新容器；构建失败时保留原容器。"
      exit 0
      ;;
    *)
      echo "未知选项: $1"
      echo "使用 --help 查看可用选项"
      exit 1
      ;;
  esac
done

echo "🚀 启动 3x-ui Docker 开发环境..."

# 每次都快速构建最新代码
BUILD_COMMAND=("$SCRIPT_DIR/docker-build-fast.sh")
if [ "$FORCE_REBUILD" = true ]; then
  echo "🔨 强制重新构建镜像（不使用缓存）..."
  BUILD_COMMAND+=(--no-cache)
else
  echo "⚡ 构建最新代码..."
fi
if ! "${BUILD_COMMAND[@]}"; then
  echo "❌ 构建失败，未更新开发容器。" >&2
  exit 1
fi

echo "📁 确保数据目录存在..."
mkdir -p db cert

# 启动开发容器
echo "🏗️ 启动开发容器..."
if ! docker compose -f "$DEV_COMPOSE_FILE" up -d --no-build --wait --wait-timeout 60 3xui; then
  echo "❌ 开发容器启动失败，请执行 docker logs 3xui_app_dev 查看日志。" >&2
  exit 1
fi

echo ""
echo "🌐 3x-ui 开发环境已启动"
echo "📍 访问地址: http://localhost:54321"
echo ""
echo "💡 查看日志: docker logs -f 3xui_app_dev"
echo "💡 清理日志: ./docker-clear-logs.sh"
echo "💡 进入容器: docker exec -it 3xui_app_dev bash"
echo "💡 重启服务: docker restart 3xui_app_dev"
echo "💡 停止环境: docker compose -f \"$DEV_COMPOSE_FILE\" down"
echo ""
echo "🔄 每次运行都会自动构建最新代码"
echo "💥 完全重建: $0 --rebuild"
