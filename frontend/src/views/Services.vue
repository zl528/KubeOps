<template>
  <div class="services-page">
    <!-- 渐变页头 -->
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">Services</h1>
        <span class="page-subtitle">管理集群中的服务资源</span>
      </div>
      <div class="header-actions">
        <div class="ns-selector">
          <el-select v-model="namespace" placeholder="选择命名空间" clearable @change="fetchData">
            <el-option label="全部命名空间" value="" />
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </div>
        <button type="button" class="btn-gradient btn-refresh" @click="showCreate">
          <el-icon :size="16"><Plus /></el-icon>
          <span>新建 Service</span>
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchData">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <!-- 毛玻璃表格容器 -->
    <div class="glass-table-container">
      <el-table
        :data="services"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        class="custom-table"
        :empty-text="'暂无 Service 数据'"
      >
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="cell-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140">
          <template #default="{ row }">
            <span class="cell-ns">{{ row.namespace }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="130">
          <template #default="{ row }">
            <el-tag :type="row.type === 'ClusterIP' ? '' : row.type === 'NodePort' ? 'warning' : 'success'" size="small">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="clusterIP" label="ClusterIP" width="140" />
        <el-table-column label="端口" min-width="200">
          <template #default="{ row }">
            <div v-for="p in (row.ports || [])" :key="p.port" class="port-item">
              {{ p.port }}/{{ p.protocol }} → {{ p.targetPort }}{{ p.nodePort ? ' (NodePort: ' + p.nodePort + ')' : '' }}
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="存活" width="80" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button type="button" class="action-btn action-detail" @click="viewDetail(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/></svg>
                详情
              </button>
              <button type="button" class="action-btn action-edit" @click="editDialog(row)">
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

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="Service 详情" width="650px" class="dark-dialog">
      <template v-if="selected">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ selected.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ selected.namespace }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ selected.type }}</el-descriptions-item>
          <el-descriptions-item label="ClusterIP">{{ selected.clusterIP }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ selected.age }}</el-descriptions-item>
        </el-descriptions>
        <el-divider content-position="left">端口</el-divider>
        <el-table :data="selected.ports || []" size="small">
          <el-table-column prop="name" label="名称" />
          <el-table-column prop="port" label="端口" />
          <el-table-column prop="protocol" label="协议" />
          <el-table-column prop="targetPort" label="目标端口" />
          <el-table-column prop="nodePort" label="NodePort" />
        </el-table>
      </template>
    </el-dialog>

    <!-- 创建 Service -->
    <el-dialog v-model="createVisible" title="新建 Service" width="700px" class="dark-dialog">
      <el-form :model="form" label-width="100px">
        <el-form-item label="命名空间">
          <el-select v-model="form.namespace" placeholder="选择命名空间" style="width: 100%">
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="my-service" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width: 100%">
            <el-option label="ClusterIP" value="ClusterIP" />
            <el-option label="NodePort" value="NodePort" />
            <el-option label="LoadBalancer" value="LoadBalancer" />
          </el-select>
        </el-form-item>
        <el-divider content-position="left">端口</el-divider>
        <el-form-item label="端口">
          <div v-for="(port, idx) in form.ports" :key="idx" class="kv-row">
            <el-input v-model="port.name" placeholder="名称" style="width: 80px; margin-right: 8px" />
            <el-input v-model.number="port.port" placeholder="端口" style="width: 90px; margin-right: 8px" />
            <el-select v-model="port.protocol" style="width: 90px; margin-right: 8px">
              <el-option label="TCP" value="TCP" />
              <el-option label="UDP" value="UDP" />
            </el-select>
            <el-input v-model.number="port.targetPort" placeholder="目标端口" style="width: 100px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="form.ports.splice(idx, 1)" :disabled="form.ports.length <= 1">删除</button>
          </div>
          <button type="button" class="action-btn action-detail" @click="form.ports.push({ name: '', port: 80, protocol: 'TCP', targetPort: 80 })" style="margin-top: 8px">+ 添加端口</button>
        </el-form-item>
        <el-divider content-position="left">Selector</el-divider>
        <el-form-item label="Selector">
          <div v-for="(sel, idx) in form.selectors" :key="idx" class="kv-row">
            <el-input v-model="sel.key" placeholder="标签键" style="width: 180px; margin-right: 8px" />
            <el-input v-model="sel.value" placeholder="标签值" style="width: 180px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="form.selectors.splice(idx, 1)" :disabled="form.selectors.length <= 1">删除</button>
          </div>
          <button type="button" class="action-btn action-detail" @click="form.selectors.push({ key: '', value: '' })" style="margin-top: 8px">+ 添加 Selector</button>
        </el-form-item>
        <el-divider content-position="left">标签 & 注解</el-divider>
        <el-form-item label="Labels">
          <div v-for="(label, idx) in form.labels" :key="idx" class="kv-row">
            <el-input v-model="label.key" placeholder="键" style="width: 180px; margin-right: 8px" />
            <el-input v-model="label.value" placeholder="值" style="width: 180px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="form.labels.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="action-btn action-detail" @click="form.labels.push({ key: '', value: '' })" style="margin-top: 8px">+ 添加标签</button>
        </el-form-item>
        <el-form-item label="Annotations">
          <div v-for="(ann, idx) in form.annotations" :key="idx" class="kv-row">
            <el-input v-model="ann.key" placeholder="键" style="width: 180px; margin-right: 8px" />
            <el-input v-model="ann.value" placeholder="值" style="width: 180px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="form.annotations.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="action-btn action-detail" @click="form.annotations.push({ key: '', value: '' })" style="margin-top: 8px">+ 添加注解</button>
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

    <!-- 编辑 Service -->
    <el-dialog v-model="editVisible" title="编辑 Service" width="700px" class="dark-dialog">
      <el-form :model="editForm" label-width="100px">
        <el-form-item label="命名空间">
          <el-input :model-value="editForm.namespace" disabled />
        </el-form-item>
        <el-form-item label="名称">
          <el-input :model-value="editForm.name" disabled />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="editForm.type" style="width: 100%">
            <el-option label="ClusterIP" value="ClusterIP" />
            <el-option label="NodePort" value="NodePort" />
            <el-option label="LoadBalancer" value="LoadBalancer" />
          </el-select>
        </el-form-item>
        <el-divider content-position="left">端口</el-divider>
        <el-form-item label="端口">
          <div v-for="(port, idx) in editForm.ports" :key="idx" class="kv-row">
            <el-input v-model="port.name" placeholder="名称" style="width: 80px; margin-right: 8px" />
            <el-input v-model.number="port.port" placeholder="端口" style="width: 90px; margin-right: 8px" />
            <el-select v-model="port.protocol" style="width: 90px; margin-right: 8px">
              <el-option label="TCP" value="TCP" />
              <el-option label="UDP" value="UDP" />
            </el-select>
            <el-input v-model.number="port.targetPort" placeholder="目标端口" style="width: 100px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="editForm.ports.splice(idx, 1)" :disabled="editForm.ports.length <= 1">删除</button>
          </div>
          <button type="button" class="action-btn action-detail" @click="editForm.ports.push({ name: '', port: 80, protocol: 'TCP', targetPort: 80 })" style="margin-top: 8px">+ 添加端口</button>
        </el-form-item>
        <el-divider content-position="left">Selector</el-divider>
        <el-form-item label="Selector">
          <div v-for="(sel, idx) in editForm.selectors" :key="idx" class="kv-row">
            <el-input v-model="sel.key" placeholder="标签键" style="width: 180px; margin-right: 8px" />
            <el-input v-model="sel.value" placeholder="标签值" style="width: 180px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="editForm.selectors.splice(idx, 1)" :disabled="editForm.selectors.length <= 1">删除</button>
          </div>
          <button type="button" class="action-btn action-detail" @click="editForm.selectors.push({ key: '', value: '' })" style="margin-top: 8px">+ 添加 Selector</button>
        </el-form-item>
        <el-divider content-position="left">标签 & 注解</el-divider>
        <el-form-item label="Labels">
          <div v-for="(label, idx) in editForm.labels" :key="idx" class="kv-row">
            <el-input v-model="label.key" placeholder="键" style="width: 180px; margin-right: 8px" />
            <el-input v-model="label.value" placeholder="值" style="width: 180px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="editForm.labels.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="action-btn action-detail" @click="editForm.labels.push({ key: '', value: '' })" style="margin-top: 8px">+ 添加标签</button>
        </el-form-item>
        <el-form-item label="Annotations">
          <div v-for="(ann, idx) in editForm.annotations" :key="idx" class="kv-row">
            <el-input v-model="ann.key" placeholder="键" style="width: 180px; margin-right: 8px" />
            <el-input v-model="ann.value" placeholder="值" style="width: 180px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="editForm.annotations.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="action-btn action-detail" @click="editForm.annotations.push({ key: '', value: '' })" style="margin-top: 8px">+ 添加注解</button>
        </el-form-item>
      </el-form>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="editVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-confirm" @click="handleEdit" :disabled="editLoading">
          <span v-if="editLoading" class="btn-spinner"></span>
          保存
        </button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useGlobalNamespace } from '../store/namespace'
import { useHighlightRow } from '../composables/useHighlightRow'
import api from '../api'

useHighlightRow()
const services = ref<any[]>([])
useHighlightRow()
const nsList = ref<any[]>([])
const { namespace } = useGlobalNamespace()
const loading = ref(false)
const detailVisible = ref(false)
const selected = ref<any>(null)
const createVisible = ref(false)
const saving = ref(false)
const editVisible = ref(false)
const editLoading = ref(false)
const editForm = reactive({
  namespace: '',
  name: '',
  type: 'ClusterIP',
  ports: [{ name: '', port: 80, protocol: 'TCP', targetPort: 80 }] as any[],
  selectors: [{ key: '', value: '' }] as { key: string; value: string }[],
  labels: [] as { key: string; value: string }[],
  annotations: [] as { key: string; value: string }[],
})

const form = reactive({
  namespace: '',
  name: '',
  type: 'ClusterIP',
  ports: [{ name: '', port: 80, protocol: 'TCP', targetPort: 80 }] as any[],
  selectors: [{ key: '', value: '' }] as { key: string; value: string }[],
  labels: [] as { key: string; value: string }[],
  annotations: [] as { key: string; value: string }[],
})

const fetchData = async () => {
  loading.value = true
  try {
    const params = namespace.value ? { namespace: namespace.value } : {}
    const res: any = await api.get('/services', { params })
    services.value = res.data || []
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

const viewDetail = (row: any) => {
  selected.value = row
  detailVisible.value = true
}

const showCreate = () => {
  form.namespace = namespace.value || ''
  form.name = ''
  form.type = 'ClusterIP'
  form.ports = [{ name: '', port: 80, protocol: 'TCP', targetPort: 80 }]
  form.selectors = [{ key: '', value: '' }]
  form.labels = []
  form.annotations = []
  createVisible.value = true
}

const handleCreate = async () => {
  if (!form.namespace || !form.name) {
    ElMessage.warning('请填写命名空间和名称')
    return
  }

  const selector: Record<string, string> = {}
  for (const s of form.selectors) {
    if (s.key) selector[s.key] = s.value
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
    await api.post('/services/create', {
      namespace: form.namespace,
      name: form.name,
      type: form.type,
      ports: form.ports,
      selector,
      labels: Object.keys(labels).length > 0 ? labels : undefined,
      annotations: Object.keys(annotations).length > 0 ? annotations : undefined,
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
  try {
    await ElMessageBox.confirm(`确定要删除 Service "${row.name}" 吗？`, '确认', { type: 'warning' })
    await api.delete(`/services/delete?namespace=${row.namespace}&name=${row.name}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const editDialog = async (row: any) => {
  try {
    const res: any = await api.get(`/services/get`, { params: { namespace: row.namespace, name: row.name } })
    const data = res.data || row
    editForm.namespace = data.namespace
    editForm.name = data.name
    editForm.type = data.type || 'ClusterIP'
    editForm.ports = (data.ports || []).map((p: any) => ({
      name: p.name || '',
      port: p.port,
      protocol: p.protocol || 'TCP',
      targetPort: p.targetPort,
    }))
    editForm.selectors = data.selector ? Object.entries(data.selector).map(([k, v]) => ({ key: k, value: String(v) })) : [{ key: '', value: '' }]
    editForm.labels = data.labels ? Object.entries(data.labels).map(([k, v]) => ({ key: k, value: String(v) })) : []
    editForm.annotations = data.annotations ? Object.entries(data.annotations).map(([k, v]) => ({ key: k, value: String(v) })) : []
  } catch (e) {
    editForm.namespace = row.namespace
    editForm.name = row.name
    editForm.type = row.type || 'ClusterIP'
    editForm.ports = (row.ports || []).map((p: any) => ({
      name: p.name || '',
      port: p.port,
      protocol: p.protocol || 'TCP',
      targetPort: p.targetPort,
    }))
    editForm.selectors = row.selector ? Object.entries(row.selector).map(([k, v]) => ({ key: k, value: String(v) })) : [{ key: '', value: '' }]
    editForm.labels = row.labels ? Object.entries(row.labels).map(([k, v]) => ({ key: k, value: String(v) })) : []
    editForm.annotations = row.annotations ? Object.entries(row.annotations).map(([k, v]) => ({ key: k, value: String(v) })) : []
  }
  editVisible.value = true
}

const handleEdit = async () => {
  editLoading.value = true
  try {
    const selector: Record<string, string> = {}
    editForm.selectors.forEach(s => { if (s.key) selector[s.key] = s.value })
    const labels: Record<string, string> = {}
    editForm.labels.forEach(l => { if (l.key) labels[l.key] = l.value })
    const annotations: Record<string, string> = {}
    editForm.annotations.forEach(a => { if (a.key) annotations[a.key] = a.value })
    const payload: any = {
      namespace: editForm.namespace,
      name: editForm.name,
      type: editForm.type,
      ports: editForm.ports.filter(p => p.port > 0),
      selector: Object.keys(selector).length > 0 ? selector : undefined,
      labels: Object.keys(labels).length > 0 ? labels : undefined,
      annotations: Object.keys(annotations).length > 0 ? annotations : undefined,
    }
    await api.post('/services/update', payload)
    ElMessage.success('更新成功')
    editVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error('更新失败')
  } finally {
    editLoading.value = false
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

onMounted(() => { fetchNs(); fetchData() })
</script>

<style scoped>
.services-page {
  padding: 24px;
  background: var(--bg-primary);
  min-height: 100vh;
}

/* 渐变页头 */
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

/* 渐变按钮 */
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

/* 毛玻璃表格容器 */
.glass-table-container {
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
}

/* 自定义表格 */
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

/* 表格单元格样式 */
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

.port-item {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.6;
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

.action-edit {
  background: rgba(99, 102, 241, 0.12);
  color: #818cf8;
}

.action-edit:hover:not(:disabled) {
  background: rgba(99, 102, 241, 0.25);
}

.action-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.action-delete:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
}

/* 加载旋转 */
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

/* 弹窗按钮 */
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

/* 表单行 */
.kv-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

/* 暗色弹窗覆盖 */
.services-page :deep(.dark-dialog .el-dialog),
.services-page :deep(.el-dialog.dark-dialog) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.services-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.services-page :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  padding: 20px 24px;
  margin: 0;
}

.services-page :deep(.el-dialog__title) {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 18px;
}

.services-page :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--text-secondary);
}

.services-page :deep(.el-dialog__body) {
  padding: 24px;
  color: var(--text-primary);
}

.services-page :deep(.el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.08);
  padding: 16px 24px;
}

.services-page :deep(.el-descriptions) {
  --el-descriptions-item-bordered-label-background: rgba(30, 41, 59, 0.8);
}

.services-page :deep(.el-descriptions__label) {
  color: var(--text-secondary);
}

.services-page :deep(.el-descriptions__content) {
  color: var(--text-primary);
}

.services-page :deep(.el-descriptions__cell) {
  border-color: rgba(148, 163, 184, 0.08) !important;
}

.services-page :deep(.el-divider__text) {
  background: #1e293b;
  color: var(--text-secondary);
}

.services-page :deep(.el-divider) {
  border-color: rgba(148, 163, 184, 0.08);
}

.services-page :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

.services-page :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.services-page :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.services-page :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.services-page :deep(.el-input__inner) {
  color: var(--text-primary);
}

.services-page :deep(.el-input__inner::placeholder) {
  color: var(--text-secondary);
}

.services-page :deep(.el-tag--info) {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(148, 163, 184, 0.1);
  color: var(--text-secondary);
}

/* Loading overlay */
.glass-table-container :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(4px);
}

.glass-table-container :deep(.el-loading-spinner .circular) {
  stroke: var(--primary);
}

/* 弹窗遮罩 */
.services-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}
</style>