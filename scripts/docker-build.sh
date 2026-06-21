#!/bin/bash
# Docker 构建脚本

set -e

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DEPLOY_DIR="$ROOT_DIR/deploy/docker"

GREEN='\033[0;32m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }

IMAGE_PREFIX="${IMAGE_PREFIX:-kubeops}"
TAG="${TAG:-latest}"

log_info "Building backend image: ${IMAGE_PREFIX}-backend:${TAG}"
docker build -t "${IMAGE_PREFIX}-backend:${TAG}" -f "$DEPLOY_DIR/Dockerfile.backend" "$ROOT_DIR/backend"

log_info "Building frontend image: ${IMAGE_PREFIX}-frontend:${TAG}"
docker build -t "${IMAGE_PREFIX}-frontend:${TAG}" -f "$DEPLOY_DIR/Dockerfile.frontend" "$ROOT_DIR/frontend"

log_info "Docker images built successfully!"
log_info "  ${IMAGE_PREFIX}-backend:${TAG}"
log_info "  ${IMAGE_PREFIX}-frontend:${TAG}"
