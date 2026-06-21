<template>
  <div class="deployments-page">
    <!-- 渐变页头 -->
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">Deployments</h1>
        <span class="page-subtitle">管理集群中的部署资源</span>
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

    <!-- 毛玻璃表格容器 -->
    <div class="glass-table-container">
      <el-table
        :data="deploys"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 Deployment 数据'"
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
        <el-table-column prop="replicas" label="副本" width="80">
          <template #default="{ row }">
            <span class="cell-metric">{{ row.replicas }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="ready" label="就绪" width="80">
          <template #default="{ row }">
            <span class="cell-metric" :class="{ 'metric-ok': row.ready === row.replicas }">{{ row.ready }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="upToDate" label="更新" width="80" />
        <el-table-column prop="available" label="可用" width="80" />
        <el-table-column prop="age" label="存活" width="80" />
        <el-table-column label="操作" width="260" fixed="right">
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
              <button type="button" class="action-btn action-scale" @click="scaleDialog(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5" rx="1"/><line x1="10" y1="12" x2="14" y2="12"/></svg>
                扩缩容
              </button>
              <button type="button" class="action-btn action-restart" @click="restartDeploy(row)" :disabled="row._restarting">
                <svg v-if="!row._restarting" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
                <span v-else class="btn-spinner"></span>
                重启
              </button>
              <button type="button" class="action-btn action-rollback" @click="rollbackDialog(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="1 4 1 10 7 10"/><path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/></svg>
                回滚
              </button>
              <button type="button" class="action-btn action-delete" @click="deleteDeploy(row)" :disabled="row._deleting">
                <svg v-if="!row._deleting" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                <span v-else class="btn-spinner"></span>
                删除
              </button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="Deployment 详情" width="750px" class="dark-dialog">
      <template v-if="selected">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="名称">{{ selected.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ selected.namespace }}</el-descriptions-item>
          <el-descriptions-item label="副本数">{{ selected.replicas }}</el-descriptions-item>
          <el-descriptions-item label="就绪">{{ selected.ready }}</el-descriptions-item>
          <el-descriptions-item label="策略">{{ selected.strategy || 'RollingUpdate' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ selected.createdAt || '-' }}</el-descriptions-item>
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

    <!-- 扩缩容弹窗 -->
    <el-dialog v-model="scaleVisible" title="扩缩容" width="400px" class="dark-dialog">
      <el-form label-width="80px">
        <el-form-item label="当前副本">
          <el-input-number v-model="scaleForm.replicas" :min="0" :max="100" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="scaleVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-confirm" @click="confirmScale" :disabled="scaleLoading">
          <span v-if="scaleLoading" class="btn-spinner"></span>
          确定
        </button>
      </template>
    </el-dialog>

    <!-- 编辑弹窗 -->
    <el-dialog v-model="editVisible" title="编辑 Deployment" width="680px" class="dark-dialog" :close-on-click-modal="false">
      <div v-loading="editLoading">
        <el-form label-width="100px" class="edit-form">
          <el-form-item label="命名空间">
            <el-input :model-value="editForm.namespace" disabled />
          </el-form-item>
          <el-form-item label="名称">
            <el-input :model-value="editForm.name" disabled />
          </el-form-item>
          <el-form-item label="容器名称">
            <el-select v-model="editForm.containerName" placeholder="选择容器" style="width: 100%" @change="onContainerChange">
              <el-option v-for="c in editForm.containers" :key="c.name" :label="c.name" :value="c.name" />
            </el-select>
          </el-form-item>
          <el-form-item label="副本数">
            <el-input-number v-model="editForm.replicas" :min="0" :max="1000" style="width: 100%" />
          </el-form-item>
          <el-form-item label="镜像">
            <el-input v-model="editForm.image" placeholder="例如 nginx:latest" />
          </el-form-item>

          <el-divider content-position="left">环境变量</el-divider>
          <div class="env-vars-section">
            <div v-for="(ev, idx) in editForm.envVars" :key="idx" class="env-var-row">
              <el-input v-model="ev.key" placeholder="变量名" class="env-input" />
              <el-input v-model="ev.value" placeholder="值" class="env-input" />
              <button type="button" class="env-remove-btn" @click="removeEnvVar(idx)">
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <button type="button" class="env-add-btn" @click="addEnvVar">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
              添加环境变量
            </button>
          </div>

          <el-divider content-position="left">资源限制</el-divider>
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
        </el-form>
      </div>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="editVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-confirm" @click="confirmEdit" :disabled="editSaving">
          <span v-if="editSaving" class="btn-spinner"></span>
          保存
        </button>
      </template>
    </el-dialog>

    <!-- 回滚弹窗 -->
    <el-dialog v-model="rollbackVisible" title="回滚版本" width="400px" class="dark-dialog">
      <el-form label-width="80px">
        <el-form-item label="目标版本">
          <el-input-number v-model="rollbackForm.revision" :min="1" :max="999" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="rollbackVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-warning" @click="confirmRollback" :disabled="rollbackLoading">
          <span v-if="rollbackLoading" class="btn-spinner"></span>
          回滚
        </button>
      </template>
    </el-dialog>

    <!-- 创建 Deployment 弹窗 -->
    <el-dialog v-model="createVisible" title="新建 Deployment" width="680px" class="dark-dialog" :close-on-click-modal="false">
      <el-form label-width="100px" class="create-form">
        <el-tabs v-model="createTab" class="create-tabs">
          <!-- 基本配置 -->
          <el-tab-pane label="基本配置" name="basic">
            <el-form-item label="命名空间" required>
              <el-select v-model="createForm.namespace" placeholder="选择命名空间" style="width: 100%">
                <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="名称" required>
              <el-input v-model="createForm.name" placeholder="my-deployment" />
            </el-form-item>
            <el-form-item label="副本数">
              <el-input-number v-model="createForm.replicas" :min="1" :max="1000" style="width: 100%" />
            </el-form-item>
            <el-form-item label="镜像" required>
              <el-input v-model="createForm.image" placeholder="nginx:latest" />
            </el-form-item>
            <el-form-item label="容器端口">
              <div v-for="(port, idx) in createForm.ports" :key="idx" class="port-row">
                <el-input v-model.number="createForm.ports[idx]" placeholder="80" style="width: 120px" />
                <button type="button" class="env-remove-btn" @click="createForm.ports.splice(idx, 1)" v-if="createForm.ports.length > 1">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
              <button type="button" class="env-add-btn" @click="createForm.ports.push(0)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                添加端口
              </button>
            </el-form-item>
            <el-form-item label="重启策略">
              <el-select v-model="createForm.restartPolicy" style="width: 100%">
                <el-option label="Always (默认)" value="Always" />
                <el-option label="OnFailure" value="OnFailure" />
                <el-option label="Never" value="Never" />
              </el-select>
            </el-form-item>
            <el-form-item label="ServiceAccount">
              <el-input v-model="createForm.serviceAccountName" placeholder="留空使用默认" />
            </el-form-item>
          </el-tab-pane>

          <!-- 环境变量 -->
          <el-tab-pane label="环境变量" name="env">
            <div class="env-vars-section">
              <div v-for="(ev, idx) in createForm.envVars" :key="idx" class="env-var-row">
                <el-input v-model="ev.key" placeholder="变量名" class="env-input" />
                <el-input v-model="ev.value" placeholder="值" class="env-input" />
                <button type="button" class="env-remove-btn" @click="removeCreateEnvVar(idx)">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
              <button type="button" class="env-add-btn" @click="addCreateEnvVar">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                添加环境变量
              </button>
            </div>
          </el-tab-pane>

          <!-- 资源限制 -->
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

          <!-- 标签与调度 -->
          <el-tab-pane label="标签与调度" name="scheduling">
            <div class="section-label">Labels</div>
            <div class="env-vars-section" style="margin-bottom: 20px">
              <div v-for="(l, idx) in createForm.labels" :key="idx" class="env-var-row">
                <el-input v-model="l.key" placeholder="键" class="env-input" />
                <el-input v-model="l.value" placeholder="值" class="env-input" />
                <button type="button" class="env-remove-btn" @click="removeCreateLabel(idx)">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
              <button type="button" class="env-add-btn" @click="addCreateLabel">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                添加标签
              </button>
            </div>
            <div class="section-label">Node Selector</div>
            <div class="env-vars-section">
              <div v-for="(n, idx) in createForm.nodeSelector" :key="idx" class="env-var-row">
                <el-input v-model="n.key" placeholder="键" class="env-input" />
                <el-input v-model="n.value" placeholder="值" class="env-input" />
                <button type="button" class="env-remove-btn" @click="removeCreateNodeSelector(idx)">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
              <button type="button" class="env-add-btn" @click="addCreateNodeSelector">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
                添加节点选择器
              </button>
            </div>
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
const deploys = ref<any[]>([])
const nsList = ref<any[]>([])
const { namespace } = useGlobalNamespace()
const loading = ref(false)

const detailVisible = ref(false)
const selected = ref<any>(null)

const scaleVisible = ref(false)
const scaleLoading = ref(false)
const scaleForm = ref({ name: '', namespace: '', replicas: 1 })

const rollbackVisible = ref(false)
const rollbackLoading = ref(false)
const rollbackForm = ref({ name: '', namespace: '', revision: 1 })

const editVisible = ref(false)
const editLoading = ref(false)
const editSaving = ref(false)
const editForm = ref({
  namespace: '',
  name: '',
  containerName: '',
  replicas: 1,
  image: '',
  envVars: [] as { key: string; value: string }[],
  cpuLimit: '',
  memoryLimit: '',
  cpuRequest: '',
  memoryRequest: '',
  containers: [] as { name: string; image: string; env: { name: string; value: string }[]; resources: any }[],
})

const createVisible = ref(false)
const createLoading = ref(false)
const createTab = ref('basic')
const createForm = ref({
  namespace: '',
  name: '',
  replicas: 1,
  image: '',
  ports: [80] as number[],
  envVars: [] as { key: string; value: string }[],
  cpuLimit: '',
  memoryLimit: '',
  cpuRequest: '',
  memoryRequest: '',
  labels: [] as { key: string; value: string }[],
  nodeSelector: [] as { key: string; value: string }[],
  serviceAccountName: '',
  restartPolicy: 'Always' as string,
})

const showCreate = () => {
  createTab.value = 'basic'
  createForm.value = {
    namespace: namespace.value || '',
    name: '',
    replicas: 1,
    image: '',
    ports: [80],
    envVars: [],
    cpuLimit: '',
    memoryLimit: '',
    cpuRequest: '',
    memoryRequest: '',
    labels: [],
    nodeSelector: [],
    serviceAccountName: '',
    restartPolicy: 'Always',
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
      ports: createForm.value.ports.filter(p => p > 0),
    }
    const envVars: Record<string, string> = {}
    createForm.value.envVars.forEach(e => { if (e.key) envVars[e.key] = e.value })
    if (Object.keys(envVars).length > 0) payload.envVars = envVars
    const labels: Record<string, string> = {}
    createForm.value.labels.forEach(l => { if (l.key) labels[l.key] = l.value })
    if (Object.keys(labels).length > 0) payload.labels = labels
    const nodeSelector: Record<string, string> = {}
    createForm.value.nodeSelector.forEach(n => { if (n.key) nodeSelector[n.key] = n.value })
    if (Object.keys(nodeSelector).length > 0) payload.nodeSelector = nodeSelector
    const resourceLimits: Record<string, string> = {}
    const resourceRequests: Record<string, string> = {}
    if (createForm.value.cpuLimit) resourceLimits.cpu = createForm.value.cpuLimit
    if (createForm.value.memoryLimit) resourceLimits.memory = createForm.value.memoryLimit
    if (createForm.value.cpuRequest) resourceRequests.cpu = createForm.value.cpuRequest
    if (createForm.value.memoryRequest) resourceRequests.memory = createForm.value.memoryRequest
    if (Object.keys(resourceLimits).length > 0) payload.resourceLimits = resourceLimits
    if (Object.keys(resourceRequests).length > 0) payload.resourceRequests = resourceRequests
    if (createForm.value.serviceAccountName) payload.serviceAccountName = createForm.value.serviceAccountName
    if (createForm.value.restartPolicy && createForm.value.restartPolicy !== 'Always') payload.restartPolicy = createForm.value.restartPolicy
    await api.post('/deployments/create', payload)
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
    const res: any = await api.get('/deployments', { params })
    deploys.value = (res.data || []).map((d: any) => ({ ...d, _restarting: false, _deleting: false }))
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

const scaleDialog = (row: any) => {
  scaleForm.value = { name: row.name, namespace: row.namespace, replicas: row.replicas || 1 }
  scaleVisible.value = true
}

const confirmScale = async () => {
  scaleLoading.value = true
  try {
    await api.post('/deployments/scale', scaleForm.value)
    ElMessage.success('扩缩容成功')
    scaleVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '扩缩容失败')
  } finally {
    scaleLoading.value = false
  }
}

const restartDeploy = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要重启 Deployment "${row.name}" 吗？`, '确认重启', { type: 'warning' })
    row._restarting = true
    await api.post('/deployments/restart', { namespace: row.namespace, name: row.name })
    ElMessage.success('重启请求已发送')
    setTimeout(fetchData, 2000)
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e.response?.data?.message || '重启失败')
  } finally {
    row._restarting = false
  }
}

const rollbackDialog = (row: any) => {
  rollbackForm.value = { name: row.name, namespace: row.namespace, revision: 1 }
  rollbackVisible.value = true
}

const confirmRollback = async () => {
  rollbackLoading.value = true
  try {
    await api.post('/deployments/rollback', rollbackForm.value)
    ElMessage.success('回滚成功')
    rollbackVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '回滚失败')
  } finally {
    rollbackLoading.value = false
  }
}

const deleteDeploy = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Deployment "${row.name}" 吗？`, '确认删除', { type: 'error' })
    row._deleting = true
    await api.delete('/deployments/delete', { params: { namespace: row.namespace, name: row.name } })
    ElMessage.success('已删除')
    fetchData()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e.response?.data?.message || '删除失败')
  } finally {
    row._deleting = false
  }
}

const editDialog = async (row: any) => {
  editVisible.value = true
  editLoading.value = true
  editForm.value = {
    namespace: row.namespace,
    name: row.name,
    containerName: '',
    replicas: row.replicas || 1,
    image: '',
    envVars: [],
    cpuLimit: '',
    memoryLimit: '',
    cpuRequest: '',
    memoryRequest: '',
    containers: [],
  }
  try {
    const res: any = await api.get('/deployments/get', { params: { namespace: row.namespace, name: row.name } })
    const dep = res.data
    const containers = (dep?.spec?.template?.spec?.containers || []).map((c: any) => ({
      name: c.name,
      image: c.image || '',
      env: (c.env || []).map((e: any) => ({ name: e.name, value: e.value || '' })),
      resources: c.resources || {},
    }))
    editForm.value.containers = containers
    if (containers.length > 0) {
      const first = containers[0]
      editForm.value.containerName = first.name
      editForm.value.image = first.image
      editForm.value.envVars = first.env.map((e: any) => ({ key: e.name, value: e.value }))
      editForm.value.cpuLimit = first.resources.limits?.cpu || ''
      editForm.value.memoryLimit = first.resources.limits?.memory || ''
      editForm.value.cpuRequest = first.resources.requests?.cpu || ''
      editForm.value.memoryRequest = first.resources.requests?.memory || ''
    }
    editForm.value.replicas = dep?.spec?.replicas || row.replicas || 1
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '获取详情失败')
  } finally {
    editLoading.value = false
  }
}

const onContainerChange = (name: string) => {
  const c = editForm.value.containers.find((x) => x.name === name)
  if (!c) return
  editForm.value.image = c.image
  editForm.value.envVars = c.env.map((e) => ({ key: e.name, value: e.value }))
  editForm.value.cpuLimit = c.resources.limits?.cpu || ''
  editForm.value.memoryLimit = c.resources.limits?.memory || ''
  editForm.value.cpuRequest = c.resources.requests?.cpu || ''
  editForm.value.memoryRequest = c.resources.requests?.memory || ''
}

const addEnvVar = () => {
  editForm.value.envVars.push({ key: '', value: '' })
}

const removeEnvVar = (idx: number) => {
  editForm.value.envVars.splice(idx, 1)
}

const addCreateEnvVar = () => { createForm.value.envVars.push({ key: '', value: '' }) }
const removeCreateEnvVar = (idx: number) => { createForm.value.envVars.splice(idx, 1) }
const addCreateLabel = () => { createForm.value.labels.push({ key: '', value: '' }) }
const removeCreateLabel = (idx: number) => { createForm.value.labels.splice(idx, 1) }
const addCreateNodeSelector = () => { createForm.value.nodeSelector.push({ key: '', value: '' }) }
const removeCreateNodeSelector = (idx: number) => { createForm.value.nodeSelector.splice(idx, 1) }

const confirmEdit = async () => {
  editSaving.value = true
  try {
    const envVars: Record<string, string> = {}
    editForm.value.envVars.forEach((e) => {
      if (e.key) envVars[e.key] = e.value
    })
    const payload: any = {
      namespace: editForm.value.namespace,
      name: editForm.value.name,
      replicas: editForm.value.replicas,
      image: editForm.value.image,
      containerName: editForm.value.containerName,
    }
    if (Object.keys(envVars).length > 0) payload.envVars = envVars
    const resourceLimits: Record<string, string> = {}
    const resourceRequests: Record<string, string> = {}
    if (editForm.value.cpuLimit) resourceLimits.cpu = editForm.value.cpuLimit
    if (editForm.value.memoryLimit) resourceLimits.memory = editForm.value.memoryLimit
    if (editForm.value.cpuRequest) resourceRequests.cpu = editForm.value.cpuRequest
    if (editForm.value.memoryRequest) resourceRequests.memory = editForm.value.memoryRequest
    if (Object.keys(resourceLimits).length > 0) payload.resourceLimits = resourceLimits
    if (Object.keys(resourceRequests).length > 0) payload.resourceRequests = resourceRequests
    await api.put('/deployments/update', payload)
    ElMessage.success('更新成功')
    editVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '更新失败')
  } finally {
    editSaving.value = false
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
.deployments-page {
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

.cell-metric {
  font-variant-numeric: tabular-nums;
  font-weight: 500;
}

.metric-ok {
  color: var(--success);
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
  background: rgba(139, 92, 246, 0.12);
  color: #a78bfa;
}

.action-edit:hover:not(:disabled) {
  background: rgba(139, 92, 246, 0.25);
}

.action-scale {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
}

.action-scale:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.25);
}

.action-restart {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
}

.action-restart:hover:not(:disabled) {
  background: rgba(148, 163, 184, 0.25);
}

.action-rollback {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
}

.action-rollback:hover:not(:disabled) {
  background: rgba(34, 197, 94, 0.25);
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

.btn-warning {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  color: white;
}

.btn-warning:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.4);
}

.btn-warning:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* 标签列表 */
.label-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

/* 暗色弹窗覆盖 */
.deployments-page :deep(.dark-dialog .el-dialog),
.deployments-page :deep(.el-dialog.dark-dialog) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.deployments-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.deployments-page :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  padding: 20px 24px;
  margin: 0;
}

.deployments-page :deep(.el-dialog__title) {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 18px;
}

.deployments-page :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--text-secondary);
}

.deployments-page :deep(.el-dialog__body) {
  padding: 24px;
  color: var(--text-primary);
}

.deployments-page :deep(.el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.08);
  padding: 16px 24px;
}

.deployments-page :deep(.el-descriptions) {
  --el-descriptions-item-bordered-label-background: rgba(30, 41, 59, 0.8);
}

.deployments-page :deep(.el-descriptions__label) {
  color: var(--text-secondary);
}

.deployments-page :deep(.el-descriptions__content) {
  color: var(--text-primary);
}

.deployments-page :deep(.el-descriptions__cell) {
  border-color: rgba(148, 163, 184, 0.08) !important;
}

.deployments-page :deep(.el-divider__text) {
  background: #1e293b;
  color: var(--text-secondary);
}

.deployments-page :deep(.el-divider) {
  border-color: rgba(148, 163, 184, 0.08);
}

.deployments-page :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

.deployments-page :deep(.el-input-number .el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.deployments-page :deep(.el-input-number .el-input__inner) {
  color: var(--text-primary);
}

.deployments-page :deep(.el-input-number .el-input-number__decrease),
.deployments-page :deep(.el-input-number .el-input-number__increase) {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(148, 163, 184, 0.1);
  color: var(--text-secondary);
}

.deployments-page :deep(.el-input-number .el-input-number__decrease:hover),
.deployments-page :deep(.el-input-number .el-input-number__increase:hover) {
  color: var(--primary);
}

.deployments-page :deep(.el-tag--info) {
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
.deployments-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

/* 编辑弹窗表单 */
.edit-form :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.edit-form :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.edit-form :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.edit-form :deep(.el-input__inner) {
  color: var(--text-primary);
}

.edit-form :deep(.el-input__inner::placeholder) {
  color: rgba(148, 163, 184, 0.5);
}

.edit-form :deep(.el-input.is-disabled .el-input__wrapper) {
  background: rgba(51, 65, 85, 0.4);
  opacity: 0.7;
}

.edit-form :deep(.el-input.is-disabled .el-input__inner) {
  color: var(--text-secondary);
}

.edit-form :deep(.el-select .el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.edit-form :deep(.el-select-dropdown) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
}

.edit-form :deep(.el-select-dropdown__item) {
  color: var(--text-primary);
}

.edit-form :deep(.el-select-dropdown__item.hover) {
  background: rgba(99, 102, 241, 0.15);
}

.edit-form :deep(.el-select-dropdown__item.selected) {
  color: #6366f1;
}

/* 环境变量 */
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

/* 创建表单 */
.port-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
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
</style>
