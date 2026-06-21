# KubeOps - Kubernetes 运维管理平台

一个轻量级的 Kubernetes 运维管理平台，提供资源管理、权限控制、监控告警、审计日志等能力。

## 技术栈

- **前端**: Vue 3 + TypeScript + Element Plus + ECharts
- **后端**: Go + client-go
- **数据库**: SQLite
- **部署**: 支持 K8s 集群内部署和独立部署

## 功能模块

### 1. 集群管理
- 多集群支持（连接、切换、断开）
- 集群概览（节点、Pod、资源使用率）
- 节点管理（查看状态、资源）
- 命名空间管理

### 2. 工作负载
- Deployment（详情、编辑、扩缩容、重启、回滚、删除）
- StatefulSet（详情、编辑、扩缩容、重启、删除）
- DaemonSet（详情、编辑、重启、删除）
- CronJob（编辑、暂停/恢复、删除）
- Job（详情、删除）

### 3. Pod 管理
- Pod 列表、详情、日志查看
- Pod 删除、重启
- Pod 终端（WebSocket）

### 4. 网络管理
- Service（详情、编辑、删除）
- Ingress（详情、删除）
- NetworkPolicy（列表、删除）
- Endpoint（列表）

### 5. 配置管理
- ConfigMap（详情、编辑、删除）
- Secret（详情、编辑、删除）

### 6. RBAC 管理
- Role / ClusterRole（编辑、删除）
- RoleBinding / ClusterRoleBinding（删除）
- ServiceAccount（删除）

### 7. 存储管理
- PersistentVolume（列表）
- PersistentVolumeClaim（详情、编辑、删除）
- StorageClass（列表、删除）

### 8. 策略管理
- HPA（列表、创建、删除）
- LimitRange（列表、创建、删除）

### 9. 用户中心
- 用户管理（创建、编辑、删除、启用/禁用）
- 角色管理（创建、编辑、删除）
- 权限控制（模块级 + 资源级）
- 集群授权（用户可访问的集群）
- 个人设置（修改信息、修改密码）

### 10. 监控告警
- 监控中心（CPU/内存/磁盘/网络指标）
- Prometheus 集成（自动检测、PromQL 查询）
- 告警规则管理（CRUD）
- 告警历史查看
- 后台告警检查

### 11. 审计日志
- 操作审计（创建、更新、删除操作记录）
- 审计日志查询

### 12. 资源关系图
- 可视化资源关联关系
- 支持拖拽、缩放
- 点击跳转到资源详情

### 13. 命名空间视图
- 命名空间资源脑图
- 支持展开/收起

## 权限体系

### 预设角色
| 角色 | 工作负载 | 网络 | 存储 | RBAC | 用户中心 |
|------|---------|------|------|------|---------|
| 管理员 | 读写 | 读写 | 读写 | 读写 | 完全访问 |
| 开发者 | 读写 | 只读 | 只读 | 无 | 个人设置 |
| 只读用户 | 只读 | 只读 | 只读 | 无 | 个人设置 |

### 集群授权
- 管理员自动拥有所有集群访问权限
- 其他用户需要管理员分配可访问的集群

## 快速开始

### 开发环境

```bash
# 初始化项目
make setup

# 启动开发环境
make dev

# 运行测试
make test
```

### 生产部署

```bash
# 使用 Helm 部署
make deploy-helm

# 或使用 K8s 清单部署
make deploy
```

## 项目结构

```
ops-kubernetes/
├── backend/                    # Go 后端
│   ├── cmd/server/            # 入口
│   ├── internal/
│   │   ├── api/               # 路由注册
│   │   ├── config/            # 配置
│   │   ├── database/          # 数据库
│   │   ├── handler/           # HTTP 处理器
│   │   ├── middleware/        # 中间件（认证、权限、审计）
│   │   ├── model/             # 数据模型
│   │   └── service/           # 业务逻辑
│   └── pkg/                   # 公共包
├── frontend/                   # Vue 3 前端
│   └── src/
│       ├── api/               # API 层
│       ├── components/        # 组件
│       ├── composables/       # 组合函数
│       ├── router/            # 路由
│       ├── store/             # 状态管理
│       └── views/             # 页面视图（34个）
├── deploy/                     # 部署配置
│   ├── docker/                # Dockerfile
│   ├── helm/                  # Helm Chart
│   └── k8s/                   # K8s 清单
├── scripts/                    # 工具脚本
└── Makefile
```

## 镜像源配置

- npm: `https://registry.npmmirror.com`
- Go: `https://goproxy.cn,direct`
