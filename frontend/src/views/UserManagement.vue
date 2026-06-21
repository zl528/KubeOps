<template>
  <div class="user-center-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">用户管理</h1>
        <span class="page-subtitle">管理系统用户和权限</span>
      </div>
      <div class="header-actions">
        <button type="button" class="btn-gradient" @click="showCreate">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          <span>新建用户</span>
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchData">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="glass-table-container">
      <el-table
        :data="users"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无用户数据'"
      >
        <el-table-column prop="username" label="用户名" min-width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="cell-name">{{ row.username }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="displayName" label="显示名" width="150">
          <template #default="{ row }">
            <span>{{ row.displayName || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="email" label="邮箱" width="200">
          <template #default="{ row }">
            <span>{{ row.email || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="roleName" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : row.roleName === '开发者' ? 'warning' : 'info'" size="small">
              {{ row.roleName || row.role }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button type="button" class="action-btn action-edit" @click="editUser(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                编辑
              </button>
              <button type="button" class="action-btn" :class="row.status === 1 ? 'action-warning' : 'action-success'" @click="toggleStatus(row)" :disabled="row.username === 'admin'">
                {{ row.status === 1 ? '禁用' : '启用' }}
              </button>
              <button type="button" class="action-btn action-delete" @click="deleteUser(row)" :disabled="row.username === 'admin'">
                删除
              </button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 创建用户弹窗 -->
    <el-dialog v-model="createVisible" title="新建用户" width="600px" :close-on-click-modal="false" class="dark-dialog">
      <el-form label-width="100px" class="create-form">
        <el-form-item label="用户名" required>
          <el-input v-model="createForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="密码" required>
          <el-input v-model="createForm.password" type="password" placeholder="请输入密码" show-password />
        </el-form-item>
        <el-form-item label="显示名">
          <el-input v-model="createForm.displayName" placeholder="请输入显示名" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="createForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色" required>
          <el-select v-model="createForm.roleId" placeholder="请选择角色" style="width: 100%">
            <el-option v-for="role in roles" :key="role.id" :label="role.name" :value="role.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="可访问集群">
          <el-select v-model="createForm.clusterNames" multiple placeholder="选择可访问的集群" style="width: 100%">
            <el-option v-for="cluster in clusters" :key="cluster.name" :label="cluster.name" :value="cluster.name" />
          </el-select>
          <div class="form-tip">管理员自动拥有所有集群访问权限</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <button type="button" class="btn-dialog btn-cancel" @click="createVisible = false">取消</button>
        <button type="button" class="btn-dialog btn-confirm" @click="handleCreate" :disabled="createLoading">
          <span v-if="createLoading" class="btn-spinner"></span>
          创建
        </button>
      </template>
    </el-dialog>

    <!-- 编辑用户弹窗 -->
    <el-dialog v-model="editVisible" title="编辑用户" width="600px" :close-on-click-modal="false" class="dark-dialog">
      <el-form label-width="100px" class="create-form">
        <el-form-item label="用户名">
          <el-input :model-value="editForm.username" disabled />
        </el-form-item>
        <el-form-item label="显示名">
          <el-input v-model="editForm.displayName" placeholder="请输入显示名" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="editForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="editForm.roleId" placeholder="请选择角色" style="width: 100%">
            <el-option v-for="role in roles" :key="role.id" :label="role.name" :value="role.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="可访问集群">
          <el-select v-model="editForm.clusterNames" multiple placeholder="选择可访问的集群" style="width: 100%">
            <el-option v-for="cluster in clusters" :key="cluster.name" :label="cluster.name" :value="cluster.name" />
          </el-select>
          <div class="form-tip">管理员自动拥有所有集群访问权限</div>
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="editForm.password" type="password" placeholder="留空则不修改密码" show-password />
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
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const users = ref<any[]>([])
const roles = ref<any[]>([])
const clusters = ref<any[]>([])
const loading = ref(false)

const createVisible = ref(false)
const createLoading = ref(false)
const createForm = ref({
  username: '',
  password: '',
  displayName: '',
  email: '',
  roleId: null as number | null,
  clusterNames: [] as string[],
})

const editVisible = ref(false)
const editLoading = ref(false)
const editForm = ref({
  userId: 0,
  username: '',
  displayName: '',
  email: '',
  roleId: null as number | null,
  clusterNames: [] as string[],
  password: '',
})

const showCreate = () => {
  createForm.value = {
    username: '',
    password: '',
    displayName: '',
    email: '',
    roleId: null,
    clusterNames: [],
  }
  createVisible.value = true
}

const editUser = async (row: any) => {
  editForm.value = {
    userId: row.id,
    username: row.username,
    displayName: row.displayName || '',
    email: row.email || '',
    roleId: row.roleId,
    clusterNames: [],
    password: '',
  }
  // Fetch user's cluster assignments
  try {
    const res: any = await api.get(`/auth/users/clusters?userId=${row.id}`)
    editForm.value.clusterNames = res.data || []
  } catch (e) {
    console.error(e)
  }
  editVisible.value = true
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await api.get('/auth/users')
    users.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const fetchRoles = async () => {
  try {
    const res: any = await api.get('/auth/roles')
    roles.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const fetchClusters = async () => {
  try {
    const res: any = await api.get('/clusters')
    clusters.value = res.clusters || []
  } catch (e) {
    console.error(e)
  }
}

const handleCreate = async () => {
  if (!createForm.value.username || !createForm.value.password) {
    ElMessage.warning('请填写用户名和密码')
    return
  }
  createLoading.value = true
  try {
    const res: any = await api.post('/auth/users/create', createForm.value)
    // Set cluster access if provided
    if (createForm.value.clusterNames.length > 0 && res.data?.userId) {
      await api.post('/auth/users/clusters/set', {
        userId: res.data.userId,
        clusterNames: createForm.value.clusterNames,
      })
    }
    ElMessage.success('创建成功')
    createVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  } finally {
    createLoading.value = false
  }
}

const handleEdit = async () => {
  editLoading.value = true
  try {
    // Update user info
    await api.post('/auth/users/update', {
      userId: editForm.value.userId,
      displayName: editForm.value.displayName,
      email: editForm.value.email,
      roleId: editForm.value.roleId,
      password: editForm.value.password,
    })
    // Update cluster assignments
    await api.post('/auth/users/clusters/set', {
      userId: editForm.value.userId,
      clusterNames: editForm.value.clusterNames,
    })
    ElMessage.success('更新成功')
    editVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '更新失败')
  } finally {
    editLoading.value = false
  }
}

const toggleStatus = async (row: any) => {
  const action = row.status === 1 ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(`确定要${action}用户 ${row.username} 吗？`, '确认', { type: 'warning' })
    await api.post('/auth/users/update', {
      userId: row.id,
      status: row.status === 1 ? 0 : 1,
    })
    ElMessage.success(`${action}成功`)
    fetchData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error(`${action}失败`)
  }
}

const deleteUser = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除用户 ${row.username} 吗？`, '确认', { type: 'warning' })
    await api.post('/auth/users/delete', { userId: row.id })
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
  fetchRoles()
  fetchClusters()
})
</script>

<style scoped>
.user-center-page {
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

.action-warning {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
}

.action-warning:hover:not(:disabled) {
  background: rgba(245, 158, 11, 0.25);
}

.action-success {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
}

.action-success:hover:not(:disabled) {
  background: rgba(34, 197, 94, 0.25);
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

.user-center-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.user-center-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.user-center-page :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.08);
  padding: 20px 24px;
  margin: 0;
}

.user-center-page :deep(.el-dialog__title) {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 18px;
}

.user-center-page :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: var(--text-secondary);
}

.user-center-page :deep(.el-dialog__body) {
  padding: 24px;
}

.user-center-page :deep(.el-dialog__footer) {
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

.form-tip {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}
</style>
