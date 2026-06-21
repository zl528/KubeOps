#!/bin/bash
# K8s 部署脚本

set -e

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DEPLOY_DIR="$ROOT_DIR/deploy/k8s"

GREEN='\033[0;32m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }

NAMESPACE="${NAMESPACE:-default}"

log_info "Deploying RBAC..."
kubectl apply -f "$DEPLOY_DIR/rbac.yaml"

log_info "Deploying backend..."
kubectl apply -f "$DEPLOY_DIR/backend.yaml"

log_info "Deploying frontend..."
kubectl apply -f "$DEPLOY_DIR/frontend.yaml"

log_info "Deployment complete!"
log_info "Check status: kubectl get pods -l app=kubeops -n $NAMESPACE"
