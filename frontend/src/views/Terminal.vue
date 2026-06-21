<template>
  <div class="terminal-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">WebTerminal</h1>
        <span class="page-subtitle">连接到集群中的 Pod 执行命令</span>
      </div>
      <div class="header-actions">
        <el-tag v-if="wsConnected" type="success" size="small" effect="dark" class="status-tag">已连接</el-tag>
        <el-tag v-else type="info" size="small" effect="dark" class="status-tag">未连接</el-tag>
      </div>
    </div>

    <div class="terminal-toolbar">
      <div class="toolbar-left">
        <el-select v-model="selectedNs" placeholder="Namespace" @change="onNsChange">
          <el-option v-for="ns in namespaces" :key="ns.name" :label="ns.name" :value="ns.name" />
        </el-select>
        <el-select v-model="selectedPod" placeholder="Pod" @change="onPodChange">
          <el-option v-for="p in pods" :key="p.name" :label="p.name" :value="p.name" />
        </el-select>
        <el-select v-model="selectedContainer" placeholder="Container" :disabled="!selectedPod">
          <el-option v-for="c in containers" :key="c" :label="c" :value="c" />
        </el-select>
        <el-select v-model="selectedShell" placeholder="Shell">
          <el-option v-for="s in shells" :key="s.cmd" :label="s.label" :value="s.cmd" />
        </el-select>
        <button type="button" class="btn-gradient" @click="connectTerminal" :disabled="!canConnect">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
          {{ wsConnected ? '已连接' : '连接' }}
        </button>
        <button type="button" class="btn-outline" @click="disconnectTerminal" :disabled="!wsConnected">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          断开
        </button>
      </div>
    </div>

    <div ref="terminalRef" class="terminal-output"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Terminal } from 'xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import 'xterm/css/xterm.css'
import api from '../api'

const route = useRoute()
const terminalRef = ref<HTMLElement>()
const selectedNs = ref('')
const selectedPod = ref('')
const selectedContainer = ref('')
const namespaces = ref<any[]>([])
const pods = ref<any[]>([])
const containers = ref<string[]>([])
const selectedShell = ref('/bin/sh')
const shells = [
  { label: '/bin/sh', cmd: '/bin/sh' },
  { label: '/bin/bash', cmd: '/bin/bash' },
  { label: '/usr/bin/sh', cmd: '/usr/bin/sh' },
  { label: '/usr/bin/bash', cmd: '/usr/bin/bash' },
  { label: '/bin/zsh', cmd: '/bin/zsh' },
  { label: '/usr/bin/zsh', cmd: '/usr/bin/zsh' },
]
const wsConnected = ref(false)

const canConnect = computed(() => selectedNs.value && selectedPod.value && !wsConnected.value)

let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let ws: WebSocket | null = null

const initTerminal = () => {
  if (!terminalRef.value || terminal) return

  terminal = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    theme: {
      background: '#0f172a',
      foreground: '#e2e8f0',
      cursor: '#818cf8',
      selectionBackground: 'rgba(99, 102, 241, 0.3)',
    },
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())

  terminal.open(terminalRef.value)
  fitAddon.fit()

  terminal.onData((data) => {
    if (!ws || ws.readyState !== WebSocket.OPEN) return
    ws.send(JSON.stringify({ type: 'input', data }))
  })

  terminal.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'resize', cols, rows }))
    }
  })

  terminal.writeln('\x1b[38;2;99;102;241mKubeOps WebTerminal\x1b[0m')
  terminal.writeln('\x1b[38;2;148;163;184m选择 Namespace 和 Pod 后点击「连接」\x1b[0m')
  terminal.writeln('')
}

const connectTerminal = () => {
  if (!selectedNs.value || !selectedPod.value) {
    ElMessage.warning('请先选择 Namespace 和 Pod')
    return
  }

  if (ws) {
    ws.close()
    ws = null
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsHost = window.location.hostname
  const token = localStorage.getItem('token') || ''
  const wsUrl = `${protocol}//${wsHost}:8080/api/pods/exec/ws?token=${token}`

  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    const cols = terminal?.cols || 80
    const rows = terminal?.rows || 24
    ws!.send(JSON.stringify({
      namespace: selectedNs.value,
      pod: selectedPod.value,
      container: selectedContainer.value,
      command: [selectedShell.value, '-il'],
      columns: cols,
      rows: rows,
    }))
    wsConnected.value = true
  }

  ws.onmessage = (event) => {
    const processData = (data: string) => {
      try {
        const json = JSON.parse(data)
        if (json.error) {
          terminal?.writeln(`\x1b[31mError: ${json.error}\x1b[0m`)
          return
        }
      } catch {}
      terminal?.write(data)
    }
    if (event.data instanceof Blob) {
      event.data.text().then(processData)
    } else if (typeof event.data === 'string') {
      processData(event.data)
    } else {
      processData(new TextDecoder().decode(event.data))
    }
  }

  ws.onerror = (e) => {
    console.error('WebSocket error:', e)
    ElMessage.error('连接失败')
    wsConnected.value = false
  }

  ws.onclose = () => {
    if (wsConnected.value) {
      terminal?.writeln('\r\n\x1b[33m连接已断开\x1b[0m')
    }
    wsConnected.value = false
    ws = null
  }
}

const disconnectTerminal = () => {
  if (ws) {
    ws.close()
    ws = null
  }
  wsConnected.value = false
}

const fetchNamespaces = async () => {
  try {
    const res: any = await api.get('/namespaces')
    namespaces.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

const onNsChange = async () => {
  selectedPod.value = ''
  selectedContainer.value = ''
  pods.value = []
  containers.value = []
  try {
    const res: any = await api.get('/pods', { params: { namespace: selectedNs.value } })
    pods.value = (res.data || []).filter((p: any) => p.status === 'Running')
  } catch (e) {
    console.error(e)
  }
}

const onPodChange = () => {
  selectedContainer.value = ''
  const pod = pods.value.find((p: any) => p.name === selectedPod.value)
  containers.value = (pod?.containers || []).map((c: any) => c.name)
  if (containers.value.length === 1) {
    selectedContainer.value = containers.value[0]
  }
}

const handleResize = () => {
  fitAddon?.fit()
}

onMounted(async () => {
  await nextTick()
  initTerminal()
  await fetchNamespaces()

  if (route.query.namespace) {
    selectedNs.value = route.query.namespace as string
    await onNsChange()
  }
  if (route.query.pod) {
    selectedPod.value = route.query.pod as string
    onPodChange()
  }

  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  if (ws) {
    ws.close()
    ws = null
  }
  terminal?.dispose()
  terminal = null
})
</script>

<style scoped>
.terminal-page {
  padding: 24px;
  background: var(--bg-primary);
  min-height: 100vh;
  display: flex;
  flex-direction: column;
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

.status-tag {
  font-weight: 500;
}

.terminal-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-bottom: none;
  border-radius: 16px 16px 0 0;
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-left :deep(.el-select) {
  width: 160px;
}

.toolbar-left :deep(.el-input__wrapper) {
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 8px;
  box-shadow: none;
}

.toolbar-left :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.toolbar-left :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.toolbar-left :deep(.el-input__inner) {
  color: var(--text-primary);
}

.toolbar-left :deep(.el-input__inner::placeholder) {
  color: var(--text-secondary);
}

.btn-gradient {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 8px;
  font-weight: 500;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-gradient:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.4), 0 0 20px rgba(99, 102, 241, 0.3);
}

.btn-gradient:active {
  transform: translateY(0);
}

.btn-gradient:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-outline {
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid rgba(148, 163, 184, 0.2);
  padding: 8px 16px;
  border-radius: 8px;
  font-weight: 500;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn-outline:hover:not(:disabled) {
  background: rgba(51, 65, 85, 0.4);
  border-color: rgba(148, 163, 184, 0.3);
  color: var(--text-primary);
}

.btn-outline:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-sm {
  padding: 6px 14px;
  font-size: 12px;
}

.terminal-output {
  flex: 1;
  background: #0f172a;
  border-left: 1px solid rgba(148, 163, 184, 0.08);
  border-right: 1px solid rgba(148, 163, 184, 0.08);
  min-height: 500px;
}

</style>
