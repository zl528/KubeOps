#!/bin/bash
# 测试脚本

set -e

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

log_info "Running backend tests..."
cd "$BACKEND_DIR"
export PATH=$PATH:/usr/local/go/bin
export GOPROXY=https://goproxy.cn,direct
if go test ./... -v -count=1; then
    log_info "Backend tests passed!"
else
    log_error "Backend tests failed!"
    exit 1
fi

log_info "Running frontend type check..."
cd "$FRONTEND_DIR"
if npx vue-tsc --noEmit; then
    log_info "Frontend type check passed!"
else
    log_error "Frontend type check failed!"
    exit 1
fi

log_info "All tests passed!"
