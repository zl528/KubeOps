<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '72px' : '260px'" class="aside" :class="{ collapsed: isCollapse }">
      <div class="logo" @click="$router.push('/dashboard')">
        <div class="logo-icon">
          <el-icon size="24" color="#fff"><Monitor /></el-icon>
        </div>
        <transition name="fade">
          <span v-if="!isCollapse" class="logo-text">KubeOps</span>
        </transition>
      </div>
      <el-scrollbar class="menu-scrollbar" :style="{ '--scrollbar-bg': 'transparent' }">
        <el-menu
          :default-active="currentRoute"
          :collapse="isCollapse"
          router
          background-color="transparent"
          text-color="#94a3b8"
          active-text-color="#818cf8"
          :collapse-transition="false"
          class="sidebar-menu"
        >
          <el-menu-item index="/dashboard" class="menu-item">
            <el-icon><Monitor /></el-icon>
            <template #title>集群概览 Dashboard</template>
          </el-menu-item>

          <el-menu-item index="/resource-graph" class="menu-item">
            <el-icon><Share /></el-icon>
            <template #title>资源关系图 Graph</template>
          </el-menu-item>

          <el-sub-menu index="workloads" class="menu-item" v-if="hasPermission('workloads')">
            <template #title>
              <el-icon><Box /></el-icon>
              <span>工作负载 Workloads</span>
            </template>
            <el-menu-item index="/deployments">Deploy 部署</el-menu-item>
            <el-menu-item index="/statefulsets">STS 有状态集</el-menu-item>
            <el-menu-item index="/daemonsets">DS 守护进程</el-menu-item>
            <el-menu-item index="/cronjobs">CronJob 定时任务</el-menu-item>
            <el-menu-item index="/jobs">Job 任务</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="network" class="menu-item" v-if="hasPermission('network')">
            <template #title>
              <el-icon><Connection /></el-icon>
              <span>网络 Network</span>
            </template>
            <el-menu-item index="/services">Service 服务</el-menu-item>
            <el-menu-item index="/ingresses">Ingress 入口</el-menu-item>
            <el-menu-item index="/networkpolicies">NetworkPolicy 网络策略</el-menu-item>
            <el-menu-item index="/endpoints">Endpoint 端点</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="config" class="menu-item" v-if="hasPermission('storage')">
            <template #title>
              <el-icon><Document /></el-icon>
              <span>配置 Config</span>
            </template>
            <el-menu-item index="/configmaps">CM 配置映射</el-menu-item>
            <el-menu-item index="/secrets">Secret 密钥</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="rbac" class="menu-item" v-if="hasPermission('rbac')">
            <template #title>
              <el-icon><User /></el-icon>
              <span>RBAC 权限</span>
            </template>
            <el-menu-item index="/roles">Role 角色</el-menu-item>
            <el-menu-item index="/clusterroles">ClusterRole 集群角色</el-menu-item>
            <el-menu-item index="/rolebindings">RoleBinding 角色绑定</el-menu-item>
            <el-menu-item index="/clusterrolebindings">ClusterRoleBinding 集群角色绑定</el-menu-item>
            <el-menu-item index="/serviceaccounts">SA 服务账号</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="storage" class="menu-item" v-if="hasPermission('storage')">
            <template #title>
              <el-icon><Box /></el-icon>
              <span>存储 Storage</span>
            </template>
            <el-menu-item index="/persistentvolumes">PV 持久卷</el-menu-item>
            <el-menu-item index="/persistentvolumeclaims">PVC 持久卷声明</el-menu-item>
            <el-menu-item index="/storageclasses">StorageClass 存储类</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="policy" class="menu-item" v-if="hasPermission('workloads')">
            <template #title>
              <el-icon><Timer /></el-icon>
              <span>策略 Policy</span>
            </template>
            <el-menu-item index="/limitranges">LimitRange 限制范围</el-menu-item>
            <el-menu-item index="/hpas">HPA 自动扩缩</el-menu-item>
          </el-sub-menu>

          <el-sub-menu index="cluster" class="menu-item" v-if="hasPermission('workloads')">
            <template #title>
              <el-icon><Cpu /></el-icon>
              <span>集群 Cluster</span>
            </template>
            <el-menu-item index="/nodes">Node 节点</el-menu-item>
            <el-menu-item index="/namespaces">Namespace 命名空间</el-menu-item>
            <el-menu-item index="/namespace-explorer">Namespace View 命名空间视图</el-menu-item>
            <el-menu-item index="/pods">Pod</el-menu-item>
            <el-menu-item index="/events">Event 事件</el-menu-item>
          </el-sub-menu>

          <div class="menu-divider"></div>

          <el-menu-item index="/monitoring" class="menu-item">
            <el-icon><DataAnalysis /></el-icon>
            <template #title>监控中心 Monitoring</template>
          </el-menu-item>

          <el-menu-item index="/alerts" class="menu-item">
            <el-icon><Warning /></el-icon>
            <template #title>告警管理 Alerts</template>
          </el-menu-item>

          <el-menu-item index="/logs" class="menu-item">
            <el-icon><Document /></el-icon>
            <template #title>日志管理 Logs</template>
          </el-menu-item>

          <el-menu-item index="/audit" class="menu-item" v-if="isAdmin">
            <el-icon><Notebook /></el-icon>
            <template #title>审计日志 Audit</template>
          </el-menu-item>

          <el-menu-item index="/backup" class="menu-item" v-if="isAdmin">
            <el-icon><Download /></el-icon>
            <template #title>备份恢复 Backup</template>
          </el-menu-item>

          <el-menu-item index="/terminal" class="menu-item">
            <el-icon><Monitor /></el-icon>
            <template #title>WebTerminal</template>
          </el-menu-item>

          <el-sub-menu index="user-center" class="menu-group" v-if="isAdmin">
            <template #title>
              <el-icon><User /></el-icon>
              <span>用户中心</span>
            </template>
            <el-menu-item index="/user-management">用户管理</el-menu-item>
            <el-menu-item index="/role-management">角色管理</el-menu-item>
          </el-sub-menu>

          <el-menu-item index="/profile" class="menu-item">
            <el-icon><Setting /></el-icon>
            <template #title>个人设置</template>
          </el-menu-item>
        </el-menu>
      </el-scrollbar>
      <div class="sidebar-footer">
        <div class="sidebar-footer-text" v-if="!isCollapse">
          <span>KubeOps v1.0</span>
        </div>
      </div>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="isCollapse = !isCollapse" size="20">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <el-breadcrumb separator="/" class="breadcrumb">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ currentTitle }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-dropdown trigger="click" @command="handleClusterSwitch" v-if="clusters.length > 0">
            <div class="cluster-selector">
              <el-icon><Connection /></el-icon>
              <span>{{ activeCluster || '选择集群' }}</span>
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="cluster in clusters"
                  :key="cluster.name"
                  :command="cluster.name"
                  :class="{ 'is-active': cluster.active }"
                >
                  <el-icon><Connection /></el-icon>
                  {{ cluster.name }}
                  <el-tag v-if="cluster.active" type="success" size="small" style="margin-left: 8px">当前</el-tag>
                </el-dropdown-item>
                <el-dropdown-item command="manage" divided>
                  <el-icon><Setting /></el-icon>管理集群
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <div class="cluster-info" v-else-if="clusterName">
            <el-icon><Connection /></el-icon>
            <span>{{ clusterName }}</span>
            <el-tag size="small" type="success">已连接</el-tag>
          </div>
          <el-dropdown trigger="click" @command="handleCommand">
            <div class="user-avatar">
              <el-avatar :size="32" class="avatar-gradient">
                <el-icon size="18"><User /></el-icon>
              </el-avatar>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="cluster" v-if="isAdmin">
                  <el-icon><Connection /></el-icon>集群管理
                </el-dropdown-item>
                <el-dropdown-item command="disconnect" v-if="isAdmin">
                  <el-icon><SwitchButton /></el-icon>断开连接
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Monitor, Cpu, Box, Connection, Fold, Expand, Document, User, DataAnalysis, Warning, Notebook, Download, Timer, Grid, Coin, Link, Share, Lock, Avatar, Suitcase, UserFilled, TrendCharts, SwitchButton, Setting, ArrowDown } from '@element-plus/icons-vue'
import api from '../api'

interface ClusterInfo {
  name: string
  server: string
  status: string
  active: boolean
}

const route = useRoute()
const router = useRouter()
const isCollapse = ref(false)
const clusterName = ref('')
const clusters = ref<ClusterInfo[]>([])
const activeCluster = ref('')

const currentRoute = computed(() => route.path)
const currentTitle = computed(() => (route.meta?.title as string) || '')

const permissions = ref<any>({})

const isAdmin = computed(() => {
  try {
    const token = localStorage.getItem('token')
    if (!token) return false
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.role === 'admin'
  } catch {
    return false
  }
})

const hasPermission = (module: string, action: string = 'view') => {
  if (isAdmin.value) return true
  const modules = permissions.value?.modules
  if (!modules) return false
  const mod = modules[module]
  if (!mod) return false
  return mod[action] === true
}

const fetchPermissions = async () => {
  if (isAdmin.value) return
  try {
    const res: any = await api.get('/auth/permissions')
    if (res.code === 0 && res.data) {
      permissions.value = res.data
    }
  } catch (e) {
    console.error(e)
  }
}

const fetchClusters = async () => {
  try {
    const res: any = await api.get('/clusters')
    if (res.success && res.clusters) {
      clusters.value = res.clusters
      const active = res.clusters.find((c: ClusterInfo) => c.active)
      if (active) {
        activeCluster.value = active.name
        clusterName.value = active.name
        localStorage.setItem('cluster_name', active.name)
      }
    }
  } catch (e) {
    console.error(e)
  }
}

onMounted(async () => {
  clusterName.value = localStorage.getItem('cluster_name') || ''
  await fetchPermissions()
  await fetchClusters()

  if (clusters.value.length === 0) {
    try {
      const res: any = await api.get('/cluster/status')
      if (res.connected) {
        clusterName.value = res.name || 'connected'
        localStorage.setItem('cluster_name', clusterName.value)
      }
    } catch (e) {
      console.error(e)
    }
  }
})

const handleClusterSwitch = async (name: string) => {
  if (name === 'manage') {
    router.push('/clusters')
    return
  }

  try {
    const res: any = await api.post('/clusters/switch', { name })
    if (res.success) {
      ElMessage.success(`已切换到集群: ${name}`)
      localStorage.setItem('cluster_name', name)
      activeCluster.value = name
      clusterName.value = name
      fetchClusters()
      window.location.reload()
    } else {
      ElMessage.error(res.message || '切换失败')
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '切换失败')
  }
}

const handleCommand = async (cmd: string) => {
  if (cmd === 'cluster') {
    router.push('/clusters')
  } else if (cmd === 'disconnect') {
    try {
      await ElMessageBox.confirm('确定要断开集群连接吗？', '确认', { type: 'warning' })
      await api.post('/cluster/disconnect')
      localStorage.removeItem('cluster_connected')
      localStorage.removeItem('cluster_name')
      clusters.value = []
      activeCluster.value = ''
      clusterName.value = ''
      ElMessage.success('已断开连接')
      router.push('/login')
    } catch (e) {
      if (e !== 'cancel') ElMessage.error('操作失败')
    }
  } else if (cmd === 'logout') {
    try {
      await ElMessageBox.confirm('确定要退出登录吗？', '确认', { type: 'warning' })
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      localStorage.removeItem('cluster_connected')
      localStorage.removeItem('cluster_name')
      ElMessage.success('已退出登录')
      router.push('/login')
    } catch (e) {
      if (e !== 'cancel') ElMessage.error('操作失败')
    }
  }
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  overflow: hidden;
}

.aside {
  background: linear-gradient(180deg, #0f172a 0%, #1a1f3a 100%);
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgba(99, 102, 241, 0.15);
  position: relative;
}

.aside::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 1px;
  height: 100%;
  background: linear-gradient(180deg, transparent, rgba(99, 102, 241, 0.3), transparent);
  pointer-events: none;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  cursor: pointer;
  border-bottom: 1px solid rgba(99, 102, 241, 0.1);
  flex-shrink: 0;
  padding: 0 16px;
  transition: all 0.3s ease;
}

.logo:hover {
  background: rgba(99, 102, 241, 0.05);
}

.logo-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-radius: 10px;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
  transition: all 0.3s ease;
}

.logo:hover .logo-icon {
  transform: scale(1.05);
  box-shadow: 0 6px 16px rgba(99, 102, 241, 0.4);
}

.logo-text {
  font-size: 20px;
  font-weight: 700;
  color: #fff;
  white-space: nowrap;
  background: linear-gradient(135deg, #6366f1, #a78bfa);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: 0.5px;
}

.menu-scrollbar {
  flex: 1;
  overflow: hidden;
  padding: 8px 0;
}

.menu-scrollbar :deep(.el-scrollbar__view) {
  padding: 0 8px;
}

.menu-scrollbar :deep(.el-scrollbar__bar.is-vertical) {
  width: 6px;
}

.menu-scrollbar :deep(.el-scrollbar__thumb) {
  background: rgba(99, 102, 241, 0.3);
  border-radius: 3px;
}

.menu-scrollbar :deep(.el-scrollbar__thumb:hover) {
  background: rgba(99, 102, 241, 0.5);
}

.sidebar-menu {
  border-right: none !important;
}

.sidebar-menu :deep(.el-menu) {
  background: transparent !important;
}

:deep(.el-menu-item),
:deep(.el-sub-menu__title) {
  height: 44px;
  line-height: 44px;
  margin: 2px 4px;
  border-radius: 8px;
  transition: all 0.2s ease;
  color: #94a3b8;
}

:deep(.el-menu-item:hover),
:deep(.el-sub-menu__title:hover) {
  background: rgba(99, 102, 241, 0.1) !important;
  color: #e2e8f0;
}

:deep(.el-menu-item.is-active) {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(139, 92, 246, 0.2)) !important;
  color: #a78bfa !important;
  position: relative;
}

:deep(.el-menu-item.is-active::before) {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: linear-gradient(180deg, #6366f1, #8b5cf6);
  border-radius: 0 3px 3px 0;
}

:deep(.el-sub-menu .el-menu-item) {
  min-width: auto;
  padding-left: 50px !important;
  height: 40px;
  line-height: 40px;
  margin: 1px 4px;
  border-radius: 6px;
  font-size: 13px;
}

:deep(.el-sub-menu .el-menu-item:hover) {
  background: rgba(99, 102, 241, 0.15) !important;
}

:deep(.el-sub-menu .el-menu-item.is-active) {
  background: rgba(99, 102, 241, 0.2) !important;
  color: #a78bfa !important;
}

:deep(.el-sub-menu__icon-arrow) {
  color: #64748b;
}

.menu-divider {
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(99, 102, 241, 0.3), transparent);
  margin: 12px 16px;
}

.sidebar-footer {
  padding: 12px 16px;
  border-top: 1px solid rgba(99, 102, 241, 0.1);
  flex-shrink: 0;
}

.sidebar-footer-text {
  text-align: center;
  font-size: 11px;
  color: #475569;
  letter-spacing: 0.5px;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 64px;
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(99, 102, 241, 0.1);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  transition: all 0.2s ease;
  color: #94a3b8;
}

.collapse-btn:hover {
  background: rgba(99, 102, 241, 0.1);
  color: #e2e8f0;
}

.breadcrumb {
  font-size: 14px;
}

.breadcrumb :deep(.el-breadcrumb__inner) {
  color: #94a3b8;
}

.breadcrumb :deep(.el-breadcrumb__item:last-child .el-breadcrumb__inner) {
  color: #e2e8f0;
  font-weight: 500;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

.cluster-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  background: rgba(99, 102, 241, 0.1);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 20px;
  font-size: 13px;
  color: #a78bfa;
}

.cluster-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(99, 102, 241, 0.1);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 20px;
  font-size: 13px;
  color: #a78bfa;
  cursor: pointer;
  transition: all 0.2s ease;
}

.cluster-selector:hover {
  background: rgba(99, 102, 241, 0.2);
  border-color: rgba(99, 102, 241, 0.4);
  box-shadow: 0 0 16px rgba(99, 102, 241, 0.2);
}

:deep(.el-dropdown-menu) {
  background: #1e293b;
  border: 1px solid rgba(99, 102, 241, 0.2);
}

:deep(.el-dropdown-menu__item) {
  color: #94a3b8;
}

:deep(.el-dropdown-menu__item:hover) {
  background: rgba(99, 102, 241, 0.1);
  color: #e2e8f0;
}

:deep(.el-dropdown-menu__item.is-active) {
  color: #a78bfa;
  background: rgba(99, 102, 241, 0.15);
}

.user-avatar {
  cursor: pointer;
  padding: 4px;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.user-avatar:hover {
  background: rgba(99, 102, 241, 0.1);
}

.avatar-gradient {
  background: linear-gradient(135deg, #6366f1, #8b5cf6) !important;
}

.main {
  background: #0f172a;
  min-height: calc(100vh - 64px);
  padding: 24px;
  overflow-y: auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@keyframes glow {
  0%, 100% {
    box-shadow: 0 0 20px rgba(99, 102, 241, 0.3);
  }
  50% {
    box-shadow: 0 0 30px rgba(99, 102, 241, 0.5);
  }
}

@media (max-width: 768px) {
  .aside:not(.collapsed) {
    position: fixed;
    z-index: 1000;
    height: 100vh;
  }

  .header {
    padding: 0 16px;
  }

  .breadcrumb {
    display: none;
  }

  .cluster-info span {
    display: none;
  }

  .main {
    padding: 16px;
  }
}
</style>
