<template>
  <div class="monitoring-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">Metrics Monitoring</h1>
        <span class="page-subtitle">实时监控集群资源使用情况</span>
      </div>
      <div class="header-actions">
        <button type="button" class="btn-gradient btn-prometheus" @click="showPrometheus = !showPrometheus">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z"/></svg>
          <span>Prometheus</span>
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchAll">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <!-- Prometheus Status Panel -->
    <div v-if="showPrometheus" class="prometheus-panel glass-card">
      <div class="panel-header">
        <span class="panel-title">Prometheus 状态</span>
        <el-tag :type="prometheusStatus.available ? 'success' : 'danger'" size="small">
          {{ prometheusStatus.available ? '可用' : '不可用' }}
        </el-tag>
      </div>
      <div v-if="prometheusStatus.available" class="prometheus-stats">
        <div class="pstat-item">
          <span class="pstat-label">采集目标</span>
          <span class="pstat-value">{{ prometheusStatus.targetsUp }} / {{ prometheusStatus.targetsTotal }}</span>
        </div>
        <div class="pstat-item">
          <span class="pstat-label">活跃告警</span>
          <span class="pstat-value" :class="{ 'alert-active': prometheusStatus.alertsCount > 0 }">{{ prometheusStatus.alertsCount }}</span>
        </div>
      </div>
      <div v-else class="prometheus-empty">
        <p>未检测到 Prometheus，监控功能受限</p>
        <p class="prometheus-hint">请确保 Prometheus 已部署并可通过 NodePort 访问</p>
      </div>
    </div>

    <div class="metrics-row">
      <div class="metric-card glass-card">
        <div class="metric-icon cpu-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/><line x1="9" y1="1" x2="9" y2="4"/><line x1="15" y1="1" x2="15" y2="4"/><line x1="9" y1="20" x2="9" y2="23"/><line x1="15" y1="20" x2="15" y2="23"/><line x1="20" y1="9" x2="23" y2="9"/><line x1="20" y1="14" x2="23" y2="14"/><line x1="1" y1="9" x2="4" y2="9"/><line x1="1" y1="14" x2="4" y2="14"/></svg>
        </div>
        <div class="metric-header">CPU 使用率</div>
        <div class="metric-value" :class="cpuClass">{{ cpuUsage.toFixed(1) }}%</div>
        <div class="metric-bar-track">
          <div class="metric-bar-fill" :class="cpuClass" :style="{ width: cpuUsage + '%' }"></div>
        </div>
      </div>
      <div class="metric-card glass-card">
        <div class="metric-icon mem-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="6" width="20" height="12" rx="2"/><line x1="6" y1="10" x2="6" y2="14"/><line x1="10" y1="10" x2="10" y2="14"/><line x1="14" y1="10" x2="14" y2="14"/><line x1="18" y1="10" x2="18" y2="14"/></svg>
        </div>
        <div class="metric-header">内存使用率</div>
        <div class="metric-value" :class="memClass">{{ memoryUsage.toFixed(1) }}%</div>
        <div class="metric-bar-track">
          <div class="metric-bar-fill" :class="memClass" :style="{ width: memoryUsage + '%' }"></div>
        </div>
      </div>
      <div class="metric-card glass-card">
        <div class="metric-icon disk-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/></svg>
        </div>
        <div class="metric-header">磁盘使用率</div>
        <div class="metric-value" :class="diskClass">{{ diskUsage.toFixed(1) }}%</div>
        <div class="metric-bar-track">
          <div class="metric-bar-fill" :class="diskClass" :style="{ width: diskUsage + '%' }"></div>
        </div>
      </div>
      <div class="metric-card glass-card">
        <div class="metric-icon pod-icon">
          <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
        </div>
        <div class="metric-header">Pod 数量</div>
        <div class="metric-value pods-value">{{ clusterMetrics?.runningPods || 0 }} / {{ clusterMetrics?.totalPods || 0 }}</div>
        <div class="metric-sub">运行中 / 总数</div>
      </div>
    </div>

    <div class="charts-row">
      <div class="glass-card">
        <div class="card-header">
          <span class="card-title">CPU / 内存趋势</span>
          <div class="card-actions">
            <span class="info-tag">每5秒采集 · 最近30个数据点</span>
            <button type="button" class="btn-action btn-clear" @click="clearHistory">清空</button>
          </div>
        </div>
        <div ref="cpuMemChartRef" class="chart-container"></div>
      </div>
      <div class="glass-card">
        <div class="card-header">
          <span class="card-title">网络流量趋势</span>
          <span class="info-tag">累计字节</span>
        </div>
        <div ref="networkChartRef" class="chart-container"></div>
      </div>
    </div>

    <div class="charts-row">
      <div class="glass-card">
        <div class="card-header">
          <span class="card-title">节点状态</span>
          <button type="button" class="btn-action" @click="fetchClusterMetrics">刷新</button>
        </div>
        <div class="node-info">
          <div class="info-stat">
            <div class="info-stat-value">{{ clusterMetrics?.totalNodes || 0 }}</div>
            <div class="info-stat-label">总节点数</div>
          </div>
          <div class="info-stat">
            <div class="info-stat-value ok">{{ clusterMetrics?.readyNodes || 0 }}</div>
            <div class="info-stat-label">就绪节点</div>
          </div>
        </div>
      </div>
      <div class="glass-card">
        <div class="card-header">
          <span class="card-title">网络流量</span>
          <button type="button" class="btn-action" @click="fetchNetworkUsage">刷新</button>
        </div>
        <div class="node-info">
          <div class="info-stat">
            <div class="info-stat-value">↓ {{ formatBytes(networkRx) }}</div>
            <div class="info-stat-label">入站流量</div>
          </div>
          <div class="info-stat">
            <div class="info-stat-value">↑ {{ formatBytes(networkTx) }}</div>
            <div class="info-stat-label">出站流量</div>
          </div>
        </div>
      </div>
    </div>

    <div class="glass-card" style="margin-top: 20px">
      <div class="card-header">
        <span class="card-title">Prometheus 查询</span>
      </div>
      <div class="quick-metrics">
        <span class="quick-label">快捷查询：</span>
        <button v-for="q in quickQueries" :key="q.query" type="button" class="btn-quick" @click="promQuery = q.query; queryPrometheus()">
          {{ q.label }}
        </button>
      </div>
      <div class="prom-input-row">
        <div class="prom-input-wrapper">
          <input v-model="promQuery" class="prom-input" placeholder="输入 PromQL 查询，如: up, node_cpu_seconds_total" @keyup.enter="queryPrometheus" @input="onPromInputChange" @focus="showSuggestions = true" @blur="hideSuggestions" />
          <div v-if="showSuggestions && filteredMetrics.length > 0" class="suggestions-dropdown">
            <div v-for="m in filteredMetrics" :key="m" class="suggestion-item" @mousedown.prevent="selectMetric(m)">
              <span class="suggestion-name">{{ m }}</span>
            </div>
          </div>
        </div>
        <button type="button" class="btn-gradient" @click="queryPrometheus">查询</button>
      </div>

      <!-- Table View - always show for instant queries -->
      <div v-if="promResult" class="prom-table-container">
        <el-table :data="paginatedTableData" :header-cell-style="headerCellStyle" :cell-style="cellStyle" class="custom-table" max-height="400">
          <el-table-column label="指标" min-width="300">
            <template #default="{ row }">
              <span class="metric-name">{{ row.name }}</span>
              <span v-if="row.labels" class="metric-labels">{{ row.labels }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="value" label="值" width="150" />
          <el-table-column prop="timestamp" label="时间" width="180" />
        </el-table>
        <div class="prom-pagination" v-if="promTableData.length > pageSize">
          <el-pagination
            v-model:current-page="currentPage"
            :page-size="pageSize"
            :total="promTableData.length"
            layout="prev, pager, next, total"
            small
          />
        </div>
      </div>

      <!-- Raw JSON fallback -->
      <pre v-if="promResult && promViewMode === 'raw'" class="prom-result">{{ promResult }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import api from '../api'

const cpuUsage = ref(0)
const memoryUsage = ref(0)
const diskUsage = ref(0)
const networkRx = ref(0)
const networkTx = ref(0)
const clusterMetrics = ref<any>(null)
const promQuery = ref('')
const promResult = ref('')
const promViewMode = ref('table')
const promTableData = ref<any[]>([])
const currentPage = ref(1)
const pageSize = 20
const metricNames = ref<string[]>([])
const showSuggestions = ref(false)

const filteredMetrics = computed(() => {
  if (!promQuery.value) return metricNames.value.slice(0, 20)
  const q = promQuery.value.toLowerCase()
  return metricNames.value.filter(m => m.toLowerCase().includes(q)).slice(0, 20)
})

const onPromInputChange = () => {
  showSuggestions.value = true
}

const hideSuggestions = () => {
  setTimeout(() => { showSuggestions.value = false }, 200)
}

const selectMetric = (name: string) => {
  promQuery.value = name
  showSuggestions.value = false
}

const fetchMetricNames = async () => {
  try {
    const res: any = await api.get('/monitor/prometheus/labels', { params: { label: '__name__' } })
    const results = res.data?.result || []
    metricNames.value = [...new Set(results.map((item: any) => item.metric?.__name__ || '').filter(Boolean))].sort() as string[]
  } catch (e) {
    console.error(e)
  }
}

const paginatedTableData = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return promTableData.value.slice(start, start + pageSize)
})

const quickQueries = [
  { label: '集群状态', query: 'up' },
  { label: 'CPU使用率', query: '100 - (avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)' },
  { label: '内存使用率', query: '(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100' },
  { label: '磁盘使用率', query: '(1 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"})) * 100' },
  { label: 'Pod数量', query: 'count(kube_pod_info)' },
  { label: '节点数量', query: 'count(kube_node_info)' },
]

const showPrometheus = ref(false)
const prometheusStatus = ref<any>({ available: false })
const capabilities = ref<any>({})

const cpuMemChartRef = ref<HTMLElement>()
const networkChartRef = ref<HTMLElement>()
let cpuMemChart: echarts.ECharts | null = null
let networkChart: echarts.ECharts | null = null

const MAX_POINTS = 30
const history = ref<{
  time: string[]
  cpu: number[]
  mem: number[]
  rx: number[]
  tx: number[]
}>({ time: [], cpu: [], mem: [], rx: [], tx: [] })

let pollTimer: ReturnType<typeof setInterval> | null = null

const cpuColor = computed(() => cpuUsage.value < 60 ? '#67c23a' : cpuUsage.value < 80 ? '#e6a23c' : '#f56c6c')
const memColor = computed(() => memoryUsage.value < 60 ? '#67c23a' : memoryUsage.value < 80 ? '#e6a23c' : '#f56c6c')
const diskColor = computed(() => diskUsage.value < 60 ? '#67c23a' : diskUsage.value < 80 ? '#e6a23c' : '#f56c6c')

const cpuClass = computed(() => cpuUsage.value < 60 ? 'low' : cpuUsage.value < 80 ? 'medium' : 'high')
const memClass = computed(() => memoryUsage.value < 60 ? 'low' : memoryUsage.value < 80 ? 'medium' : 'high')
const diskClass = computed(() => diskUsage.value < 60 ? 'low' : diskUsage.value < 80 ? 'medium' : 'high')

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatTime = (d: Date) => {
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}:${String(d.getSeconds()).padStart(2, '0')}`
}

const initCharts = () => {
  if (cpuMemChartRef.value) {
    cpuMemChart = echarts.init(cpuMemChartRef.value)
    cpuMemChart.setOption({
      tooltip: { trigger: 'axis' },
      legend: { data: ['CPU', '内存'], bottom: 0, textStyle: { color: '#94a3b8' } },
      grid: { top: 20, right: 20, bottom: 40, left: 50 },
      xAxis: { type: 'category', data: [], axisLine: { lineStyle: { color: 'rgba(148, 163, 184, 0.15)' } }, axisLabel: { color: '#94a3b8' } },
      yAxis: { type: 'value', min: 0, max: 100, axisLabel: { formatter: '{value}%', color: '#94a3b8' }, splitLine: { lineStyle: { color: 'rgba(148, 163, 184, 0.08)' } } },
      series: [
        { name: 'CPU', type: 'line', smooth: true, data: [], areaStyle: { opacity: 0.15 }, itemStyle: { color: '#6366f1' } },
        { name: '内存', type: 'line', smooth: true, data: [], areaStyle: { opacity: 0.15 }, itemStyle: { color: '#8b5cf6' } },
      ],
    })
  }

  if (networkChartRef.value) {
    networkChart = echarts.init(networkChartRef.value)
    networkChart.setOption({
      tooltip: { trigger: 'axis', formatter: (params: any) => {
        let s = params[0].axisValue + '<br/>'
        params.forEach((p: any) => {
          s += `${p.marker} ${p.seriesName}: ${formatBytes(p.value)}<br/>`
        })
        return s
      }},
      legend: { data: ['入站', '出站'], bottom: 0, textStyle: { color: '#94a3b8' } },
      grid: { top: 20, right: 20, bottom: 40, left: 60 },
      xAxis: { type: 'category', data: [], axisLine: { lineStyle: { color: 'rgba(148, 163, 184, 0.15)' } }, axisLabel: { color: '#94a3b8' } },
      yAxis: { type: 'value', axisLabel: { formatter: (v: number) => formatBytes(v), color: '#94a3b8' }, splitLine: { lineStyle: { color: 'rgba(148, 163, 184, 0.08)' } } },
      series: [
        { name: '入站', type: 'line', smooth: true, data: [], areaStyle: { opacity: 0.15 }, itemStyle: { color: '#6366f1' } },
        { name: '出站', type: 'line', smooth: true, data: [], areaStyle: { opacity: 0.15 }, itemStyle: { color: '#f59e0b' } },
      ],
    })
  }
}

const updateCharts = () => {
  const h = history.value
  cpuMemChart?.setOption({
    xAxis: { data: h.time },
    series: [{ data: h.cpu }, { data: h.mem }],
  })
  networkChart?.setOption({
    xAxis: { data: h.time },
    series: [{ data: h.rx }, { data: h.tx }],
  })
}

const pushDataPoint = () => {
  const now = formatTime(new Date())
  const h = history.value
  h.time.push(now)
  h.cpu.push(cpuUsage.value)
  h.mem.push(memoryUsage.value)
  h.rx.push(networkRx.value)
  h.tx.push(networkTx.value)

  if (h.time.length > MAX_POINTS) {
    h.time.shift()
    h.cpu.shift()
    h.mem.shift()
    h.rx.shift()
    h.tx.shift()
  }

  updateCharts()
}

const clearHistory = () => {
  history.value = { time: [], cpu: [], mem: [], rx: [], tx: [] }
  updateCharts()
}

const fetchClusterMetrics = async () => {
  try {
    const res: any = await api.get('/monitor/cluster')
    clusterMetrics.value = res.data
  } catch (e) {
    console.error(e)
  }
}

const fetchCPUUsage = async () => {
  try {
    const res: any = await api.get('/monitor/cpu')
    cpuUsage.value = res.data?.usage || 0
  } catch (e) {
    console.error(e)
  }
}

const fetchMemoryUsage = async () => {
  try {
    const res: any = await api.get('/monitor/memory')
    memoryUsage.value = res.data?.usage || 0
  } catch (e) {
    console.error(e)
  }
}

const fetchDiskUsage = async () => {
  try {
    const res: any = await api.get('/monitor/disk')
    diskUsage.value = res.data?.usage || 0
  } catch (e) {
    console.error(e)
  }
}

const fetchNetworkUsage = async () => {
  try {
    const res: any = await api.get('/monitor/network')
    networkRx.value = res.data?.rxBytes || 0
    networkTx.value = res.data?.txBytes || 0
  } catch (e) {
    console.error(e)
  }
}

const fetchCapabilities = async () => {
  try {
    const res: any = await api.get('/monitor/capabilities')
    capabilities.value = res.data || {}
  } catch (e) {
    console.error(e)
  }
}

const fetchPrometheusStatus = async () => {
  try {
    const res: any = await api.get('/monitor/prometheus/status')
    prometheusStatus.value = res.data || { available: false }
  } catch (e) {
    prometheusStatus.value = { available: false }
  }
}

const deployPrometheus = () => {}

const fetchAll = async () => {
  await Promise.all([
    fetchCPUUsage(),
    fetchMemoryUsage(),
    fetchDiskUsage(),
    fetchNetworkUsage(),
  ])
  pushDataPoint()
}

const queryPrometheus = async () => {
  if (!promQuery.value) return
  try {
    const res: any = await api.get('/monitor/prometheus', { params: { query: promQuery.value } })
    promResult.value = JSON.stringify(res.data, null, 2)

    // Parse results into table data - similar to Prometheus UI
    const results = res.data?.result || []
    promTableData.value = results.map((item: any) => {
      const metric = item.metric || {}
      const name = metric.__name__ || ''
      const labels = Object.entries(metric)
        .filter(([k]) => k !== '__name__')
        .map(([k, v]) => `${k}="${v}"`)
        .join(', ')
      const value = item.value?.[1] || '-'
      const timestamp = item.value?.[0] ? new Date(item.value[0] * 1000).toLocaleString() : '-'
      return { name, labels, value, timestamp }
    })
  } catch (e) {
    promResult.value = '查询失败: ' + (e as Error).message
    promTableData.value = []
  }
}

const handleResize = () => {
  cpuMemChart?.resize()
  networkChart?.resize()
}

onMounted(async () => {
  fetchCapabilities()
  fetchPrometheusStatus()
  fetchMetricNames()
  fetchClusterMetrics()
  await nextTick()
  initCharts()
  await fetchAll()
  pollTimer = setInterval(fetchAll, 5000)
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  if (pollTimer) clearInterval(pollTimer)
  window.removeEventListener('resize', handleResize)
  cpuMemChart?.dispose()
  networkChart?.dispose()
})
</script>

<style scoped>
.monitoring-page {
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

.glass-card {
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
}

.metrics-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 20px;
}

.metric-card {
  text-align: center;
  position: relative;
}

.metric-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 12px;
}

.cpu-icon {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(139, 92, 246, 0.15));
  color: #6366f1;
}

.mem-icon {
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.2), rgba(168, 85, 247, 0.15));
  color: #8b5cf6;
}

.disk-icon {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.2), rgba(22, 163, 74, 0.15));
  color: #4ade80;
}

.pod-icon {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(37, 99, 235, 0.15));
  color: #60a5fa;
}

.metric-header {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.metric-value {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 12px;
  font-variant-numeric: tabular-nums;
}

.metric-value.low { color: #4ade80; }
.metric-value.medium { color: #fbbf24; }
.metric-value.high { color: #f87171; }
.metric-value.pods-value { color: var(--text-primary); }

.metric-sub {
  font-size: 12px;
  color: var(--text-secondary);
}

.metric-bar-track {
  height: 4px;
  background: rgba(51, 65, 85, 0.6);
  border-radius: 2px;
}

.metric-bar-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.5s ease;
}

.metric-bar-fill.low { background: linear-gradient(90deg, #4ade80, #22c55e); }
.metric-bar-fill.medium { background: linear-gradient(90deg, #fbbf24, #f59e0b); }
.metric-bar-fill.high { background: linear-gradient(90deg, #f87171, #ef4444); }

.charts-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-tag {
  font-size: 12px;
  color: var(--text-secondary);
  background: rgba(51, 65, 85, 0.5);
  padding: 4px 10px;
  border-radius: 6px;
}

.btn-action {
  background: rgba(51, 65, 85, 0.6);
  color: var(--text-secondary);
  border: 1px solid rgba(148, 163, 184, 0.1);
  padding: 6px 14px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-action:hover {
  background: rgba(51, 65, 85, 0.9);
  color: var(--text-primary);
  border-color: rgba(148, 163, 184, 0.2);
}

.btn-clear {
  background: rgba(99, 102, 241, 0.12);
  color: #818cf8;
  border-color: rgba(99, 102, 241, 0.2);
}

.btn-clear:hover {
  background: rgba(99, 102, 241, 0.25);
}

.chart-container {
  width: 100%;
  height: 280px;
}

.node-info {
  display: flex;
  justify-content: space-around;
  padding: 16px 0;
}

.info-stat {
  text-align: center;
}

.info-stat-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
  font-variant-numeric: tabular-nums;
}

.info-stat-value.ok {
  color: #4ade80;
}

.info-stat-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.prom-input-row {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.prom-input {
  flex: 1;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 8px;
  padding: 10px 16px;
  color: var(--text-primary);
  font-size: 14px;
  font-family: monospace;
  outline: none;
  transition: border-color 0.2s;
}

.prom-input:focus {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.prom-input::placeholder {
  color: var(--text-secondary);
}

.prom-result {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.08);
  padding: 16px;
  border-radius: 8px;
  font-family: monospace;
  font-size: 12px;
  color: var(--text-primary);
  max-height: 400px;
  overflow-y: auto;
  line-height: 1.6;
}

.prometheus-panel {
  margin-bottom: 24px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.prometheus-stats {
  display: flex;
  gap: 24px;
}

.pstat-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pstat-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.pstat-value {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
}

.pstat-value.alert-active {
  color: #f87171;
}

.prometheus-empty {
  text-align: center;
  padding: 20px;
}

.prometheus-empty p {
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.prometheus-hint {
  font-size: 12px;
  color: var(--text-secondary);
  opacity: 0.7;
}

.btn-small {
  padding: 6px 14px;
  font-size: 12px;
}

.btn-prometheus {
  background: linear-gradient(135deg, #f59e0b, #ef4444);
}

.quick-metrics {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.quick-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.btn-quick {
  padding: 4px 12px;
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 16px;
  background: rgba(99, 102, 241, 0.1);
  color: #818cf8;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-quick:hover {
  background: rgba(99, 102, 241, 0.2);
  border-color: rgba(99, 102, 241, 0.5);
}

.prom-input-row {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.prom-input-wrapper {
  flex: 1;
  position: relative;
}

.suggestions-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: #1e293b;
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 8px;
  max-height: 300px;
  overflow-y: auto;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.suggestion-item {
  padding: 8px 12px;
  cursor: pointer;
  transition: background 0.15s ease;
}

.suggestion-item:hover {
  background: rgba(99, 102, 241, 0.15);
}

.suggestion-name {
  font-family: monospace;
  font-size: 13px;
  color: #818cf8;
}

.prom-input-row .btn-action {
  padding: 8px 12px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 6px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
}

.prom-input-row .btn-action.active {
  background: rgba(99, 102, 241, 0.15);
  border-color: rgba(99, 102, 241, 0.4);
  color: #818cf8;
}

.prom-table-container {
  max-height: 400px;
  overflow-y: auto;
}

.metric-name {
  font-weight: 600;
  color: #818cf8 !important;
  margin-right: 8px;
}

.metric-labels {
  font-size: 12px;
  color: #94a3b8 !important;
  font-family: monospace;
}

.prom-pagination {
  display: flex;
  justify-content: center;
  padding: 12px 0;
}

.prom-chart-container {
  height: 300px;
  width: 100%;
}
</style>
