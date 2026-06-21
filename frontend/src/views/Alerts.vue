<template>
  <div class="alerts-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">告警管理</h1>
        <span class="page-subtitle">监控集群告警规则和告警历史</span>
      </div>
      <div class="header-actions">
        <button type="button" class="btn-gradient" @click="showCreateRule">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          <span>新建规则</span>
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchRules">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="stats-row">
      <div class="stat-card glass-card">
        <div class="stat-icon total-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
        </div>
        <div class="stat-value">{{ stats.total || 0 }}</div>
        <div class="stat-label">告警总数</div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon firing-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z"/></svg>
        </div>
        <div class="stat-value firing">{{ stats.firing || 0 }}</div>
        <div class="stat-label">触发中</div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon resolved-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
        </div>
        <div class="stat-value resolved">{{ stats.resolved || 0 }}</div>
        <div class="stat-label">已解决</div>
      </div>
      <div class="stat-card glass-card">
        <div class="stat-icon rules-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
        </div>
        <div class="stat-value">{{ rules.length }}</div>
        <div class="stat-label">告警规则</div>
      </div>
    </div>

    <div class="glass-table-container">
      <div class="table-header">
        <span class="table-title">告警规则</span>
      </div>
      <el-table :data="rules" v-loading="loading" :header-cell-style="headerCellStyle" :cell-style="cellStyle" :row-class-name="rowClassName" class="custom-table" :empty-text="'暂无告警规则'">
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="cell-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="metric" label="指标" width="150" />
        <el-table-column prop="condition" label="条件" width="120" />
        <el-table-column prop="threshold" label="阈值" width="100" />
        <el-table-column prop="severity" label="级别" width="100">
          <template #default="{ row }">
            <el-tag :type="severityType(row.severity)" size="small">{{ row.severity }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" label="状态" width="80">
          <template #default="{ row }">
            <div class="switch-wrapper">
              <el-switch v-model="row.enabled" @change="toggleRule(row)" />
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button type="button" class="action-btn action-edit" @click="editRule(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                编辑
              </button>
              <button type="button" class="action-btn action-delete" @click="handleDeleteRule(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                删除
              </button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="glass-table-container" style="margin-top: 24px;">
      <div class="table-header">
        <span class="table-title">告警历史</span>
        <div class="filter-group">
          <el-select v-model="filterSeverity" placeholder="筛选级别" clearable @change="fetchAlerts" class="severity-filter">
            <el-option label="全部" value="" />
            <el-option label="Warning" value="warning" />
            <el-option label="Critical" value="critical" />
          </el-select>
        </div>
      </div>
      <el-table :data="alerts" v-loading="alertsLoading" :header-cell-style="headerCellStyle" :cell-style="cellStyle" :row-class-name="rowClassName" class="custom-table" :empty-text="'暂无告警历史'">
        <el-table-column prop="ruleName" label="规则名称" min-width="200" show-overflow-tooltip />
        <el-table-column prop="severity" label="级别" width="100">
          <template #default="{ row }">
            <el-tag :type="severityType(row.severity)" size="small">{{ row.severity }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="message" label="消息" min-width="250" show-overflow-tooltip />
        <el-table-column prop="value" label="当前值" width="100" />
        <el-table-column prop="threshold" label="阈值" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'firing' ? 'danger' : 'success'" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="firedAt" label="触发时间" width="180" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button v-if="row.status === 'firing'" type="button" class="action-btn action-resolve" @click="handleResolve(row)">
                解决
              </button>
              <button type="button" class="action-btn action-delete" @click="handleDeleteAlert(row)">
                删除
              </button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 创建/编辑规则弹窗 -->
    <el-dialog v-model="ruleDialogVisible" :title="editingRule ? '编辑规则' : '新建规则'" width="550px" :close-on-click-modal="false" class="dark-dialog">
      <el-form :model="ruleForm" label-width="100px" :rules="ruleFormRules" ref="ruleFormRef" class="create-form">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="ruleForm.name" placeholder="CPU使用率告警" />
        </el-form-item>
        <el-form-item label="监控指标" prop="metric">
          <el-select v-model="ruleForm.metric" placeholder="选择指标" style="width: 100%">
            <el-option label="CPU 使用率" value="cpu_usage" />
            <el-option label="内存使用率" value="memory_usage" />
            <el-option label="磁盘使用率" value="disk_usage" />
            <el-option label="Pod 重启次数" value="pod_restarts" />
            <el-option label="网络接收字节" value="network_rx" />
            <el-option label="网络发送字节" value="network_tx" />
          </el-select>
        </el-form-item>
        <el-form-item label="触发条件" prop="condition">
          <el-select v-model="ruleForm.condition" placeholder="选择条件" style="width: 100%">
            <el-option label="大于 (>)" value="gt" />
            <el-option label="大于等于 (>=)" value="gte" />
            <el-option label="小于 (<)" value="lt" />
            <el-option label="小于等于 (<=)" value="lte" />
            <el-option label="等于 (=)" value="eq" />
          </el-select>
        </el-form-item>
        <el-form-item label="阈值" prop="threshold">
          <el-input-number v-model="ruleForm.threshold" :min="0" :max="1000000" :step="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="告警级别" prop="severity">
          <el-radio-group v-model="ruleForm.severity">
            <el-radio label="info">Info</el-radio>
            <el-radio label="warning">Warning</el-radio>
            <el-radio label="critical">Critical</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="ruleForm.description" type="textarea" :rows="2" placeholder="可选描述信息" />
        </el-form-item>
      </el-form>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="ruleDialogVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-confirm" @click="saveRule" :disabled="ruleSaving">
          <span v-if="ruleSaving" class="btn-spinner"></span>
          {{ editingRule ? '保存' : '创建' }}
        </button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import api from '../api'

const rules = ref<any[]>([])
const alerts = ref<any[]>([])
const stats = ref<any>({})
const loading = ref(false)
const alertsLoading = ref(false)
const filterSeverity = ref('')

const ruleDialogVisible = ref(false)
const ruleSaving = ref(false)
const editingRule = ref<any>(null)
const ruleFormRef = ref<FormInstance>()

const ruleForm = reactive({
  name: '',
  metric: '',
  condition: 'gt',
  threshold: 80,
  severity: 'warning',
  description: '',
})

const ruleFormRules: FormRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  metric: [{ required: true, message: '请选择监控指标', trigger: 'change' }],
  condition: [{ required: true, message: '请选择触发条件', trigger: 'change' }],
  threshold: [{ required: true, message: '请输入阈值', trigger: 'blur' }],
  severity: [{ required: true, message: '请选择告警级别', trigger: 'change' }],
}

const severityType = (s: string) => {
  const map: Record<string, string> = { warning: 'warning', critical: 'danger', info: 'info' }
  return map[s] || 'info'
}

const fetchRules = async () => {
  loading.value = true
  try {
    const res: any = await api.get('/alerts/rules')
    rules.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const fetchAlerts = async () => {
  alertsLoading.value = true
  try {
    const params = filterSeverity.value ? { severity: filterSeverity.value } : {}
    const res: any = await api.get('/alerts/history', { params })
    alerts.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    alertsLoading.value = false
  }
}

const fetchStats = async () => {
  try {
    const res: any = await api.get('/alerts/stats')
    stats.value = res.data || {}
  } catch (e) {
    console.error(e)
  }
}

const resetRuleForm = () => {
  ruleForm.name = ''
  ruleForm.metric = ''
  ruleForm.condition = 'gt'
  ruleForm.threshold = 80
  ruleForm.severity = 'warning'
  ruleForm.description = ''
  editingRule.value = null
}

const showCreateRule = () => {
  resetRuleForm()
  ruleDialogVisible.value = true
}

const editRule = (row: any) => {
  editingRule.value = row
  ruleForm.name = row.name
  ruleForm.metric = row.metric
  ruleForm.condition = row.condition
  ruleForm.threshold = row.threshold
  ruleForm.severity = row.severity
  ruleForm.description = row.description || ''
  ruleDialogVisible.value = true
}

const saveRule = async () => {
  if (!ruleFormRef.value) return
  await ruleFormRef.value.validate()

  ruleSaving.value = true
  try {
    if (editingRule.value) {
      await api.post('/alerts/rules/update', {
        id: editingRule.value.id,
        ...ruleForm,
      })
      ElMessage.success('规则已更新')
    } else {
      await api.post('/alerts/rules/create', { ...ruleForm })
      ElMessage.success('规则已创建')
    }
    ruleDialogVisible.value = false
    fetchRules()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '操作失败')
  } finally {
    ruleSaving.value = false
  }
}

const toggleRule = async (row: any) => {
  try {
    await api.post(`/alerts/rules/enable?id=${row.id}&enabled=${row.enabled}`)
    ElMessage.success('更新成功')
  } catch (e) {
    row.enabled = !row.enabled
    ElMessage.error('更新失败')
  }
}

const handleDeleteRule = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除规则 "${row.name}" 吗？`, '确认', { type: 'warning' })
    await api.delete(`/alerts/rules/delete?id=${row.id}`)
    ElMessage.success('删除成功')
    fetchRules()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

const handleResolve = async (row: any) => {
  try {
    await api.post(`/alerts/history/resolve?id=${row.id}`)
    ElMessage.success('已标记为解决')
    fetchAlerts()
  } catch (e) {
    ElMessage.error('操作失败')
  }
}

const handleDeleteAlert = async (row: any) => {
  try {
    await ElMessageBox.confirm('确定要删除此告警记录吗？', '确认', { type: 'warning' })
    await api.delete(`/alerts/history/delete?id=${row.id}`)
    ElMessage.success('删除成功')
    fetchAlerts()
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

onMounted(() => {
  fetchRules()
  fetchAlerts()
  fetchStats()
})
</script>

<style scoped>
.alerts-page {
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

.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.glass-card {
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 12px;
  padding: 20px;
}

.stat-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.total-icon {
  background: rgba(99, 102, 241, 0.15);
  color: #818cf8;
}

.firing-icon {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

.resolved-icon {
  background: rgba(34, 197, 94, 0.15);
  color: #4ade80;
}

.rules-icon {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--text-primary);
}

.stat-value.firing {
  color: #f87171;
}

.stat-value.resolved {
  color: #4ade80;
}

.stat-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.glass-table-container {
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
}

.table-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 8px;
}

.severity-filter {
  width: 140px;
}

.severity-filter :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.severity-filter :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.severity-filter :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
}

.severity-filter :deep(.el-input__inner) {
  color: var(--text-primary);
}

.glass-table-container :deep(.el-table) {
  background: transparent;
  --el-table-bg-color: transparent;
  --el-table-header-bg-color: rgba(30, 41, 59, 0.9);
  --el-table-header-text-color: #94a3b8;
  --el-table-text-color: #f1f5f9;
  --el-table-border-color: rgba(148, 163, 184, 0.06);
  --el-table-row-hover-bg-color: rgba(51, 65, 85, 0.4);
}

.glass-table-container :deep(.el-table th.el-table__cell) {
  background: rgba(30, 41, 59, 0.9) !important;
  color: #94a3b8 !important;
}

.glass-table-container :deep(.el-table td.el-table__cell) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.06) !important;
}

.cell-name {
  font-weight: 500;
  color: #e2e8f0;
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

.action-edit {
  background: rgba(99, 102, 241, 0.12);
  color: #818cf8;
}

.action-edit:hover {
  background: rgba(99, 102, 241, 0.25);
}

.action-resolve {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
}

.action-resolve:hover {
  background: rgba(34, 197, 94, 0.25);
}

.action-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.action-delete:hover {
  background: rgba(239, 68, 68, 0.25);
}

.alerts-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.alerts-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.alerts-page :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  padding: 20px 24px;
}

.alerts-page :deep(.el-dialog__title) {
  color: var(--text-primary);
}

.alerts-page :deep(.el-dialog__body) {
  padding: 24px;
}

.alerts-page :deep(.el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.08);
  padding: 16px 24px;
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
}

.create-form :deep(.el-input__inner) {
  color: var(--text-primary);
}

.create-form :deep(.el-select .el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
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

.create-form :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

.create-form :deep(.el-textarea__inner) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  color: var(--text-primary);
}

.create-form :deep(.el-radio__label) {
  color: var(--text-secondary);
}

.create-form :deep(.el-radio__input.is-checked + .el-radio__label) {
  color: #818cf8;
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

.switch-wrapper :deep(.el-switch__core) {
  background: #4b5563 !important;
  border-color: #6b7280 !important;
}

.switch-wrapper :deep(.el-switch.is-checked .el-switch__core) {
  background: #22c55e !important;
  border-color: #16a34a !important;
}

.switch-wrapper :deep(.el-switch__action) {
  background: #ffffff !important;
}
</style>
