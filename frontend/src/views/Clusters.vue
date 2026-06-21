<template>
  <div class="page-container">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">集群管理</h1>
        <span class="page-subtitle">管理 Kubernetes 集群连接</span>
      </div>
      <div class="header-actions">
        <button type="button" class="btn-gradient btn-create" @click="showAddDialog">
          <el-icon :size="16"><Plus /></el-icon>
          <span>添加集群</span>
        </button>
      </div>
    </div>

    <div class="glass-table-container">
      <el-table
        :data="clusters"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无集群连接'"
      >
        <el-table-column prop="name" label="集群名称" min-width="150">
          <template #default="{ row }">
            <div class="cluster-name">
              <el-icon :style="{ color: row.active ? '#4ade80' : '#94a3b8' }">
                <Connection />
              </el-icon>
              <span class="cell-name">{{ row.name }}</span>
              <span v-if="row.active" class="cell-status status-active">当前</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="server" label="API Server" min-width="250">
          <template #default="{ row }">
            <span class="cell-mono">{{ row.server || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <span class="cell-status" :class="row.status === 'connected' ? 'status-active' : 'status-inactive'">
              {{ row.status === 'connected' ? '已连接' : '未连接' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button v-if="!row.active" class="action-btn action-switch" @click="switchCluster(row.name)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/><polyline points="21 16 21 21 16 21"/><line x1="15" y1="15" x2="21" y2="21"/><line x1="4" y1="4" x2="9" y2="9"/></svg>
                切换
              </button>
              <button v-if="!row.active" class="action-btn action-delete" @click="removeCluster(row.name)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                移除
              </button>
              <span v-if="row.active" class="cell-active-label">当前集群</span>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="clusters.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无集群连接">
          <button type="button" class="btn-gradient" @click="showAddDialog">添加集群</button>
        </el-empty>
      </div>
    </div>

    <!-- 添加集群弹窗 -->
    <el-dialog
      v-model="addDialogVisible"
      title="添加集群"
      width="500px"
      :close-on-click-modal="false"
      class="dark-dialog"
    >
      <el-form :model="addForm" label-width="100px">
        <el-form-item label="集群名称" required>
          <el-input v-model="addForm.name" placeholder="my-cluster" />
        </el-form-item>
        <el-form-item label="连接方式">
          <el-radio-group v-model="addForm.method">
            <el-radio label="token">Token</el-radio>
            <el-radio label="kubeconfig">KubeConfig</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="addForm.method === 'token'" label="API Server" required>
          <el-input v-model="addForm.server" placeholder="https://k8s-server:6443" />
        </el-form-item>
        <el-form-item v-if="addForm.method === 'token'" label="Token">
          <el-input v-model="addForm.token" type="password" show-password placeholder="Bearer Token" />
        </el-form-item>
        <el-form-item v-if="addForm.method === 'kubeconfig'" label="KubeConfig" required>
          <el-input v-model="addForm.kubeconfig" type="textarea" :rows="8" placeholder="粘贴 kubeconfig 内容..." />
        </el-form-item>
      </el-form>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="addDialogVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-confirm" @click="addCluster" :disabled="adding">
          <span v-if="adding" class="btn-spinner"></span>
          确定
        </button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Connection } from '@element-plus/icons-vue'
import api from '../api'

interface ClusterInfo {
  name: string
  server: string
  status: string
  active: boolean
}

const clusters = ref<ClusterInfo[]>([])
const loading = ref(false)
const addDialogVisible = ref(false)
const adding = ref(false)

const addForm = ref({
  name: '',
  method: 'token',
  server: '',
  token: '',
  kubeconfig: '',
})

const fetchClusters = async () => {
  loading.value = true
  try {
    const res: any = await api.get('/clusters')
    if (res.success) {
      clusters.value = res.clusters || []
    }
  } catch (e) {
    console.error('Failed to fetch clusters:', e)
  } finally {
    loading.value = false
  }
}

const showAddDialog = () => {
  addForm.value = {
    name: '',
    method: 'token',
    server: '',
    token: '',
    kubeconfig: '',
  }
  addDialogVisible.value = true
}

const addCluster = async () => {
  if (!addForm.value.name) {
    ElMessage.warning('请输入集群名称')
    return
  }

  if (addForm.value.method === 'token' && !addForm.value.server) {
    ElMessage.warning('请输入 API Server 地址')
    return
  }

  if (addForm.value.method === 'kubeconfig' && !addForm.value.kubeconfig) {
    ElMessage.warning('请粘贴 KubeConfig 内容')
    return
  }

  adding.value = true
  try {
    const payload: any = {
      name: addForm.value.name,
    }

    if (addForm.value.method === 'token') {
      payload.server = addForm.value.server
      payload.token = addForm.value.token
    } else {
      payload.kubeconfig = addForm.value.kubeconfig
    }

    const res: any = await api.post('/clusters/add', payload)
    if (res.success) {
      ElMessage.success('集群添加成功')
      addDialogVisible.value = false
      fetchClusters()
    } else {
      ElMessage.error(res.message || '添加失败')
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '添加失败')
  } finally {
    adding.value = false
  }
}

const switchCluster = async (name: string) => {
  try {
    const res: any = await api.post('/clusters/switch', { name })
    if (res.success) {
      ElMessage.success(`已切换到集群: ${name}`)
      localStorage.setItem('cluster_name', name)
      fetchClusters()
    } else {
      ElMessage.error(res.message || '切换失败')
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '切换失败')
  }
}

const removeCluster = async (name: string) => {
  try {
    await ElMessageBox.confirm(
      `确定要移除集群 "${name}" 吗？`,
      '确认移除',
      { type: 'warning' }
    )

    const res: any = await api.post('/clusters/remove', { name })
    if (res.success) {
      ElMessage.success('集群已移除')
      fetchClusters()
    } else {
      ElMessage.error(res.message || '移除失败')
    }
  } catch (e: any) {
    if (e !== 'cancel') {
      ElMessage.error(e.response?.data?.message || '移除失败')
    }
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

onMounted(() => {
  fetchClusters()
})
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

.empty-state {
  padding: 40px 0;
}

.cluster-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.cell-name {
  font-weight: 500;
  color: #e2e8f0;
}

.cell-mono {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
  color: var(--text-secondary);
}

.cell-status {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.status-active {
  color: #4ade80;
  background: rgba(34, 197, 94, 0.12);
}

.status-inactive {
  color: #94a3b8;
  background: rgba(148, 163, 184, 0.12);
}

.cell-active-label {
  font-size: 12px;
  color: var(--text-secondary);
  opacity: 0.6;
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

.action-switch {
  background: rgba(99, 102, 241, 0.12);
  color: #818cf8;
}

.action-switch:hover:not(:disabled) {
  background: rgba(99, 102, 241, 0.25);
}

.action-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.action-delete:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
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

.page-container :deep(.dark-dialog .el-dialog),
.page-container :deep(.el-dialog.dark-dialog) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.page-container :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.page-container :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  padding: 20px 24px;
  margin: 0;
}

.page-container :deep(.el-dialog__title) {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 18px;
}

.page-container :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--text-secondary);
}

.page-container :deep(.el-dialog__body) {
  padding: 24px;
  color: var(--text-primary);
}

.page-container :deep(.el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.08);
  padding: 16px 24px;
}

.page-container :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

.page-container :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.page-container :deep(.el-input__inner) {
  color: var(--text-primary);
}

.page-container :deep(.el-textarea__inner) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
  color: var(--text-primary);
}

.page-container :deep(.el-radio__label) {
  color: var(--text-secondary);
}

.page-container :deep(.el-radio__input.is-checked + .el-radio__label) {
  color: var(--primary);
}

.page-container :deep(.el-radio__input.is-checked .el-radio__inner) {
  background: var(--primary);
  border-color: var(--primary);
}

.page-container :deep(.el-tag--success) {
  background: rgba(34, 197, 94, 0.12);
  border-color: rgba(34, 197, 94, 0.3);
  color: #4ade80;
}

.page-container :deep(.el-tag--info) {
  background: rgba(148, 163, 184, 0.12);
  border-color: rgba(148, 163, 184, 0.3);
  color: #94a3b8;
}

.page-container :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}
</style>
