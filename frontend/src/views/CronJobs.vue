<template>
  <div class="cronjobs-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">CronJobs</h1>
        <span class="page-subtitle">管理集群中的定时任务</span>
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
        :data="cronjobs"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 CronJob 数据'"
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
        <el-table-column prop="schedule" label="调度规则" width="200">
          <template #default="{ row }">
            <span class="cell-schedule">{{ row.schedule }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="suspend" label="暂停" width="80">
          <template #default="{ row }">
            <span class="cell-status" :class="row.suspend ? 'status-warning' : 'status-ok'">{{ row.suspend ? '已暂停' : '运行中' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="active" label="活跃" width="80">
          <template #default="{ row }">
            <span class="cell-metric">{{ row.active }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="lastScheduleTime" label="上次调度" width="180" />
        <el-table-column prop="age" label="存活时间" width="100" />
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button type="button" class="action-btn action-edit" @click="handleEdit(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                编辑
              </button>
              <button type="button" class="action-btn" :class="row.suspend ? 'action-suspend' : 'action-resume'" @click="handleSuspend(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
                  <template v-if="row.suspend">
                    <polygon points="5 3 19 12 5 21 5 3"/>
                  </template>
                  <template v-else>
                    <rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/>
                  </template>
                </svg>
                {{ row.suspend ? '恢复' : '暂停' }}
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

    <el-dialog v-model="editDialogVisible" title="编辑 CronJob" width="600px" :close-on-click-modal="false" class="cronjobs-dialog">
      <el-form :model="editForm" label-width="120px" label-position="left">
        <el-form-item label="调度规则">
          <el-input v-model="editForm.schedule" placeholder="*/5 * * * *" />
        </el-form-item>
        <el-form-item label="暂停">
          <el-switch v-model="editForm.suspend" />
        </el-form-item>
        <el-form-item label="并发策略">
          <el-select v-model="editForm.concurrencyPolicy" placeholder="选择策略">
            <el-option label="Allow (允许)" value="Allow" />
            <el-option label="Forbid (禁止)" value="Forbid" />
            <el-option label="Replace (替换)" value="Replace" />
          </el-select>
        </el-form-item>
        <el-form-item label="完成数">
          <el-input-number v-model="editForm.completions" :min="1" :max="1000" />
        </el-form-item>
        <el-form-item label="并行数">
          <el-input-number v-model="editForm.parallelism" :min="1" :max="1000" />
        </el-form-item>
        <el-form-item label="失败重试">
          <el-input-number v-model="editForm.backoffLimit" :min="0" :max="100" />
        </el-form-item>
        <el-divider content-position="left">容器配置</el-divider>
        <template v-if="editForm.containers.length > 0">
          <div v-for="(container, idx) in editForm.containers" :key="idx" class="container-edit-item">
            <el-form-item :label="container.name">
              <el-input v-model="container.image" placeholder="镜像地址" />
            </el-form-item>
          </div>
        </template>
        <template v-else>
          <div style="color: var(--text-secondary); font-size: 13px; padding: 8px 0;">无容器配置</div>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editLoading" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 创建 CronJob 弹窗 -->
    <el-dialog v-model="createDialogVisible" title="新建 CronJob" width="700px" :close-on-click-modal="false" class="cronjobs-dialog">
      <el-form :model="createForm" label-width="140px" label-position="left">
        <el-form-item label="命名空间">
          <el-select v-model="createForm.namespace" placeholder="选择命名空间" style="width: 100%">
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="createForm.name" placeholder="my-cronjob" />
        </el-form-item>
        <el-form-item label="调度规则">
          <el-input v-model="createForm.schedule" placeholder="*/5 * * * *" />
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
        <el-divider content-position="left">调度选项</el-divider>
        <el-form-item label="并发策略">
          <el-select v-model="createForm.concurrencyPolicy" placeholder="选择策略" style="width: 100%">
            <el-option label="Allow (允许)" value="Allow" />
            <el-option label="Forbid (禁止)" value="Forbid" />
            <el-option label="Replace (替换)" value="Replace" />
          </el-select>
        </el-form-item>
        <el-form-item label="暂停">
          <el-switch v-model="createForm.suspend" />
        </el-form-item>
        <el-form-item label="重启策略">
          <el-select v-model="createForm.restartPolicy" style="width: 100%">
            <el-option label="OnFailure" value="OnFailure" />
            <el-option label="Never" value="Never" />
          </el-select>
        </el-form-item>
        <el-form-item label="成功历史限制">
          <el-input-number v-model="createForm.successfulJobsHistoryLimit" :min="0" :max="1000" style="width: 100%" />
        </el-form-item>
        <el-form-item label="失败历史限制">
          <el-input-number v-model="createForm.failedJobsHistoryLimit" :min="0" :max="1000" style="width: 100%" />
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Refresh, Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useGlobalNamespace } from '../store/namespace'
import api from '../api'

const cronjobs = ref<any[]>([])
const nsList = ref<any[]>([])
const { namespace } = useGlobalNamespace()
const loading = ref(false)
const editDialogVisible = ref(false)
const editLoading = ref(false)
const editForm = ref({
  namespace: '',
  name: '',
  schedule: '',
  suspend: false,
  concurrencyPolicy: 'Allow',
  completions: 1,
  parallelism: 1,
  backoffLimit: 6,
  containers: [] as Array<{ name: string; image: string }>
})

const createDialogVisible = ref(false)
const createLoading = ref(false)
const createForm = ref({
  namespace: '',
  name: '',
  schedule: '',
  image: '',
  command: '',
  concurrencyPolicy: 'Allow',
  suspend: false,
  successfulJobsHistoryLimit: 3,
  failedJobsHistoryLimit: 1,
  restartPolicy: 'OnFailure',
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
    schedule: '',
    image: '',
    command: '',
    concurrencyPolicy: 'Allow',
    suspend: false,
    successfulJobsHistoryLimit: 3,
    failedJobsHistoryLimit: 1,
    restartPolicy: 'OnFailure',
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
  if (!createForm.value.namespace || !createForm.value.name || !createForm.value.image || !createForm.value.schedule) {
    ElMessage.warning('请填写命名空间、名称、镜像和调度规则')
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
    const payload: any = {
      namespace: createForm.value.namespace,
      name: createForm.value.name,
      schedule: createForm.value.schedule,
      image: createForm.value.image,
      command: createForm.value.command,
      concurrencyPolicy: createForm.value.concurrencyPolicy,
      suspend: createForm.value.suspend,
      successfulJobsHistoryLimit: createForm.value.successfulJobsHistoryLimit,
      failedJobsHistoryLimit: createForm.value.failedJobsHistoryLimit,
      restartPolicy: createForm.value.restartPolicy,
      activeDeadlineSeconds: createForm.value.activeDeadlineSeconds || undefined,
      ttlSecondsAfterFinished: createForm.value.ttlSecondsAfterFinished || undefined,
      labels: Object.keys(labels).length > 0 ? labels : undefined,
      envVars: Object.keys(envVars).length > 0 ? envVars : undefined,
      resourceRequests: Object.keys(resourceRequests).length > 0 ? resourceRequests : undefined,
      resourceLimits: Object.keys(resourceLimits).length > 0 ? resourceLimits : undefined,
    }
    await api.post('/cronjobs/create', payload)
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  } finally {
    createLoading.value = false
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = namespace.value ? { namespace: namespace.value } : {}
    const res: any = await api.get('/cronjobs', { params })
    cronjobs.value = res.data || []
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

const handleSuspend = async (row: any) => {
  const action = row.suspend ? '恢复' : '暂停'
  try {
    await ElMessageBox.confirm(`确定要${action} CronJob ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.post(`/cronjobs/suspend?namespace=${row.namespace}&name=${row.name}&suspend=${!row.suspend}`)
    ElMessage.success(`${action}成功`)
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(`${action}失败`)
  }
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 CronJob ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.delete(`/cronjobs/delete?namespace=${row.namespace}&name=${row.name}`)
    ElMessage.success('删除成功')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const handleEdit = async (row: any) => {
  try {
    const res: any = await api.get(`/cronjobs/get?namespace=${row.namespace}&name=${row.name}`)
    const cj = res.data
    const spec = cj.spec || {}
    const jobSpec = spec.jobTemplate?.spec || {}
    const containers = (jobSpec.template?.spec?.containers || []).map((c: any) => ({
      name: c.name,
      image: c.image || ''
    }))
    editForm.value = {
      namespace: row.namespace,
      name: row.name,
      schedule: spec.schedule || '',
      suspend: spec.suspend || false,
      concurrencyPolicy: spec.concurrencyPolicy || 'Allow',
      completions: jobSpec.completions || 1,
      parallelism: jobSpec.parallelism || 1,
      backoffLimit: jobSpec.backoffLimit ?? 6,
      containers
    }
    editDialogVisible.value = true
  } catch (e) {
    ElMessage.error('获取 CronJob 详情失败')
  }
}

const submitEdit = async () => {
  editLoading.value = true
  try {
    const payload: any = {
      namespace: editForm.value.namespace,
      name: editForm.value.name,
      schedule: editForm.value.schedule,
      suspend: editForm.value.suspend,
      concurrencyPolicy: editForm.value.concurrencyPolicy,
      jobTemplateSpec: {
        completions: editForm.value.completions,
        parallelism: editForm.value.parallelism,
        backoffLimit: editForm.value.backoffLimit,
        template: {
          containers: editForm.value.containers.map(c => ({
            name: c.name,
            image: c.image
          }))
        }
      }
    }
    await api.put('/cronjobs/update', payload)
    ElMessage.success('更新成功')
    editDialogVisible.value = false
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
.cronjobs-page {
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

.cell-schedule {
  font-family: monospace;
  font-size: 13px;
  color: #e2e8f0;
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

.metric-ok {
  color: var(--success);
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

.action-suspend {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
}

.action-suspend:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.25);
}

.action-resume {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
}

.action-resume:hover:not(:disabled) {
  background: rgba(34, 197, 94, 0.25);
}

.action-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.action-delete:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
}

.action-edit {
  background: rgba(99, 102, 241, 0.12);
  color: #818cf8;
}

.action-edit:hover:not(:disabled) {
  background: rgba(99, 102, 241, 0.25);
}

.container-edit-item {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 8px;
  padding: 8px 12px;
  margin-bottom: 8px;
}

.kv-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.action-detail {
  background: rgba(59, 130, 246, 0.12);
  color: #60a5fa;
}

.action-detail:hover:not(:disabled) {
  background: rgba(59, 130, 246, 0.25);
}

.glass-table-container :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(4px);
}

.glass-table-container :deep(.el-loading-spinner .circular) {
  stroke: var(--primary);
}

.cronjobs-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.cronjobs-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.cronjobs-page :deep(.el-message-box) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
}

.cronjobs-page :deep(.el-message-box__title) {
  color: var(--text-primary);
}

.cronjobs-page :deep(.el-message-box__content) {
  color: var(--text-secondary);
}

.cronjobs-page :deep(.el-message-box__btns .el-button--primary) {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-color: #6366f1;
}

.cronjobs-dialog :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  padding: 16px 24px;
}

.cronjobs-dialog :deep(.el-dialog__title) {
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 600;
}

.cronjobs-dialog :deep(.el-dialog__body) {
  padding: 24px;
}

.cronjobs-dialog :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

.cronjobs-dialog :deep(.el-input__wrapper) {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.cronjobs-dialog :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.cronjobs-dialog :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.cronjobs-dialog :deep(.el-input__inner) {
  color: var(--text-primary);
}

.cronjobs-dialog :deep(.el-select .el-input__wrapper) {
  background: rgba(15, 23, 42, 0.6);
}

.cronjobs-dialog :deep(.el-divider__text) {
  color: var(--text-secondary);
  background: #1e293b;
}

.cronjobs-dialog :deep(.el-divider) {
  border-color: rgba(148, 163, 184, 0.1);
}

.cronjobs-dialog :deep(.el-input-number .el-input__wrapper) {
  background: rgba(15, 23, 42, 0.6);
}

.cronjobs-dialog :deep(.el-input-number__decrease),
.cronjobs-dialog :deep(.el-input-number__increase) {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(148, 163, 184, 0.1);
  color: var(--text-secondary);
}

.cronjobs-dialog :deep(.el-input-number__decrease:hover),
.cronjobs-dialog :deep(.el-input-number__increase:hover) {
  color: var(--primary);
}
</style>
