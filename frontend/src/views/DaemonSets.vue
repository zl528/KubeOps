<template>
  <div class="daemonsets-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">DaemonSets</h1>
        <span class="page-subtitle">管理集群中的守护进程集</span>
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
        :data="daemonsets"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 DaemonSet 数据'"
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
        <el-table-column prop="desired" label="期望" width="80">
          <template #default="{ row }">
            <span class="cell-metric">{{ row.desired }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="current" label="当前" width="80">
          <template #default="{ row }">
            <span class="cell-metric">{{ row.current }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="ready" label="就绪" width="80">
          <template #default="{ row }">
            <span class="cell-metric" :class="{ 'metric-ok': row.ready === row.desired }">{{ row.ready }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="upToDate" label="已更新" width="80" />
        <el-table-column prop="available" label="可用" width="80" />
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

    <!-- 创建 DaemonSet 弹窗 -->
    <el-dialog v-model="createVisible" title="新建 DaemonSet" width="680px" :close-on-click-modal="false" class="dark-dialog">
      <el-form label-width="100px" class="create-form">
        <el-tabs v-model="createTab" class="create-tabs">
          <el-tab-pane label="基本配置" name="basic">
            <el-form-item label="命名空间" required>
              <el-select v-model="createForm.namespace" placeholder="选择命名空间" style="width: 100%">
                <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="名称" required>
              <el-input v-model="createForm.name" placeholder="my-daemonset" />
            </el-form-item>
            <el-form-item label="镜像" required>
              <el-input v-model="createForm.image" placeholder="fluentd:latest" />
            </el-form-item>
            <el-form-item label="容器端口">
              <el-input-number v-model="createForm.containerPort" :min="0" :max="65535" style="width: 100%" placeholder="留空则不暴露端口" />
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

          <el-tab-pane label="标签与调度" name="scheduling">
            <div class="section-label">Labels</div>
            <div class="env-vars-section" style="margin-bottom: 20px">
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
            <div class="section-label">Node Selector</div>
            <div class="env-vars-section" style="margin-bottom: 20px">
              <div v-for="(sel, idx) in createForm.nodeSelectorList" :key="idx" class="env-var-row">
                <el-input v-model="sel.key" placeholder="kubernetes.io/os" class="env-input" />
                <el-input v-model="sel.value" placeholder="linux" class="env-input" />
                <button type="button" class="env-remove-btn" @click="createForm.nodeSelectorList.splice(idx, 1)">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
              <button type="button" class="env-add-btn" @click="createForm.nodeSelectorList.push({ key: '', value: '' })">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                添加节点选择器
              </button>
            </div>
            <div class="section-label">Tolerations</div>
            <div v-for="(tol, idx) in createForm.tolerations" :key="idx" class="toleration-item">
              <div class="toleration-fields">
                <el-form-item label="Key" label-width="60px">
                  <el-input v-model="tol.key" placeholder="node-role.kubernetes.io/master" />
                </el-form-item>
                <el-form-item label="Operator" label-width="60px">
                  <el-select v-model="tol.operator" style="width: 100%">
                    <el-option label="Equal" value="Equal" />
                    <el-option label="Exists" value="Exists" />
                  </el-select>
                </el-form-item>
                <el-form-item label="Value" label-width="60px">
                  <el-input v-model="tol.value" placeholder="" :disabled="tol.operator === 'Exists'" />
                </el-form-item>
                <el-form-item label="Effect" label-width="60px">
                  <el-select v-model="tol.effect" style="width: 100%">
                    <el-option label="NoSchedule" value="NoSchedule" />
                    <el-option label="PreferNoSchedule" value="PreferNoSchedule" />
                    <el-option label="NoExecute" value="NoExecute" />
                  </el-select>
                </el-form-item>
              </div>
              <button type="button" class="toleration-remove" @click="createForm.tolerations.splice(idx, 1)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <button type="button" class="toleration-add" @click="createForm.tolerations.push({ key: '', operator: 'Equal', value: '', effect: 'NoSchedule' })">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              添加容忍
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
    <el-dialog v-model="detailVisible" title="DaemonSet 详情" width="700px" :close-on-click-modal="false" class="dark-dialog">
      <div v-if="detailLoading" class="detail-loading">加载中...</div>
      <div v-else-if="detailData" class="detail-content">
        <div class="detail-section">
          <h4 class="detail-section-title">基本信息</h4>
          <div class="detail-grid">
            <div class="detail-item"><span class="detail-label">名称</span><span class="detail-value">{{ detailData.name }}</span></div>
            <div class="detail-item"><span class="detail-label">命名空间</span><span class="detail-value">{{ detailData.namespace }}</span></div>
            <div class="detail-item"><span class="detail-label">期望副本</span><span class="detail-value">{{ detailData.desired }}</span></div>
            <div class="detail-item"><span class="detail-label">当前副本</span><span class="detail-value">{{ detailData.current }}</span></div>
            <div class="detail-item"><span class="detail-label">就绪副本</span><span class="detail-value">{{ detailData.ready }}</span></div>
            <div class="detail-item"><span class="detail-label">可用副本</span><span class="detail-value">{{ detailData.available }}</span></div>
            <div class="detail-item"><span class="detail-label">更新策略</span><span class="detail-value">{{ detailData.updateStrategy || 'RollingUpdate' }}</span></div>
            <div class="detail-item"><span class="detail-label">创建时间</span><span class="detail-value">{{ detailData.creationTimestamp || detailData.age }}</span></div>
          </div>
        </div>
        <div class="detail-section" v-if="detailData.containers?.length">
          <h4 class="detail-section-title">容器</h4>
          <div v-for="(c, i) in detailData.containers" :key="i" class="container-item">
            <div class="detail-grid">
              <div class="detail-item"><span class="detail-label">名称</span><span class="detail-value">{{ c.name }}</span></div>
              <div class="detail-item"><span class="detail-label">镜像</span><span class="detail-value">{{ c.image }}</span></div>
              <div class="detail-item" v-if="c.ports?.length"><span class="detail-label">端口</span><span class="detail-value">{{ c.ports.map((p: any) => p.containerPort).join(', ') }}</span></div>
            </div>
          </div>
        </div>
        <div class="detail-section" v-if="detailData.labels">
          <h4 class="detail-section-title">标签</h4>
          <div class="labels-grid">
            <span v-for="(v, k) in detailData.labels" :key="k" class="label-tag">{{ k }}={{ v }}</span>
          </div>
        </div>
        <div class="detail-section" v-if="detailData.nodeSelector">
          <h4 class="detail-section-title">节点选择器</h4>
          <div class="labels-grid">
            <span v-for="(v, k) in detailData.nodeSelector" :key="k" class="label-tag">{{ k }}={{ v }}</span>
          </div>
        </div>
      </div>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="detailVisible = false">关闭</button>
      </template>
    </el-dialog>

    <!-- 编辑弹窗 -->
    <el-dialog v-model="editVisible" title="编辑 DaemonSet" width="600px" :close-on-click-modal="false" class="dark-dialog">
      <el-form label-width="100px" class="create-form">
        <el-form-item label="名称">
          <el-input :model-value="editForm.name" disabled />
        </el-form-item>
        <el-form-item label="命名空间">
          <el-input :model-value="editForm.namespace" disabled />
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

useHighlightRow()
const daemonsets = ref<any[]>([])
useHighlightRow()
const nsList = ref<any[]>([])
const { namespace } = useGlobalNamespace()
const loading = ref(false)

const createVisible = ref(false)
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
const createTab = ref('basic')
const createForm = ref({
  namespace: '',
  name: '',
  image: '',
  containerPort: 0,
  nodeSelectorList: [] as Array<{ key: string; value: string }>,
  envVars: [] as { key: string; value: string }[],
  cpuLimit: '',
  memoryLimit: '',
  cpuRequest: '',
  memoryRequest: '',
  labels: [] as { key: string; value: string }[],
  tolerations: [] as Array<{ key: string; operator: string; value: string; effect: string }>,
})

const showCreate = () => {
  createTab.value = 'basic'
  createForm.value = {
    namespace: namespace.value || '',
    name: '',
    image: '',
    containerPort: 0,
    nodeSelectorList: [],
    envVars: [],
    cpuLimit: '',
    memoryLimit: '',
    cpuRequest: '',
    memoryRequest: '',
    labels: [],
    tolerations: [],
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
      image: createForm.value.image,
    }
    if (createForm.value.containerPort > 0) {
      payload.containerPort = createForm.value.containerPort
    }
    const validSelectors = createForm.value.nodeSelectorList.filter(s => s.key && s.value)
    if (validSelectors.length > 0) {
      payload.nodeSelector = Object.fromEntries(validSelectors.map(s => [s.key, s.value]))
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
    const validTolerations = createForm.value.tolerations.filter(t => t.key && t.effect)
    if (validTolerations.length > 0) {
      payload.tolerations = validTolerations.map(t => ({
        key: t.key,
        operator: t.operator,
        value: t.operator === 'Exists' ? '' : t.value,
        effect: t.effect,
      }))
    }
    await api.post('/daemonsets/create', payload)
    ElMessage.success('创建成功')
    createVisible.value = false
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
    const res: any = await api.get('/daemonsets', { params })
    daemonsets.value = res.data || []
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
    await ElMessageBox.confirm(`确定要重启 DaemonSet ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.post(`/daemonsets/restart?namespace=${row.namespace}&name=${row.name}`)
    ElMessage.success('重启成功')
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('重启失败')
  }
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 DaemonSet ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.delete(`/daemonsets/delete?namespace=${row.namespace}&name=${row.name}`)
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
    const res: any = await api.get(`/daemonsets`, { params: { namespace: row.namespace } })
    const ds = (res.data || []).find((d: any) => d.name === row.name)
    if (ds) {
      detailData.value = {
        name: ds.name,
        namespace: ds.namespace,
        desired: ds.desired || ds.desiredNumberScheduled,
        current: ds.current || ds.currentNumberScheduled,
        ready: ds.ready || ds.numberReady,
        available: ds.available || ds.numberAvailable,
        updateStrategy: ds.updateStrategy || 'RollingUpdate',
        creationTimestamp: ds.creationTimestamp || ds.age,
        containers: ds.containers || [{ name: ds.name, image: ds.image || '-' }],
        labels: ds.labels || { app: ds.name },
        nodeSelector: ds.nodeSelector || {},
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
    const res: any = await api.get(`/daemonsets/get`, { params: { namespace: row.namespace, name: row.name } })
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
    await api.post('/daemonsets/update', payload)
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
.daemonsets-page {
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

.glass-table-container :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(4px);
}

.glass-table-container :deep(.el-loading-spinner .circular) {
  stroke: var(--primary);
}

.daemonsets-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.daemonsets-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.daemonsets-page :deep(.el-message-box) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
}

.daemonsets-page :deep(.el-message-box__title) {
  color: var(--text-primary);
}

.daemonsets-page :deep(.el-message-box__content) {
  color: var(--text-secondary);
}

.daemonsets-page :deep(.el-message-box__btns .el-button--primary) {
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

.section-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  margin-bottom: 10px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
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

.toleration-item {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.toleration-fields {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.toleration-fields .el-form-item {
  margin-bottom: 4px;
}

.toleration-remove {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  border: none;
  border-radius: 6px;
  padding: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
  margin-top: 6px;
}

.toleration-remove:hover {
  background: rgba(239, 68, 68, 0.25);
}

.toleration-add {
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

.toleration-add:hover {
  background: rgba(99, 102, 241, 0.2);
  border-color: rgba(99, 102, 241, 0.5);
}

.node-selector-item {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 8px;
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.node-selector-item .el-form-item {
  margin-bottom: 0;
}

.node-selector-remove {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  border: none;
  border-radius: 6px;
  padding: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
  margin-top: 6px;
}

.node-selector-remove:hover {
  background: rgba(239, 68, 68, 0.25);
}

.node-selector-add {
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

.node-selector-add:hover {
  background: rgba(99, 102, 241, 0.2);
  border-color: rgba(99, 102, 241, 0.5);
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

.daemonsets-page :deep(.el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.08);
  padding: 16px 24px;
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
