import { createRouter, createWebHistory } from 'vue-router'
import Layout from '../components/Layout.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { title: '集群连接', public: true }
  },
  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { title: '集群概览', icon: 'Monitor' }
      },
      {
        path: 'resource-graph',
        name: 'ResourceGraph',
        component: () => import('../views/ResourceGraph.vue'),
        meta: { title: '资源关系图', icon: 'Share' }
      },
      {
        path: 'nodes',
        name: 'Nodes',
        component: () => import('../views/Nodes.vue'),
        meta: { title: '节点管理', icon: 'Cpu' }
      },
      {
        path: 'namespaces',
        name: 'Namespaces',
        component: () => import('../views/Namespaces.vue'),
        meta: { title: '命名空间', icon: 'Folder' }
      },
      {
        path: 'namespace-explorer',
        name: 'NamespaceExplorer',
        component: () => import('../views/NamespaceExplorer.vue'),
        meta: { title: '命名空间视图', icon: 'FolderOpened' }
      },
      {
        path: 'pods',
        name: 'Pods',
        component: () => import('../views/Pods.vue'),
        meta: { title: 'Pod 管理', icon: 'Box' }
      },
      {
        path: 'deployments',
        name: 'Deployments',
        component: () => import('../views/Deployments.vue'),
        meta: { title: 'Deployment', icon: 'Upload' }
      },
      {
        path: 'statefulsets',
        name: 'StatefulSets',
        component: () => import('../views/StatefulSets.vue'),
        meta: { title: 'StatefulSet', icon: 'Coin' }
      },
      {
        path: 'daemonsets',
        name: 'DaemonSets',
        component: () => import('../views/DaemonSets.vue'),
        meta: { title: 'DaemonSet', icon: 'Grid' }
      },
      {
        path: 'services',
        name: 'Services',
        component: () => import('../views/Services.vue'),
        meta: { title: 'Service', icon: 'Connection' }
      },
      {
        path: 'ingresses',
        name: 'Ingresses',
        component: () => import('../views/Ingresses.vue'),
        meta: { title: 'Ingress', icon: 'Link' }
      },
      {
        path: 'networkpolicies',
        name: 'NetworkPolicies',
        component: () => import('../views/NetworkPolicies.vue'),
        meta: { title: 'NetworkPolicy', icon: 'Lock' }
      },
      {
        path: 'endpoints',
        name: 'Endpoints',
        component: () => import('../views/Endpoints.vue'),
        meta: { title: 'Endpoint', icon: 'Share' }
      },
      {
        path: 'configmaps',
        name: 'ConfigMaps',
        component: () => import('../views/ConfigMaps.vue'),
        meta: { title: 'ConfigMap', icon: 'Document' }
      },
      {
        path: 'secrets',
        name: 'Secrets',
        component: () => import('../views/Secrets.vue'),
        meta: { title: 'Secret', icon: 'Lock' }
      },
      {
        path: 'cronjobs',
        name: 'CronJobs',
        component: () => import('../views/CronJobs.vue'),
        meta: { title: 'CronJob', icon: 'Timer' }
      },
      {
        path: 'jobs',
        name: 'Jobs',
        component: () => import('../views/Jobs.vue'),
        meta: { title: 'Job', icon: 'Suitcase' }
      },
      {
        path: 'roles',
        name: 'Roles',
        component: () => import('../views/Roles.vue'),
        meta: { title: 'Role', icon: 'User' }
      },
      {
        path: 'clusterroles',
        name: 'ClusterRoles',
        component: () => import('../views/ClusterRoles.vue'),
        meta: { title: 'ClusterRole', icon: 'UserFilled' }
      },
      {
        path: 'rolebindings',
        name: 'RoleBindings',
        component: () => import('../views/RoleBindings.vue'),
        meta: { title: 'RoleBinding', icon: 'Link' }
      },
      {
        path: 'clusterrolebindings',
        name: 'ClusterRoleBindings',
        component: () => import('../views/ClusterRoleBindings.vue'),
        meta: { title: 'ClusterRoleBinding', icon: 'Link' }
      },
      {
        path: 'serviceaccounts',
        name: 'ServiceAccounts',
        component: () => import('../views/ServiceAccounts.vue'),
        meta: { title: 'ServiceAccount', icon: 'Avatar' }
      },
      {
        path: 'persistentvolumes',
        name: 'PersistentVolumes',
        component: () => import('../views/PersistentVolumes.vue'),
        meta: { title: 'PersistentVolume', icon: 'Box' }
      },
      {
        path: 'persistentvolumeclaims',
        name: 'PersistentVolumeClaims',
        component: () => import('../views/PersistentVolumeClaims.vue'),
        meta: { title: 'PersistentVolumeClaim', icon: 'Box' }
      },
      {
        path: 'storageclasses',
        name: 'StorageClasses',
        component: () => import('../views/StorageClasses.vue'),
        meta: { title: 'StorageClass', icon: 'Box' }
      },
      {
        path: 'limitranges',
        name: 'LimitRanges',
        component: () => import('../views/LimitRanges.vue'),
        meta: { title: 'LimitRange', icon: 'Timer' }
      },
      {
        path: 'hpas',
        name: 'HPAs',
        component: () => import('../views/HPAs.vue'),
        meta: { title: 'HPA', icon: 'TrendCharts' }
      },
      {
        path: 'events',
        name: 'Events',
        component: () => import('../views/Events.vue'),
        meta: { title: '事件', icon: 'Bell' }
      },
      {
        path: 'monitoring',
        name: 'Monitoring',
        component: () => import('../views/Monitoring.vue'),
        meta: { title: '监控中心', icon: 'DataAnalysis' }
      },
      {
        path: 'alerts',
        name: 'Alerts',
        component: () => import('../views/Alerts.vue'),
        meta: { title: '告警管理', icon: 'Warning' }
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('../views/Logs.vue'),
        meta: { title: '日志管理', icon: 'Document' }
      },
      {
        path: 'audit',
        name: 'AuditLogs',
        component: () => import('../views/AuditLogs.vue'),
        meta: { title: '审计日志', icon: 'Notebook' }
      },
      {
        path: 'backup',
        name: 'Backup',
        component: () => import('../views/Backup.vue'),
        meta: { title: '备份恢复', icon: 'Download' }
      },
      {
        path: 'clusters',
        name: 'Clusters',
        component: () => import('../views/Clusters.vue'),
        meta: { title: '集群管理', icon: 'Connection', requireAdmin: true }
      },
      {
        path: 'terminal',
        name: 'Terminal',
        component: () => import('../views/Terminal.vue'),
        meta: { title: 'WebTerminal', icon: 'Monitor' }
      },
      {
        path: 'user-management',
        name: 'UserManagement',
        component: () => import('../views/UserManagement.vue'),
        meta: { title: '用户管理', icon: 'User', requireAdmin: true }
      },
      {
        path: 'role-management',
        name: 'RoleManagement',
        component: () => import('../views/RoleManagement.vue'),
        meta: { title: '角色管理', icon: 'UserFilled', requireAdmin: true }
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('../views/Profile.vue'),
        meta: { title: '个人设置', icon: 'Setting' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  
  if (to.path === '/login') {
    if (token) {
      next('/dashboard')
    } else {
      next()
    }
    return
  }

  if (!token) {
    next('/login')
    return
  }

  // Check admin permission for protected routes
  if (to.meta?.requireAdmin) {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]))
      if (payload.role !== 'admin') {
        next('/dashboard')
        return
      }
    } catch {
      next('/login')
      return
    }
  }

  next()
})

export default router
