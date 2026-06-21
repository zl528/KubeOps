# Namespace Explorer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use compose:subagent (recommended) or compose:execute to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign frontend to show namespace-scoped resources grouped by namespace in a left-right panel layout, with a unified resource table component.

**Architecture:** New `NamespaceExplorer.vue` page with left panel listing namespaces and right panel showing resource tabs. A reusable `ResourceTable.vue` component renders different resource types with dynamic columns and action buttons. Existing pages remain unchanged.

**Tech Stack:** Vue 3 + TypeScript + Element Plus

---

## File Structure

| Action | File | Purpose |
|--------|------|---------|
| Create | `frontend/src/components/ResourceTable.vue` | Unified resource table with dynamic columns/actions |
| Create | `frontend/src/views/NamespaceExplorer.vue` | Main namespace explorer page (left-right layout) |
| Modify | `frontend/src/router/index.ts` | Add `/namespaces/explorer` route |
| Modify | `frontend/src/components/Layout.vue` | Add "命名空间" menu item |

---

### Task 1: Create ResourceTable.vue Component

**Covers:** Unified list component for all resource types

**Files:**
- Create: `frontend/src/components/ResourceTable.vue`

- [ ] **Step 1: Create ResourceTable component with props and dynamic columns**

```vue
<template>
  <div class="resource-table">
    <div class="table-header">
      <div class="header-left">
        <slot name="actions" />
      </div>
      <div class="header-right">
        <el-input
          v-model="searchText"
          placeholder="搜索..."
          clearable
          style="width: 200px"
          :prefix-icon="Search"
        />
        <el-button @click="$emit('refresh')" :icon="Refresh" style="margin-left: 8px">刷新</el-button>
      </div>
    </div>
    <el-table
      :data="filteredData"
      v-loading="loading"
      stripe
      highlight-current-row
      style="width: 100%"
      @row-click="$emit('row-click', $event)"
    >
      <el-table-column
        v-for="col in columns"
        :key="col.prop"
        :prop="col.prop"
        :label="col.label"
        :width="col.width"
        :min-width="col.minWidth"
        :show-overflow-tooltip="col.tooltip !== false"
      >
        <template #default="{ row }" v-if="col.slot">
          <slot :name="col.slot" :row="row" />
        </template>
      </el-table-column>
      <el-table-column
        v-if="actions.length > 0"
        label="操作"
        :width="actionsWidth"
        fixed="right"
      >
        <template #default="{ row }">
          <el-button
            v-for="action in actions"
            :key="action.name"
            :type="action.type"
            link
            size="small"
            @click.stop="$emit('action', { action: action.name, row })"
            :loading="row[`_loading_${action.name}`]"
          >
            {{ action.label }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Search, Refresh } from '@element-plus/icons-vue'

export interface Column {
  prop: string
  label: string
  width?: number
  minWidth?: number
  slot?: string
  tooltip?: boolean
}

export interface Action {
  name: string
  label: string
  type?: 'primary' | 'warning' | 'info' | 'success' | 'danger'
}

const props = defineProps<{
  columns: Column[]
  actions: Action[]
  data: any[]
  loading?: boolean
  actionsWidth?: number
}>()

defineEmits<{
  (e: 'refresh'): void
  (e: 'action', payload: { action: string; row: any }): void
  (e: 'row-click', row: any): void
}>()

const searchText = ref('')

const filteredData = computed(() => {
  if (!searchText.value) return props.data
  const keyword = searchText.value.toLowerCase()
  return props.data.filter(row =>
    Object.values(row).some(val =>
      String(val).toLowerCase().includes(keyword)
    )
  )
})
</script>

<style scoped>
.resource-table {
  width: 100%;
}
.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.header-left {
  display: flex;
  gap: 8px;
}
.header-right {
  display: flex;
  align-items: center;
}
</style>
```

- [ ] **Step 2: Verify component compiles**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: No errors

---

### Task 2: Create NamespaceExplorer.vue Page

**Covers:** Main namespace explorer with left-right panel layout

**Files:**
- Create: `frontend/src/views/NamespaceExplorer.vue`

- [ ] **Step 1: Create NamespaceExplorer with namespace list and resource tabs**

```vue
<template>
  <div class="namespace-explorer">
    <!-- Left Panel: Namespace List -->
    <div class="ns-panel">
      <div class="ns-panel-header">
        <span class="ns-title">命名空间</span>
        <el-tag size="small" type="info">{{ namespaces.length }}</el-tag>
      </div>
      <el-input
        v-model="nsSearch"
        placeholder="搜索..."
        clearable
        size="small"
        style="margin: 0 12px 8px"
        :prefix-icon="Search"
      />
      <el-scrollbar class="ns-list">
        <div
          v-for="ns in filteredNamespaces"
          :key="ns.name"
          class="ns-item"
          :class="{ active: selectedNs === ns.name }"
          @click="selectNamespace(ns.name)"
        >
          <el-icon><Folder /></el-icon>
          <span class="ns-name">{{ ns.name }}</span>
          <el-tag size="small" :type="ns.status === 'Active' ? 'success' : 'danger'">
            {{ ns.status }}
          </el-tag>
        </div>
      </el-scrollbar>
    </div>

    <!-- Right Panel: Resource Tabs -->
    <div class="resource-panel" v-if="selectedNs">
      <div class="resource-header">
        <el-tag size="large" type="primary">{{ selectedNs }}</el-tag>
      </div>
      <el-tabs v-model="activeTab" type="border-card">
        <!-- Pods Tab -->
        <el-tab-pane label="Pods" name="pods">
          <ResourceTable
            :columns="podColumns"
            :actions="podActions"
            :data="pods"
            :loading="loading.pods"
            @refresh="fetchPods"
            @action="handlePodAction"
          >
            <template #status="{ row }">
              <el-tag :type="podStatusType(row.status)" size="small">{{ row.status }}</el-tag>
            </template>
            <template #restarts="{ row }">
              <span :class="{ 'text-danger': row.restarts > 5 }">{{ row.restarts }}</span>
            </template>
          </ResourceTable>
        </el-tab-pane>

        <!-- Deployments Tab -->
        <el-tab-pane label="Deployments" name="deployments">
          <ResourceTable
            :columns="deploymentColumns"
            :actions="deploymentActions"
            :data="deployments"
            :loading="loading.deployments"
            @refresh="fetchDeployments"
            @action="handleDeploymentAction"
          />
        </el-tab-pane>

        <!-- Services Tab -->
        <el-tab-pane label="Services" name="services">
          <ResourceTable
            :columns="serviceColumns"
            :actions="serviceActions"
            :data="services"
            :loading="loading.services"
            @refresh="fetchServices"
            @action="handleServiceAction"
          />
        </el-tab-pane>

        <!-- ConfigMaps Tab -->
        <el-tab-pane label="ConfigMaps" name="configmaps">
          <ResourceTable
            :columns="configmapColumns"
            :actions="configmapActions"
            :data="configmaps"
            :loading="loading.configmaps"
            @refresh="fetchConfigMaps"
            @action="handleConfigMapAction"
          />
        </el-tab-pane>

        <!-- Secrets Tab -->
        <el-tab-pane label="Secrets" name="secrets">
          <ResourceTable
            :columns="secretColumns"
            :actions="secretActions"
            :data="secrets"
            :loading="loading.secrets"
            @refresh="fetchSecrets"
            @action="handleSecretAction"
          />
        </el-tab-pane>

        <!-- Other tabs with similar structure -->
        <el-tab-pane label="Ingresses" name="ingresses">
          <ResourceTable
            :columns="ingressColumns"
            :actions="[{ name: 'delete', label: '删除', type: 'danger' }]"
            :data="ingresses"
            :loading="loading.ingresses"
            @refresh="fetchIngresses"
            @action="handleIngressAction"
          />
        </el-tab-pane>

        <el-tab-pane label="Events" name="events">
          <ResourceTable
            :columns="eventColumns"
            :actions="[]"
            :data="events"
            :loading="loading.events"
            @refresh="fetchEvents"
          />
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- Empty State -->
    <div class="empty-state" v-else>
      <el-empty description="请选择一个命名空间">
        <el-icon size="64" color="#c0c4cc"><Folder /></el-icon>
      </el-empty>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Folder } from '@element-plus/icons-vue'
import ResourceTable from '../components/ResourceTable.vue'
import api from '../api'
import type { Column, Action } from '../components/ResourceTable.vue'

const router = useRouter()

// Namespace state
const namespaces = ref<any[]>([])
const selectedNs = ref('')
const nsSearch = ref('')

// Tab state
const activeTab = ref('pods')

// Resource data
const pods = ref<any[]>([])
const deployments = ref<any[]>([])
const services = ref<any[]>([])
const configmaps = ref<any[]>([])
const secrets = ref<any[]>([])
const ingresses = ref<any[]>([])
const events = ref<any[]>([])

// Loading states
const loading = reactive({
  pods: false,
  deployments: false,
  services: false,
  configmaps: false,
  secrets: false,
  ingresses: false,
  events: false,
})

// Column definitions
const podColumns: Column[] = [
  { prop: 'name', label: '名称', minWidth: 200 },
  { prop: 'status', label: '状态', width: 100, slot: 'status' },
  { prop: 'restarts', label: '重启', width: 80, slot: 'restarts' },
  { prop: 'node', label: '节点', minWidth: 150 },
  { prop: 'age', label: '创建时间', width: 120 },
]

const podActions: Action[] = [
  { name: 'terminal', label: '终端', type: 'success' },
  { name: 'logs', label: '日志', type: 'info' },
  { name: 'restart', label: '重启', type: 'warning' },
  { name: 'delete', label: '删除', type: 'danger' },
]

const deploymentColumns: Column[] = [
  { prop: 'name', label: '名称', minWidth: 200 },
  { prop: 'replicas', label: '副本', width: 80 },
  { prop: 'available', label: '可用', width: 80 },
  { prop: 'age', label: '创建时间', width: 120 },
]

const deploymentActions: Action[] = [
  { name: 'scale', label: '扩缩容', type: 'primary' },
  { name: 'restart', label: '重启', type: 'warning' },
  { name: 'rollback', label: '回滚', type: 'info' },
  { name: 'delete', label: '删除', type: 'danger' },
]

const serviceColumns: Column[] = [
  { prop: 'name', label: '名称', minWidth: 200 },
  { prop: 'type', label: '类型', width: 120 },
  { prop: 'clusterIP', label: 'ClusterIP', width: 140 },
  { prop: 'ports', label: '端口', minWidth: 150 },
  { prop: 'age', label: '创建时间', width: 120 },
]

const serviceActions: Action[] = [
  { name: 'detail', label: '详情', type: 'primary' },
  { name: 'delete', label: '删除', type: 'danger' },
]

const configmapColumns: Column[] = [
  { prop: 'name', label: '名称', minWidth: 200 },
  { prop: 'keys', label: '键数', width: 80 },
  { prop: 'age', label: '创建时间', width: 120 },
]

const configmapActions: Action[] = [
  { name: 'view', label: '查看', type: 'primary' },
  { name: 'edit', label: '编辑', type: 'warning' },
  { name: 'delete', label: '删除', type: 'danger' },
]

const secretColumns: Column[] = [
  { prop: 'name', label: '名称', minWidth: 200 },
  { prop: 'type', label: '类型', width: 150 },
  { prop: 'keys', label: '键数', width: 80 },
  { prop: 'age', label: '创建时间', width: 120 },
]

const secretActions: Action[] = [
  { name: 'view', label: '查看', type: 'primary' },
  { name: 'edit', label: '编辑', type: 'warning' },
  { name: 'delete', label: '删除', type: 'danger' },
]

const ingressColumns: Column[] = [
  { prop: 'name', label: '名称', minWidth: 200 },
  { prop: 'host', label: 'Host', minWidth: 150 },
  { prop: 'tls', label: 'TLS', width: 80 },
  { prop: 'age', label: '创建时间', width: 120 },
]

const eventColumns: Column[] = [
  { prop: 'type', label: '类型', width: 100 },
  { prop: 'reason', label: '原因', width: 120 },
  { prop: 'object', label: '对象', minWidth: 150 },
  { prop: 'message', label: '消息', minWidth: 200 },
  { prop: 'age', label: '时间', width: 120 },
]

// Computed
const filteredNamespaces = computed(() => {
  if (!nsSearch.value) return namespaces.value
  return namespaces.value.filter(ns =>
    ns.name.toLowerCase().includes(nsSearch.value.toLowerCase())
  )
})

// Methods
const selectNamespace = (name: string) => {
  selectedNs.value = name
  activeTab.value = 'pods'
  fetchAllResources()
}

const fetchNamespaces = async () => {
  try {
    const res: any = await api.get('/namespaces')
    namespaces.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const fetchAllResources = async () => {
  if (!selectedNs.value) return
  fetchPods()
  fetchDeployments()
  fetchServices()
  fetchConfigMaps()
  fetchSecrets()
  fetchIngresses()
  fetchEvents()
}

const fetchPods = async () => {
  loading.pods = true
  try {
    const res: any = await api.get('/pods', { params: { namespace: selectedNs.value } })
    pods.value = res.data || []
  } catch (e) { console.error(e) }
  loading.pods = false
}

const fetchDeployments = async () => {
  loading.deployments = true
  try {
    const res: any = await api.get('/deployments', { params: { namespace: selectedNs.value } })
    deployments.value = res.data || []
  } catch (e) { console.error(e) }
  loading.deployments = false
}

const fetchServices = async () => {
  loading.services = true
  try {
    const res: any = await api.get('/services', { params: { namespace: selectedNs.value } })
    services.value = res.data || []
  } catch (e) { console.error(e) }
  loading.services = false
}

const fetchConfigMaps = async () => {
  loading.configmaps = true
  try {
    const res: any = await api.get('/configmaps', { params: { namespace: selectedNs.value } })
    configmaps.value = res.data || []
  } catch (e) { console.error(e) }
  loading.configmaps = false
}

const fetchSecrets = async () => {
  loading.secrets = true
  try {
    const res: any = await api.get('/secrets', { params: { namespace: selectedNs.value } })
    secrets.value = res.data || []
  } catch (e) { console.error(e) }
  loading.secrets = false
}

const fetchIngresses = async () => {
  loading.ingresses = true
  try {
    const res: any = await api.get('/ingresses', { params: { namespace: selectedNs.value } })
    ingresses.value = res.data || []
  } catch (e) { console.error(e) }
  loading.ingresses = false
}

const fetchEvents = async () => {
  loading.events = true
  try {
    const res: any = await api.get('/events', { params: { namespace: selectedNs.value } })
    events.value = res.data || []
  } catch (e) { console.error(e) }
  loading.events = false
}

const podStatusType = (status: string) => {
  const map: Record<string, string> = {
    Running: 'success',
    Pending: 'warning',
    Succeeded: 'success',
    Failed: 'danger',
    Unknown: 'info',
  }
  return map[status] || 'info'
}

// Action handlers
const handlePodAction = ({ action, row }: { action: string; row: any }) => {
  switch (action) {
    case 'terminal':
      router.push({ path: '/terminal', query: { namespace: selectedNs.value, pod: row.name } })
      break
    case 'logs':
      router.push({ path: '/logs', query: { namespace: selectedNs.value, pod: row.name } })
      break
    case 'restart':
      restartPod(row)
      break
    case 'delete':
      deletePod(row)
      break
  }
}

const handleDeploymentAction = ({ action, row }: { action: string; row: any }) => {
  switch (action) {
    case 'scale':
      // TODO: show scale dialog
      break
    case 'restart':
      restartDeployment(row)
      break
    case 'rollback':
      // TODO: show rollback dialog
      break
    case 'delete':
      deleteDeployment(row)
      break
  }
}

const handleServiceAction = ({ action, row }: { action: string; row: any }) => {
  if (action === 'delete') deleteService(row)
}

const handleConfigMapAction = ({ action, row }: { action: string; row: any }) => {
  if (action === 'delete') deleteConfigMap(row)
}

const handleSecretAction = ({ action, row }: { action: string; row: any }) => {
  if (action === 'delete') deleteSecret(row)
}

const handleIngressAction = ({ action, row }: { action: string; row: any }) => {
  if (action === 'delete') deleteIngress(row)
}

// CRUD operations
const restartPod = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要重启 Pod "${row.name}" 吗？`, '确认', { type: 'warning' })
    await api.post(`/pods/restart?namespace=${selectedNs.value}&name=${row.name}`)
    ElMessage.success('重启成功')
    fetchPods()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('重启失败')
  }
}

const deletePod = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Pod "${row.name}" 吗？`, '确认', { type: 'error' })
    await api.delete(`/pods/delete?namespace=${selectedNs.value}&name=${row.name}`)
    ElMessage.success('已删除')
    fetchPods()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const restartDeployment = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要重启 Deployment "${row.name}" 吗？`, '确认', { type: 'warning' })
    await api.post(`/deployments/restart?namespace=${selectedNs.value}&name=${row.name}`)
    ElMessage.success('重启成功')
    fetchDeployments()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('重启失败')
  }
}

const deleteDeployment = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Deployment "${row.name}" 吗？`, '确认', { type: 'error' })
    await api.delete(`/deployments/delete?namespace=${selectedNs.value}&name=${row.name}`)
    ElMessage.success('已删除')
    fetchDeployments()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const deleteService = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Service "${row.name}" 吗？`, '确认', { type: 'error' })
    await api.delete(`/services/delete?namespace=${selectedNs.value}&name=${row.name}`)
    ElMessage.success('已删除')
    fetchServices()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const deleteConfigMap = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 ConfigMap "${row.name}" 吗？`, '确认', { type: 'error' })
    await api.delete(`/configmaps/delete?namespace=${selectedNs.value}&name=${row.name}`)
    ElMessage.success('已删除')
    fetchConfigMaps()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const deleteSecret = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Secret "${row.name}" 吗？`, '确认', { type: 'error' })
    await api.delete(`/secrets/delete?namespace=${selectedNs.value}&name=${row.name}`)
    ElMessage.success('已删除')
    fetchSecrets()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const deleteIngress = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Ingress "${row.name}" 吗？`, '确认', { type: 'error' })
    await api.delete(`/ingresses/delete?namespace=${selectedNs.value}&name=${row.name}`)
    ElMessage.success('已删除')
    fetchIngresses()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => {
  fetchNamespaces()
})
</script>

<style scoped>
.namespace-explorer {
  display: flex;
  height: calc(100vh - 100px);
  gap: 16px;
}

.ns-panel {
  width: 280px;
  background: #fff;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.ns-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #ebeef5;
}

.ns-title {
  font-weight: 600;
  font-size: 16px;
}

.ns-list {
  flex: 1;
  overflow: hidden;
}

.ns-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.2s;
}

.ns-item:hover {
  background: #f5f7fa;
}

.ns-item.active {
  background: #ecf5ff;
  color: #409eff;
}

.ns-name {
  flex: 1;
  font-size: 14px;
}

.resource-panel {
  flex: 1;
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.resource-header {
  margin-bottom: 16px;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border-radius: 8px;
}

.text-danger {
  color: #f56c6c;
  font-weight: 600;
}
</style>
```

- [ ] **Step 2: Verify component compiles**

Run: `cd frontend && npx vue-tsc --noEmit`
Expected: No errors

---

### Task 3: Add Route and Menu Item

**Covers:** Integration with existing navigation

**Files:**
- Modify: `frontend/src/router/index.ts:118-121`
- Modify: `frontend/src/components/Layout.vue:76-81`

- [ ] **Step 1: Add route for NamespaceExplorer**

In `frontend/src/router/index.ts`, add after the `namespaces` route:

```typescript
{
  path: 'namespace-explorer',
  name: 'NamespaceExplorer',
  component: () => import('../views/NamespaceExplorer.vue'),
  meta: { title: '命名空间视图', icon: 'FolderOpened' }
},
```

- [ ] **Step 2: Add menu item in Layout.vue**

In `frontend/src/components/Layout.vue`, find the cluster submenu and add:

```vue
<el-menu-item index="/namespace-explorer">命名空间视图</el-menu-item>
```

- [ ] **Step 3: Verify build**

Run: `cd frontend && npm run build`
Expected: Build succeeds

---

### Task 4: Test End-to-End

**Covers:** Verify all functionality works

- [ ] **Step 1: Start dev server**

Run: `cd frontend && npm run dev`

- [ ] **Step 2: Verify namespace list loads**

Navigate to `/namespace-explorer`, verify left panel shows namespaces

- [ ] **Step 3: Verify resource tabs load**

Click a namespace, verify all tabs load data

- [ ] **Step 4: Verify actions work**

Test terminal, logs, restart, delete buttons

- [ ] **Step 5: Verify existing pages still work**

Navigate to `/pods`, `/deployments`, etc. - verify unchanged

---

## Summary

This plan creates:
1. **ResourceTable.vue** - Reusable component with dynamic columns/actions
2. **NamespaceExplorer.vue** - Main page with left-right panel layout
3. **Route + Menu** - Integration with existing navigation

Existing pages remain unchanged, ensuring no regression.
