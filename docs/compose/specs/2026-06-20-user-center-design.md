# 用户中心设计方案

## [S1] 概述

为KubeOps平台添加完整的用户中心功能，包括用户管理、角色权限管理、个人设置和操作审计。

## [S2] 功能模块

### 用户管理（仅管理员）
- 用户列表展示
- 创建/编辑/删除用户
- 分配角色
- 重置密码
- 启用/禁用用户

### 角色管理（仅管理员）
- 预设角色：管理员、开发者、只读用户
- 自定义角色创建
- 权限配置（模块级 + 资源级）

### 个人设置（所有用户）
- 修改个人信息（显示名、邮箱）
- 修改密码

### 操作审计（管理员可看全部，普通用户看自己）
- 关键操作记录
- 审计日志查询

## [S3] 预设角色权限

| 角色 | 工作负载 | 网络 | 存储 | RBAC | 用户中心 |
|------|---------|------|------|------|---------|
| 管理员 | 读写 | 读写 | 读写 | 读写 | 完全访问 |
| 开发者 | 读写 | 只读 | 只读 | 无 | 个人设置 |
| 只读用户 | 只读 | 只读 | 只读 | 无 | 个人设置 |

## [S4] 权限粒度

- **模块级**：工作负载、网络、存储、RBAC、用户中心
- **资源级**：每个模块下的具体资源类型（Pod、Deployment、Service等）
- **操作级**：查看、创建、编辑、删除

## [S5] 数据模型

### 用户表 (users)
```sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  email TEXT,
  display_name TEXT,
  role_id INTEGER,
  status INTEGER DEFAULT 1,  -- 1:启用 0:禁用
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (role_id) REFERENCES roles(id)
);
```

### 角色表 (roles)
```sql
CREATE TABLE roles (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT UNIQUE NOT NULL,
  description TEXT,
  is_preset INTEGER DEFAULT 0,  -- 1:预设 0:自定义
  permissions TEXT,  -- JSON格式权限配置
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 权限JSON结构
```json
{
  "modules": {
    "workloads": { "view": true, "create": true, "edit": true, "delete": true },
    "network": { "view": true, "create": false, "edit": false, "delete": false },
    "storage": { "view": true, "create": false, "edit": false, "delete": false },
    "rbac": { "view": false, "create": false, "edit": false, "delete": false },
    "usercenter": { "view": true, "create": false, "edit": false, "delete": false }
  },
  "resources": {
    "pods": { "view": true, "create": false, "edit": false, "delete": false },
    "deployments": { "view": true, "create": true, "edit": true, "delete": true },
    "services": { "view": true, "create": false, "edit": false, "delete": false }
  }
}
```

### 审计日志表 (audit_logs)
```sql
CREATE TABLE audit_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER,
  username TEXT,
  action TEXT,  -- create/update/delete/login/logout
  resource_type TEXT,
  resource_name TEXT,
  namespace TEXT,
  details TEXT,
  ip TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## [S6] API设计

### 用户管理
- `GET /api/users` - 用户列表（管理员）
- `GET /api/users/get?username=xxx` - 获取用户详情
- `POST /api/users/create` - 创建用户（管理员）
- `PUT /api/users/update` - 更新用户（管理员）
- `DELETE /api/users/delete?username=xxx` - 删除用户（管理员）
- `PUT /api/users/password` - 修改密码（自己或管理员）
- `PUT /api/users/status` - 启用/禁用用户（管理员）

### 角色管理
- `GET /api/roles/list` - 角色列表
- `GET /api/roles/get?id=xxx` - 获取角色详情
- `POST /api/roles/create` - 创建角色（管理员）
- `PUT /api/roles/update` - 更新角色（管理员）
- `DELETE /api/roles/delete?id=xxx` - 删除角色（管理员，预设角色不可删）

### 审计日志
- `GET /api/audit/logs` - 审计日志列表（管理员看全部，普通用户看自己）

### 认证
- `POST /api/auth/login` - 登录
- `POST /api/auth/logout` - 登出
- `GET /api/auth/me` - 获取当前用户信息

## [S7] 页面设计

### 侧边栏菜单结构
```
管理员视图：
├── 集群概览
├── 工作负载
│   ├── Deployments
│   ├── StatefulSets
│   └── ...
├── 网络
├── 存储
├── RBAC
└── 用户中心
    ├── 用户管理
    ├── 角色管理
    └── 操作审计

普通用户视图：
├── 集群概览
├── 工作负载
├── 网络
├── 存储
└── 个人设置
    ├── 修改信息
    └── 修改密码
```

### 用户管理页面
- 表格列：用户名、显示名、邮箱、角色、状态、创建时间、操作
- 操作按钮：编辑、重置密码、启用/禁用、删除
- 创建用户对话框：用户名、密码、显示名、邮箱、角色选择

### 角色管理页面
- 表格列：角色名、描述、是否预设、权限概览、操作
- 操作按钮：编辑权限、复制角色、删除（预设角色不可删）
- 权限配置页面：模块级开关 + 资源级开关 + 操作级开关

### 操作审计页面
- 表格列：时间、用户、操作、资源类型、资源名称、命名空间、详情
- 筛选：时间范围、用户、操作类型、资源类型

## [S8] 设计风格

与平台其他页面保持一致的渐变现代风格：
- 深色主题
- 渐变卡片
- 毛玻璃效果
- Element Plus组件库

## [S9] 实现阶段

### 阶段1：后端基础
1. 数据库表创建
2. 用户CRUD API
3. 角色CRUD API
4. 审计日志API
5. 认证中间件增强

### 阶段2：前端页面
1. 用户管理页面
2. 角色管理页面
3. 权限配置页面
4. 个人设置页面
5. 操作审计页面

### 阶段3：权限集成
1. 路由守卫
2. 菜单权限控制
3. 按钮权限控制
4. API权限验证

### 阶段4：审计集成
1. 关键操作记录
2. 审计日志展示
3. 日志导出功能
