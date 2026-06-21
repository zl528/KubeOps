#!/bin/bash
# 项目初始化脚本

set -e

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查依赖
check_deps() {
    log_info "Checking dependencies..."

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.22+"
        exit 1
    fi

    if ! command -v node &> /dev/null; then
        log_error "Node.js is not installed. Please install Node.js 20+"
        exit 1
    fi

    if ! command -v npm &> /dev/null; then
        log_error "npm is not installed."
        exit 1
    fi

    log_info "All dependencies found!"
}

# 安装 Go 依赖
setup_backend() {
    log_info "Setting up backend..."
    cd "$ROOT_DIR/backend"
    export GOPROXY=https://goproxy.cn,direct
    go mod tidy
    log_info "Backend dependencies installed."
}

# 安装前端依赖
setup_frontend() {
    log_info "Setting up frontend..."
    cd "$ROOT_DIR/frontend"
    npm install --registry=https://registry.npmmirror.com
    log_info "Frontend dependencies installed."
}

# 设置脚本权限
setup_scripts() {
    log_info "Setting up scripts..."
    chmod +x "$ROOT_DIR/scripts/"*.sh
    log_info "Scripts are now executable."
}

# 主流程
main() {
    log_info "Initializing KubeOps project..."
    echo ""

    check_deps
    setup_backend
    setup_frontend
    setup_scripts

    echo ""
    log_info "Project initialized successfully!"
    echo ""
    log_info "Next steps:"
    log_info "  1. Start development: ./scripts/dev.sh"
    log_info "  2. Run tests:         ./scripts/test.sh"
    log_info "  3. Build:             ./scripts/build.sh"
    log_info "  4. Docker build:      ./scripts/docker-build.sh"
    log_info "  5. Deploy to K8s:     ./scripts/deploy-k8s.sh"
}

main
