<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1 class="page-title">集群概览</h1>
      <span class="page-subtitle">实时监控 Kubernetes 集群状态</span>
    </div>

    <div class="stat-grid">
      <div class="stat-card stat-card--nodes">
        <div class="stat-icon">
          <el-icon size="28"><Cpu /></el-icon>
        </div>
        <div class="stat-body">
          <div class="stat-value">{{ overview?.totalNodes || 0 }}</div>
          <div class="stat-label">节点总数</div>
        </div>
        <div class="stat-glow"></div>
      </div>

      <div class="stat-card stat-card--running">
        <div class="stat-icon">
          <el-icon size="28"><Box /></el-icon>
        </div>
        <div class="stat-body">
          <div class="stat-value">{{ overview?.runningPods || 0 }}</div>
          <div class="stat-label">运行中 Pod</div>
        </div>
        <div class="stat-glow"></div>
      </div>

      <div class="stat-card stat-card--pending">
        <div class="stat-icon">
          <el-icon size="28"><Warning /></el-icon>
        </div>
        <div class="stat-body">
          <div class="stat-value">{{ overview?.pendingPods || 0 }}</div>
          <div class="stat-label">Pending Pod</div>
        </div>
        <div class="stat-glow"></div>
      </div>

      <div class="stat-card stat-card--failed">
        <div class="stat-icon">
          <el-icon size="28"><CircleCloseFilled /></el-icon>
        </div>
        <div class="stat-body">
          <div class="stat-value">{{ overview?.failedPods || 0 }}</div>
          <div class="stat-label">Failed Pod</div>
        </div>
        <div class="stat-glow"></div>
      </div>
    </div>

    <div class="glass-section">
      <div class="section-grid">
        <div class="usage-card">
          <div class="usage-header">
            <span class="usage-title">CPU 使用率</span>
            <span class="usage-percent">{{ cpuPercent }}%</span>
          </div>
          <el-progress
            :percentage="cpuPercent"
            :color="progressColor"
            :stroke-width="16"
            :text-inside="true"
            class="custom-progress"
          />
          <div class="usage-detail">
            <span>已用: {{ overview?.cpuUsage.used || '0' }}</span>
            <span>总量: {{ overview?.cpuUsage.total || '0' }}</span>
          </div>
        </div>

        <div class="usage-card">
          <div class="usage-header">
            <span class="usage-title">内存使用率</span>
            <span class="usage-percent">{{ memPercent }}%</span>
          </div>
          <el-progress
            :percentage="memPercent"
            :color="progressColor"
            :stroke-width="16"
            :text-inside="true"
            class="custom-progress"
          />
          <div class="usage-detail">
            <span>已用: {{ overview?.memoryUsage.used || '0' }}</span>
            <span>总量: {{ overview?.memoryUsage.total || '0' }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="glass-section">
      <div class="section-grid section-grid--half">
        <div class="info-card">
          <div class="info-header">
            <span class="info-title">节点状态</span>
          </div>
          <div class="status-list">
            <div class="status-item status-item--ready">
              <span class="status-dot"></span>
              <span class="status-text">Ready</span>
              <span class="status-count">{{ overview?.readyNodes || 0 }}</span>
            </div>
            <div class="status-item status-item--notready">
              <span class="status-dot"></span>
              <span class="status-text">NotReady</span>
              <span class="status-count">{{ (overview?.totalNodes || 0) - (overview?.readyNodes || 0) }}</span>
            </div>
          </div>
        </div>

        <div class="info-card">
          <div class="info-header">
            <span class="info-title">命名空间</span>
          </div>
          <div class="status-list">
            <div class="status-item status-item--info">
              <span class="status-dot"></span>
              <span class="status-text">总数</span>
              <span class="status-count">{{ overview?.namespaces || 0 }}</span>
            </div>
            <div class="status-item status-item--info">
              <span class="status-dot"></span>
              <span class="status-text">Pod 总数</span>
              <span class="status-count">{{ overview?.totalPods || 0 }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useClusterStore } from '../store/cluster'
import { Cpu, Box, Warning, CircleCloseFilled } from '@element-plus/icons-vue'

const store = useClusterStore()
const overview = computed(() => store.overview)

const cpuPercent = computed(() => {
  const pct = overview.value?.cpuUsage.percentage || '0%'
  return parseInt(pct) || 0
})

const memPercent = computed(() => {
  const pct = overview.value?.memoryUsage.percentage || '0%'
  return parseInt(pct) || 0
})

const progressColor = computed(() => {
  const pct = cpuPercent.value
  if (pct < 60) return '#67c23a'
  if (pct < 80) return '#e6a23c'
  return '#f56c6c'
})

onMounted(() => {
  store.fetchOverview()
})
</script>

<style scoped>
.dashboard {
  padding: 0;
  min-height: 100vh;
}

.dashboard-header {
  margin-bottom: 32px;
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 4px 0;
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  position: relative;
  background: var(--bg-card);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: var(--radius-lg);
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.stat-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.stat-card--nodes::before {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.12), transparent 60%);
}
.stat-card--running::before {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.12), transparent 60%);
}
.stat-card--pending::before {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.12), transparent 60%);
}
.stat-card--failed::before {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.12), transparent 60%);
}

.stat-card:hover::before {
  opacity: 1;
}

.stat-card:hover {
  transform: translateY(-4px);
  border-color: rgba(99, 102, 241, 0.25);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.4), 0 0 24px rgba(99, 102, 241, 0.08);
}

.stat-glow {
  position: absolute;
  top: -50%;
  right: -50%;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  opacity: 0;
  transition: opacity 0.3s ease;
  pointer-events: none;
}

.stat-card--nodes .stat-glow {
  background: radial-gradient(circle, rgba(99, 102, 241, 0.15), transparent 70%);
}
.stat-card--running .stat-glow {
  background: radial-gradient(circle, rgba(34, 197, 94, 0.15), transparent 70%);
}
.stat-card--pending .stat-glow {
  background: radial-gradient(circle, rgba(245, 158, 11, 0.15), transparent 70%);
}
.stat-card--failed .stat-glow {
  background: radial-gradient(circle, rgba(239, 68, 68, 0.15), transparent 70%);
}

.stat-card:hover .stat-glow {
  opacity: 1;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

.stat-card--nodes .stat-icon {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.35);
}
.stat-card--running .stat-icon {
  background: linear-gradient(135deg, #22c55e, #16a34a);
  box-shadow: 0 4px 12px rgba(34, 197, 94, 0.35);
}
.stat-card--pending .stat-icon {
  background: linear-gradient(135deg, #f59e0b, #d97706);
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.35);
}
.stat-card--failed .stat-icon {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.35);
}

.stat-body {
  position: relative;
  z-index: 1;
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.1;
}

.stat-label {
  font-size: 14px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.glass-section {
  background: var(--bg-glass);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: var(--radius-lg);
  padding: 24px;
  margin-bottom: 24px;
  transition: border-color 0.3s ease;
}

.glass-section:hover {
  border-color: rgba(148, 163, 184, 0.15);
}

.section-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24px;
}

.section-grid--half {
  gap: 24px;
}

.usage-card {
  padding: 0;
}

.usage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.usage-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.usage-percent {
  font-size: 20px;
  font-weight: 700;
  color: var(--primary-light);
}

.usage-detail {
  display: flex;
  justify-content: space-between;
  margin-top: 12px;
  color: var(--text-secondary);
  font-size: 13px;
}

.info-card {
  padding: 0;
}

.info-header {
  margin-bottom: 16px;
}

.info-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  background: rgba(15, 23, 42, 0.5);
  border-radius: var(--radius-md);
  border: 1px solid rgba(148, 163, 184, 0.06);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-item--ready .status-dot {
  background: var(--success);
  box-shadow: 0 0 8px rgba(34, 197, 94, 0.5);
}
.status-item--notready .status-dot {
  background: var(--danger);
  box-shadow: 0 0 8px rgba(239, 68, 68, 0.5);
}
.status-item--info .status-dot {
  background: var(--info);
  box-shadow: 0 0 8px rgba(59, 130, 246, 0.5);
}

.status-text {
  font-size: 14px;
  color: var(--text-secondary);
  flex: 1;
}

.status-count {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
}

:deep(.el-progress) {
  --el-progress-bar-color: var(--primary);
}

:deep(.el-progress__inner) {
  border-radius: 8px;
}

:deep(.el-progress-bar__outer) {
  background: rgba(15, 23, 42, 0.6);
  border-radius: 8px;
}

:deep(.el-progress__text) {
  color: var(--text-primary);
  font-weight: 600;
}
</style>
