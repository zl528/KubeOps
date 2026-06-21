<template>
  <div class="backup-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">Backup Management</h1>
        <span class="page-subtitle">管理集群资源备份与恢复</span>
      </div>
      <div class="header-actions">
        <button type="button" class="btn-action btn-create" @click="showCreateDialog">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          创建备份
        </button>
        <button type="button" class="btn-action btn-import" @click="showImportDialog">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          导入备份
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchBackups">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="glass-card">
      <div class="glass-table-container">
        <el-table :data="backups" v-loading="loading" :header-cell-style="headerCellStyle" :cell-style="cellStyle" :row-class-name="rowClassName" class="custom-table" :empty-text="'暂无备份数据'">
          <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="cell-name">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="namespace" label="命名空间" width="120">
            <template #default="{ row }">
              <span class="cell-ns">{{ row.namespace || '全部' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <span class="cell-status" :class="'status-' + row.status">{{ row.status }}</span>
            </template>
          </el-table-column>
          <el-table-column label="资源类型" min-width="200">
            <template #default="{ row }">
              <div class="tag-list">
                <span v-for="r in (row.resources || [])" :key="r" class="resource-tag">{{ r }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="size" label="大小" width="100">
            <template #default="{ row }">
              <span class="cell-metric">{{ formatSize(row.size) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180">
            <template #default="{ row }">
              <span class="cell-time">{{ formatTime(row.createdAt) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="280" fixed="right">
            <template #default="{ row }">
              <div class="action-cell">
                <button type="button" class="action-btn action-restore" @click="showRestoreDialog(row)" :disabled="row.status !== 'completed'">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
                  恢复
                </button>
                <button type="button" class="action-btn action-export" @click="exportBackup(row)">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
                  导出
                </button>
                <button type="button" class="action-btn action-view" @click="viewContent(row)">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/></svg>
                  查看
                </button>
                <button type="button" class="action-btn action-delete" @click="handleDelete(row)">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  删除
                </button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <div class="modal-overlay" v-if="createVisible" @click.self="createVisible = false">
      <div class="modal-content glass-card">
        <div class="modal-header">
          <h3 class="modal-title">创建备份</h3>
          <button type="button" class="modal-close" @click="createVisible = false">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">备份名称 <span class="required">*</span></label>
            <input v-model="createForm.name" class="form-input" placeholder="请输入备份名称" />
          </div>
          <div class="form-item">
            <label class="form-label">命名空间</label>
            <select v-model="createForm.namespace" class="form-select">
              <option value="">全部命名空间</option>
              <option v-for="ns in nsList" :key="ns.name" :value="ns.name">{{ ns.name }}</option>
            </select>
          </div>
          <div class="form-item full-width">
            <label class="form-label">描述</label>
            <textarea v-model="createForm.description" class="form-textarea" placeholder="备份描述"></textarea>
          </div>
          <div class="form-item full-width">
            <label class="form-label">资源类型</label>
            <div class="checkbox-group">
              <label v-for="r in availableResources" :key="r" class="checkbox-label">
                <input type="checkbox" :value="r" v-model="createForm.resources" />
                <span class="checkbox-text">{{ r }}</span>
              </label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn-dialog btn-cancel" @click="createVisible = false">取消</button>
          <button type="button" class="btn-dialog btn-confirm" @click="handleCreate" :disabled="creating">
            <span v-if="creating" class="btn-spinner"></span>
            创建
          </button>
        </div>
      </div>
    </div>

    <div class="modal-overlay" v-if="restoreVisible" @click.self="restoreVisible = false">
      <div class="modal-content glass-card">
        <div class="modal-header">
          <h3 class="modal-title">恢复备份</h3>
          <button type="button" class="modal-close" @click="restoreVisible = false">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="form-grid">
          <div class="form-item full-width">
            <label class="form-label">备份</label>
            <input class="form-input" :value="restoreForm.backupName" disabled />
          </div>
          <div class="form-item full-width">
            <label class="form-label">目标命名空间</label>
            <select v-model="restoreForm.namespace" class="form-select">
              <option value="">使用原始命名空间</option>
              <option v-for="ns in nsList" :key="ns.name" :value="ns.name">{{ ns.name }}</option>
            </select>
          </div>
          <div class="form-item full-width">
            <label class="form-label">资源类型</label>
            <select v-model="restoreForm.resources" multiple class="form-select form-select-multi">
              <option v-for="r in availableResources" :key="r" :value="r">{{ r }}</option>
            </select>
          </div>
          <div class="form-item full-width">
            <label class="form-label">选项</label>
            <div class="checkbox-group">
              <label class="checkbox-label">
                <input type="checkbox" v-model="restoreForm.overwrite" />
                <span class="checkbox-text">覆盖已有资源</span>
              </label>
              <label class="checkbox-label">
                <input type="checkbox" v-model="restoreForm.dryRun" />
                <span class="checkbox-text">模拟运行</span>
              </label>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn-dialog btn-cancel" @click="restoreVisible = false">取消</button>
          <button type="button" class="btn-dialog btn-confirm" @click="handleRestore" :disabled="restoring">
            <span v-if="restoring" class="btn-spinner"></span>
            恢复
          </button>
        </div>
      </div>
    </div>

    <div class="modal-overlay" v-if="importVisible" @click.self="importVisible = false">
      <div class="modal-content glass-card">
        <div class="modal-header">
          <h3 class="modal-title">导入备份</h3>
          <button type="button" class="modal-close" @click="importVisible = false">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="form-grid">
          <div class="form-item full-width">
            <label class="form-label">备份名称 <span class="required">*</span></label>
            <input v-model="importForm.name" class="form-input" placeholder="请输入备份名称" />
          </div>
          <div class="form-item full-width">
            <label class="form-label">备份文件 <span class="required">*</span></label>
            <div class="upload-area">
              <input ref="uploadInputRef" type="file" accept=".json" @change="handleFileChange" style="display: none" />
              <button type="button" class="btn-action btn-upload" @click="triggerUpload">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
                选择文件
              </button>
              <span v-if="importForm.file" class="file-name">{{ importForm.file.name }}</span>
              <span v-else class="file-placeholder">未选择文件</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn-dialog btn-cancel" @click="importVisible = false">取消</button>
          <button type="button" class="btn-dialog btn-confirm" @click="handleImport" :disabled="importing">
            <span v-if="importing" class="btn-spinner"></span>
            导入
          </button>
        </div>
      </div>
    </div>

    <div class="modal-overlay" v-if="contentVisible" @click.self="contentVisible = false">
      <div class="modal-content modal-lg glass-card">
        <div class="modal-header">
          <h3 class="modal-title">备份内容</h3>
          <button type="button" class="modal-close" @click="contentVisible = false">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <pre class="content-pre">{{ contentData }}</pre>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const backups = ref<any[]>([])
const nsList = ref<any[]>([])
const loading = ref(false)
const creating = ref(false)
const restoring = ref(false)
const importing = ref(false)

const createVisible = ref(false)
const restoreVisible = ref(false)
const importVisible = ref(false)
const contentVisible = ref(false)

const uploadInputRef = ref<HTMLInputElement>()

const createForm = ref({
  name: '',
  namespace: '',
  description: '',
  resources: ['deployments', 'services', 'configmaps', 'secrets'],
})

const restoreForm = ref({
  backupId: '',
  backupName: '',
  namespace: '',
  resources: [],
  overwrite: false,
  dryRun: false,
})

const importForm = ref({
  name: '',
  file: null as File | null,
})

const contentData = ref('')

const availableResources = [
  'deployments', 'statefulsets', 'daemonsets',
  'services', 'configmaps', 'secrets',
  'ingresses', 'roles', 'rolebindings',
  'serviceaccounts', 'persistentvolumeclaims',
  'cronjobs', 'jobs',
]

const statusType = (status: string) => {
  const map: Record<string, string> = {
    running: 'warning',
    completed: 'success',
    failed: 'danger',
  }
  return map[status] || 'info'
}

const formatSize = (bytes: number) => {
  if (!bytes) return '-'
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  while (bytes >= 1024 && i < units.length - 1) {
    bytes /= 1024
    i++
  }
  return `${bytes.toFixed(1)} ${units[i]}`
}

const formatTime = (ts: string) => {
  if (!ts) return '-'
  return new Date(ts).toLocaleString()
}

const fetchBackups = async () => {
  loading.value = true
  try {
    const res: any = await api.get('/backup/list')
    backups.value = res.data || []
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

const showCreateDialog = () => {
  createForm.value = {
    name: '',
    namespace: '',
    description: '',
    resources: ['deployments', 'services', 'configmaps', 'secrets'],
  }
  createVisible.value = true
}

const handleCreate = async () => {
  if (!createForm.value.name) {
    ElMessage.warning('请输入备份名称')
    return
  }

  creating.value = true
  try {
    await api.post('/backup/create', createForm.value)
    ElMessage.success('备份创建成功')
    createVisible.value = false
    fetchBackups()
  } catch (e) {
    ElMessage.error('创建失败')
  } finally {
    creating.value = false
  }
}

const showRestoreDialog = (row: any) => {
  restoreForm.value = {
    backupId: row.id,
    backupName: row.name,
    namespace: '',
    resources: [],
    overwrite: false,
    dryRun: false,
  }
  restoreVisible.value = true
}

const handleRestore = async () => {
  restoring.value = true
  try {
    const res: any = await api.post('/backup/restore', restoreForm.value)
    const result = res.data
    ElMessage.success(`恢复完成：成功 ${result.restored}，跳过 ${result.skipped}，失败 ${result.failed}`)
    restoreVisible.value = false
  } catch (e) {
    ElMessage.error('恢复失败')
  } finally {
    restoring.value = false
  }
}

const exportBackup = async (row: any) => {
  try {
    const res = await api.get(`/backup/export?id=${row.id}`, { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([res as any]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `backup-${row.name}.json`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (e) {
    ElMessage.error('导出失败')
  }
}

const viewContent = async (row: any) => {
  try {
    const res: any = await api.get(`/backup/content?id=${row.id}`)
    contentData.value = JSON.stringify(res.data, null, 2)
    contentVisible.value = true
  } catch (e) {
    ElMessage.error('获取内容失败')
  }
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除备份 ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.delete(`/backup/delete?id=${row.id}`)
    ElMessage.success('删除成功')
    fetchBackups()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const showImportDialog = () => {
  importForm.value = { name: '', file: null }
  importVisible.value = true
}

const triggerUpload = () => {
  uploadInputRef.value?.click()
}

const handleFileChange = (event: Event) => {
  const input = event.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    importForm.value.file = input.files[0]
  }
}

const handleImport = async () => {
  if (!importForm.value.name || !importForm.value.file) {
    ElMessage.warning('请填写完整信息')
    return
  }

  importing.value = true
  try {
    const reader = new FileReader()
    reader.onload = async (e) => {
      const data = e.target?.result
      if (data) {
        await api.post(`/backup/import?name=${importForm.value.name}`, data, {
          headers: { 'Content-Type': 'application/json' },
        })
        ElMessage.success('导入成功')
        importVisible.value = false
        fetchBackups()
      }
      importing.value = false
    }
    reader.readAsText(importForm.value.file)
  } catch (e) {
    ElMessage.error('导入失败')
    importing.value = false
  }
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
  fetchBackups()
  fetchNs()
})
</script>

<style scoped>
.backup-page {
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

.btn-action {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-create {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
  border: 1px solid rgba(34, 197, 94, 0.2);
}

.btn-create:hover {
  background: rgba(34, 197, 94, 0.25);
  transform: translateY(-1px);
}

.btn-import {
  background: rgba(59, 130, 246, 0.12);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.2);
}

.btn-import:hover {
  background: rgba(59, 130, 246, 0.25);
  transform: translateY(-1px);
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

.cell-status {
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
  text-transform: capitalize;
}

.status-running {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
}

.status-completed {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
}

.status-failed {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.resource-tag {
  font-size: 11px;
  color: var(--text-secondary);
  background: rgba(51, 65, 85, 0.5);
  padding: 2px 8px;
  border-radius: 4px;
}

.cell-metric {
  font-variant-numeric: tabular-nums;
  font-weight: 500;
}

.cell-time {
  font-size: 13px;
  color: var(--text-secondary);
}

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

.action-restore {
  background: rgba(99, 102, 241, 0.12);
  color: #818cf8;
}

.action-restore:hover:not(:disabled) {
  background: rgba(99, 102, 241, 0.25);
}

.action-export {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
}

.action-export:hover:not(:disabled) {
  background: rgba(34, 197, 94, 0.25);
}

.action-view {
  background: rgba(59, 130, 246, 0.12);
  color: #60a5fa;
}

.action-view:hover:not(:disabled) {
  background: rgba(59, 130, 246, 0.25);
}

.action-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.action-delete:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
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

.modal-content.modal-lg {
  width: 800px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
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

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-item.full-width {
  grid-column: 1 / -1;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.required {
  color: #f87171;
}

.form-input,
.form-select,
.form-textarea {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 8px;
  padding: 10px 14px;
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.form-input:focus,
.form-select:focus,
.form-textarea:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.form-input::placeholder,
.form-textarea::placeholder {
  color: var(--text-secondary);
}

.form-select option {
  background: #1e293b;
  color: var(--text-primary);
}

.form-select-multi {
  min-height: 120px;
}

.form-textarea {
  min-height: 80px;
  resize: vertical;
  font-family: inherit;
}

.checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-primary);
}

.checkbox-label input[type="checkbox"] {
  accent-color: #6366f1;
  width: 16px;
  height: 16px;
}

.checkbox-text {
  color: var(--text-primary);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid rgba(148, 163, 184, 0.08);
}

.btn-dialog {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-cancel {
  background: rgba(51, 65, 85, 0.6);
  color: var(--text-secondary);
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.btn-cancel:hover {
  background: rgba(51, 65, 85, 0.9);
  color: var(--text-primary);
}

.btn-confirm {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
}

.btn-confirm:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.4);
}

.btn-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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

.upload-area {
  display: flex;
  align-items: center;
  gap: 12px;
}

.btn-upload {
  background: rgba(99, 102, 241, 0.12);
  color: #818cf8;
  border: 1px solid rgba(99, 102, 241, 0.2);
  padding: 8px 14px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.15s ease;
}

.btn-upload:hover {
  background: rgba(99, 102, 241, 0.25);
}

.file-name {
  font-size: 13px;
  color: var(--text-primary);
}

.file-placeholder {
  font-size: 13px;
  color: var(--text-secondary);
}

.content-pre {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.08);
  padding: 16px;
  border-radius: 8px;
  font-family: monospace;
  font-size: 12px;
  color: var(--text-primary);
  max-height: 500px;
  overflow-y: auto;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
