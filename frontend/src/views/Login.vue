<template>
  <div class="login-container">
    <div class="login-bg-orbs">
      <div class="orb orb-1"></div>
      <div class="orb orb-2"></div>
      <div class="orb orb-3"></div>
    </div>

    <div class="login-card">
      <div class="login-header">
        <div class="logo-icon">
          <el-icon size="40"><Monitor /></el-icon>
        </div>
        <h1>KubeOps</h1>
        <p>Kubernetes 运维管理平台</p>
      </div>

      <el-tabs v-model="activeTab" class="login-tabs">
        <el-tab-pane label="登录" name="login">
          <el-form :model="loginForm" label-position="top" @submit.prevent="handleLogin">
            <el-form-item label="用户名">
              <el-input v-model="loginForm.username" placeholder="admin" prefix-icon="User" size="large" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="loginForm.password" type="password" show-password placeholder="输入密码" prefix-icon="Lock" size="large" @keyup.enter="handleLogin" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="添加集群" name="cluster" v-if="isLoggedIn">
          <el-form :model="clusterForm" label-position="top">
            <el-form-item label="集群名称">
              <el-input v-model="clusterForm.name" placeholder="my-cluster" prefix-icon="Folder" />
            </el-form-item>
            <el-form-item label="连接方式">
              <el-radio-group v-model="clusterForm.method">
                <el-radio label="token">Token</el-radio>
                <el-radio label="kubeconfig">KubeConfig</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item v-if="clusterForm.method === 'token'" label="API Server" required>
              <el-input v-model="clusterForm.server" placeholder="https://k8s-server:6443" prefix-icon="Link" />
            </el-form-item>
            <el-form-item v-if="clusterForm.method === 'token'" label="Token">
              <el-input v-model="clusterForm.token" type="password" show-password placeholder="Bearer Token" prefix-icon="Key" />
            </el-form-item>
            <el-form-item v-if="clusterForm.method === 'kubeconfig'" label="KubeConfig" required>
              <el-input v-model="clusterForm.kubeconfig" type="textarea" :rows="8" placeholder="粘贴 kubeconfig 内容..." />
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <div class="login-actions">
        <el-button v-if="activeTab === 'login'" type="primary" size="large" :loading="loading" @click="handleLogin" class="btn-primary-gradient">
          {{ loading ? '登录中...' : '登 录' }}
        </el-button>
        <el-button v-else type="success" size="large" :loading="connecting" @click="handleConnect" class="btn-primary-gradient btn-success-gradient">
          {{ connecting ? '添加中...' : '添加集群' }}
        </el-button>
      </div>

      <div v-if="errorMsg" class="error-msg">
        <el-alert :title="errorMsg" type="error" show-icon :closable="true" @close="errorMsg=''" />
      </div>

      <div class="login-footer" v-if="activeTab === 'login'">
        <p>默认账号: admin / admin123</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Monitor } from '@element-plus/icons-vue'
import api from '../api'

const router = useRouter()
const activeTab = ref('login')
const loading = ref(false)
const connecting = ref(false)
const errorMsg = ref('')
const isLoggedIn = ref(false)

const loginForm = ref({
  username: '',
  password: '',
})

const clusterForm = ref({
  name: '',
  method: 'token',
  server: '',
  token: '',
  kubeconfig: '',
})

onMounted(async () => {
  const token = localStorage.getItem('token')
  if (token) {
    isLoggedIn.value = true
    try {
      const clusterRes: any = await api.get('/clusters')
      if (clusterRes.clusters && clusterRes.clusters.length > 0) {
        router.push('/dashboard')
        return
      }
    } catch {}
    activeTab.value = 'cluster'
  }
})

const handleLogin = async () => {
  errorMsg.value = ''
  if (!loginForm.value.username || !loginForm.value.password) {
    errorMsg.value = '请输入用户名和密码'
    return
  }

  loading.value = true
  try {
    const res: any = await api.post('/auth/login', loginForm.value)
    if (res.code === 0 && res.data) {
      localStorage.setItem('token', res.data.token)
      localStorage.setItem('user', JSON.stringify(res.data.user))
      ElMessage.success('登录成功')
      isLoggedIn.value = true
      try {
        const clusterRes: any = await api.get('/clusters')
        if (clusterRes.clusters && clusterRes.clusters.length > 0) {
          router.push('/dashboard')
          return
        }
      } catch {}
      activeTab.value = 'cluster'
    } else {
      errorMsg.value = res.message || '登录失败'
    }
  } catch (e: any) {
    errorMsg.value = e.response?.data?.message || e.message || '登录失败'
  } finally {
    loading.value = false
  }
}

const handleConnect = async () => {
  errorMsg.value = ''
  connecting.value = true

  try {
    let payload: any = {}

    if (clusterForm.value.method === 'token') {
      if (!clusterForm.value.server) {
        errorMsg.value = '请输入 API Server 地址'
        connecting.value = false
        return
      }
      payload = {
        name: clusterForm.value.name || 'remote-cluster',
        server: clusterForm.value.server,
        token: clusterForm.value.token,
      }
    } else {
      if (!clusterForm.value.kubeconfig) {
        errorMsg.value = '请粘贴 KubeConfig 内容'
        connecting.value = false
        return
      }
      payload = {
        name: clusterForm.value.name || 'imported-cluster',
        kubeconfig: clusterForm.value.kubeconfig,
      }
    }

    const res: any = await api.post('/clusters/add', payload)

    if (res.success) {
      ElMessage.success('集群添加成功')
      localStorage.setItem('cluster_name', res.name || 'default')
      localStorage.setItem('cluster_connected', 'true')
      router.push('/dashboard')
    } else {
      errorMsg.value = res.message || '连接失败'
    }
  } catch (e: any) {
    errorMsg.value = e.response?.data?.message || e.message || '连接失败'
  } finally {
    connecting.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary, #0f172a);
  padding: 20px;
  position: relative;
}

.login-bg-orbs {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.4;
  animation: orbFloat 20s ease-in-out infinite;
}

.orb-1 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  top: -100px;
  left: -100px;
  animation-delay: 0s;
}

.orb-2 {
  width: 300px;
  height: 300px;
  background: linear-gradient(135deg, #3b82f6, #06b6d4);
  bottom: -80px;
  right: -80px;
  animation-delay: -7s;
}

.orb-3 {
  width: 250px;
  height: 250px;
  background: linear-gradient(135deg, #8b5cf6, #ec4899);
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  animation-delay: -14s;
}

@keyframes orbFloat {
  0%, 100% { transform: translate(0, 0) scale(1); }
  25% { transform: translate(30px, -20px) scale(1.05); }
  50% { transform: translate(-20px, 30px) scale(0.95); }
  75% { transform: translate(20px, 20px) scale(1.02); }
}

.login-card {
  width: 100%;
  max-width: 440px;
  background: var(--bg-glass, rgba(30, 41, 59, 0.6));
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: var(--border-default);
  border-radius: var(--radius-xl, 16px);
  padding: 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4), var(--shadow-glow);
  position: relative;
  z-index: 1;
  animation: fadeIn 0.5s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 16px;
  background: var(--primary-gradient);
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  box-shadow: var(--shadow-glow);
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.logo-icon:hover {
  transform: scale(1.05);
  box-shadow: 0 0 30px rgba(99, 102, 241, 0.5);
}

.login-header h1 {
  font-size: 28px;
  font-weight: 700;
  margin: 0 0 8px 0;
  background: var(--primary-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.login-header p {
  color: var(--text-secondary);
  margin: 0;
  font-size: 14px;
}

.login-tabs {
  margin-bottom: 24px;
}

.login-actions {
  margin-top: 24px;
}

.btn-primary-gradient {
  width: 100%;
  height: 44px;
  font-size: 16px;
  font-weight: 600;
  background: var(--primary-gradient);
  border: none;
  border-radius: var(--radius-md, 8px);
  color: #fff;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(99, 102, 241, 0.4);
}

.btn-primary-gradient:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 25px rgba(99, 102, 241, 0.6);
}

.btn-primary-gradient:active {
  transform: translateY(0);
}

.btn-success-gradient {
  background: linear-gradient(135deg, #22c55e, #06b6d4);
  box-shadow: 0 4px 15px rgba(34, 197, 94, 0.4);
}

.btn-success-gradient:hover {
  box-shadow: 0 6px 25px rgba(34, 197, 94, 0.6);
}

.error-msg {
  margin-top: 16px;
}

.login-footer {
  margin-top: 20px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 12px;
}

:deep(.el-tabs__nav-wrap::after) {
  display: none;
}

:deep(.el-tabs__item) {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-secondary);
}

:deep(.el-tabs__item.is-active) {
  color: var(--primary);
}

:deep(.el-tabs__active-bar) {
  background: var(--primary-gradient);
  height: 3px;
  border-radius: 2px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--text-primary);
}

:deep(.el-input__wrapper) {
  background: rgba(15, 23, 42, 0.6);
  border-radius: var(--radius-md, 8px);
  box-shadow: none;
  border: var(--border-default);
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
}

:deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.3);
}

:deep(.el-input__wrapper.is-focus) {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
}

:deep(.el-input__inner) {
  color: var(--text-primary);
}

:deep(.el-input__inner::placeholder) {
  color: var(--text-disabled);
}

:deep(.el-textarea__inner) {
  background: rgba(15, 23, 42, 0.6);
  border: var(--border-default);
  border-radius: var(--radius-md, 8px);
  color: var(--text-primary);
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
}

:deep(.el-textarea__inner:hover) {
  border-color: rgba(99, 102, 241, 0.3);
}

:deep(.el-textarea__inner:focus) {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.15);
}

:deep(.el-radio__label) {
  color: var(--text-secondary);
}

:deep(.el-radio__input.is-checked + .el-radio__label) {
  color: var(--primary);
}

:deep(.el-radio__input.is-checked .el-radio__inner) {
  background: var(--primary);
  border-color: var(--primary);
}

:deep(.el-alert) {
  border-radius: var(--radius-md, 8px);
}

:deep(.el-tabs__header) {
  margin-bottom: 24px;
}
</style>
