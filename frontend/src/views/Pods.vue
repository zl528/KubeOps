<template>
  <div class="pods-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">Pods</h1>
        <span class="page-subtitle">管理集群中的 Pod 资源</span>
      </div>
      <div class="header-actions">
        <div class="ns-selector">
          <el-select v-model="namespace" placeholder="选择命名空间" clearable @change="fetchData">
            <el-option label="全部命名空间" value="" />
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </div>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchData">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="glass-table-container">
      <el-table
        :data="pods"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 Pod 数据'"
      >
        <el-table-column prop="name" label="名称" min-width="250" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="cell-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140">
          <template #default="{ row }">
            <span class="cell-ns">{{ row.namespace }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <span class="status-badge" :class="'status-' + statusClass(row.status)">{{ row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="node" label="节点" width="160" show-overflow-tooltip />
        <el-table-column prop="restarts" label="重启" width="70">
          <template #default="{ row }">
            <span class="cell-metric" :class="{ 'metric-warn': row.restarts > 0 }">{{ row.restarts }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="存活" width="80" />
        <el-table-column label="容器" min-width="180">
          <template #default="{ row }">
            <span v-for="c in (row.containers || [])" :key="c.name" class="container-tag" :class="c.ready ? 'ct-ready' : 'ct-unready'">
              {{ c.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button type="button" class="action-btn action-detail" @click="viewPod(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/></svg>
                详情
              </button>
              <button type="button" class="action-btn action-terminal" @click="openTerminal(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>
                终端
              </button>
              <button type="button" class="action-btn action-logs" @click="viewLogs(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
                日志
              </button>
              <button type="button" class="action-btn action-restart" @click="restartPod(row)" :disabled="row._restarting">
                <svg v-if="!row._restarting" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                <span v-else class="btn-spinner"></span>
                重启
              </button>
              <button type="button" class="action-btn action-delete" @click="deletePod(row)" :disabled="row._deleting">
                <svg v-if="!row._deleting" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                <span v-else class="btn-spinner"></span>
                删除
              </button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- Pod 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="Pod 详情" width="800px" class="dark-dialog">
      <template v-if="selectedPod">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ selectedPod.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ selectedPod.namespace }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <span class="status-badge" :class="'status-' + statusClass(selectedPod.status)">{{ selectedPod.status }}</span>
          </el-descriptions-item>
          <el-descriptions-item label="节点">{{ selectedPod.node }}</el-descriptions-item>
          <el-descriptions-item label="IP">{{ selectedPod.ip || '-' }}</el-descriptions-item>
          <el-descriptions-item label="重启次数">{{ selectedPod.restarts }}</el-descriptions-item>
          <el-descriptions-item label="QoS">{{ selectedPod.qos || '-' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ selectedPod.createdAt || '-' }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">容器</el-divider>
        <el-table :data="selectedPod.containers || []" size="small" class="detail-table">
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="image" label="镜像" show-overflow-tooltip />
          <el-table-column label="就绪" width="80">
            <template #default="{ row }">
              <span class="status-badge" :class="row.ready ? 'status-running' : 'status-failed'">{{ row.ready ? '是' : '否' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="重启" width="80">
            <template #default="{ row }">{{ row.restarts || 0 }}</template>
          </el-table-column>
          <el-table-column label="资源请求" width="150">
            <template #default="{ row }">
              <div v-if="row.resources?.requests">
                CPU: {{ row.resources.requests.cpu || '-' }}<br/>
                Mem: {{ row.resources.requests.memory || '-' }}
              </div>
              <span v-else>-</span>
            </template>
          </el-table-column>
        </el-table>

        <template v-if="selectedPod.labels && Object.keys(selectedPod.labels).length">
          <el-divider content-position="left">标签</el-divider>
          <div class="label-list">
            <el-tag v-for="(v, k) in selectedPod.labels" :key="k" size="small" type="info" style="margin: 2px">
              {{ k }}={{ v }}
            </el-tag>
          </div>
        </template>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useHighlightRow } from '../composables/useHighlightRow'
import api from '../api'
import { useGlobalNamespace } from '../store/namespace'

const router = useRouter()
useHighlightRow()
const pods = ref<any[]>([])
useHighlightRow()
const nsList = ref<any[]>([])
const { namespace } = useGlobalNamespace()
const loading = ref(false)
const detailVisible = ref(false)
const selectedPod = ref<any>(null)

const statusType = (s: string) => {
  const map: Record<string, string> = { Running: 'success', Pending: 'warning', Failed: 'danger', Succeeded: 'info', Error: 'danger', Unknown: 'info' }
  return map[s] || 'info'
}

const statusClass = (s: string) => {
  const map: Record<string, string> = { Running: 'running', Pending: 'pending', Failed: 'failed', Succeeded: 'succeeded', Error: 'error', Unknown: 'unknown' }
  return map[s] || 'unknown'
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = namespace.value ? { namespace: namespace.value } : {}
    const res: any = await api.get('/pods', { params })
    pods.value = (res.data || []).map((p: any) => ({ ...p, _restarting: false, _deleting: false }))
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const fetchNs = async () => {
  try {
    const res: any = await api.get('/namespaces')
    nsList.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const viewPod = async (row: any) => {
  try {
    const res: any = await api.get('/pods/get', { params: { namespace: row.namespace, name: row.name } })
    selectedPod.value = res.data || row
    detailVisible.value = true
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '获取详情失败')
  }
}

const restartPod = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要重启 Pod "${row.name}" 吗？`, '确认重启', { type: 'warning' })
    row._restarting = true
    await api.post('/pods/restart', { namespace: row.namespace, name: row.name })
    ElMessage.success('重启请求已发送')
    setTimeout(fetchData, 2000)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e.response?.data?.message || '重启失败')
  } finally {
    row._restarting = false
  }
}

const deletePod = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Pod "${row.name}" 吗？此操作不可恢复。`, '确认删除', { type: 'error' })
    row._deleting = true
    await api.delete('/pods/delete', { params: { namespace: row.namespace, name: row.name } })
    ElMessage.success('已删除')
    fetchData()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e.response?.data?.message || '删除失败')
  } finally {
    row._deleting = false
  }
}

const viewLogs = (row: any) => {
  router.push({ path: '/logs', query: { namespace: row.namespace, pod: row.name } })
}

const openTerminal = (row: any) => {
  router.push({ path: '/terminal', query: { namespace: row.namespace, pod: row.name } })
}

const headerCellStyle = () => ({
  background: 'rgba(30, 41, 59, 0.9)',
  color: '#94a3b8',
  borderBottom: '1px solid rgba(148, 163, 184, 0.1)',
  fontSize: '12px',
  fontWeight: '500',
  textTransform: 'uppercase',
  letterSpacing: '0.5px',
})

const cellStyle = () => ({
  background: 'transparent',
  color: '#f1f5f9',
  borderBottom: '1px solid rgba(148, 163, 184, 0.06)',
  fontSize: '14px',
})

const rowClassName = ({ rowIndex }: { rowIndex: number }) =>
  rowIndex % 2 === 0 ? 'row-even' : 'row-odd'

onMounted(() => { fetchNs(); fetchData() })
</script>

<style scoped>
.pods-page {
  padding: 24px;
  background: var(--bg-primary);
  min-height: 100vh;
}

.page-header-gradient {
  overflow: hidden !important;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.15), rgba(139, 92, 246, 0.1));
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 16px;
  padding: 28px 32px;
  margin-bottom: 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  position: relative;
}

.page-header-gradient::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -10%;
  width: 300px;
  height: 300px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.12) 0%, transparent 70%);
  pointer-events: none;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  font-size: 26px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  letter-spacing: -0.5px;
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  position: relative;
  z-index: 1;
}

.ns-selector :deep(.el-select) {
  width: 200px;
}

.ns-selector :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 8px;
  box-shadow: none;
}

.ns-selector :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.ns-selector :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.ns-selector :deep(.el-input__inner) {
  color: var(--text-primary);
}

.ns-selector :deep(.el-input__inner::placeholder) {
  color: var(--text-secondary);
}

.btn-gradient {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 8px;
  font-weight: 500;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-gradient:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.4), 0 0 20px rgba(99, 102, 241, 0.3);
}

.btn-gradient:active {
  transform: translateY(0);
}

.glass-table-container {
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
}

.glass-table-container :deep(.el-table) {
  background: transparent;
  --el-table-bg-color: transparent;
  --el-table-header-bg-color: rgba(30, 41, 59, 0.9);
  --el-table-header-text-color: #94a3b8;
  --el-table-text-color: #f1f5f9;
  --el-table-border-color: rgba(148, 163, 184, 0.06);
  --el-table-row-hover-bg-color: rgba(51, 65, 85, 0.4);
  --el-table-current-row-bg-color: rgba(99, 102, 241, 0.1);
}

.glass-table-container :deep(.el-table th.el-table__cell) {
  background: rgba(30, 41, 59, 0.9) !important;
  color: #94a3b8 !important;
  font-size: 12px;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1) !important;
}

.glass-table-container :deep(.el-table td.el-table__cell) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.06) !important;
}

.glass-table-container :deep(.el-table--striped .el-table__body tr.el-table__row--striped td.el-table__cell) {
  background: rgba(30, 41, 59, 0.3);
}

.glass-table-container :deep(.el-table__empty-block) {
  background: transparent;
}

.glass-table-container :deep(.el-table__empty-text) {
  color: var(--text-secondary);
}

.glass-table-container :deep(.el-table .cell) {
  padding: 0 16px;
}

.glass-table-container :deep(.el-table .el-table__row) {
  transition: background 0.15s ease;
}

.glass-table-container :deep(.el-table .el-table__row:hover > td.el-table__cell) {
  background: rgba(51, 65, 85, 0.4) !important;
}

.cell-name {
  font-weight: 500;
  color: #e2e8f0;
}

.cell-ns {
  font-size: 12px;
  color: var(--text-secondary);
  background: rgba(51, 65, 85, 0.5);
  padding: 2px 8px;
  border-radius: 4px;
}

.cell-metric {
  font-variant-numeric: tabular-nums;
  font-weight: 500;
}

.metric-warn {
  color: var(--warning);
}

/* 状态徽章 */
.status-badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
}

.status-running {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
  border: 1px solid rgba(34, 197, 94, 0.2);
}

.status-pending {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.status-failed {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.status-succeeded {
  background: rgba(59, 130, 246, 0.12);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.2);
}

.status-error {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.status-unknown {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
  border: 1px solid rgba(148, 163, 184, 0.2);
}

/* 容器标签 */
.container-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  margin: 2px;
}

.ct-ready {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
}

.ct-unready {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

/* 操作按钮 */
.action-cell {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 10px;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-detail {
  background: rgba(59, 130, 246, 0.12);
  color: #60a5fa;
}

.action-detail:hover:not(:disabled) {
  background: rgba(59, 130, 246, 0.25);
}

.action-terminal {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
}

.action-terminal:hover:not(:disabled) {
  background: rgba(34, 197, 94, 0.25);
}

.action-logs {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
}

.action-logs:hover:not(:disabled) {
  background: rgba(148, 163, 184, 0.25);
}

.action-restart {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
}

.action-restart:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.25);
}

.action-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.action-delete:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
}

.btn-spinner {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.label-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

/* 暗色弹窗 */
.pods-page :deep(.dark-dialog .el-dialog),
.pods-page :deep(.el-dialog.dark-dialog) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.pods-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.pods-page :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  padding: 20px 24px;
  margin: 0;
}

.pods-page :deep(.el-dialog__title) {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 18px;
}

.pods-page :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--text-secondary);
}

.pods-page :deep(.el-dialog__body) {
  padding: 24px;
  color: var(--text-primary);
}

.pods-page :deep(.el-descriptions) {
  --el-descriptions-item-bordered-label-background: rgba(30, 41, 59, 0.8);
}

.pods-page :deep(.el-descriptions__label) {
  color: var(--text-secondary);
}

.pods-page :deep(.el-descriptions__content) {
  color: var(--text-primary);
}

.pods-page :deep(.el-descriptions__cell) {
  border-color: rgba(148, 163, 184, 0.08) !important;
}

.pods-page :deep(.el-divider__text) {
  background: #1e293b;
  color: var(--text-secondary);
}

.pods-page :deep(.el-divider) {
  border-color: rgba(148, 163, 184, 0.08);
}

.pods-page :deep(.el-tag--info) {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(148, 163, 184, 0.1);
  color: var(--text-secondary);
}

/* 详情表格 */
.detail-table {
  background: transparent;
  --el-table-bg-color: transparent;
  --el-table-header-bg-color: rgba(30, 41, 59, 0.9);
  --el-table-header-text-color: #94a3b8;
  --el-table-text-color: #f1f5f9;
  --el-table-border-color: rgba(148, 163, 184, 0.06);
}

.detail-table :deep(.el-table th.el-table__cell) {
  background: rgba(30, 41, 59, 0.9) !important;
  color: #94a3b8 !important;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1) !important;
}

.detail-table :deep(.el-table td.el-table__cell) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.06) !important;
}

.glass-table-container :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(4px);
}

.glass-table-container :deep(.el-loading-spinner .circular) {
  stroke: var(--primary);
}

.pods-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}
</style>
