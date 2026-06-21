# KubeOps - Kubernetes 运维管理平台

一个功能完整的 Kubernetes 运维管理平台，提供集群管理、应用部署、CI/CD 流水线、监控告警、安全审计等能力。

## 技术栈

- **前端**: Vue 3 + TypeScript + Element Plus
- **后端**: Go + client-go
- **部署**: 支持 K8s 集群内部署和独立部署

## 功能模块

### 1. 集群管理
- 节点管理（查看、状态、版本、资源）
- 命名空间管理（列表、标签）
- 资源配额管理（CRUD）
- 集群概览（资源使用率、健康状态）

### 2. 工作负载
- Deployment（列表、扩缩容、重启、回滚、删除）
- StatefulSet（列表、扩缩容、重启、删除）
- DaemonSet（列表、重启、删除）
- CronJob（列表、暂停/恢复、删除）
- Job（列表、删除）

### 3. Pod 管理
- Pod 列表、详情、日志查看
- Pod 删除、重启
- Pod Exec 命令执行
- Pod Port Forward 端口转发
- 实时日志流

### 4. 网络管理
- Service（列表、类型、端口）
- Ingress（列表、Host、TLS）
- NetworkPolicy（列表、Ingress/Egress 规则）
- Endpoint（列表、地址、端口）

### 5. 配置管理
- ConfigMap（列表、详情、删除）
- Secret（列表、类型、数据键）
- LimitRange（列表、限制规则）
- ResourceQuota（CRUD）

### 6. RBAC 管理
- Role / ClusterRole（列表、规则、删除）
- RoleBinding / ClusterRoleBinding（列表、绑定对象、删除）
- ServiceAccount（列表、删除）

### 7. 存储管理
- PersistentVolume（列表、容量、状态、回收策略）
- PersistentVolumeClaim（列表、状态、关联 PV）
- StorageClass（列表、Provisioner、绑定模式）

### 8. 策略管理
- HPA（列表、副本范围、指标、当前/期望副本）
- LimitRange（列表、资源限制）

### 9. 事件管理
- 事件列表（类型、原因、对象、消息）
- 按命名空间过滤

## API 接口

### 集群
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/health | 健康检查 |
| GET | /api/cluster/overview | 集群概览 |
| GET | /api/nodes | 节点列表 |
| GET | /api/nodes/get?name=xxx | 节点详情 |
| GET | /api/namespaces | 命名空间列表 |

### Pod
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/pods?namespace=xxx | Pod 列表 |
| GET | /api/pods/get?namespace=xxx&name=xxx | Pod 详情 |
| GET | /api/pods/logs?namespace=xxx&pod=xxx | Pod 日志 |
| POST | /api/pods/exec | 执行命令 |
| POST | /api/pods/portforward | 端口转发 |
| POST | /api/pods/restart?namespace=xxx&name=xxx | 重启 Pod |
| DELETE | /api/pods/delete?namespace=xxx&name=xxx | 删除 Pod |

### Deployment
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/deployments?namespace=xxx | Deployment 列表 |
| POST | /api/deployments/scale | 扩缩容 |
| POST | /api/deployments/rollback | 回滚 |
| POST | /api/deployments/restart?namespace=xxx&name=xxx | 重启 |
| DELETE | /api/deployments/delete?namespace=xxx&name=xxx | 删除 |

### RBAC
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/roles?namespace=xxx | Role 列表 |
| GET | /api/clusterroles | ClusterRole 列表 |
| GET | /api/rolebindings?namespace=xxx | RoleBinding 列表 |
| GET | /api/clusterrolebindings | ClusterRoleBinding 列表 |
| GET | /api/serviceaccounts?namespace=xxx | ServiceAccount 列表 |

### 存储
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/persistentvolumes | PV 列表 |
| GET | /api/persistentvolumeclaims?namespace=xxx | PVC 列表 |
| GET | /api/storageclasses | StorageClass 列表 |

### 策略
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/limitranges?namespace=xxx | LimitRange 列表 |
| GET | /api/hpas?namespace=xxx | HPA 列表 |
| GET | /api/resourcequotas?namespace=xxx | ResourceQuota 列表 |
| POST | /api/resourcequotas/create | 创建 ResourceQuota |
| POST | /api/resourcequotas/update | 更新 ResourceQuota |

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
│   │   ├── handler/           # HTTP 处理器
│   │   ├── middleware/        # 中间件
│   │   ├── model/             # 数据模型
│   │   └── service/           # 业务逻辑
│   └── pkg/                   # 公共包
├── frontend/                   # Vue 3 前端
│   └── src/
│       ├── api/               # API 层
│       ├── components/        # 组件
│       ├── router/            # 路由
│       ├── store/             # 状态管理
│       └── views/             # 页面视图
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
