#!/bin/bash

# Docker日志清理脚本
echo "🧹 Docker日志清理工具"

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
  echo "❌ Docker 未运行，请先启动 Docker"
  exit 1
fi

# 解析参数
CONTAINER_NAME="3xui_app_dev"
CLEAR_ALL=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --all)
      CLEAR_ALL=true
      shift
      ;;
    --container|-c)
      CONTAINER_NAME="$2"
      shift 2
      ;;
    --help|-h)
      echo "用法: $0 [选项]"
      echo "选项:"
      echo "  --all              清理所有容器的日志"
      echo "  --container, -c    指定容器名称（默认: 3xui_app_dev）"
      echo "  --help, -h         显示此帮助信息"
      echo ""
      echo "示例:"
      echo "  $0                 # 清理3xui_app_dev容器日志"
      echo "  $0 --all           # 清理所有容器日志"
      echo "  $0 -c my_container # 清理指定容器日志"
      exit 0
      ;;
    *)
      echo "未知选项: $1"
      echo "使用 --help 查看可用选项"
      exit 1
      ;;
  esac
done

# 清理指定容器日志
clear_container_logs() {
  local container=$1
  
  if ! docker ps -a --format "{{.Names}}" | grep -q "^${container}$"; then
    echo "❌ 容器 '${container}' 不存在"
    return 1
  fi
  
  echo "🧹 清理容器 '${container}' 的日志..."
  
  # 方法1: 尝试直接清理日志文件（需要root权限）
  local log_path=$(docker inspect --format='{{.LogPath}}' "$container" 2>/dev/null)
  if [ -n "$log_path" ] && [ -f "$log_path" ]; then
    if echo '' > "$log_path" 2>/dev/null; then
      echo "✅ 成功清理容器 '${container}' 的日志"
      return 0
    fi
  fi
  
  # 方法2: 重启容器来清理日志
  echo "🔄 尝试通过重启容器来清理日志..."
  if docker restart "$container" > /dev/null 2>&1; then
    echo "✅ 容器 '${container}' 已重启，日志已清理"
  else
    echo "❌ 无法重启容器 '${container}'"
    return 1
  fi
}

# 清理所有容器日志
clear_all_logs() {
  echo "🧹 清理所有容器的日志..."
  
  local containers=$(docker ps -a --format "{{.Names}}")
  if [ -z "$containers" ]; then
    echo "📭 没有找到任何容器"
    return 0
  fi
  
  local success_count=0
  local total_count=0
  
  while IFS= read -r container; do
    if [ -n "$container" ]; then
      total_count=$((total_count + 1))
      if clear_container_logs "$container"; then
        success_count=$((success_count + 1))
      fi
    fi
  done <<< "$containers"
  
  echo ""
  echo "📊 清理完成: ${success_count}/${total_count} 个容器日志已清理"
}

# 显示当前日志大小
show_log_sizes() {
  echo "📊 当前容器日志大小:"
  
  if [ "$CLEAR_ALL" = true ]; then
    docker ps -a --format "table {{.Names}}\t{{.Status}}" | head -1
    docker ps -a --format "{{.Names}}" | while read container; do
      if [ -n "$container" ]; then
        local log_path=$(docker inspect --format='{{.LogPath}}' "$container" 2>/dev/null)
        local size="未知"
        if [ -f "$log_path" ]; then
          size=$(du -h "$log_path" 2>/dev/null | cut -f1 || echo "未知")
        fi
        printf "%-20s\t%s\n" "$container" "$size"
      fi
    done
  else
    local log_path=$(docker inspect --format='{{.LogPath}}' "$CONTAINER_NAME" 2>/dev/null)
    if [ -f "$log_path" ]; then
      local size=$(du -h "$log_path" 2>/dev/null | cut -f1 || echo "未知")
      echo "容器 '${CONTAINER_NAME}': ${size}"
    else
      echo "容器 '${CONTAINER_NAME}': 日志文件不存在或无法访问"
    fi
  fi
  echo ""
}

# 主逻辑
echo ""
show_log_sizes

if [ "$CLEAR_ALL" = true ]; then
  clear_all_logs
else
  clear_container_logs "$CONTAINER_NAME"
fi

echo ""
echo "🎯 提示:"
echo "   - 如果清理失败，可能需要sudo权限"
echo "   - 也可以通过重启Docker服务来清理所有日志"
echo "   - 使用 'docker system prune' 可以清理更多Docker数据"
