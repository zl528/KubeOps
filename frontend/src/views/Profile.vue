<template>
  <div class="profile-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">个人设置</h1>
        <span class="page-subtitle">管理您的个人信息</span>
      </div>
    </div>

    <div class="profile-content">
      <div class="profile-card">
        <h3 class="card-title">基本信息</h3>
        <el-form label-width="100px" class="profile-form">
          <el-form-item label="用户名">
            <el-input :model-value="userInfo.username" disabled />
          </el-form-item>
          <el-form-item label="显示名">
            <el-input v-model="userInfo.displayName" placeholder="请输入显示名" />
          </el-form-item>
          <el-form-item label="邮箱">
            <el-input v-model="userInfo.email" placeholder="请输入邮箱" />
          </el-form-item>
          <el-form-item label="角色">
            <el-tag :type="userInfo.role === 'admin' ? 'danger' : 'info'" size="large">
              {{ userInfo.roleName || userInfo.role }}
            </el-tag>
          </el-form-item>
          <el-form-item>
            <button type="button" class="btn-gradient" @click="updateProfile" :disabled="updating">
              <span v-if="updating" class="btn-spinner"></span>
              保存信息
            </button>
          </el-form-item>
        </el-form>
      </div>

      <div class="profile-card">
        <h3 class="card-title">修改密码</h3>
        <el-form label-width="100px" class="profile-form">
          <el-form-item label="当前密码">
            <el-input v-model="passwordForm.oldPassword" type="password" placeholder="请输入当前密码" show-password />
          </el-form-item>
          <el-form-item label="新密码">
            <el-input v-model="passwordForm.newPassword" type="password" placeholder="请输入新密码" show-password />
          </el-form-item>
          <el-form-item label="确认密码">
            <el-input v-model="passwordForm.confirmPassword" type="password" placeholder="请再次输入新密码" show-password />
          </el-form-item>
          <el-form-item>
            <button type="button" class="btn-gradient" @click="updatePassword" :disabled="changingPassword">
              <span v-if="changingPassword" class="btn-spinner"></span>
              修改密码
            </button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'

const userInfo = reactive({
  id: 0,
  username: '',
  displayName: '',
  email: '',
  role: '',
  roleName: '',
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
})

const updating = ref(false)
const changingPassword = ref(false)

const fetchUserInfo = async () => {
  try {
    const res: any = await api.get('/auth/me')
    const user = res.data
    userInfo.id = user.id
    userInfo.username = user.username
    userInfo.displayName = user.displayName || ''
    userInfo.email = user.email || ''
    userInfo.role = user.role
    userInfo.roleName = user.roleName || ''
  } catch (e) {
    console.error(e)
  }
}

const updateProfile = async () => {
  updating.value = true
  try {
    await api.post('/auth/users/update', {
      userId: userInfo.id,
      displayName: userInfo.displayName,
      email: userInfo.email,
    })
    ElMessage.success('保存成功')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '保存失败')
  } finally {
    updating.value = false
  }
}

const updatePassword = async () => {
  if (!passwordForm.oldPassword || !passwordForm.newPassword) {
    ElMessage.warning('请填写当前密码和新密码')
    return
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }
  changingPassword.value = true
  try {
    await api.post('/auth/password', {
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword,
    })
    ElMessage.success('密码修改成功')
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '密码修改失败')
  } finally {
    changingPassword.value = false
  }
}

onMounted(() => {
  fetchUserInfo()
})
</script>

<style scoped>
.profile-page {
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

.profile-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.profile-card {
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
}

.card-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 24px 0;
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
}

.profile-form :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.profile-form :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.profile-form :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.profile-form :deep(.el-input__inner) {
  color: var(--text-primary);
}

.profile-form :deep(.el-input__inner::placeholder) {
  color: rgba(148, 163, 184, 0.5);
}

.profile-form :deep(.el-form-item__label) {
  color: var(--text-secondary);
}

.btn-gradient {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border: none;
  padding: 10px 24px;
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

.btn-gradient:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
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
