<template>
  <div class="page-container">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">Namespaces</h1>
        <span class="page-subtitle">管理集群中的命名空间资源</span>
      </div>
      <div class="header-actions">
        <button type="button" class="btn-gradient btn-create" @click="showCreate">
          <el-icon :size="16"><Plus /></el-icon>
          <span>新建命名空间</span>
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchData">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="glass-table-container">
      <el-table
        :data="namespaces"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无命名空间数据'"
      >
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="cell-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <span class="cell-status" :class="row.status === 'Active' ? 'status-active' : 'status-inactive'">{{ row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="存活" width="100">
          <template #default="{ row }">
            <span class="cell-metric">{{ row.age }}</span>
          </template>
        </el-table-column>
        <el-table-column label="标签" min-width="250">
          <template #default="{ row }">
            <template v-if="row.labels && Object.keys(row.labels).length">
              <el-tag v-for="(v, k) in getTopLabels(row.labels)" :key="k" size="small" style="margin: 2px">
                {{ k }}={{ v }}
              </el-tag>
              <el-tag v-if="Object.keys(row.labels).length > 3" size="small" type="info" style="margin: 2px">
                +{{ Object.keys(row.labels).length - 3 }}
              </el-tag>
            </template>
            <span v-else class="cell-empty">-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button type="button" class="action-btn action-detail" @click="viewDetail(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/></svg>
                详情
              </button>
              <button type="button" class="action-btn action-delete" :disabled="isSystemNs(row.name)" @click="handleDelete(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                删除
              </button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="命名空间详情" width="600px" class="dark-dialog">
      <template v-if="selected">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ selected.name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="selected.status === 'Active' ? 'success' : 'info'">{{ selected.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ selected.age }}</el-descriptions-item>
        </el-descriptions>
        <template v-if="selected.labels && Object.keys(selected.labels).length">
          <el-divider content-position="left">标签</el-divider>
          <div class="label-list">
            <el-tag v-for="(v, k) in selected.labels" :key="k" size="small" type="info" style="margin: 2px">
              {{ k }}={{ v }}
            </el-tag>
          </div>
        </template>
      </template>
    </el-dialog>

    <!-- 创建命名空间弹窗 -->
    <el-dialog v-model="createVisible" title="新建命名空间" width="500px" class="dark-dialog">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="my-namespace" />
        </el-form-item>
        <el-form-item label="标签">
          <div v-for="(label, idx) in form.labels" :key="idx" class="kv-row">
            <el-input v-model="label.key" placeholder="键" style="width: 180px; margin-right: 8px" />
            <el-input v-model="label.value" placeholder="值" style="width: 180px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="form.labels.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="btn-add-label" @click="form.labels.push({ key: '', value: '' })">+ 添加标签</button>
        </el-form-item>
        <el-form-item label="注解">
          <div v-for="(ann, idx) in form.annotations" :key="idx" class="kv-row">
            <el-input v-model="ann.key" placeholder="键" style="width: 180px; margin-right: 8px" />
            <el-input v-model="ann.value" placeholder="值" style="width: 180px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="form.annotations.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="btn-add-label" @click="form.annotations.push({ key: '', value: '' })">+ 添加注解</button>
        </el-form-item>
      </el-form>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="createVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-confirm" @click="handleCreate" :disabled="saving">
          <span v-if="saving" class="btn-spinner"></span>
          创建
        </button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const SYSTEM_NS = ['kube-system', 'kube-public', 'kube-node-lease', 'default']

const namespaces = ref<any[]>([])
const loading = ref(false)
const detailVisible = ref(false)
const selected = ref<any>(null)
const createVisible = ref(false)
const saving = ref(false)

const form = reactive({
  name: '',
  labels: [{ key: '', value: '' }] as { key: string; value: string }[],
  annotations: [] as { key: string; value: string }[],
})

const isSystemNs = (name: string) => SYSTEM_NS.includes(name)

const getTopLabels = (labels: Record<string, string>) => {
  const entries = Object.entries(labels).slice(0, 3)
  return Object.fromEntries(entries)
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await api.get('/namespaces')
    namespaces.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const viewDetail = (row: any) => {
  selected.value = row
  detailVisible.value = true
}

const showCreate = () => {
  form.name = ''
  form.labels = [{ key: '', value: '' }]
  form.annotations = []
  createVisible.value = true
}

const handleCreate = async () => {
  if (!form.name) {
    ElMessage.warning('请输入命名空间名称')
    return
  }

  const labels: Record<string, string> = {}
  for (const l of form.labels) {
    if (l.key) labels[l.key] = l.value
  }
  const annotations: Record<string, string> = {}
  for (const a of form.annotations) {
    if (a.key) annotations[a.key] = a.value
  }

  saving.value = true
  try {
    await api.post('/namespaces/create', {
      name: form.name,
      labels,
      annotations: Object.keys(annotations).length ? annotations : undefined,
    })
    ElMessage.success('创建成功')
    createVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  } finally {
    saving.value = false
  }
}

const handleDelete = async (row: any) => {
  if (isSystemNs(row.name)) {
    ElMessage.warning('系统命名空间不能删除')
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定要删除命名空间 "${row.name}" 吗？该命名空间下的所有资源都将被删除！`,
      '确认删除',
      { type: 'error' }
    )
    await api.delete(`/namespaces/delete?name=${row.name}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
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

onMounted(fetchData)
</script>

<style scoped>
.page-container {
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

.cell-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.status-active {
  color: #4ade80;
  background: rgba(34, 197, 94, 0.12);
}

.status-inactive {
  color: #94a3b8;
  background: rgba(148, 163, 184, 0.12);
}

.cell-metric {
  font-variant-numeric: tabular-nums;
  font-weight: 500;
}

.cell-empty {
  color: var(--text-secondary);
}

.label-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.kv-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.btn-add-label {
  background: none;
  border: none;
  color: #6366f1;
  font-size: 13px;
  cursor: pointer;
  padding: 4px 0;
  margin-top: 8px;
}

.btn-add-label:hover {
  color: #8b5cf6;
}

.action-cell {
  display: flex;
  gap: 4px;
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

.page-container :deep(.dark-dialog .el-dialog),
.page-container :deep(.el-dialog.dark-dialog) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.page-container :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.page-container :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  padding: 20px 24px;
  margin: 0;
}

.page-container :deep(.el-dialog__title) {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 18px;
}

.page-container :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--text-secondary);
}

.page-container :deep(.el-dialog__body) {
  padding: 24px;
  color: var(--text-primary);
}

.page-container :deep(.el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.08);
  padding: 16px 24px;
}

.page-container :deep(.el-descriptions) {
  --el-descriptions-item-bordered-label-background: rgba(30, 41, 59, 0.8);
}

.page-container :deep(.el-descriptions__label) {
  color: var(--text-secondary);
}

.page-container :deep(.el-descriptions__content) {
  color: var(--text-primary);
}

.page-container :deep(.el-descriptions__cell) {
  border-color: rgba(148, 163, 184, 0.08) !important;
}

.page-container :deep(.el-divider__text) {
  background: #1e293b;
  color: var(--text-secondary);
}

.page-container :deep(.el-divider) {
  border-color: rgba(148, 163, 184, 0.08);
}

.page-container :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

.page-container :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.page-container :deep(.el-input__inner) {
  color: var(--text-primary);
}

.page-container :deep(.el-tag--info) {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(148, 163, 184, 0.1);
  color: var(--text-secondary);
}

.page-container :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}
</style>
