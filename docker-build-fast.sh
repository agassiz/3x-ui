#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
DEV_COMPOSE_FILE="$SCRIPT_DIR/docker-compose.dev.yml"
cd "$SCRIPT_DIR"

# 解析参数
NO_CACHE=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --no-cache)
      NO_CACHE=true
      shift
      ;;
    --verbose|-v)
      VERBOSE=true
      shift
      ;;
    --help|-h)
      echo "用法: $0 [选项]"
      echo "选项:"
      echo "  --no-cache       不使用构建缓存"
      echo "  --verbose, -v    显示详细构建信息"
      echo "  --help, -h       显示此帮助信息"
      echo ""
      echo "通过 docker-compose.dev.yml 构建，架构由 Docker 目标平台决定。"
      exit 0
      ;;
    *)
      echo "未知选项: $1"
      echo "使用 --help 查看可用选项"
      exit 1
      ;;
  esac
done

echo "⚡ 3x-ui 快速构建脚本"
if ! docker info > /dev/null 2>&1; then
  echo "❌ Docker 未运行，请先启动 Docker" >&2
  exit 1
fi
if ! docker compose version > /dev/null 2>&1; then
  echo "❌ 需要支持 docker compose 的 Docker Compose 插件" >&2
  exit 1
fi

# 显示构建信息
echo "📋 构建配置:"
echo "   - Compose: $DEV_COMPOSE_FILE"
echo "   - 目标架构: Docker 目标平台"

COMPOSE_ARGS=(-f "$DEV_COMPOSE_FILE")
BUILD_ARGS=(build)
if [ "$NO_CACHE" = true ]; then
  BUILD_ARGS+=(--no-cache)
fi
if [ "$VERBOSE" = true ]; then
  COMPOSE_ARGS+=(--progress plain)
  echo "   - 详细模式: 启用"
fi
BUILD_ARGS+=(3xui)

echo ""

# 开始构建
echo "🔨 开始构建..."
start_time=$(date +%s)

if docker compose "${COMPOSE_ARGS[@]}" "${BUILD_ARGS[@]}"; then
  end_time=$(date +%s)
  duration=$((end_time - start_time))
  
  echo ""
  echo "✅ 构建成功！"
  echo "⏱️  构建时间: ${duration}秒"
  echo ""
  echo "🚀 启动开发环境:"
  echo "   ./docker-dev.sh"
  echo ""
  echo "🔄 如需重新构建:"
  echo "   ./docker-dev.sh --rebuild"
else
  echo ""
  echo "❌ 构建失败！" >&2
  echo "💡 尝试清理缓存重新构建:"
  echo "   $0 --no-cache"
  exit 1
fi
