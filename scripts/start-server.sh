#!/bin/bash
# 持久化启动 KubeOps 服务器

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SERVER="$ROOT_DIR/bin/kubeops-server"
LOG="/tmp/kubeops.log"
SESSION="kubeops"

# 检查是否已运行
if curl -s --connect-timeout 2 http://localhost:8080/api/health > /dev/null 2>&1; then
    echo "Server already running on :8080"
    exit 0
fi

# 清理旧会话
screen -S "$SESSION" -X quit 2>/dev/null

# 在 screen 会话中启动
screen -dmS "$SESSION" bash -c "cd $ROOT_DIR && exec ./bin/kubeops-server > $LOG 2>&1"

# 等待启动
for i in $(seq 1 10); do
    sleep 1
    if curl -s --connect-timeout 2 http://localhost:8080/api/health > /dev/null 2>&1; then
        echo "Server started on :8080"
        exit 0
    fi
done

echo "Server failed to start. Check $LOG"
exit 1
