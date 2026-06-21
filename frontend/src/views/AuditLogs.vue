<template>
  <div class="audit-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">Audit Logs</h1>
        <span class="page-subtitle">查看集群操作审计记录</span>
      </div>
      <div class="header-actions">
        <button type="button" class="btn-gradient btn-refresh" @click="fetchLogs">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="stats-row">
      <div class="stat-card glass-card">
        <div class="stat-icon total-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/><polyline points="10 9 9 9 8 9"/></svg>
        </div>
        <div class="stat-value">{{ stats.total || 0 }}</div>
        <div class="stat-label">总操作数</div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon recent-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
        </div>
        <div class="stat-value">{{ stats.recentCount || 0 }}</div>
        <div class="stat-label">最近1小时</div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon user-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
        </div>
        <div class="stat-value">{{ userCount }}</div>
        <div class="stat-label">操作用户数</div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon resource-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
        </div>
        <div class="stat-value">{{ resourceCount }}</div>
        <div class="stat-label">资源类型数</div>
      </div>
    </div>

    <div class="glass-card">
      <div class="card-header">
        <span class="card-title">审计日志</span>
      </div>

      <div class="filter-bar">
        <div class="filter-group">
          <select v-model="filterUser" class="filter-select" @change="fetchLogs">
            <option value="">全部用户</option>
            <option v-for="u in users" :key="u" :value="u">{{ u }}</option>
          </select>
          <select v-model="filterAction" class="filter-select" @change="fetchLogs">
            <option value="">全部操作</option>
            <option v-for="a in actions" :key="a" :value="a">{{ a }}</option>
          </select>
          <select v-model="filterStatus" class="filter-select" @change="fetchLogs">
            <option value="">全部状态</option>
            <option value="success">成功</option>
            <option value="failed">失败</option>
          </select>
          <input v-model="searchKeyword" class="filter-input" placeholder="搜索..." @keyup.enter="fetchLogs" />
          <button type="button" class="btn-gradient" @click="fetchLogs">查询</button>
        </div>
        <div class="filter-group">
          <button type="button" class="btn-action btn-export-json" @click="exportLogs('json')">导出 JSON</button>
          <button type="button" class="btn-action btn-export-csv" @click="exportLogs('csv')">导出 CSV</button>
          <button type="button" class="btn-action btn-cleanup" @click="handleCleanup">清理旧日志</button>
        </div>
      </div>

      <div class="glass-table-container">
        <el-table :data="logs" v-loading="loading" :header-cell-style="headerCellStyle" :cell-style="cellStyle" :row-class-name="rowClassName" class="custom-table" :empty-text="'暂无审计日志数据'">
          <el-table-column prop="timestamp" label="时间" width="180">
            <template #default="{ row }">
              <span class="cell-time">{{ formatTime(row.timestamp) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="user" label="用户" width="120">
            <template #default="{ row }">
              <span class="cell-user">{{ row.user }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="action" label="操作" width="120">
            <template #default="{ row }">
              <span class="cell-action" :class="'action-' + row.action">{{ row.action }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="resource" label="资源类型" width="120" />
          <el-table-column prop="name" label="资源名称" min-width="150" show-overflow-tooltip />
          <el-table-column prop="namespace" label="命名空间" width="120">
            <template #default="{ row }">
              <span class="cell-ns">{{ row.namespace || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <span class="cell-status" :class="row.status === 'success' ? 'status-ok' : 'status-fail'">{{ row.status }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="ip" label="IP" width="140">
            <template #default="{ row }">
              <span class="cell-ip">{{ row.ip || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100" fixed="right">
            <template #default="{ row }">
              <button type="button" class="action-btn action-detail" @click="showDetail(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/></svg>
                详情
              </button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="pagination-row">
        <span class="pagination-info">共 {{ total }} 条</span>
        <div class="pagination-controls">
          <select v-model="pageSize" class="page-size-select" @change="fetchLogs">
            <option :value="10">10 条/页</option>
            <option :value="20">20 条/页</option>
            <option :value="50">50 条/页</option>
            <option :value="100">100 条/页</option>
          </select>
          <button type="button" class="page-btn" :disabled="currentPage <= 1" @click="currentPage--; fetchLogs()">上一页</button>
          <span class="page-current">{{ currentPage }}</span>
          <button type="button" class="page-btn" :disabled="currentPage * pageSize >= total" @click="currentPage++; fetchLogs()">下一页</button>
        </div>
      </div>
    </div>

    <div class="modal-overlay" v-if="detailVisible" @click.self="detailVisible = false">
      <div class="modal-content glass-card">
        <div class="modal-header">
          <h3 class="modal-title">审计日志详情</h3>
          <button type="button" class="modal-close" @click="detailVisible = false">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="detail-grid">
          <div class="detail-item"><span class="detail-label">ID</span><span class="detail-value">{{ currentLog.id }}</span></div>
          <div class="detail-item"><span class="detail-label">时间</span><span class="detail-value">{{ formatTime(currentLog.timestamp) }}</span></div>
          <div class="detail-item"><span class="detail-label">用户</span><span class="detail-value">{{ currentLog.user }}</span></div>
          <div class="detail-item"><span class="detail-label">操作</span><span class="detail-value">{{ currentLog.action }}</span></div>
          <div class="detail-item"><span class="detail-label">资源类型</span><span class="detail-value">{{ currentLog.resource }}</span></div>
          <div class="detail-item"><span class="detail-label">资源名称</span><span class="detail-value">{{ currentLog.name }}</span></div>
          <div class="detail-item"><span class="detail-label">命名空间</span><span class="detail-value">{{ currentLog.namespace || '-' }}</span></div>
          <div class="detail-item"><span class="detail-label">状态</span><span class="detail-value">{{ currentLog.status }}</span></div>
          <div class="detail-item"><span class="detail-label">IP</span><span class="detail-value">{{ currentLog.ip || '-' }}</span></div>
          <div class="detail-item full-width"><span class="detail-label">UserAgent</span><span class="detail-value">{{ currentLog.userAgent || '-' }}</span></div>
          <div class="detail-item full-width">
            <span class="detail-label">详情</span>
            <pre v-if="currentLog.detail" class="detail-pre">{{ JSON.stringify(currentLog.detail, null, 2) }}</pre>
            <span v-else class="detail-value">-</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const logs = ref<any[]>([])
const stats = ref<any>({})
const users = ref<string[]>([])
const actions = ref<string[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const filterUser = ref('')
const filterAction = ref('')
const filterStatus = ref('')
const searchKeyword = ref('')
const detailVisible = ref(false)
const currentLog = ref<any>({})

const userCount = computed(() => Object.keys(stats.value?.byUser || {}).length)
const resourceCount = computed(() => Object.keys(stats.value?.byResource || {}).length)

const actionType = (action: string) => {
  const map: Record<string, string> = {
    create: 'success',
    delete: 'danger',
    update: 'warning',
    get: 'info',
    list: 'info',
  }
  return map[action] || 'info'
}

const formatTime = (ts: string) => {
  if (!ts) return '-'
  return new Date(ts).toLocaleString()
}

const fetchLogs = async () => {
  loading.value = true
  try {
    const query: any = {
      page: currentPage.value,
      pageSize: pageSize.value,
    }
    if (filterUser.value) query.user = filterUser.value
    if (filterAction.value) query.action = filterAction.value
    if (filterStatus.value) query.status = filterStatus.value
    if (searchKeyword.value) query.keyword = searchKeyword.value

    const res: any = await api.post('/audit/query', query)
    logs.value = res.data?.logs || []
    total.value = res.data?.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const res: any = await api.get('/audit/stats')
    stats.value = res.data || {}
  } catch (e) {
    console.error(e)
  }
}

const fetchUsers = async () => {
  try {
    const res: any = await api.get('/audit/users')
    users.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const fetchActions = async () => {
  try {
    const res: any = await api.get('/audit/actions')
    actions.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const exportLogs = async (format: string) => {
  try {
    const params = new URLSearchParams()
    if (filterUser.value) params.append('user', filterUser.value)
    if (filterAction.value) params.append('action', filterAction.value)
    params.append('format', format)

    const res = await api.get(`/audit/export?${params.toString()}`, { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([res as any]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `audit-logs.${format}`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (e) {
    console.error(e)
    ElMessage.error('导出失败')
  }
}

const handleCleanup = async () => {
  try {
    await ElMessageBox.confirm('确定要清理30天前的审计日志吗？', '确认', { type: 'warning' })
    const res: any = await api.delete('/audit/cleanup?days=30')
    ElMessage.success(`已清理 ${res.data?.removed || 0} 条日志`)
    fetchLogs()
    fetchStats()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('清理失败')
  }
}

const showDetail = (row: any) => {
  currentLog.value = row
  detailVisible.value = true
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

onMounted(() => {
  fetchLogs()
  fetchStats()
  fetchUsers()
  fetchActions()
})
</script>

<style scoped>
.audit-page {
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

.glass-card {
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  text-align: center;
  position: relative;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 12px;
}

.total-icon {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(139, 92, 246, 0.15));
  color: #6366f1;
}

.recent-icon {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.2), rgba(22, 163, 74, 0.15));
  color: #4ade80;
}

.user-icon {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.2), rgba(217, 119, 6, 0.15));
  color: #fbbf24;
}

.resource-icon {
  background: linear-gradient(135deg, rgba(236, 72, 153, 0.2), rgba(219, 39, 119, 0.15));
  color: #f472b6;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
  font-variant-numeric: tabular-nums;
}

.stat-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 20px;
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.filter-select,
.filter-input {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 8px;
  padding: 8px 12px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  transition: border-color 0.2s;
}

.filter-select:focus,
.filter-input:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.filter-select option {
  background: #1e293b;
  color: var(--text-primary);
}

.filter-input::placeholder {
  color: var(--text-secondary);
}

.btn-action {
  background: rgba(51, 65, 85, 0.6);
  color: var(--text-secondary);
  border: 1px solid rgba(148, 163, 184, 0.1);
  padding: 8px 14px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.btn-action:hover {
  background: rgba(51, 65, 85, 0.9);
  color: var(--text-primary);
  border-color: rgba(148, 163, 184, 0.2);
}

.btn-export-json {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
  border-color: rgba(34, 197, 94, 0.2);
}

.btn-export-json:hover {
  background: rgba(34, 197, 94, 0.25);
}

.btn-export-csv {
  background: rgba(59, 130, 246, 0.12);
  color: #60a5fa;
  border-color: rgba(59, 130, 246, 0.2);
}

.btn-export-csv:hover {
  background: rgba(59, 130, 246, 0.25);
}

.btn-cleanup {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  border-color: rgba(239, 68, 68, 0.2);
}

.btn-cleanup:hover {
  background: rgba(239, 68, 68, 0.25);
}

.glass-table-container {
  background: rgba(30, 41, 59, 0.4);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 12px;
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

.glass-table-container :deep(.el-table .el-table__row:hover > td.el-table__cell) {
  background: rgba(51, 65, 85, 0.4) !important;
}

.glass-table-container :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(4px);
}

.glass-table-container :deep(.el-loading-spinner .circular) {
  stroke: var(--primary);
}

.cell-time {
  font-size: 13px;
  color: var(--text-secondary);
}

.cell-user {
  font-weight: 500;
  color: #e2e8f0;
}

.cell-action {
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
  text-transform: capitalize;
}

.cell-action.action-create {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
}

.cell-action.action-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.cell-action.action-update {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
}

.cell-action.action-get,
.cell-action.action-list {
  background: rgba(59, 130, 246, 0.12);
  color: #60a5fa;
}

.cell-ns {
  font-size: 12px;
  color: var(--text-secondary);
  background: rgba(51, 65, 85, 0.5);
  padding: 2px 8px;
  border-radius: 4px;
}

.cell-status {
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
}

.status-ok {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
}

.status-fail {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.cell-ip {
  font-size: 13px;
  color: var(--text-secondary);
  font-variant-numeric: tabular-nums;
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

.action-detail {
  background: rgba(59, 130, 246, 0.12);
  color: #60a5fa;
}

.action-detail:hover {
  background: rgba(59, 130, 246, 0.25);
}

.pagination-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
}

.pagination-info {
  font-size: 13px;
  color: var(--text-secondary);
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-size-select {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 6px;
  padding: 6px 10px;
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
}

.page-size-select option {
  background: #1e293b;
}

.page-btn {
  background: rgba(51, 65, 85, 0.6);
  color: var(--text-secondary);
  border: 1px solid rgba(148, 163, 184, 0.1);
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.page-btn:hover:not(:disabled) {
  background: rgba(51, 65, 85, 0.9);
  color: var(--text-primary);
}

.page-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.page-current {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  width: 600px;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
}

.modal-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.modal-close {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
  border-radius: 6px;
  transition: all 0.15s ease;
}

.modal-close:hover {
  background: rgba(51, 65, 85, 0.6);
  color: var(--text-primary);
}

.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item.full-width {
  grid-column: 1 / -1;
}

.detail-label {
  font-size: 12px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-value {
  font-size: 14px;
  color: var(--text-primary);
  word-break: break-all;
}

.detail-pre {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.08);
  padding: 12px;
  border-radius: 8px;
  font-family: monospace;
  font-size: 12px;
  color: var(--text-primary);
  max-height: 200px;
  overflow-y: auto;
  line-height: 1.6;
}
</style>
