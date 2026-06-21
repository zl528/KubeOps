<template>
  <div class="jobs-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">Jobs</h1>
        <span class="page-subtitle">管理集群中的一次性任务</span>
      </div>
      <div class="header-actions">
        <div class="ns-selector">
          <el-select v-model="namespace" placeholder="选择命名空间" clearable @change="fetchData">
            <el-option label="全部命名空间" value="" />
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </div>
        <button type="button" class="btn-gradient btn-create" @click="showCreate">
          <el-icon :size="16"><Plus /></el-icon>
          <span>新建</span>
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchData">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="glass-table-container">
      <el-table
        :data="jobs"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 Job 数据'"
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
        <el-table-column prop="completions" label="完成数" width="120">
          <template #default="{ row }">
            <span class="cell-metric">{{ row.completions }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <span class="cell-status" :class="statusClass(row.status)">{{ row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="存活时间" width="100" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button type="button" class="action-btn action-detail" @click="viewDetail(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
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

    <!-- 创建 Job 弹窗 -->
    <el-dialog v-model="createDialogVisible" title="新建 Job" width="700px" :close-on-click-modal="false" class="jobs-dialog">
      <el-form :model="createForm" label-width="120px" label-position="left">
        <el-form-item label="命名空间">
          <el-select v-model="createForm.namespace" placeholder="选择命名空间" style="width: 100%">
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="createForm.name" placeholder="my-job" />
        </el-form-item>
        <el-form-item label="镜像">
          <el-input v-model="createForm.image" placeholder="busybox:latest" />
        </el-form-item>
        <el-form-item label="命令">
          <el-input v-model="createForm.command" placeholder="echo hello" />
        </el-form-item>
        <el-divider content-position="left">标签</el-divider>
        <el-form-item label="Labels">
          <div v-for="(label, idx) in createForm.labels" :key="idx" class="kv-row">
            <el-input v-model="label.key" placeholder="键" style="width: 180px; margin-right: 8px" />
            <el-input v-model="label.value" placeholder="值" style="width: 180px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="createForm.labels.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="action-btn action-detail" @click="createForm.labels.push({ key: '', value: '' })" style="margin-top: 8px">+ 添加标签</button>
        </el-form-item>
        <el-divider content-position="left">环境变量</el-divider>
        <el-form-item label="Env Vars">
          <div v-for="(env, idx) in createForm.envVars" :key="idx" class="kv-row">
            <el-input v-model="env.key" placeholder="变量名" style="width: 180px; margin-right: 8px" />
            <el-input v-model="env.value" placeholder="变量值" style="width: 180px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="createForm.envVars.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="action-btn action-detail" @click="createForm.envVars.push({ key: '', value: '' })" style="margin-top: 8px">+ 添加环境变量</button>
        </el-form-item>
        <el-divider content-position="left">资源配额</el-divider>
        <el-form-item label="CPU 请求">
          <el-input v-model="createForm.resourceRequestsCpu" placeholder="100m" />
        </el-form-item>
        <el-form-item label="内存请求">
          <el-input v-model="createForm.resourceRequestsMem" placeholder="128Mi" />
        </el-form-item>
        <el-form-item label="CPU 限制">
          <el-input v-model="createForm.resourceLimitsCpu" placeholder="500m" />
        </el-form-item>
        <el-form-item label="内存限制">
          <el-input v-model="createForm.resourceLimitsMem" placeholder="256Mi" />
        </el-form-item>
        <el-divider content-position="left">任务配置</el-divider>
        <el-form-item label="完成数">
          <el-input-number v-model="createForm.completions" :min="1" :max="1000" style="width: 100%" />
        </el-form-item>
        <el-form-item label="并行数">
          <el-input-number v-model="createForm.parallelism" :min="1" :max="1000" style="width: 100%" />
        </el-form-item>
        <el-form-item label="失败重试">
          <el-input-number v-model="createForm.backoffLimit" :min="0" :max="100" style="width: 100%" />
        </el-form-item>
        <el-form-item label="重启策略">
          <el-select v-model="createForm.restartPolicy" style="width: 100%">
            <el-option label="Never" value="Never" />
            <el-option label="OnFailure" value="OnFailure" />
          </el-select>
        </el-form-item>
        <el-form-item label="Active Deadline">
          <el-input-number v-model="createForm.activeDeadlineSeconds" :min="0" :max="86400" style="width: 100%" placeholder="秒" />
        </el-form-item>
        <el-form-item label="TTL After Finish">
          <el-input-number v-model="createForm.ttlSecondsAfterFinished" :min="0" :max="604800" style="width: 100%" placeholder="秒" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="Job 详情" width="700px" :close-on-click-modal="false" class="dark-dialog">
      <div v-if="detailLoading" class="detail-loading">加载中...</div>
      <div v-else-if="detailData" class="detail-content">
        <div class="detail-section">
          <h4 class="detail-section-title">基本信息</h4>
          <div class="detail-grid">
            <div class="detail-item"><span class="detail-label">名称</span><span class="detail-value">{{ detailData.name }}</span></div>
            <div class="detail-item"><span class="detail-label">命名空间</span><span class="detail-value">{{ detailData.namespace }}</span></div>
            <div class="detail-item"><span class="detail-label">状态</span><span class="detail-value">{{ detailData.status }}</span></div>
            <div class="detail-item"><span class="detail-label">完成数</span><span class="detail-value">{{ detailData.completions }}</span></div>
            <div class="detail-item"><span class="detail-label">并行数</span><span class="detail-value">{{ detailData.parallelism }}</span></div>
            <div class="detail-item"><span class="detail-label">失败重试</span><span class="detail-value">{{ detailData.backoffLimit }}</span></div>
            <div class="detail-item"><span class="detail-label">成功 Pods</span><span class="detail-value">{{ detailData.succeeded || 0 }}</span></div>
            <div class="detail-item"><span class="detail-label">失败 Pods</span><span class="detail-value">{{ detailData.failed || 0 }}</span></div>
            <div class="detail-item"><span class="detail-label">活跃 Pods</span><span class="detail-value">{{ detailData.active || 0 }}</span></div>
            <div class="detail-item"><span class="detail-label">创建时间</span><span class="detail-value">{{ detailData.creationTimestamp || detailData.age }}</span></div>
          </div>
        </div>
        <div class="detail-section" v-if="detailData.labels">
          <h4 class="detail-section-title">标签</h4>
          <div class="labels-grid">
            <span v-for="(v, k) in detailData.labels" :key="k" class="label-tag">{{ k }}={{ v }}</span>
          </div>
        </div>
      </div>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="detailVisible = false">关闭</button>
      </template>
    </el-dialog>

    <!-- 编辑弹窗 -->
    <el-dialog v-model="editVisible" title="编辑 Job" width="600px" :close-on-click-modal="false" class="dark-dialog">
      <el-form label-width="100px" class="create-form">
        <el-form-item label="名称">
          <el-input :model-value="editForm.name" disabled />
        </el-form-item>
        <el-form-item label="命名空间">
          <el-input :model-value="editForm.namespace" disabled />
        </el-form-item>
        <el-form-item label="镜像">
          <el-input v-model="editForm.image" placeholder="busybox:latest" />
        </el-form-item>
        <el-form-item label="CPU 限制">
          <el-input v-model="editForm.cpuLimit" placeholder="例如 100m, 1" />
        </el-form-item>
        <el-form-item label="内存限制">
          <el-input v-model="editForm.memoryLimit" placeholder="例如 128Mi, 1Gi" />
        </el-form-item>
        <el-form-item label="CPU 请求">
          <el-input v-model="editForm.cpuRequest" placeholder="例如 50m, 0.5" />
        </el-form-item>
        <el-form-item label="内存请求">
          <el-input v-model="editForm.memoryRequest" placeholder="例如 64Mi, 512Mi" />
        </el-form-item>
        <el-form-item label="标签">
          <div v-for="(l, idx) in editForm.labels" :key="idx" class="env-var-row">
            <el-input v-model="l.key" placeholder="键" class="env-input" />
            <el-input v-model="l.value" placeholder="值" class="env-input" />
            <button type="button" class="env-remove-btn" @click="editForm.labels.splice(idx, 1)">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          <button type="button" class="env-add-btn" @click="editForm.labels.push({ key: '', value: '' })">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            添加标签
          </button>
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
import { ref, onMounted } from 'vue'
import { Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useGlobalNamespace } from '../store/namespace'
import api from '../api'

const jobs = ref<any[]>([])
const nsList = ref<any[]>([])
const { namespace } = useGlobalNamespace()
const loading = ref(false)

const createDialogVisible = ref(false)
const createLoading = ref(false)
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailData = ref<any>(null)
const editVisible = ref(false)
const editLoading = ref(false)
const editForm = ref({
  namespace: '',
  name: '',
  image: '',
  cpuLimit: '',
  memoryLimit: '',
  cpuRequest: '',
  memoryRequest: '',
  labels: [] as { key: string; value: string }[],
})
const createForm = ref({
  namespace: '',
  name: '',
  image: '',
  command: '',
  completions: 1,
  parallelism: 1,
  backoffLimit: 6,
  restartPolicy: 'Never',
  activeDeadlineSeconds: undefined as number | undefined,
  ttlSecondsAfterFinished: undefined as number | undefined,
  labels: [] as Array<{ key: string; value: string }>,
  envVars: [] as Array<{ key: string; value: string }>,
  resourceRequestsCpu: '',
  resourceRequestsMem: '',
  resourceLimitsCpu: '',
  resourceLimitsMem: '',
})

const showCreate = () => {
  createForm.value = {
    namespace: namespace.value || '',
    name: '',
    image: '',
    command: '',
    completions: 1,
    parallelism: 1,
    backoffLimit: 6,
    restartPolicy: 'Never',
    activeDeadlineSeconds: undefined,
    ttlSecondsAfterFinished: undefined,
    labels: [],
    envVars: [],
    resourceRequestsCpu: '',
    resourceRequestsMem: '',
    resourceLimitsCpu: '',
    resourceLimitsMem: '',
  }
  createDialogVisible.value = true
}

const handleCreate = async () => {
  if (!createForm.value.namespace || !createForm.value.name || !createForm.value.image) {
    ElMessage.warning('请填写命名空间、名称和镜像')
    return
  }
  createLoading.value = true
  try {
    const labels: Record<string, string> = {}
    for (const l of createForm.value.labels) {
      if (l.key) labels[l.key] = l.value
    }
    const envVars: Record<string, string> = {}
    for (const e of createForm.value.envVars) {
      if (e.key) envVars[e.key] = e.value
    }
    const resourceRequests: Record<string, string> = {}
    if (createForm.value.resourceRequestsCpu) resourceRequests['cpu'] = createForm.value.resourceRequestsCpu
    if (createForm.value.resourceRequestsMem) resourceRequests['memory'] = createForm.value.resourceRequestsMem
    const resourceLimits: Record<string, string> = {}
    if (createForm.value.resourceLimitsCpu) resourceLimits['cpu'] = createForm.value.resourceLimitsCpu
    if (createForm.value.resourceLimitsMem) resourceLimits['memory'] = createForm.value.resourceLimitsMem
    await api.post('/jobs/create', {
      namespace: createForm.value.namespace,
      name: createForm.value.name,
      image: createForm.value.image,
      command: createForm.value.command,
      completions: createForm.value.completions,
      parallelism: createForm.value.parallelism,
      backoffLimit: createForm.value.backoffLimit,
      restartPolicy: createForm.value.restartPolicy,
      activeDeadlineSeconds: createForm.value.activeDeadlineSeconds || undefined,
      ttlSecondsAfterFinished: createForm.value.ttlSecondsAfterFinished || undefined,
      labels: Object.keys(labels).length > 0 ? labels : undefined,
      envVars: Object.keys(envVars).length > 0 ? envVars : undefined,
      resourceRequests: Object.keys(resourceRequests).length > 0 ? resourceRequests : undefined,
      resourceLimits: Object.keys(resourceLimits).length > 0 ? resourceLimits : undefined,
    })
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  } finally {
    createLoading.value = false
  }
}

const statusClass = (s: string) => {
  const map: Record<string, string> = { Complete: 'status-ok', Running: 'status-warning', Failed: 'status-danger' }
  return map[s] || 'status-info'
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = namespace.value ? { namespace: namespace.value } : {}
    const res: any = await api.get('/jobs', { params })
    jobs.value = res.data || []
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

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Job ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.delete(`/jobs/delete?namespace=${row.namespace}&name=${row.name}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const viewDetail = async (row: any) => {
  detailVisible.value = true
  detailLoading.value = true
  detailData.value = null
  try {
    const res: any = await api.get(`/jobs`, { params: { namespace: row.namespace } })
    const job = (res.data || []).find((j: any) => j.name === row.name)
    if (job) {
      detailData.value = {
        name: job.name,
        namespace: job.namespace,
        status: job.status,
        completions: job.completions,
        parallelism: job.parallelism,
        backoffLimit: job.backoffLimit,
        succeeded: job.succeeded,
        failed: job.failed,
        active: job.active,
        creationTimestamp: job.creationTimestamp || job.age,
        labels: job.labels || {},
      }
    } else {
      detailData.value = row
    }
  } catch (e) {
    detailData.value = row
  } finally {
    detailLoading.value = false
  }
}

const editDialog = async (row: any) => {
  try {
    const res: any = await api.get(`/jobs/get`, { params: { namespace: row.namespace, name: row.name } })
    const data = res.data || row
    editForm.value = {
      namespace: data.namespace,
      name: data.name,
      image: data.image || '',
      cpuLimit: data.resourceLimits?.cpu || '',
      memoryLimit: data.resourceLimits?.memory || '',
      cpuRequest: data.resourceRequests?.cpu || '',
      memoryRequest: data.resourceRequests?.memory || '',
      labels: data.labels ? Object.entries(data.labels).map(([k, v]) => ({ key: k, value: String(v) })) : [],
    }
  } catch (e) {
    editForm.value = {
      namespace: row.namespace,
      name: row.name,
      image: row.image || '',
      cpuLimit: '',
      memoryLimit: '',
      cpuRequest: '',
      memoryRequest: '',
      labels: [],
    }
  }
  editVisible.value = true
}

const handleEdit = async () => {
  editLoading.value = true
  try {
    const payload: any = {
      namespace: editForm.value.namespace,
      name: editForm.value.name,
    }
    if (editForm.value.image) payload.image = editForm.value.image
    const labels: Record<string, string> = {}
    editForm.value.labels.forEach(l => { if (l.key) labels[l.key] = l.value })
    if (Object.keys(labels).length > 0) payload.labels = labels
    const resourceLimits: Record<string, string> = {}
    const resourceRequests: Record<string, string> = {}
    if (editForm.value.cpuLimit) resourceLimits.cpu = editForm.value.cpuLimit
    if (editForm.value.memoryLimit) resourceLimits.memory = editForm.value.memoryLimit
    if (editForm.value.cpuRequest) resourceRequests.cpu = editForm.value.cpuRequest
    if (editForm.value.memoryRequest) resourceRequests.memory = editForm.value.memoryRequest
    if (Object.keys(resourceLimits).length > 0) payload.resourceLimits = resourceLimits
    if (Object.keys(resourceRequests).length > 0) payload.resourceRequests = resourceRequests
    await api.post('/jobs/update', payload)
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

const rowClassName = ({ rowIndex }: { rowIndex: number }) =>
  rowIndex % 2 === 0 ? 'row-even' : 'row-odd'

onMounted(() => { fetchNs(); fetchData() })
</script>

<style scoped>
.jobs-page {
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

.status-warning {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
}

.status-danger {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.status-info {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
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

.detail-loading {
  text-align: center;
  padding: 40px;
  color: var(--text-secondary);
}

.detail-content {
  max-height: 60vh;
  overflow-y: auto;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--primary);
  margin: 0 0 12px 0;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.detail-value {
  font-size: 14px;
  color: var(--text-primary);
  word-break: break-all;
}

.labels-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.label-tag {
  background: rgba(99, 102, 241, 0.15);
  color: #818cf8;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-family: monospace;
}

.kv-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.jobs-dialog :deep(.el-divider__text) {
  color: var(--text-secondary);
  background: #1e293b;
}

.jobs-dialog :deep(.el-divider) {
  border-color: rgba(148, 163, 184, 0.1);
}

.jobs-dialog :deep(.el-select .el-input__wrapper) {
  background: rgba(15, 23, 42, 0.6);
}

.jobs-dialog :deep(.el-input-number .el-input__wrapper) {
  background: rgba(15, 23, 42, 0.6);
}

.jobs-dialog :deep(.el-input-number__decrease),
.jobs-dialog :deep(.el-input-number__increase) {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(148, 163, 184, 0.1);
  color: var(--text-secondary);
}

.glass-table-container :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(4px);
}

.glass-table-container :deep(.el-loading-spinner .circular) {
  stroke: var(--primary);
}

.jobs-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.jobs-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.jobs-page :deep(.el-message-box) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
}

.jobs-page :deep(.el-message-box__title) {
  color: var(--text-primary);
}

.jobs-page :deep(.el-message-box__content) {
  color: var(--text-secondary);
}

.jobs-page :deep(.el-message-box__btns .el-button--primary) {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-color: #6366f1;
}

.jobs-dialog :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  padding: 16px 24px;
}

.jobs-dialog :deep(.el-dialog__title) {
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 600;
}

.jobs-dialog :deep(.el-dialog__body) {
  padding: 24px;
}

.jobs-dialog :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

.jobs-dialog :deep(.el-input__wrapper) {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.jobs-dialog :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.jobs-dialog :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.jobs-dialog :deep(.el-input__inner) {
  color: var(--text-primary);
}

.jobs-dialog :deep(.el-select .el-input__wrapper) {
  background: rgba(15, 23, 42, 0.6);
}

.jobs-page :deep(.el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.08);
  padding: 16px 24px;
}
</style>
