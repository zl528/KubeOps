# AGENTS.md - KubeOps Development Guide

## Project Overview

Kubernetes operations management platform with Go backend and Vue 3 frontend.

## Repository Structure

```
backend/          # Go API server (client-go)
  cmd/server/     # Entry point
  internal/       # Core logic (handler/service/model/middleware)
  pkg/            # Shared utilities (client, logger)
frontend/         # Vue 3 + TypeScript + Element Plus
  src/views/      # Page components
  src/router/     # Routes
  src/api/        # Axios client
deploy/           # K8s manifests, Helm chart, Dockerfiles
scripts/          # Dev/test/build scripts
```

## Quick Commands

```bash
# Setup (installs Go deps + npm deps)
make setup

# Run dev environment (backend :8080, frontend :3000)
make dev

# Run all tests
make test

# Build for production
make build

# Build Docker images
IMAGE_PREFIX=kubeops TAG=latest make docker-build
```

## Backend (Go)

**Run tests:**
```bash
cd backend
export GOPROXY=https://goproxy.cn,direct
go test ./... -v
```

**Run single test:**
```bash
go test ./internal/service -run TestAlertService_ListRules -v
```

**Build:**
```bash
go build ./cmd/server
```

**Key dependencies:** k8s.io/client-go v0.36.1, k8s.io/api v0.36.1

**Environment variables:**
- `IN_CLUSTER=false` - Run outside K8s (uses kubeconfig)
- `KUBECONFIG` - Path to kubeconfig (defaults to ~/.kube/config)
- `PORT=8080` - Server port
- `MODE=development|release`

**Adding new resources:**
1. Create service in `internal/service/<name>.go`
2. Create handler in `internal/handler/<name>.go`
3. Register routes in `internal/api/server.go`
4. Add tests in `internal/service/<name>_test.go`

## Frontend (Vue 3)

**Run dev server:**
```bash
cd frontend
npm install --registry=https://registry.npmmirror.com
npm run dev
```

**Type check:**
```bash
npx vue-tsc --noEmit
```

**Build:**
```bash
npm run build
```

**Adding new pages:**
1. Create view in `src/views/<Name>.vue`
2. Add route in `src/router/index.ts`
3. Add menu item in `src/components/Layout.vue`

**Proxy config:** Frontend dev server proxies `/api/*` to `http://localhost:8080`

## Tests

- Backend tests use `k8s.io/client-go/kubernetes/fake` for mocking
- Run `make test` to execute both backend tests and frontend type check
- Tests are in `*_test.go` files alongside source

## Server Startup (IMPORTANT)

**Always use screen to start the server so it persists across bash sessions:**

```bash
# Start server in persistent screen session
screen -dmS kubeops bash -c "cd /root/ops-kubernetes && exec ./bin/kubeops-server > /tmp/kubeops.log 2>&1"

# Check if server is running
curl -s http://localhost:8080/api/health

# List screen sessions
screen -ls

# Attach to screen session (for debugging)
screen -r kubeops
```

**After rebuilding the server, restart it:**
```bash
pkill -f kubeops-server 2>/dev/null
screen -S kubeops -X quit 2>/dev/null
sleep 1
screen -dmS kubeops bash -c "cd /root/ops-kubernetes && exec ./bin/kubeops-server > /tmp/kubeops.log 2>&1"
sleep 2
curl -s http://localhost:8080/api/health
```

**Do NOT use `./bin/kubeops-server &` directly** - the process will be killed when the bash session ends.

## Conventions

- Backend follows Go standard project layout
- Frontend uses Composition API with `<script setup>`
- API responses use `model.Response{Code, Message, Data}` format
- All API routes are prefixed with `/api/`
- Chinese comments and UI text throughout
