<template>
  <div class="statefulsets-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">StatefulSets</h1>
        <span class="page-subtitle">管理集群中的有状态应用</span>
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
        :data="statefulsets"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 StatefulSet 数据'"
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
        <el-table-column prop="replicas" label="副本数" width="120">
          <template #default="{ row }">
            <span class="cell-metric">{{ row.replicas }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="ready" label="就绪" width="120">
          <template #default="{ row }">
            <span class="cell-metric" :class="{ 'metric-ok': row.ready === row.replicas }">{{ row.ready }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="存活时间" width="100" />
        <el-table-column label="操作" width="280" fixed="right">
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
              <button type="button" class="action-btn action-scale" @click="scaleDialog(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 3H3v18h18V3z"/><path d="M12 8v8"/><path d="M8 12h8"/></svg>
                扩缩容
              </button>
              <button type="button" class="action-btn action-restart" @click="handleRestart(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                重启
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

    <!-- 创建 StatefulSet 弹窗 -->
    <el-dialog v-model="createVisible" title="新建 StatefulSet" width="680px" :close-on-click-modal="false" class="dark-dialog">
      <el-form label-width="100px" class="create-form">
        <el-tabs v-model="createTab" class="create-tabs">
          <el-tab-pane label="基本配置" name="basic">
            <el-form-item label="命名空间" required>
              <el-select v-model="createForm.namespace" placeholder="选择命名空间" style="width: 100%">
                <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="名称" required>
              <el-input v-model="createForm.name" placeholder="my-statefulset" />
            </el-form-item>
            <el-form-item label="副本数">
              <el-input-number v-model="createForm.replicas" :min="1" :max="1000" style="width: 100%" />
            </el-form-item>
            <el-form-item label="镜像" required>
              <el-input v-model="createForm.image" placeholder="nginx:latest" />
            </el-form-item>
            <el-form-item label="容器端口">
              <el-input-number v-model="createForm.containerPort" :min="0" :max="65535" style="width: 100%" placeholder="留空则不暴露端口" />
            </el-form-item>
            <el-form-item label="Service名称">
              <el-input v-model="createForm.serviceName" placeholder="留空则使用资源名称" />
            </el-form-item>
          </el-tab-pane>

          <el-tab-pane label="环境变量" name="env">
            <div class="env-vars-section">
              <div v-for="(ev, idx) in createForm.envVars" :key="idx" class="env-var-row">
                <el-input v-model="ev.key" placeholder="变量名" class="env-input" />
                <el-input v-model="ev.value" placeholder="值" class="env-input" />
                <button type="button" class="env-remove-btn" @click="createForm.envVars.splice(idx, 1)">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
              <button type="button" class="env-add-btn" @click="createForm.envVars.push({ key: '', value: '' })">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                添加环境变量
              </button>
            </div>
          </el-tab-pane>

          <el-tab-pane label="资源限制" name="resources">
            <el-form-item label="CPU 限制">
              <el-input v-model="createForm.cpuLimit" placeholder="例如 100m, 1" />
            </el-form-item>
            <el-form-item label="内存限制">
              <el-input v-model="createForm.memoryLimit" placeholder="例如 128Mi, 1Gi" />
            </el-form-item>
            <el-form-item label="CPU 请求">
              <el-input v-model="createForm.cpuRequest" placeholder="例如 50m, 0.5" />
            </el-form-item>
            <el-form-item label="内存请求">
              <el-input v-model="createForm.memoryRequest" placeholder="例如 64Mi, 512Mi" />
            </el-form-item>
          </el-tab-pane>

          <el-tab-pane label="标签" name="labels">
            <div class="env-vars-section">
              <div v-for="(l, idx) in createForm.labels" :key="idx" class="env-var-row">
                <el-input v-model="l.key" placeholder="键" class="env-input" />
                <el-input v-model="l.value" placeholder="值" class="env-input" />
                <button type="button" class="env-remove-btn" @click="createForm.labels.splice(idx, 1)">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
              <button type="button" class="env-add-btn" @click="createForm.labels.push({ key: '', value: '' })">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                添加标签
              </button>
            </div>
          </el-tab-pane>

          <el-tab-pane label="存储卷" name="storage">
            <div v-for="(vc, idx) in createForm.volumeClaims" :key="idx" class="volume-claim-item">
              <div class="volume-claim-fields">
                <el-form-item label="名称" label-width="60px">
                  <el-input v-model="vc.name" placeholder="data" />
                </el-form-item>
                <el-form-item label="大小" label-width="60px">
                  <el-input v-model="vc.size" placeholder="10Gi" />
                </el-form-item>
                <el-form-item label="存储类" label-width="60px">
                  <el-select v-model="vc.storageClass" placeholder="选择存储类" filterable style="width: 100%">
                    <el-option label="不指定" value="" />
                    <el-option v-for="sc in storageClasses" :key="sc.name" :label="sc.name" :value="sc.name" />
                  </el-select>
                </el-form-item>
              </div>
              <button type="button" class="volume-claim-remove" @click="createForm.volumeClaims.splice(idx, 1)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <button type="button" class="volume-claim-add" @click="createForm.volumeClaims.push({ name: '', size: '10Gi', storageClass: '' })">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              添加存储卷
            </button>
          </el-tab-pane>
        </el-tabs>
      </el-form>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="createVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-confirm" @click="handleCreate" :disabled="createLoading">
          <span v-if="createLoading" class="btn-spinner"></span>
          创建
        </button>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="StatefulSet 详情" width="700px" :close-on-click-modal="false" class="dark-dialog">
      <div v-if="detailLoading" class="detail-loading">加载中...</div>
      <div v-else-if="detailData" class="detail-content">
        <div class="detail-section">
          <h4 class="detail-section-title">基本信息</h4>
          <div class="detail-grid">
            <div class="detail-item"><span class="detail-label">名称</span><span class="detail-value">{{ detailData.name }}</span></div>
            <div class="detail-item"><span class="detail-label">命名空间</span><span class="detail-value">{{ detailData.namespace }}</span></div>
            <div class="detail-item"><span class="detail-label">副本数</span><span class="detail-value">{{ detailData.replicas }}</span></div>
            <div class="detail-item"><span class="detail-label">就绪副本</span><span class="detail-value">{{ detailData.readyReplicas || 0 }}</span></div>
            <div class="detail-item"><span class="detail-label">可用副本</span><span class="detail-value">{{ detailData.availableReplicas || 0 }}</span></div>
            <div class="detail-item"><span class="detail-label">更新策略</span><span class="detail-value">{{ detailData.updateStrategy || 'RollingUpdate' }}</span></div>
            <div class="detail-item"><span class="detail-label">Service 名称</span><span class="detail-value">{{ detailData.serviceName || '-' }}</span></div>
            <div class="detail-item"><span class="detail-label">创建时间</span><span class="detail-value">{{ detailData.creationTimestamp || '-' }}</span></div>
          </div>
        </div>
        <div class="detail-section" v-if="detailData.containers?.length">
          <h4 class="detail-section-title">容器</h4>
          <div v-for="(c, i) in detailData.containers" :key="i" class="container-item">
            <div class="detail-grid">
              <div class="detail-item"><span class="detail-label">名称</span><span class="detail-value">{{ c.name }}</span></div>
              <div class="detail-item"><span class="detail-label">镜像</span><span class="detail-value">{{ c.image }}</span></div>
              <div class="detail-item" v-if="c.ports?.length"><span class="detail-label">端口</span><span class="detail-value">{{ c.ports.map(p => p.containerPort).join(', ') }}</span></div>
            </div>
          </div>
        </div>
        <div class="detail-section" v-if="detailData.labels">
          <h4 class="detail-section-title">标签</h4>
          <div class="labels-grid">
            <span v-for="(v, k) in detailData.labels" :key="k" class="label-tag">{{ k }}={{ v }}</span>
          </div>
        </div>
        <div class="detail-section" v-if="detailData.conditions?.length">
          <h4 class="detail-section-title">状态条件</h4>
          <div v-for="(cond, i) in detailData.conditions" :key="i" class="condition-item">
            <span class="condition-type">{{ cond.type }}</span>
            <span class="condition-status" :class="cond.status === 'True' ? 'status-ok' : 'status-warning'">{{ cond.status }}</span>
            <span class="condition-message">{{ cond.message || '-' }}</span>
          </div>
        </div>
      </div>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="detailVisible = false">关闭</button>
      </template>
    </el-dialog>

    <!-- 扩缩容弹窗 -->
    <el-dialog v-model="scaleVisible" title="扩缩容 StatefulSet" width="400px" :close-on-click-modal="false" class="dark-dialog">
      <el-form label-width="80px" class="create-form">
        <el-form-item label="名称">
          <el-input :model-value="scaleForm.name" disabled />
        </el-form-item>
        <el-form-item label="命名空间">
          <el-input :model-value="scaleForm.namespace" disabled />
        </el-form-item>
        <el-form-item label="副本数" required>
          <el-input-number v-model="scaleForm.replicas" :min="0" :max="1000" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="scaleVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-confirm" @click="handleScale" :disabled="scaleLoading">
          <span v-if="scaleLoading" class="btn-spinner"></span>
          确定
        </button>
      </template>
    </el-dialog>

    <!-- 编辑弹窗 -->
    <el-dialog v-model="editVisible" title="编辑 StatefulSet" width="600px" :close-on-click-modal="false" class="dark-dialog">
      <el-form label-width="100px" class="create-form">
        <el-form-item label="名称">
          <el-input :model-value="editForm.name" disabled />
        </el-form-item>
        <el-form-item label="命名空间">
          <el-input :model-value="editForm.namespace" disabled />
        </el-form-item>
        <el-form-item label="副本数">
          <el-input-number v-model="editForm.replicas" :min="0" :max="1000" style="width: 100%" />
        </el-form-item>
        <el-form-item label="镜像">
          <el-input v-model="editForm.image" placeholder="nginx:latest" />
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
import { useHighlightRow } from '../composables/useHighlightRow'
import api from '../api'

const { namespace } = useGlobalNamespace()
useHighlightRow()
const statefulsets = ref<any[]>([])
useHighlightRow()
const nsList = ref<any[]>([])
useHighlightRow()
const storageClasses = ref<any[]>([])

const loading = ref(false)

const createVisible = ref(false)
const createLoading = ref(false)
const createTab = ref('basic')
const createForm = ref({
  namespace: '',
  name: '',
  replicas: 1,
  image: '',
  serviceName: '',
  containerPort: 0,
  volumeClaims: [] as Array<{ name: string; size: string; storageClass: string }>,
  envVars: [] as { key: string; value: string }[],
  cpuLimit: '',
  memoryLimit: '',
  cpuRequest: '',
  memoryRequest: '',
  labels: [] as { key: string; value: string }[],
})

const detailVisible = ref(false)
const detailLoading = ref(false)
const detailData = ref<any>(null)

const scaleVisible = ref(false)
const scaleLoading = ref(false)
const scaleForm = ref({ namespace: '', name: '', replicas: 1 })

const editVisible = ref(false)
const editLoading = ref(false)
const editForm = ref({
  namespace: '',
  name: '',
  replicas: 1,
  image: '',
  cpuLimit: '',
  memoryLimit: '',
  cpuRequest: '',
  memoryRequest: '',
  labels: [] as { key: string; value: string }[],
})

const showCreate = () => {
  createTab.value = 'basic'
  createForm.value = {
    namespace: namespace.value || '',
    name: '',
    replicas: 1,
    image: '',
    serviceName: '',
    containerPort: 0,
    volumeClaims: [],
    envVars: [],
    cpuLimit: '',
    memoryLimit: '',
    cpuRequest: '',
    memoryRequest: '',
    labels: [],
  }
  createVisible.value = true
}

const handleCreate = async () => {
  if (!createForm.value.namespace || !createForm.value.name || !createForm.value.image) {
    ElMessage.warning('请填写命名空间、名称和镜像')
    return
  }
  createLoading.value = true
  try {
    const payload: any = {
      namespace: createForm.value.namespace,
      name: createForm.value.name,
      replicas: createForm.value.replicas,
      image: createForm.value.image,
      serviceName: createForm.value.serviceName || createForm.value.name,
    }
    if (createForm.value.containerPort > 0) {
      payload.containerPort = createForm.value.containerPort
    }
    if (createForm.value.volumeClaims.length > 0) {
      payload.volumeClaims = createForm.value.volumeClaims.filter(vc => vc.name && vc.size)
    }
    const envVars: Record<string, string> = {}
    createForm.value.envVars.forEach(e => { if (e.key) envVars[e.key] = e.value })
    if (Object.keys(envVars).length > 0) payload.envVars = envVars
    const labels: Record<string, string> = {}
    createForm.value.labels.forEach(l => { if (l.key) labels[l.key] = l.value })
    if (Object.keys(labels).length > 0) payload.labels = labels
    const resourceLimits: Record<string, string> = {}
    const resourceRequests: Record<string, string> = {}
    if (createForm.value.cpuLimit) resourceLimits.cpu = createForm.value.cpuLimit
    if (createForm.value.memoryLimit) resourceLimits.memory = createForm.value.memoryLimit
    if (createForm.value.cpuRequest) resourceRequests.cpu = createForm.value.cpuRequest
    if (createForm.value.memoryRequest) resourceRequests.memory = createForm.value.memoryRequest
    if (Object.keys(resourceLimits).length > 0) payload.resourceLimits = resourceLimits
    if (Object.keys(resourceRequests).length > 0) payload.resourceRequests = resourceRequests
    await api.post('/statefulsets/create', payload)
    ElMessage.success('创建成功')
    createVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  } finally {
    createLoading.value = false
  }
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
    const params = namespace.value ? { namespace: namespace.value } : {}
    const res: any = await api.get('/statefulsets', { params })
    statefulsets.value = res.data || []
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

const handleRestart = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要重启 StatefulSet ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.post(`/statefulsets/restart?namespace=${row.namespace}&name=${row.name}`)
    ElMessage.success('重启成功')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('重启失败')
  }
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 StatefulSet ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.delete(`/statefulsets/delete?namespace=${row.namespace}&name=${row.name}`)
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
    const res: any = await api.get(`/statefulsets`, { params: { namespace: row.namespace } })
    const sts = (res.data || []).find((s: any) => s.name === row.name)
    if (sts) {
      detailData.value = {
        name: sts.name,
        namespace: sts.namespace,
        replicas: sts.replicas,
        readyReplicas: sts.ready,
        availableReplicas: sts.ready,
        serviceName: sts.serviceName || sts.name,
        updateStrategy: sts.updateStrategy || 'RollingUpdate',
        creationTimestamp: sts.creationTimestamp || sts.age,
        containers: sts.containers || [{ name: sts.name, image: sts.image || '-' }],
        labels: sts.labels || { app: sts.name },
        conditions: sts.conditions || [],
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

const scaleDialog = (row: any) => {
  scaleForm.value = { namespace: row.namespace, name: row.name, replicas: row.replicas || 1 }
  scaleVisible.value = true
}

const handleScale = async () => {
  scaleLoading.value = true
  try {
    await api.post(`/statefulsets/scale?namespace=${scaleForm.value.namespace}&name=${scaleForm.value.name}&replicas=${scaleForm.value.replicas}`)
    ElMessage.success('扩缩容成功')
    scaleVisible.value = false
    fetchData()
  } catch (e) {
    ElMessage.error('扩缩容失败')
  } finally {
    scaleLoading.value = false
  }
}

const editDialog = async (row: any) => {
  try {
    const res: any = await api.get(`/statefulsets/get`, { params: { namespace: row.namespace, name: row.name } })
    const data = res.data || row
    editForm.value = {
      namespace: data.namespace,
      name: data.name,
      replicas: data.replicas || 1,
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
      replicas: row.replicas || 1,
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
      replicas: editForm.value.replicas,
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
    await api.post('/statefulsets/update', payload)
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

onMounted(() => { fetchNs(); fetchData(); fetchStorageClasses() })
</script>

<style scoped>
.statefulsets-page {
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

.action-restart {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
}

.action-restart:hover:not(:disabled) {
  background: rgba(148, 163, 184, 0.25);
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

.action-scale {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
}

.action-scale:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.25);
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

.container-item {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
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

.condition-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid rgba(148, 163, 184, 0.06);
}

.condition-type {
  font-weight: 500;
  color: var(--text-primary);
  min-width: 120px;
}

.condition-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
}

.condition-message {
  font-size: 12px;
  color: var(--text-secondary);
  flex: 1;
}

.glass-table-container :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(4px);
}

.glass-table-container :deep(.el-loading-spinner .circular) {
  stroke: var(--primary);
}

.statefulsets-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.statefulsets-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.statefulsets-page :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  padding: 20px 24px;
  margin: 0;
}

.statefulsets-page :deep(.el-dialog__title) {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 18px;
}

.statefulsets-page :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--text-secondary);
}

:deep(.dark-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.2) !important;
}

:deep(.dark-dialog .el-dialog__header) {
  background: #1e293b !important;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1) !important;
}

:deep(.dark-dialog .el-dialog__title) {
  color: #ffffff !important;
}

.dark-dialog :deep(.el-dialog__body) {
  padding: 24px;
  color: var(--text-primary);
}

.statefulsets-page :deep(.el-message-box) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
}

.statefulsets-page :deep(.el-message-box__title) {
  color: var(--text-primary);
}

.statefulsets-page :deep(.el-message-box__content) {
  color: var(--text-secondary);
}

.statefulsets-page :deep(.el-message-box__btns .el-button--primary) {
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

.cronjobs-dialog :deep(.el-input-number .el-input__wrapper) {
  background: rgba(15, 23, 42, 0.6);
}

.cronjobs-dialog :deep(.el-divider__text) {
  color: var(--text-secondary);
  background: #1e293b;
}

.cronjobs-dialog :deep(.el-divider) {
  border-color: rgba(148, 163, 184, 0.1);
}

.volume-claim-item {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.volume-claim-fields {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.volume-claim-remove {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  border: none;
  border-radius: 6px;
  padding: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
}

.volume-claim-remove:hover {
  background: rgba(239, 68, 68, 0.25);
}

.volume-claim-add {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(99, 102, 241, 0.12);
  color: #818cf8;
  border: 1px dashed rgba(99, 102, 241, 0.3);
  border-radius: 8px;
  padding: 8px 16px;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s ease;
  width: 100%;
  justify-content: center;
}

.volume-claim-add:hover {
  background: rgba(99, 102, 241, 0.2);
  border-color: rgba(99, 102, 241, 0.5);
}

.statefulsets-page :deep(.el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.08);
  padding: 16px 24px;
}

.statefulsets-page :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--text-secondary);
}

.statefulsets-page :deep(.el-divider__text) {
  background: #1e293b;
  color: var(--text-secondary);
}

.statefulsets-page :deep(.el-divider) {
  border-color: rgba(148, 163, 184, 0.08);
}

.create-form :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.create-form :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.create-form :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.create-form :deep(.el-input__inner) {
  color: var(--text-primary);
}

.create-form :deep(.el-input__inner::placeholder) {
  color: rgba(148, 163, 184, 0.5);
}

.create-form :deep(.el-select .el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.create-form :deep(.el-select-dropdown) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.create-form :deep(.el-select-dropdown__item) {
  color: var(--text-primary);
}

.create-form :deep(.el-select-dropdown__item.hover) {
  background: rgba(99, 102, 241, 0.15);
}

.create-form :deep(.el-select-dropdown__item.selected) {
  color: #6366f1;
}

.create-form :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

.create-form :deep(.el-input-number .el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.env-vars-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.env-var-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.env-input {
  flex: 1;
}

.env-input :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.env-input :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.env-input :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.env-input :deep(.el-input__inner) {
  color: var(--text-primary);
}

.env-input :deep(.el-input__inner::placeholder) {
  color: rgba(148, 163, 184, 0.5);
}

.env-remove-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  min-width: 32px;
  border: none;
  border-radius: 6px;
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  cursor: pointer;
  transition: all 0.15s ease;
}

.env-remove-btn:hover {
  background: rgba(239, 68, 68, 0.25);
}

.env-add-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px dashed rgba(99, 102, 241, 0.3);
  border-radius: 6px;
  background: transparent;
  color: #818cf8;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s ease;
  align-self: flex-start;
}

.env-add-btn:hover {
  border-color: rgba(99, 102, 241, 0.6);
  background: rgba(99, 102, 241, 0.08);
}

.create-tabs :deep(.el-tabs__header) {
  margin-bottom: 16px;
}

.create-tabs :deep(.el-tabs__item) {
  color: var(--text-secondary);
}

.create-tabs :deep(.el-tabs__item.is-active) {
  color: #818cf8;
}

.create-tabs :deep(.el-tabs__active-bar) {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
}

.create-tabs :deep(.el-tabs__nav-wrap::after) {
  background: rgba(148, 163, 184, 0.08);
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
</style>
