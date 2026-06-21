#!/bin/bash
# 启动脚本

set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

# 停止旧进程
pkill -f kubeops-server 2>/dev/null || true
pkill -f "vite.*3000" 2>/dev/null || true
sleep 1

# 构建后端
echo "正在构建后端..."
cd "$ROOT_DIR/backend"
export GOPROXY=https://goproxy.cn,direct
go build -o "$ROOT_DIR/bin/kubeops-server" ./cmd/server

# 启动后端
echo "正在启动后端..."
cd "$ROOT_DIR"
export IN_CLUSTER=false
export MODE=development
nohup ./bin/kubeops-server > /tmp/backend.log 2>&1 &
echo "Backend PID: $!"

# 等待后端启动
sleep 3

# 启动前端
echo "正在启动前端..."
cd "$ROOT_DIR/frontend"
nohup node_modules/.bin/vite --host 0.0.0.0 --port 3000 > /tmp/frontend.log 2>&1 &
echo "Frontend PID: $!"

sleep 2

# 验证
echo ""
echo "=== 服务状态 ==="
ss -tunlp | grep -E "8080|3000"
echo ""
echo "=== 后端健康检查 ==="
curl -s http://localhost:8080/api/health
echo ""
echo ""
echo "=== 访问地址 ==="
echo "前端: http://localhost:3000"
echo "后端: http://localhost:8080"
echo "登录: admin / admin123"
