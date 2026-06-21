<template>
  <div class="secrets-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">Secrets</h1>
        <span class="page-subtitle">管理集群中的密钥资源</span>
      </div>
      <div class="header-actions">
        <div class="ns-selector">
          <el-select v-model="namespace" placeholder="选择命名空间" clearable @change="fetchData">
            <el-option label="全部命名空间" value="" />
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </div>
        <button type="button" class="btn-gradient" @click="showCreate">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          新建 Secret
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchData">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="glass-table-container">
      <el-table
        :data="secrets"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 Secret 数据'"
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
        <el-table-column prop="type" label="类型" width="180">
          <template #default="{ row }">
            <span class="cell-type">{{ row.type }}</span>
          </template>
        </el-table-column>
        <el-table-column label="数据键" min-width="250">
          <template #default="{ row }">
            <el-tag v-for="key in (row.dataKeys || [])" :key="key" size="small" style="margin: 2px">{{ key }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="存活" width="80" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button type="button" class="action-btn action-detail" @click="viewDetail(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/></svg>
                查看
              </button>
              <button type="button" class="action-btn action-edit" @click="editItem(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                编辑
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

    <!-- 查看详情 -->
    <el-dialog v-model="detailVisible" title="Secret 详情" width="700px" class="dark-dialog">
      <template v-if="selected">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ selected.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ selected.namespace }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ selected.type }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ selected.age }}</el-descriptions-item>
        </el-descriptions>
        <template v-if="selected.labels && Object.keys(selected.labels).length > 0">
          <el-divider content-position="left">标签</el-divider>
          <div class="meta-tags">
            <el-tag v-for="(v, k) in selected.labels" :key="k" size="small" type="info" class="meta-tag">{{ k }}={{ v }}</el-tag>
          </div>
        </template>
        <template v-if="selected.annotations && Object.keys(selected.annotations).length > 0">
          <el-divider content-position="left">注解</el-divider>
          <div v-for="(v, k) in selected.annotations" :key="k" class="data-item">
            <div class="data-key">{{ k }}</div>
            <pre class="data-value">{{ v }}</pre>
          </div>
        </template>
        <el-divider content-position="left">数据键</el-divider>
        <div class="meta-tags">
          <el-tag v-for="key in (selected.dataKeys || [])" :key="key" size="small" type="info" class="meta-tag">{{ key }}</el-tag>
        </div>
      </template>
    </el-dialog>

    <!-- 创建/编辑 -->
    <el-dialog v-model="editVisible" :title="isEdit ? '编辑 Secret' : '新建 Secret'" width="750px" class="dark-dialog">
      <el-form :model="form" label-width="100px">
        <el-form-item label="命名空间">
          <el-select v-model="form.namespace" placeholder="选择命名空间" :disabled="isEdit" style="width: 100%">
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="form.name" :disabled="isEdit" placeholder="my-secret" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" :disabled="isEdit" style="width: 100%">
            <el-option label="Opaque" value="Opaque" />
            <el-option label="kubernetes.io/tls" value="kubernetes.io/tls" />
            <el-option label="kubernetes.io/dockerconfigjson" value="kubernetes.io/dockerconfigjson" />
            <el-option label="kubernetes.io/basic-auth" value="kubernetes.io/basic-auth" />
            <el-option label="kubernetes.io/ssh-auth" value="kubernetes.io/ssh-auth" />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">标签 (Labels)</el-divider>
        <div v-for="(item, idx) in form.labels" :key="'l'+idx" class="kv-row">
          <el-input v-model="item.key" placeholder="键" style="width: 200px" />
          <el-input v-model="item.value" placeholder="值" style="flex: 1" />
          <button type="button" class="kv-remove-btn" @click="form.labels.splice(idx, 1)">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <button type="button" class="btn-add-kv" @click="form.labels.push({ key: '', value: '' })">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加标签
        </button>

        <el-divider content-position="left">注解 (Annotations)</el-divider>
        <div v-for="(item, idx) in form.annotations" :key="'a'+idx" class="kv-row">
          <el-input v-model="item.key" placeholder="键" style="width: 200px" />
          <el-input v-model="item.value" placeholder="值" style="flex: 1" />
          <button type="button" class="kv-remove-btn" @click="form.annotations.splice(idx, 1)">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <button type="button" class="btn-add-kv" @click="form.annotations.push({ key: '', value: '' })">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加注解
        </button>

        <el-divider content-position="left">数据</el-divider>
        <div class="import-bar">
          <button type="button" class="btn-import" @click="triggerFileImport">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
            从文件导入
          </button>
          <input ref="fileInput" type="file" style="display:none" @change="handleFileImport" multiple />
        </div>
        <div v-for="(item, idx) in form.items" :key="idx" class="kv-row">
          <el-input v-model="item.key" placeholder="键" style="width: 200px" />
          <el-input v-model="item.value" type="textarea" :autosize="{ minRows: 1, maxRows: 8 }" placeholder="值 (明文，将自动 base64 编码)" style="flex: 1" />
          <button type="button" class="kv-remove-btn" @click="form.items.splice(idx, 1)" :disabled="form.items.length <= 1">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <button type="button" class="btn-add-kv" @click="form.items.push({ key: '', value: '' })">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加键值对
        </button>
      </el-form>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="editVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-confirm" @click="save" :disabled="saving">
          <span v-if="saving" class="btn-spinner"></span>
          {{ isEdit ? '保存' : '创建' }}
        </button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useGlobalNamespace } from '../store/namespace'
import { useHighlightRow } from '../composables/useHighlightRow'
import api from '../api'

useHighlightRow()
const secrets = ref<any[]>([])
useHighlightRow()
const nsList = ref<any[]>([])
const { namespace } = useGlobalNamespace()
const loading = ref(false)
const detailVisible = ref(false)
const selected = ref<any>(null)
const editVisible = ref(false)
const isEdit = ref(false)
const saving = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const form = reactive({
  namespace: '',
  name: '',
  type: 'Opaque',
  items: [{ key: '', value: '' }] as { key: string; value: string }[],
  labels: [] as { key: string; value: string }[],
  annotations: [] as { key: string; value: string }[],
})

const kvToMap = (arr: { key: string; value: string }[]) => {
  const m: Record<string, string> = {}
  for (const item of arr) {
    if (item.key) m[item.key] = item.value
  }
  return m
}

const mapToKv = (m?: Record<string, string>) => {
  if (!m || Object.keys(m).length === 0) return []
  return Object.entries(m).map(([k, v]) => ({ key: k, value: v }))
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = namespace.value ? { namespace: namespace.value } : {}
    const res: any = await api.get('/secrets', { params })
    secrets.value = res.data || []
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

const viewDetail = async (row: any) => {
  selected.value = row
  detailVisible.value = true
}

const showCreate = () => {
  isEdit.value = false
  form.namespace = namespace.value || ''
  form.name = ''
  form.type = 'Opaque'
  form.items = [{ key: '', value: '' }]
  form.labels = []
  form.annotations = []
  editVisible.value = true
}

const editItem = (row: any) => {
  isEdit.value = true
  form.namespace = row.namespace
  form.name = row.name
  form.type = row.type || 'Opaque'
  form.items = (row.dataKeys || []).map((k: string) => ({ key: k, value: '' }))
  if (form.items.length === 0) form.items = [{ key: '', value: '' }]
  form.labels = mapToKv(row.labels)
  form.annotations = mapToKv(row.annotations)
  editVisible.value = true
}

const triggerFileImport = () => {
  fileInput.value?.click()
}

const handleFileImport = (e: Event) => {
  const files = (e.target as HTMLInputElement).files
  if (!files) return
  for (const file of Array.from(files)) {
    const reader = new FileReader()
    reader.onload = (ev) => {
      const content = ev.target?.result as string
      const key = file.name
      const existing = form.items.find(i => i.key === key && !i.value)
      if (existing) {
        existing.value = content
      } else {
        const emptyIdx = form.items.findIndex(i => !i.key && !i.value)
        if (emptyIdx >= 0) {
          form.items[emptyIdx].key = key
          form.items[emptyIdx].value = content
        } else {
          form.items.push({ key, value: content })
        }
      }
    }
    reader.readAsText(file)
  }
  if (fileInput.value) fileInput.value.value = ''
  ElMessage.success(`已导入 ${files.length} 个文件`)
}

const save = async () => {
  if (!form.namespace || !form.name) {
    ElMessage.warning('请填写命名空间和名称')
    return
  }
  const data = kvToMap(form.items)
  const labels = kvToMap(form.labels)
  const annotations = kvToMap(form.annotations)

  saving.value = true
  try {
    if (isEdit.value) {
      await api.post('/secrets/update', { namespace: form.namespace, name: form.name, type: form.type, data, labels, annotations })
      ElMessage.success('更新成功')
    } else {
      await api.post('/secrets/create', { namespace: form.namespace, name: form.name, type: form.type, data, labels, annotations })
      ElMessage.success('创建成功')
    }
    editVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '操作失败')
  } finally {
    saving.value = false
  }
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Secret "${row.name}" 吗？`, '确认', { type: 'warning' })
    await api.delete(`/secrets/delete?namespace=${row.namespace}&name=${row.name}`)
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

onMounted(() => { fetchNs(); fetchData() })
</script>

<style scoped>
.secrets-page {
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

.cell-type {
  font-size: 12px;
  color: #c4b5fd;
  background: rgba(139, 92, 246, 0.12);
  padding: 2px 8px;
  border-radius: 4px;
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

.action-detail {
  background: rgba(59, 130, 246, 0.12);
  color: #60a5fa;
}

.action-detail:hover:not(:disabled) {
  background: rgba(59, 130, 246, 0.25);
}

.action-edit {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
}

.action-edit:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.25);
}

.action-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.action-delete:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
}

.data-item {
  margin-bottom: 12px;
}

.data-key {
  font-weight: 600;
  font-size: 13px;
  color: #818cf8;
  margin-bottom: 4px;
}

.data-value {
  background: rgba(15, 23, 42, 0.6);
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-family: monospace;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 200px;
  overflow-y: auto;
  margin: 0;
  color: var(--text-primary);
  border: 1px solid rgba(148, 163, 184, 0.08);
}

.kv-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 8px;
  width: 100%;
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

.secrets-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.secrets-page :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  padding: 20px 24px;
  margin: 0;
}

.secrets-page :deep(.el-dialog__title) {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 18px;
}

.secrets-page :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--text-secondary);
}

.secrets-page :deep(.el-dialog__body) {
  padding: 24px;
  color: var(--text-primary);
}

.secrets-page :deep(.el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.08);
  padding: 16px 24px;
}

.secrets-page :deep(.el-descriptions) {
  --el-descriptions-item-bordered-label-background: rgba(30, 41, 59, 0.8);
}

.secrets-page :deep(.el-descriptions__label) {
  color: var(--text-secondary);
}

.secrets-page :deep(.el-descriptions__content) {
  color: var(--text-primary);
}

.secrets-page :deep(.el-descriptions__cell) {
  border-color: rgba(148, 163, 184, 0.08) !important;
}

.secrets-page :deep(.el-divider__text) {
  background: #1e293b;
  color: var(--text-secondary);
}

.secrets-page :deep(.el-divider) {
  border-color: rgba(148, 163, 184, 0.08);
}

.secrets-page :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

.secrets-page :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.secrets-page :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.secrets-page :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.secrets-page :deep(.el-input__inner) {
  color: var(--text-primary);
}

.secrets-page :deep(.el-input__inner::placeholder) {
  color: var(--text-secondary);
}

.secrets-page :deep(.el-textarea__inner) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  color: var(--text-primary);
  box-shadow: none;
}

.secrets-page :deep(.el-textarea__inner:focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.secrets-page :deep(.el-tag--info) {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(148, 163, 184, 0.1);
  color: var(--text-secondary);
}

.secrets-page :deep(.el-empty__description p) {
  color: var(--text-secondary);
}

.secrets-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.btn-add-kv {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding: 6px 14px;
  background: rgba(99, 102, 241, 0.1);
  border: 1px dashed rgba(99, 102, 241, 0.3);
  border-radius: 6px;
  color: #818cf8;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-add-kv:hover {
  background: rgba(99, 102, 241, 0.2);
  border-color: rgba(99, 102, 241, 0.5);
}

.kv-remove-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  border-radius: 6px;
  color: #f87171;
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
}

.kv-remove-btn:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
}

.kv-remove-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.import-bar {
  margin-bottom: 12px;
}

.btn-import {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  background: rgba(16, 185, 129, 0.1);
  border: 1px solid rgba(16, 185, 129, 0.2);
  border-radius: 6px;
  color: #34d399;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-import:hover {
  background: rgba(16, 185, 129, 0.2);
  border-color: rgba(16, 185, 129, 0.4);
}

.meta-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.meta-tag {
  background: rgba(51, 65, 85, 0.6) !important;
  border-color: rgba(148, 163, 184, 0.1) !important;
  color: var(--text-secondary) !important;
}

.secrets-page :deep(.el-divider__text) {
  background: #1e293b;
  color: var(--text-secondary);
  font-size: 13px;
}
</style>
