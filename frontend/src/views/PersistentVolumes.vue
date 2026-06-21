<template>
  <div class="pvs-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">PersistentVolumes</h1>
        <span class="page-subtitle">管理集群中的持久卷资源</span>
      </div>
      <div class="header-actions">
        <button type="button" class="btn-gradient" @click="openCreateDialog">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          创建
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchData">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="glass-table-container">
      <el-table
        :data="pvs"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 PV 数据'"
      >
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="cell-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="capacity" label="容量" width="120">
          <template #default="{ row }">
            <span class="cell-metric">{{ row.capacity }}</span>
          </template>
        </el-table-column>
        <el-table-column label="访问模式" width="150">
          <template #default="{ row }">
            <el-tag v-for="mode in (row.accessModes || [])" :key="mode" size="small" style="margin: 2px">{{ mode }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reclaimPolicy" label="回收策略" width="120" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="storageClass" label="存储类" width="150" />
        <el-table-column prop="claim" label="PVC" min-width="200">
          <template #default="{ row }">
            <span class="cell-name">{{ row.claim }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="存活时间" width="100" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button type="button" class="action-btn action-delete" @click="handleDelete(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                删除
              </button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- Create Dialog -->
    <el-dialog v-model="createVisible" title="创建 PersistentVolume" width="600px" :close-on-click-modal="false" class="dark-dialog">
      <div class="dialog-form">
        <div class="form-row">
          <label class="form-label">名称 <span class="required">*</span></label>
          <el-input v-model="createForm.name" placeholder="请输入 PV 名称" />
        </div>
        <div class="form-row">
          <label class="form-label">容量 <span class="required">*</span></label>
          <el-input v-model="createForm.capacity" placeholder="例如: 10Gi" />
        </div>
        <div class="form-row">
          <label class="form-label">访问模式 <span class="required">*</span></label>
          <el-select v-model="createForm.accessModes" multiple placeholder="选择访问模式" style="width: 100%">
            <el-option label="ReadWriteOnce" value="ReadWriteOnce" />
            <el-option label="ReadOnlyMany" value="ReadOnlyMany" />
            <el-option label="ReadWriteMany" value="ReadWriteMany" />
            <el-option label="ReadWriteOncePod" value="ReadWriteOncePod" />
          </el-select>
        </div>
        <div class="form-row">
          <label class="form-label">回收策略</label>
          <el-select v-model="createForm.reclaimPolicy" placeholder="选择回收策略" clearable style="width: 100%">
            <el-option label="Retain" value="Retain" />
            <el-option label="Delete" value="Delete" />
            <el-option label="Recycle" value="Recycle" />
          </el-select>
        </div>
        <div class="form-row">
          <label class="form-label">存储类</label>
          <el-select v-model="createForm.storageClass" placeholder="选择存储类（可选）" filterable clearable style="width: 100%">
            <el-option label="不指定" value="" />
            <el-option v-for="sc in storageClasses" :key="sc.name" :label="sc.name" :value="sc.name" />
          </el-select>
        </div>
        <div class="form-row">
          <label class="form-label">Host Path</label>
          <el-input v-model="createForm.hostPath" placeholder="例如: /mnt/data" />
        </div>
        <div class="form-row">
          <label class="form-label">卷模式</label>
          <el-select v-model="createForm.volumeMode" placeholder="选择卷模式" clearable style="width: 100%">
            <el-option label="Filesystem" value="Filesystem" />
            <el-option label="Block" value="Block" />
          </el-select>
        </div>
        <div class="form-row">
          <label class="form-label">挂载选项</label>
          <div v-for="(opt, idx) in createForm.mountOptions" :key="idx" class="kv-row">
            <el-input v-model="createForm.mountOptions[idx]" placeholder="例如: nfsvers=4.1" style="flex: 1" />
            <button type="button" class="action-btn action-delete" @click="createForm.mountOptions.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="btn-add-label" @click="createForm.mountOptions.push('')">+ 添加挂载选项</button>
        </div>
        <div class="form-row">
          <label class="form-label">标签</label>
          <div v-for="(label, idx) in createForm.labels" :key="idx" class="kv-row">
            <el-input v-model="label.key" placeholder="键" style="width: 180px" />
            <el-input v-model="label.value" placeholder="值" style="width: 180px" />
            <button type="button" class="action-btn action-delete" @click="createForm.labels.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="btn-add-label" @click="createForm.labels.push({ key: '', value: '' })">+ 添加标签</button>
        </div>
        <div class="form-row">
          <label class="form-label">注解</label>
          <div v-for="(ann, idx) in createForm.annotations" :key="idx" class="kv-row">
            <el-input v-model="ann.key" placeholder="键" style="width: 180px" />
            <el-input v-model="ann.value" placeholder="值" style="width: 180px" />
            <button type="button" class="action-btn action-delete" @click="createForm.annotations.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="btn-add-label" @click="createForm.annotations.push({ key: '', value: '' })">+ 添加注解</button>
        </div>
      </div>
      <template #footer>
        <button type="button" class="btn-cancel" @click="createVisible = false">取消</button>
        <button type="button" class="btn-gradient" @click="submitCreate" :disabled="createLoading">
          <span v-if="createLoading">创建中...</span>
          <span v-else>创建</span>
        </button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const pvs = ref<any[]>([])
const storageClasses = ref<any[]>([])
const loading = ref(false)

const createVisible = ref(false)
const createLoading = ref(false)
const createForm = ref({
  name: '',
  capacity: '',
  accessModes: [] as string[],
  reclaimPolicy: '',
  storageClass: '',
  hostPath: '',
  volumeMode: '',
  mountOptions: [] as string[],
  labels: [] as { key: string; value: string }[],
  annotations: [] as { key: string; value: string }[],
})

const statusType = (s: string) => {
  const map: Record<string, string> = { Bound: 'success', Available: 'info', Released: 'warning', Failed: 'danger' }
  return map[s] || 'info'
}

const fetchStorageClasses = async () => {
  try {
    const res: any = await api.get('/storageclasses')
    storageClasses.value = res.data || []
  } catch (e) { console.error(e) }
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await api.get('/persistentvolumes')
    pvs.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  createForm.value = { name: '', capacity: '', accessModes: [], reclaimPolicy: '', storageClass: '', hostPath: '', volumeMode: '', mountOptions: [], labels: [], annotations: [] }
  createVisible.value = true
}

const submitCreate = async () => {
  if (!createForm.value.name || !createForm.value.capacity || createForm.value.accessModes.length === 0) {
    ElMessage.warning('请填写名称、容量和访问模式')
    return
  }
  createLoading.value = true
  try {
    const labels: Record<string, string> = {}
    for (const l of createForm.value.labels) { if (l.key) labels[l.key] = l.value }
    const annotations: Record<string, string> = {}
    for (const a of createForm.value.annotations) { if (a.key) annotations[a.key] = a.value }
    await api.post('/persistentvolumes/create', {
      ...createForm.value,
      labels: Object.keys(labels).length ? labels : undefined,
      annotations: Object.keys(annotations).length ? annotations : undefined,
      mountOptions: createForm.value.mountOptions.filter(Boolean).length ? createForm.value.mountOptions.filter(Boolean) : undefined,
    })
    ElMessage.success('创建成功')
    createVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error('创建失败')
  } finally {
    createLoading.value = false
  }
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 PV ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.delete(`/persistentvolumes/delete?name=${row.name}`)
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

onMounted(() => { fetchData(); fetchStorageClasses() })
</script>

<style scoped>
.pvs-page {
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

.btn-gradient:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
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

.action-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.action-delete:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
}

.glass-table-container :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(4px);
}

.glass-table-container :deep(.el-loading-spinner .circular) {
  stroke: var(--primary);
}

:deep(.dark-dialog) {
  background: rgba(30, 41, 59, 0.95);
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 16px;
}

:deep(.dark-dialog .el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  padding: 20px 24px;
}

:deep(.dark-dialog .el-dialog__title) {
  color: #f1f5f9;
  font-weight: 600;
}

:deep(.dark-dialog .el-dialog__body) {
  padding: 24px;
}

:deep(.dark-dialog .el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.1);
  padding: 16px 24px;
}

:deep(.dark-dialog .el-input__wrapper) {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.15);
  box-shadow: none;
}

:deep(.dark-dialog .el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

:deep(.dark-dialog .el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

:deep(.dark-dialog .el-input__inner) {
  color: #f1f5f9;
}

:deep(.dark-dialog .el-select .el-select__wrapper) {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.15);
  box-shadow: none;
}

:deep(.dark-dialog .el-select .el-select__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

:deep(.dark-dialog .el-select .el-select__wrapper.is-focused) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.dialog-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-label {
  font-size: 12px;
  font-weight: 500;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.required {
  color: #f87171;
}

.btn-cancel {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
  border: none;
  padding: 10px 20px;
  border-radius: 8px;
  font-weight: 500;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-right: 8px;
}

.btn-cancel:hover {
  background: rgba(148, 163, 184, 0.2);
  color: #f1f5f9;
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
  margin-top: 4px;
}

.btn-add-label:hover {
  color: #8b5cf6;
}
</style>
