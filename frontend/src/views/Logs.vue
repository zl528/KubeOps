<template>
  <div class="page-container">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">日志查看器</h1>
        <span class="page-subtitle">查看 Pod 容器日志和实时流</span>
      </div>
      <div class="header-actions">
        <button type="button" class="btn-gradient btn-refresh" @click="fetchLogs">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="glass-card">
      <div class="filter-row">
        <div class="filter-bar">
          <el-select v-model="namespace" placeholder="命名空间" class="filter-select">
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
          <el-select v-model="pod" placeholder="选择 Pod" class="filter-select-lg" @change="fetchContainers">
            <el-option v-for="p in podList" :key="p.name" :label="p.name" :value="p.name" />
          </el-select>
          <el-select v-model="container" placeholder="容器" class="filter-select" clearable>
            <el-option v-for="c in containers" :key="c" :label="c" :value="c" />
          </el-select>
          <el-input v-model="searchKeyword" placeholder="搜索日志..." class="filter-input" @keyup.enter="fetchLogs">
            <template #append>
              <button type="button" class="btn-search" @click="fetchLogs">搜索</button>
            </template>
          </el-input>
          <el-select v-model="logLevel" placeholder="日志级别" class="filter-select-sm" clearable @change="fetchLogs">
            <el-option label="全部" value="" />
            <el-option label="Error" value="error" />
            <el-option label="Warn" value="warn" />
            <el-option label="Info" value="info" />
            <el-option label="Debug" value="debug" />
          </el-select>
          <el-input-number v-model="tailLines" :min="100" :max="10000" :step="100" class="filter-number" />
        </div>
        <div class="action-bar">
          <button type="button" class="btn-gradient btn-refresh" @click="downloadLogs">
            <el-icon :size="16"><Download /></el-icon>
            <span>下载</span>
          </button>
          <button :class="['btn-gradient', isStreaming ? 'btn-streaming' : 'btn-stream']" @click="toggleStream">
            <span v-if="isStreaming" class="btn-spinner"></span>
            <span>{{ isStreaming ? '停止' : '实时' }}</span>
          </button>
        </div>
      </div>

      <div class="log-container" ref="logContainer">
        <div v-if="logs.length === 0 && !loading" class="empty-log">暂无日志</div>
        <div v-for="(log, idx) in logs" :key="idx" class="log-line" :class="log.level">
          <span v-if="log.timestamp" class="log-time">{{ log.timestamp }}</span>
          <span v-if="log.level" class="log-level" :class="log.level">{{ log.level.toUpperCase() }}</span>
          <span v-if="log.source" class="log-source">{{ log.source }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
      </div>

      <div class="log-footer">
        <span>共 {{ logs.length }} 条日志</span>
        <span v-if="isStreaming" class="streaming-indicator">● 实时采集中</span>
      </div>
    </div>

    <div class="glass-card" style="margin-top: 24px">
      <div class="card-section-header">
        <span class="section-title">上一次容器日志</span>
        <button type="button" class="btn-outline" @click="fetchPreviousLogs" :disabled="prevLoading">
          <span v-if="prevLoading" class="btn-spinner"></span>
          查看上一次日志
        </button>
      </div>
      <div v-if="previousLogs.length > 0" class="log-container log-container-sm">
        <div v-for="(log, idx) in previousLogs" :key="idx" class="log-line" :class="log.level">
          <span v-if="log.timestamp" class="log-time">{{ log.timestamp }}</span>
          <span v-if="log.level" class="log-level" :class="log.level">{{ log.level.toUpperCase() }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue'
import { Refresh, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import api from '../api'
import { useGlobalNamespace } from '../store/namespace'

interface LogEntry {
  timestamp?: string
  level?: string
  source?: string
  message: string
}

const { namespace } = useGlobalNamespace()
const pod = ref('')
const container = ref('')
const searchKeyword = ref('')
const logLevel = ref('')
const tailLines = ref(1000)
const logs = ref<LogEntry[]>([])
const previousLogs = ref<LogEntry[]>([])
const nsList = ref<any[]>([])
const podList = ref<any[]>([])
const containers = ref<string[]>([])
const loading = ref(false)
const prevLoading = ref(false)
const isStreaming = ref(false)
const logContainer = ref<HTMLElement>()
let streamInterval: number | null = null

const fetchNs = async () => {
  try {
    const res: any = await api.get('/namespaces')
    nsList.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const fetchPods = async () => {
  if (!namespace.value) return
  try {
    const res: any = await api.get('/pods', { params: { namespace: namespace.value } })
    podList.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const fetchContainers = async () => {
  if (!namespace.value || !pod.value) return
  try {
    const res: any = await api.get('/logs/containers', { params: { namespace: namespace.value, pod: pod.value } })
    containers.value = res.data || []
    if (containers.value.length > 0) {
      container.value = containers.value[0]
    }
  } catch (e) {
    console.error(e)
  }
}

const fetchLogs = async () => {
  if (!namespace.value || !pod.value) {
    ElMessage.warning('请选择命名空间和 Pod')
    return
  }

  loading.value = true
  try {
    const params: any = {
      namespace: namespace.value,
      pod: pod.value,
      tailLines: tailLines.value,
    }
    if (container.value) params.container = container.value
    if (searchKeyword.value) params.search = searchKeyword.value
    if (logLevel.value) params.level = logLevel.value

    const res: any = await api.get('/logs', { params })
    logs.value = res.data?.lines || []
    await nextTick()
    scrollToBottom()
  } catch (e) {
    console.error(e)
    ElMessage.error('获取日志失败')
  } finally {
    loading.value = false
  }
}

const fetchPreviousLogs = async () => {
  if (!namespace.value || !pod.value) {
    ElMessage.warning('请选择命名空间和 Pod')
    return
  }

  prevLoading.value = true
  try {
    const params: any = {
      namespace: namespace.value,
      pod: pod.value,
      tailLines: 100,
    }
    if (container.value) params.container = container.value

    const res: any = await api.get('/logs/previous', { params })
    previousLogs.value = res.data?.lines || []
  } catch (e) {
    console.error(e)
    ElMessage.error('获取上一次日志失败')
  } finally {
    prevLoading.value = false
  }
}

const downloadLogs = async () => {
  if (!namespace.value || !pod.value) {
    ElMessage.warning('请选择命名空间和 Pod')
    return
  }

  try {
    const params: any = {
      namespace: namespace.value,
      pod: pod.value,
    }
    if (container.value) params.container = container.value

    const res = await api.get('/logs/download', { params, responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([res as any]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `${pod.value}.log`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    ElMessage.success('下载成功')
  } catch (e) {
    console.error(e)
    ElMessage.error('下载失败')
  }
}

const toggleStream = () => {
  if (isStreaming.value) {
    stopStream()
  } else {
    startStream()
  }
}

const startStream = () => {
  if (!namespace.value || !pod.value) {
    ElMessage.warning('请选择命名空间和 Pod')
    return
  }

  isStreaming.value = true
  streamInterval = window.setInterval(async () => {
    try {
      const params: any = {
        namespace: namespace.value,
        pod: pod.value,
        tailLines: 50,
      }
      if (container.value) params.container = container.value

      const res: any = await api.get('/logs', { params })
      const newLogs = res.data?.lines || []
      if (newLogs.length > 0) {
        logs.value = [...logs.value.slice(-500), ...newLogs]
        await nextTick()
        scrollToBottom()
      }
    } catch (e) {
      console.error(e)
    }
  }, 3000)
}

const stopStream = () => {
  isStreaming.value = false
  if (streamInterval) {
    clearInterval(streamInterval)
    streamInterval = null
  }
}

const scrollToBottom = () => {
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

watch(namespace, () => {
  pod.value = ''
  container.value = ''
  logs.value = []
  fetchPods()
})

onMounted(() => {
  fetchNs()
  fetchPods()
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

.glass-card {
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
}

.filter-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
}

.filter-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-select {
  width: 150px;
}

.filter-select-lg {
  width: 200px;
}

.filter-select-sm {
  width: 120px;
}

.filter-input {
  width: 200px;
}

.filter-number {
  width: 150px;
}

.filter-bar :deep(.el-input__wrapper),
.filter-bar :deep(.el-input-number .el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 8px;
  box-shadow: none;
}

.filter-bar :deep(.el-input__wrapper:hover),
.filter-bar :deep(.el-input-number .el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.filter-bar :deep(.el-input__wrapper.is-focus),
.filter-bar :deep(.el-input-number .el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.filter-bar :deep(.el-input__inner),
.filter-bar :deep(.el-input-number .el-input__inner) {
  color: var(--text-primary);
}

.filter-bar :deep(.el-input__inner::placeholder) {
  color: var(--text-secondary);
}

.filter-bar :deep(.el-input-number .el-input-number__decrease),
.filter-bar :deep(.el-input-number .el-input-number__increase) {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(148, 163, 184, 0.1);
  color: var(--text-secondary);
}

.filter-bar :deep(.el-input-number .el-input-number__decrease:hover),
.filter-bar :deep(.el-input-number .el-input-number__increase:hover) {
  color: var(--primary);
}

.btn-search {
  background: rgba(51, 65, 85, 0.6);
  color: var(--text-secondary);
  border: none;
  padding: 6px 16px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.15s ease;
}

.btn-search:hover {
  background: rgba(51, 65, 85, 0.9);
  color: var(--text-primary);
}

.action-bar {
  display: flex;
  gap: 8px;
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

.btn-stream {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}

.btn-stream:hover {
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.4), 0 0 20px rgba(245, 158, 11, 0.3);
}

.btn-streaming {
  background: linear-gradient(135deg, #ef4444, #dc2626);
}

.btn-streaming:hover {
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.4), 0 0 20px rgba(239, 68, 68, 0.3);
}

.btn-outline {
  background: rgba(51, 65, 85, 0.4);
  color: var(--text-secondary);
  border: 1px solid rgba(148, 163, 184, 0.1);
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-outline:hover:not(:disabled) {
  background: rgba(51, 65, 85, 0.7);
  color: var(--text-primary);
  border-color: rgba(99, 102, 241, 0.3);
}

.btn-outline:disabled {
  opacity: 0.5;
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

.log-container {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.06);
  border-radius: 12px;
  padding: 16px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  line-height: 1.6;
  max-height: 600px;
  overflow-y: auto;
}

.log-container-sm {
  max-height: 300px;
  margin-top: 16px;
}

.log-line {
  display: flex;
  gap: 8px;
  padding: 2px 0;
  transition: background 0.1s ease;
  border-radius: 4px;
}

.log-line:hover {
  background: rgba(51, 65, 85, 0.3);
}

.log-line.error {
  color: #f48771;
}

.log-line.warn {
  color: #cca700;
}

.log-line.info {
  color: #d4d4d4;
}

.log-line.debug {
  color: #808080;
}

.log-time {
  color: #60a5fa;
  white-space: nowrap;
}

.log-level {
  font-weight: bold;
  min-width: 50px;
}

.log-level.error { color: #f87171; }
.log-level.warn { color: #fbbf24; }
.log-level.info { color: #4ade80; }
.log-level.debug { color: #94a3b8; }

.log-source {
  color: #a78bfa;
  white-space: nowrap;
}

.log-message {
  flex: 1;
  word-break: break-all;
}

.log-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  font-size: 12px;
  color: var(--text-secondary);
}

.streaming-indicator {
  color: #4ade80;
  animation: blink 1s infinite;
  font-weight: 500;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.empty-log {
  text-align: center;
  color: var(--text-secondary);
  padding: 32px;
}

.card-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}
</style>
