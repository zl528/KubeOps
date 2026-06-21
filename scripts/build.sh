#!/bin/bash
# 构建脚本

set -e

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"
DEPLOY_DIR="$ROOT_DIR/deploy"

GREEN='\033[0;32m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }

log_info "Building backend..."
cd "$BACKEND_DIR"
export PATH=$PATH:/usr/local/go/bin
export GOPROXY=https://goproxy.cn,direct
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$ROOT_DIR/bin/kubeops-server" ./cmd/server
log_info "Backend built: bin/kubeops-server"

log_info "Building frontend..."
cd "$FRONTEND_DIR"
npm run build
log_info "Frontend built: frontend/dist/"

log_info "Build complete!"
