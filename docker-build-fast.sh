#!/bin/bash

# 快速构建脚本 - 专为开发环境优化
echo "⚡ 3x-ui 快速构建脚本"

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
  echo "❌ Docker 未运行，请先启动 Docker"
  exit 1
fi

# 解析参数
CACHE_FROM=""
BUILD_ARGS=""
VERBOSE=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --no-cache)
      BUILD_ARGS="$BUILD_ARGS --no-cache"
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
      echo "此脚本使用优化的Dockerfile.dev进行快速构建："
      echo "1. 缓存Go依赖下载"
      echo "2. 缓存外部资源下载"
      echo "3. 只在代码变更时重新编译"
      echo "4. 自动清理容器内build目录"
      exit 0
      ;;
    *)
      echo "未知选项: $1"
      echo "使用 --help 查看可用选项"
      exit 1
      ;;
  esac
done

# 显示构建信息
echo "📋 构建配置:"
echo "   - Dockerfile: Dockerfile.dev"
echo "   - 缓存策略: 分层缓存（Go依赖 + 外部资源 + 代码）"
echo "   - 目标架构: amd64"

if [ "$VERBOSE" = true ]; then
  BUILD_ARGS="$BUILD_ARGS --progress=plain"
  echo "   - 详细模式: 启用"
fi

echo ""

# 开始构建
echo "🔨 开始构建..."
start_time=$(date +%s)

if docker build -f Dockerfile.dev -t 3x-ui:dev $BUILD_ARGS .; then
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
  echo "❌ 构建失败！"
  echo "💡 尝试清理缓存重新构建:"
  echo "   $0 --no-cache"
  exit 1
fi
