#!/bin/bash
# 开发环境启动脚本

set -e

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

cleanup() {
    log_info "Stopping services..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
    wait $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
    log_info "All services stopped."
}

trap cleanup EXIT INT TERM

# 启动后端
log_info "Starting backend on port 8080..."
cd "$BACKEND_DIR"
export PATH=$PATH:/usr/local/go/bin
export GOPROXY=https://goproxy.cn,direct
export IN_CLUSTER=false
export MODE=development
go run ./cmd/server &
BACKEND_PID=$!

# 等待后端启动
sleep 2

# 启动前端
log_info "Starting frontend on port 3000..."
cd "$FRONTEND_DIR"
npm run dev &
FRONTEND_PID=$!

log_info "Development environment started!"
log_info "  Backend:  http://localhost:8080"
log_info "  Frontend: http://localhost:3000"
log_info "  API Docs: http://localhost:8080/api/health"
log_info ""
log_info "Press Ctrl+C to stop all services."

wait
