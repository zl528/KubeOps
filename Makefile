.PHONY: help setup dev test build docker-build deploy deploy-helm clean

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## 初始化项目环境
	./scripts/setup.sh

dev: ## 启动开发环境
	./scripts/dev.sh

test: ## 运行测试
	./scripts/test.sh

build: ## 构建项目
	./scripts/build.sh

docker-build: ## 构建 Docker 镜像
	./scripts/docker-build.sh

deploy: ## 部署到 K8s
	./scripts/deploy-k8s.sh

deploy-helm: ## 使用 Helm 部署
	./scripts/deploy-helm.sh

clean: ## 清理构建产物
	rm -rf bin/ frontend/dist/
	@echo "Cleaned."
