<template>
  <div class="hpas-page">
    <!-- 渐变页头 -->
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">HPAs</h1>
        <span class="page-subtitle">管理集群中的水平自动伸缩</span>
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
        :data="hpas"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 HPA 数据'"
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
        <el-table-column prop="target" label="目标" min-width="200">
          <template #default="{ row }">
            <span class="cell-metric">{{ row.target }}</span>
          </template>
        </el-table-column>
        <el-table-column label="副本范围" width="150">
          <template #default="{ row }">
            <span class="cell-metric">{{ row.minReplicas }} - {{ row.maxReplicas }}</span>
          </template>
        </el-table-column>
        <el-table-column label="当前/期望" width="120">
          <template #default="{ row }">
            <span class="cell-metric" :class="{ 'metric-ok': row.currentReplicas === row.desiredReplicas }">{{ row.currentReplicas }} / {{ row.desiredReplicas }}</span>
          </template>
        </el-table-column>
        <el-table-column label="指标" min-width="250">
          <template #default="{ row }">
            <div v-for="(m, idx) in (row.metrics || [])" :key="idx">
              <el-tag size="small" style="margin: 2px">{{ m.type }}: {{ m.name }}</el-tag>
              <span class="cell-metric" v-if="m.target"> → {{ m.target }}</span>
            </div>
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

    <!-- 创建 HPA 弹窗 -->
    <el-dialog v-model="createVisible" title="新建 HPA" width="650px" class="dark-dialog">
      <el-form :model="form" label-width="110px">
        <el-form-item label="命名空间" required>
          <el-select v-model="form.namespace" placeholder="选择命名空间" style="width: 100%">
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="form.name" placeholder="my-hpa" />
        </el-form-item>
        <el-form-item label="目标类型" required>
          <el-select v-model="form.targetKind" placeholder="选择类型" style="width: 100%">
            <el-option label="Deployment" value="Deployment" />
            <el-option label="StatefulSet" value="StatefulSet" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标名称" required>
          <el-input v-model="form.targetName" placeholder="目标资源名称" />
        </el-form-item>
        <el-form-item label="最小副本" required>
          <el-input-number v-model="form.minReplicas" :min="1" :max="100" />
        </el-form-item>
        <el-form-item label="最大副本" required>
          <el-input-number v-model="form.maxReplicas" :min="1" :max="1000" />
        </el-form-item>
        <el-form-item label="指标名称">
          <el-select v-model="form.metricName" placeholder="选择指标" style="width: 100%">
            <el-option label="CPU" value="cpu" />
            <el-option label="Memory" value="memory" />
          </el-select>
        </el-form-item>
        <el-form-item label="目标利用率">
          <el-input v-model="form.metricTarget" placeholder="例如: 80" />
          <span style="margin-left: 8px; color: var(--text-secondary); font-size: 12px">%</span>
        </el-form-item>
        <el-form-item label="标签">
          <div v-for="(label, idx) in form.labels" :key="idx" class="kv-row">
            <el-input v-model="label.key" placeholder="键" style="width: 180px; margin-right: 8px" />
            <el-input v-model="label.value" placeholder="值" style="width: 180px; margin-right: 8px" />
            <button type="button" class="action-btn action-delete" @click="form.labels.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="btn-add-label" @click="form.labels.push({ key: '', value: '' })">+ 添加标签</button>
        </el-form-item>
        <el-divider content-position="left">扩缩容行为</el-divider>
        <el-form-item label="扩容稳定窗口">
          <el-input-number v-model="form.behavior.scaleUp.stabilizationWindowSeconds" :min="0" :max="3600" :step="60" />
          <span style="margin-left: 8px; color: var(--text-secondary); font-size: 12px">秒</span>
        </el-form-item>
        <el-form-item label="扩容策略">
          <div v-for="(policy, idx) in form.behavior.scaleUp.policies" :key="idx" class="policy-row">
            <el-select v-model="policy.type" style="width: 120px">
              <el-option label="Pods" value="Pods" />
              <el-option label="Percent" value="Percent" />
            </el-select>
            <el-input-number v-model="policy.value" :min="1" style="width: 120px" />
            <span style="color: var(--text-secondary); font-size: 12px">每</span>
            <el-input-number v-model="policy.periodSeconds" :min="1" :max="1800" style="width: 120px" />
            <span style="color: var(--text-secondary); font-size: 12px">秒</span>
            <button type="button" class="action-btn action-delete" @click="form.behavior.scaleUp.policies.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="btn-add-label" @click="form.behavior.scaleUp.policies.push({ type: 'Pods', value: 4, periodSeconds: 60 })">+ 添加策略</button>
        </el-form-item>
        <el-form-item label="缩容稳定窗口">
          <el-input-number v-model="form.behavior.scaleDown.stabilizationWindowSeconds" :min="0" :max="3600" :step="60" />
          <span style="margin-left: 8px; color: var(--text-secondary); font-size: 12px">秒</span>
        </el-form-item>
        <el-form-item label="缩容策略">
          <div v-for="(policy, idx) in form.behavior.scaleDown.policies" :key="idx" class="policy-row">
            <el-select v-model="policy.type" style="width: 120px">
              <el-option label="Pods" value="Pods" />
              <el-option label="Percent" value="Percent" />
            </el-select>
            <el-input-number v-model="policy.value" :min="1" style="width: 120px" />
            <span style="color: var(--text-secondary); font-size: 12px">每</span>
            <el-input-number v-model="policy.periodSeconds" :min="1" :max="1800" style="width: 120px" />
            <span style="color: var(--text-secondary); font-size: 12px">秒</span>
            <button type="button" class="action-btn action-delete" @click="form.behavior.scaleDown.policies.splice(idx, 1)">删除</button>
          </div>
          <button type="button" class="btn-add-label" @click="form.behavior.scaleDown.policies.push({ type: 'Pods', value: 2, periodSeconds: 60 })">+ 添加策略</button>
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
import { useGlobalNamespace } from '../store/namespace'
import api from '../api'

const hpas = ref<any[]>([])
const nsList = ref<any[]>([])
const { namespace } = useGlobalNamespace()
const loading = ref(false)
const createVisible = ref(false)
const saving = ref(false)

const form = reactive({
  namespace: '',
  name: '',
  targetKind: 'Deployment',
  targetName: '',
  minReplicas: 1,
  maxReplicas: 10,
  metricName: 'cpu',
  metricTarget: '',
  labels: [] as { key: string; value: string }[],
  behavior: {
    scaleUp: {
      stabilizationWindowSeconds: 0,
      policies: [] as { type: string; value: number; periodSeconds: number }[],
    },
    scaleDown: {
      stabilizationWindowSeconds: 300,
      policies: [] as { type: string; value: number; periodSeconds: number }[],
    },
  },
})

const showCreate = () => {
  form.namespace = namespace.value || ''
  form.name = ''
  form.targetKind = 'Deployment'
  form.targetName = ''
  form.minReplicas = 1
  form.maxReplicas = 10
  form.metricName = 'cpu'
  form.metricTarget = ''
  form.labels = []
  form.behavior = {
    scaleUp: { stabilizationWindowSeconds: 0, policies: [] },
    scaleDown: { stabilizationWindowSeconds: 300, policies: [] },
  }
  createVisible.value = true
}

const handleCreate = async () => {
  if (!form.namespace || !form.name || !form.targetKind || !form.targetName) {
    ElMessage.warning('请填写必填字段')
    return
  }
  saving.value = true
  try {
    const labels: Record<string, string> = {}
    for (const l of form.labels) { if (l.key) labels[l.key] = l.value }
    const payload: any = {
      namespace: form.namespace,
      name: form.name,
      targetKind: form.targetKind,
      targetName: form.targetName,
      minReplicas: form.minReplicas,
      maxReplicas: form.maxReplicas,
      metricName: form.metricName,
      metricTarget: form.metricTarget,
      labels: Object.keys(labels).length ? labels : undefined,
    }
    const hasScaleUp = form.behavior.scaleUp.policies.length > 0
    const hasScaleDown = form.behavior.scaleDown.policies.length > 0
    if (hasScaleUp || hasScaleDown) {
      payload.behavior = {}
      if (hasScaleUp) {
        payload.behavior.scaleUp = {
          stabilizationWindowSeconds: form.behavior.scaleUp.stabilizationWindowSeconds,
          policies: form.behavior.scaleUp.policies,
        }
      }
      if (hasScaleDown) {
        payload.behavior.scaleDown = {
          stabilizationWindowSeconds: form.behavior.scaleDown.stabilizationWindowSeconds,
          policies: form.behavior.scaleDown.policies,
        }
      }
    }
    await api.post('/hpas/create', payload)
    ElMessage.success('创建成功')
    createVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  } finally {
    saving.value = false
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = namespace.value ? { namespace: namespace.value } : {}
    const res: any = await api.get('/hpas', { params })
    hpas.value = res.data || []
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
    await ElMessageBox.confirm(`确定要删除 HPA ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.delete(`/hpas/delete?namespace=${row.namespace}&name=${row.name}`)
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
.hpas-page {
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

.policy-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
</style>
