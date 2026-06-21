<template>
  <div class="page-container">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">RoleBindings</h1>
        <span class="page-subtitle">管理命名空间级别的角色绑定</span>
      </div>
      <div class="header-actions">
        <div class="ns-selector">
          <el-select v-model="namespace" placeholder="选择命名空间" clearable @change="fetchData">
            <el-option label="全部命名空间" value="" />
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </div>
        <button type="button" class="btn-gradient" @click="showCreate">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          <span>新建 RoleBinding</span>
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchData">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="glass-table-container">
      <el-table
        :data="bindings"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 RoleBinding 数据'"
      >
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="cell-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="150">
          <template #default="{ row }">
            <span class="cell-ns">{{ row.namespace }}</span>
          </template>
        </el-table-column>
        <el-table-column label="关联角色" width="200">
          <template #default="{ row }">
            <span class="cell-tag">{{ row.roleRef?.kind }}: {{ row.roleRef?.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="绑定对象" min-width="250">
          <template #default="{ row }">
            <div v-for="s in (row.subjects || [])" :key="s.name" class="subject-item">
              <span class="cell-tag">{{ s.kind }}: {{ s.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="存活时间" width="120">
          <template #default="{ row }">
            <span class="cell-metric">{{ row.age }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
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
    <el-dialog v-model="createVisible" title="新建 RoleBinding" width="650px" :close-on-click-modal="false" class="dark-dialog">
      <div class="dialog-form">
        <div class="form-row">
          <label class="form-label">命名空间</label>
          <el-select v-model="createForm.namespace" placeholder="选择命名空间" style="width: 100%">
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </div>
        <div class="form-row">
          <label class="form-label">名称</label>
          <el-input v-model="createForm.name" placeholder="my-rolebinding" />
        </div>
        <div class="form-row">
          <label class="form-label">角色引用类型</label>
          <el-select v-model="createForm.roleRefKind" style="width: 100%">
            <el-option label="Role" value="Role" />
            <el-option label="ClusterRole" value="ClusterRole" />
          </el-select>
        </div>
        <div class="form-row">
          <label class="form-label">角色引用名称</label>
          <el-input v-model="createForm.roleRefName" placeholder="role-name" />
        </div>
        <div class="kv-section">
          <div class="rules-header">
            <label class="form-label">标签 (Labels)</label>
            <button type="button" class="btn-gradient btn-small" @click="addCreateLabel">+ 添加标签</button>
          </div>
          <div v-for="(lbl, idx) in createForm.labels" :key="idx" class="kv-row">
            <el-input v-model="lbl.key" placeholder="键" class="kv-input" />
            <el-input v-model="lbl.value" placeholder="值" class="kv-input" />
            <button type="button" class="btn-remove" @click="removeCreateLabel(idx)">删除</button>
          </div>
        </div>
        <div class="subjects-section">
          <div class="rules-header">
            <label class="form-label">绑定对象 (Subjects)</label>
            <button type="button" class="btn-gradient btn-small" @click="addSubject">+ 添加对象</button>
          </div>
          <div v-for="(sub, idx) in createForm.subjects" :key="idx" class="rule-card">
            <div class="rule-card-header">
              <span class="rule-index">对象 {{ idx + 1 }}</span>
              <button type="button" class="btn-remove" @click="removeSubject(idx)">删除</button>
            </div>
            <div class="form-row">
              <label class="form-label">类型</label>
              <el-select v-model="sub.kind" style="width: 100%">
                <el-option label="User" value="User" />
                <el-option label="Group" value="Group" />
                <el-option label="ServiceAccount" value="ServiceAccount" />
              </el-select>
            </div>
            <div class="form-row">
              <label class="form-label">名称</label>
              <el-input v-model="sub.name" placeholder="subject-name" />
            </div>
            <div class="form-row" v-if="sub.kind === 'ServiceAccount'">
              <label class="form-label">命名空间</label>
              <el-input v-model="sub.namespace" placeholder="default" />
            </div>
          </div>
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
import { useGlobalNamespace } from '../store/namespace'
import api from '../api'

const bindings = ref<any[]>([])
const nsList = ref<any[]>([])
const { namespace } = useGlobalNamespace()
const loading = ref(false)

const createVisible = ref(false)
const createLoading = ref(false)
const createForm = ref({
  namespace: '',
  name: '',
  roleRefKind: 'Role',
  roleRefName: '',
  labels: [] as { key: string; value: string }[],
  subjects: [] as any[]
})

const fetchData = async () => {
  loading.value = true
  try {
    const params = namespace.value ? { namespace: namespace.value } : {}
    const res: any = await api.get('/rolebindings', { params })
    bindings.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const showCreate = () => {
  createForm.value = {
    namespace: namespace.value || '',
    name: '',
    roleRefKind: 'Role',
    roleRefName: '',
    labels: [],
    subjects: [{ kind: 'ServiceAccount', name: '', namespace: '' }]
  }
  createVisible.value = true
}

const addCreateLabel = () => {
  createForm.value.labels.push({ key: '', value: '' })
}

const removeCreateLabel = (idx: number) => {
  createForm.value.labels.splice(idx, 1)
}

const fetchNs = async () => {
  try {
    const res: any = await api.get('/namespaces')
    nsList.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 RoleBinding ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.delete(`/rolebindings/delete?namespace=${row.namespace}&name=${row.name}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const addSubject = () => {
  createForm.value.subjects.push({ kind: 'ServiceAccount', name: '', namespace: '' })
}

const removeSubject = (idx: number) => {
  createForm.value.subjects.splice(idx, 1)
}

const submitCreate = async () => {
  if (!createForm.value.namespace || !createForm.value.name || !createForm.value.roleRefName) {
    ElMessage.warning('请填写命名空间、名称和角色引用名称')
    return
  }
  createLoading.value = true
  try {
    const labels: Record<string, string> = {}
    for (const l of createForm.value.labels) {
      if (l.key) labels[l.key] = l.value
    }
    await api.post('/rolebindings/create', {
      namespace: createForm.value.namespace,
      name: createForm.value.name,
      labels,
      roleRef: {
        kind: createForm.value.roleRefKind,
        name: createForm.value.roleRefName
      },
      subjects: createForm.value.subjects.filter((s: any) => s.name)
    })
    ElMessage.success('创建成功')
    createVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  } finally {
    createLoading.value = false
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

onMounted(() => { fetchNs(); fetchData() })
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

.cell-tag {
  font-size: 12px;
  color: var(--text-secondary);
  background: rgba(51, 65, 85, 0.5);
  padding: 2px 8px;
  border-radius: 4px;
  display: inline-block;
}

.subject-item {
  margin: 2px 0;
}

.cell-metric {
  font-variant-numeric: tabular-nums;
  font-weight: 500;
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

.action-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.action-delete:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
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

.btn-gradient.btn-small {
  padding: 6px 14px;
  font-size: 12px;
}

.subjects-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rule-card {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 10px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rule-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.rule-index {
  font-size: 13px;
  font-weight: 600;
  color: #818cf8;
}

.btn-remove {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  border: none;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-remove:hover {
  background: rgba(239, 68, 68, 0.25);
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

.kv-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.kv-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.kv-input {
  flex: 1;
}
</style>
