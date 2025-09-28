#!/bin/bash

# 3x-ui Docker开发环境启动脚本
echo "🚀 启动 3x-ui Docker 开发环境..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
  echo "❌ Docker 未运行，请先启动 Docker"
  exit 1
fi

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
      echo "默认行为: 每次都会快速构建最新代码（约5秒）"
      echo "注意: Docker构建过程中会自动清理容器内的build目录"
      exit 0
      ;;
    *)
      echo "未知选项: $1"
      echo "使用 --help 查看可用选项"
      exit 1
      ;;
  esac
done

# 创建必要的目录
echo "📁 确保数据目录存在..."
mkdir -p db cert

# 停止现有容器
if docker ps -q -f name=3xui_app_dev | grep -q .; then
  echo "🛑 停止现有开发容器..."
  docker-compose -f docker-compose.dev.yml down
fi

# 每次都快速构建最新代码
if [ "$FORCE_REBUILD" = true ]; then
  echo "🔨 强制重新构建镜像（不使用缓存）..."
  ./docker-build-fast.sh --no-cache
else
  echo "⚡ 快速构建最新代码（约5秒）..."
  ./docker-build-fast.sh
fi

# 启动开发容器
echo "🏗️ 启动开发容器..."
docker-compose -f docker-compose.dev.yml up -d

echo ""
echo "🌐 3x-ui 开发环境已启动"
echo "📍 访问地址: http://localhost:54321"
echo ""
echo "💡 查看日志: docker logs -f 3xui_app_dev"
echo "💡 清理日志: ./docker-clear-logs.sh"
echo "💡 进入容器: docker exec -it 3xui_app_dev bash"
echo "💡 重启服务: docker restart 3xui_app_dev"
echo "💡 停止环境: docker-compose -f docker-compose.dev.yml down"
echo ""
echo "🔄 每次运行都会自动构建最新代码"
echo "💥 完全重建: $0 --rebuild"
echo "🧹 Docker构建会自动清理容器内build目录"