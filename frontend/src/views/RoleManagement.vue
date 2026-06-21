<template>
  <div class="role-management-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">角色管理</h1>
        <span class="page-subtitle">管理系统角色和权限</span>
      </div>
      <div class="header-actions">
        <button type="button" class="btn-gradient" @click="showCreate">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          <span>新建角色</span>
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchData">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="glass-table-container">
      <el-table
        :data="roles"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无角色数据'"
      >
        <el-table-column prop="name" label="角色名称" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="cell-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="250">
          <template #default="{ row }">
            <span>{{ row.description || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="isPreset" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="row.isPreset === 1 ? 'success' : 'info'" size="small">
              {{ row.isPreset === 1 ? '预设' : '自定义' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="权限概览" min-width="200">
          <template #default="{ row }">
            <div class="permission-tags">
              <el-tag v-for="(value, key) in getPermissionSummary(row.permissions)" :key="key" 
                :type="value ? 'success' : 'info'" size="small" class="permission-tag">
                {{ getModuleName(key) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button type="button" class="action-btn action-edit" @click="editRole(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                编辑
              </button>
              <button type="button" class="action-btn action-delete" @click="deleteRole(row)" :disabled="row.isPreset === 1">
                删除
              </button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 创建/编辑角色弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit ? '编辑角色' : '新建角色'" width="700px" :close-on-click-modal="false" class="dark-dialog">
      <el-form label-width="100px" class="create-form">
        <el-form-item label="角色名称" required>
          <el-input v-model="form.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入角色描述" />
        </el-form-item>
        <el-divider content-position="left">模块权限</el-divider>
        <div class="permission-grid">
          <div v-for="(module, key) in form.permissions.modules" :key="key" class="permission-module" :class="{ 'module-active': module.view }">
            <div class="module-header">
              <div class="module-info">
                <span class="module-icon">{{ getModuleIcon(key) }}</span>
                <span class="module-name">{{ getModuleName(key) }}</span>
              </div>
              <div class="toggle-switch" :class="{ active: module.view }" @click="module.view = !module.view; toggleModuleView(key, module.view)">
                <div class="toggle-track" :style="{ background: module.view ? '#22c55e !important' : '#4b5563 !important' }">
                  <div class="toggle-thumb" :style="{ left: module.view ? '22px' : '2px', background: '#ffffff !important' }"></div>
                </div>
              </div>
            </div>
            <div class="module-operations" v-if="module.view">
              <div class="operation-row">
                <span class="operation-label">查看</span>
                <span class="operation-badge enabled">已启用</span>
              </div>
              <div class="operation-row">
                <span class="operation-label">创建</span>
                <div class="toggle-switch small" :class="{ active: module.create }" @click="module.create = !module.create">
                  <div class="toggle-track" :style="{ background: module.create ? '#22c55e !important' : '#4b5563 !important' }">
                    <div class="toggle-thumb" :style="{ left: module.create ? '18px' : '2px', background: '#ffffff !important' }"></div>
                  </div>
                </div>
              </div>
              <div class="operation-row">
                <span class="operation-label">编辑</span>
                <div class="toggle-switch small" :class="{ active: module.edit }" @click="module.edit = !module.edit">
                  <div class="toggle-track" :style="{ background: module.edit ? '#22c55e !important' : '#4b5563 !important' }">
                    <div class="toggle-thumb" :style="{ left: module.edit ? '18px' : '2px', background: '#ffffff !important' }"></div>
                  </div>
                </div>
              </div>
              <div class="operation-row">
                <span class="operation-label">删除</span>
                <div class="toggle-switch small" :class="{ active: module.delete }" @click="module.delete = !module.delete">
                  <div class="toggle-track" :style="{ background: module.delete ? '#22c55e !important' : '#4b5563 !important' }">
                    <div class="toggle-thumb" :style="{ left: module.delete ? '18px' : '2px', background: '#ffffff !important' }"></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-form>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="dialogVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-confirm" @click="handleSubmit" :disabled="submitLoading">
          <span v-if="submitLoading" class="btn-spinner"></span>
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
import api from '../api'

const roles = ref<any[]>([])
const loading = ref(false)

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitLoading = ref(false)
const editRoleId = ref<number | null>(null)

const defaultPermissions = {
  modules: {
    workloads: { view: true, create: false, edit: false, delete: false },
    network: { view: false, create: false, edit: false, delete: false },
    storage: { view: false, create: false, edit: false, delete: false },
    rbac: { view: false, create: false, edit: false, delete: false },
    usercenter: { view: false, create: false, edit: false, delete: false },
  }
}

const form = reactive({
  name: '',
  description: '',
  permissions: JSON.parse(JSON.stringify(defaultPermissions)),
})

const moduleNames: Record<string, string> = {
  workloads: '工作负载',
  network: '网络',
  storage: '存储',
  rbac: 'RBAC',
  usercenter: '用户中心',
}

const moduleIcons: Record<string, string> = {
  workloads: '🚀',
  network: '🌐',
  storage: '💾',
  rbac: '🔐',
  usercenter: '👤',
}

const getModuleName = (key: string) => moduleNames[key] || key
const getModuleIcon = (key: string) => moduleIcons[key] || '📦'

const getPermissionSummary = (permissionsStr: string) => {
  try {
    const permissions = JSON.parse(permissionsStr)
    return permissions.modules || {}
  } catch {
    return {}
  }
}

const toggleModuleView = (module: string, view: boolean) => {
  if (!view) {
    form.permissions.modules[module].create = false
    form.permissions.modules[module].edit = false
    form.permissions.modules[module].delete = false
  }
}

const showCreate = () => {
  isEdit.value = false
  editRoleId.value = null
  form.name = ''
  form.description = ''
  form.permissions = JSON.parse(JSON.stringify(defaultPermissions))
  dialogVisible.value = true
}

const editRole = (row: any) => {
  isEdit.value = true
  editRoleId.value = row.id
  form.name = row.name
  form.description = row.description || ''
  try {
    form.permissions = JSON.parse(row.permissions)
  } catch {
    form.permissions = JSON.parse(JSON.stringify(defaultPermissions))
  }
  dialogVisible.value = true
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await api.get('/auth/roles')
    roles.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  if (!form.name) {
    ElMessage.warning('请填写角色名称')
    return
  }
  submitLoading.value = true
  try {
    const payload = {
      name: form.name,
      description: form.description,
      permissions: JSON.stringify(form.permissions),
    }
    if (isEdit.value) {
      await api.post('/auth/roles/update', { roleId: editRoleId.value, ...payload })
      ElMessage.success('更新成功')
    } else {
      await api.post('/auth/roles/create', payload)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '操作失败')
  } finally {
    submitLoading.value = false
  }
}

const deleteRole = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除角色 ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.post('/auth/roles/delete', { roleId: row.id })
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

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.role-management-page {
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

.cell-name {
  font-weight: 500;
  color: #e2e8f0;
}

.permission-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.permission-tag {
  font-size: 11px;
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

.glass-table-container :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(4px);
}

.glass-table-container :deep(.el-loading-spinner .circular) {
  stroke: var(--primary);
}

.role-management-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.role-management-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.role-management-page :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  padding: 20px 24px;
  margin: 0;
}

.role-management-page :deep(.el-dialog__title) {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 18px;
}

.role-management-page :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--text-secondary);
}

.role-management-page :deep(.el-dialog__body) {
  padding: 24px;
}

.role-management-page :deep(.el-dialog__footer) {
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
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.create-form :deep(.el-input__inner) {
  color: var(--text-primary);
}

.create-form :deep(.el-input__inner::placeholder) {
  color: rgba(148, 163, 184, 0.5);
}

.create-form :deep(.el-textarea__inner) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  color: var(--text-primary);
}

.create-form :deep(.el-textarea__inner:focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.create-form :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

.permission-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.permission-module {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
  padding: 16px;
  transition: all 0.2s ease;
}

.permission-module:hover {
  border-color: rgba(99, 102, 241, 0.3);
}

.permission-module.module-active {
  border-color: rgba(99, 102, 241, 0.4);
  background: rgba(99, 102, 241, 0.05);
}

.module-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.module-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.module-icon {
  font-size: 20px;
}

.module-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 14px;
}

.module-operations {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid rgba(148, 163, 184, 0.1);
}

.operation-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
}

.operation-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.operation-badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
}

.operation-badge.enabled {
  background: rgba(34, 197, 94, 0.15);
  color: #4ade80;
}

.toggle-switch {
  cursor: pointer;
  user-select: none;
}

.toggle-track {
  position: relative;
  width: 44px;
  height: 24px;
  border-radius: 12px;
  transition: background 0.3s ease;
}

.toggle-switch.small .toggle-track {
  width: 36px;
  height: 20px;
  border-radius: 10px;
}

.toggle-thumb {
  position: absolute;
  top: 2px;
  width: 20px;
  height: 20px;
  background: #ffffff;
  border-radius: 50%;
  transition: left 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.toggle-switch.small .toggle-thumb {
  width: 16px;
  height: 16px;
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
