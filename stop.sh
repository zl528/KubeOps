#!/bin/bash
# 停止脚本

echo "正在停止服务..."

# 停止后端
pkill -f kubeops-server 2>/dev/null && echo "后端已停止" || echo "后端未运行"

# 停止前端
pkill -f "vite.*3000" 2>/dev/null && echo "前端已停止" || echo "前端未运行"

# 清理 screen 会话
screen -S backend -X quit 2>/dev/null
screen -S frontend -X quit 2>/dev/null

sleep 1

# 验证
echo ""
echo "=== 服务状态 ==="
if ss -tunlp | grep -qE "8080|3000"; then
  ss -tunlp | grep -E "8080|3000"
else
  echo "所有服务已停止"
fi
