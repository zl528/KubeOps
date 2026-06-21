#!/bin/bash
# Helm 部署脚本

set -e

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
HELM_DIR="$ROOT_DIR/deploy/helm/kubeops"

GREEN='\033[0;32m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }

RELEASE="${RELEASE:-kubeops}"
NAMESPACE="${NAMESPACE:-default}"

log_info "Installing Helm chart..."
helm install "$RELEASE" "$HELM_DIR" \
    --namespace "$NAMESPACE" \
    --create-namespace \
    --wait

log_info "Helm deployment complete!"
log_info "Check status: kubectl get pods -l app=$RELEASE -n $NAMESPACE"
