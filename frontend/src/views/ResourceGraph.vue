<template>
  <div class="graph-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">资源关系图</h1>
        <span class="page-subtitle">可视化展示Kubernetes资源关联关系</span>
      </div>
      <div class="header-actions">
        <div class="ns-selector">
          <el-select v-model="namespace" placeholder="选择命名空间" clearable @change="fetchGraph">
            <el-option label="全部命名空间" value="" />
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </div>
        <div class="resource-selector">
          <el-select v-model="selectedResource" placeholder="选择资源查看关联" clearable @change="focusResource" class="resource-select">
            <el-option-group v-for="group in resourceGroups" :key="group.label" :label="group.label">
              <el-option v-for="item in group.items" :key="item.id" :label="item.name" :value="item.id">
                <span>{{ group.icon }} {{ item.name }}</span>
                <span class="option-ns">{{ item.namespace }}</span>
              </el-option>
            </el-option-group>
          </el-select>
        </div>
        <div class="layout-switch">
          <button type="button" class="layout-btn" :class="{ active: layoutMode === 'hierarchical' }" @click="setLayout('hierarchical')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v6m0 8v6M2 12h6m8 0h6"/></svg>
            层级
          </button>
          <button type="button" class="layout-btn" :class="{ active: layoutMode === 'force' }" @click="setLayout('force')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><circle cx="4" cy="6" r="2"/><circle cx="20" cy="6" r="2"/><circle cx="4" cy="18" r="2"/><circle cx="20" cy="18" r="2"/></svg>
            力导向
          </button>
          <button type="button" class="layout-btn" :class="{ active: layoutMode === 'radial' }" @click="setLayout('radial')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="3"/></svg>
            环形
          </button>
        </div>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchGraph">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="graph-container" ref="graphRef">
      <svg :width="svgWidth" :height="svgHeight" class="graph-svg"
        @wheel.prevent="onWheel"
        @mousedown="onSvgMouseDown"
        @mousemove="onSvgMouseMove"
        @mouseup="onSvgMouseUp"
        @mouseleave="onSvgMouseUp"
      >
        <defs>
          <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="10" refY="3.5" orient="auto">
            <polygon points="0 0, 10 3.5, 0 7" fill="#6b7280" />
          </marker>
          <filter id="node-shadow">
            <feDropShadow dx="0" dy="2" stdDeviation="3" flood-opacity="0.3" />
          </filter>
        </defs>

        <g :transform="`translate(${panX}, ${panY}) scale(${scale})`">

        <!-- Edges -->
        <g class="edges">
          <path v-for="edge in graphEdges" :key="edge.id"
            :d="edge.path"
            :stroke="edge.color"
            :stroke-width="edge.width"
            fill="none"
            stroke-dasharray="6,3"
            opacity="0.6"
            marker-end="url(#arrowhead)"
            class="edge-path"
            :class="{ 'edge-highlighted': highlightedEdge === edge.id }"
          />
          <text v-for="edge in graphEdges" :key="'label-' + edge.id"
            :x="edge.labelX"
            :y="edge.labelY"
            :fill="edge.color"
            font-size="10"
            text-anchor="middle"
            class="edge-label"
          >
            {{ edge.label }}
          </text>
        </g>

        <!-- Nodes -->
        <g class="nodes">
          <g v-for="node in graphNodes" :key="node.id"
            :transform="`translate(${node.x}, ${node.y})`"
            class="node-group"
            :class="{ 'node-highlighted': highlightedNode === node.id }"
            @click="handleNodeClick(node)"
            @mouseenter="handleNodeHover(node)"
            @mouseleave="handleNodeLeave"
            @mousedown="startDrag($event, node)"
          >
            <!-- Deployment/StatefulSet/DaemonSet -->
            <g v-if="['deployment', 'statefulset', 'daemonset'].includes(node.type)">
              <rect :x="-node.width/2" :y="-node.height/2" :width="node.width" :height="node.height"
                rx="12" :fill="getNodeColor(node)" :stroke="getNodeBorder(node)" stroke-width="2" filter="url(#node-shadow)" />
              <text x="0" :y="-node.height/2 + 20" text-anchor="middle" fill="white" font-size="12" font-weight="600">
                {{ getNodeIcon(node.type) }} {{ node.name }}
              </text>
              <text x="0" :y="-node.height/2 + 36" text-anchor="middle" :fill="getStatusColor(node.status)" font-size="11">
                {{ node.replicas || '-' }}
              </text>
              <!-- Embedded Pods -->
              <g v-for="(pod, idx) in getNodePods(node)" :key="pod.id">
                <rect :x="-node.width/2 + 10 + idx * 55" :y="-node.height/2 + 44" width="50" height="24"
                  rx="6" :fill="getPodColor(pod)" stroke="rgba(255,255,255,0.2)" stroke-width="1" />
                <text :x="-node.width/2 + 35 + idx * 55" :y="-node.height/2 + 60" text-anchor="middle"
                  fill="white" font-size="9">
                  {{ pod.name.slice(-4) }}
                </text>
              </g>
            </g>

            <!-- Service -->
            <g v-else-if="node.type === 'service'">
              <rect :x="-node.width/2" :y="-node.height/2" :width="node.width" :height="node.height"
                rx="10" fill="#1e3a5f" stroke="#3b82f6" stroke-width="2" filter="url(#node-shadow)" />
              <text x="0" :y="-6" text-anchor="middle" fill="#60a5fa" font-size="12" font-weight="600">
                🌐 {{ node.name }}
              </text>
              <text x="0" y="12" text-anchor="middle" fill="#93c5fd" font-size="10">
                {{ node.ip || '-' }}:{{ node.ports?.[0] || '-' }}
              </text>
            </g>

            <!-- Ingress -->
            <g v-else-if="node.type === 'ingress'">
              <rect :x="-node.width/2" :y="-node.height/2" :width="node.width" :height="node.height"
                rx="10" fill="#2d1b4e" stroke="#8b5cf6" stroke-width="2" filter="url(#node-shadow)" />
              <text x="0" :y="-6" text-anchor="middle" fill="#a78bfa" font-size="12" font-weight="600">
                🔗 {{ node.name }}
              </text>
              <text x="0" y="12" text-anchor="middle" fill="#c4b5fd" font-size="10">
                {{ node.ip || '-' }}
              </text>
            </g>

            <!-- Pod (standalone) -->
            <g v-else-if="node.type === 'pod'">
              <rect :x="-node.width/2" :y="-node.height/2" :width="node.width" :height="node.height"
                rx="8" :fill="getPodColor(node)" stroke="rgba(255,255,255,0.2)" stroke-width="1" filter="url(#node-shadow)" />
              <text x="0" :y="-4" text-anchor="middle" fill="white" font-size="11" font-weight="500">
                {{ node.name.slice(0, 15) }}
              </text>
              <text x="0" y="10" text-anchor="middle" fill="rgba(255,255,255,0.7)" font-size="9">
                {{ node.ip || '-' }}
              </text>
            </g>

            <!-- ConfigMap/Secret -->
            <g v-else-if="['configmap', 'secret'].includes(node.type)">
              <rect :x="-node.width/2" :y="-node.height/2" :width="node.width" :height="node.height"
                rx="6" fill="#1a2332" stroke="#6b7280" stroke-width="1" filter="url(#node-shadow)" />
              <text x="0" y="4" text-anchor="middle" fill="#9ca3af" font-size="10">
                {{ node.type === 'configmap' ? '📋' : '🔑' }} {{ node.name.slice(0, 12) }}
              </text>
            </g>

            <!-- PVC -->
            <g v-else-if="node.type === 'pvc'">
              <rect :x="-node.width/2" :y="-node.height/2" :width="node.width" :height="node.height"
                rx="6" fill="#1a2e1a" stroke="#22c55e" stroke-width="1" filter="url(#node-shadow)" />
              <text x="0" y="4" text-anchor="middle" fill="#4ade80" font-size="10">
                💾 {{ node.name.slice(0, 12) }}
              </text>
            </g>

            <!-- HPA -->
            <g v-else-if="node.type === 'hpa'">
              <rect :x="-node.width/2" :y="-node.height/2" :width="node.width" :height="node.height"
                rx="6" fill="#2d1a00" stroke="#f59e0b" stroke-width="1" filter="url(#node-shadow)" />
              <text x="0" y="4" text-anchor="middle" fill="#fbbf24" font-size="10">
                📈 {{ node.name.slice(0, 12) }}
              </text>
            </g>

            <!-- Default -->
            <g v-else>
              <rect :x="-node.width/2" :y="-node.height/2" :width="node.width" :height="node.height"
                rx="6" fill="#1a2332" stroke="#6b7280" stroke-width="1" filter="url(#node-shadow)" />
              <text x="0" y="4" text-anchor="middle" fill="#9ca3af" font-size="10">
                {{ node.name.slice(0, 15) }}
              </text>
            </g>
          </g>
        </g>
        </g>
      </svg>

      <!-- Tooltip -->
      <div v-if="tooltip.visible" class="node-tooltip" :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }">
        <div class="tooltip-header">
          <span class="tooltip-icon">{{ getNodeIcon(tooltip.node?.type) }}</span>
          <span class="tooltip-name">{{ tooltip.node?.name }}</span>
        </div>
        <div class="tooltip-body">
          <div class="tooltip-row"><span>类型:</span><span>{{ tooltip.node?.type }}</span></div>
          <div class="tooltip-row"><span>命名空间:</span><span>{{ tooltip.node?.namespace }}</span></div>
          <div class="tooltip-row" v-if="tooltip.node?.status"><span>状态:</span><span :style="{ color: getStatusColor(tooltip.node.status) }">{{ tooltip.node.status }}</span></div>
          <div class="tooltip-row" v-if="tooltip.node?.ip"><span>IP:</span><span>{{ tooltip.node.ip }}</span></div>
          <div class="tooltip-row" v-if="tooltip.node?.replicas"><span>副本:</span><span>{{ tooltip.node.replicas }}</span></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { useGlobalNamespace } from '../store/namespace'
import api from '../api'

const router = useRouter()
const { namespace } = useGlobalNamespace()
const graphRef = ref<HTMLElement>()
const svgWidth = ref(1200)
const svgHeight = ref(800)
const nsList = ref<any[]>([])
const graphData = ref<any>({ nodes: [], edges: [] })
const layoutMode = ref('force')
const highlightedNode = ref('')
const highlightedEdge = ref('')
const selectedResource = ref('')
const focusedNodes = ref<Set<string>>(new Set())
const dragTick = ref(0)

// Pan and zoom
const scale = ref(1)
const panX = ref(0)
const panY = ref(0)
let isPanning = false
let panStartX = 0
let panStartY = 0
let panStartPanX = 0
let panStartPanY = 0

const tooltip = ref({
  visible: false,
  x: 0,
  y: 0,
  node: null as any,
})

interface LayoutNode {
  id: string
  type: string
  name: string
  namespace: string
  status: string
  labels: Record<string, string>
  ports: number[]
  ip: string
  replicas: string
  children: string[]
  x: number
  y: number
  width: number
  height: number
}

const graphNodes = computed<LayoutNode[]>(() => {
  const nodes = graphData.value.nodes || []
  const edges = graphData.value.edges || []
  // If a resource is focused, only show related nodes
  if (focusedNodes.value.size > 0) {
    const filteredNodes = nodes.filter((n: any) => focusedNodes.value.has(n.id))
    return layoutNodes(filteredNodes, edges)
  }
  return layoutNodes(nodes, edges)
})

const resourceGroups = computed(() => {
  const nodes = graphData.value.nodes || []
  const groups: Record<string, any[]> = {
    deployment: [],
    statefulset: [],
    daemonset: [],
    service: [],
    ingress: [],
    pod: [],
  }
  nodes.forEach((n: any) => {
    if (groups[n.type]) {
      groups[n.type].push(n)
    }
  })
  const groupMeta: Record<string, { label: string; icon: string }> = {
    deployment: { label: 'Deployment', icon: '🚀' },
    statefulset: { label: 'StatefulSet', icon: '📦' },
    daemonset: { label: 'DaemonSet', icon: '🔄' },
    service: { label: 'Service', icon: '🌐' },
    ingress: { label: 'Ingress', icon: '🔗' },
    pod: { label: 'Pod', icon: '🟢' },
  }
  return Object.entries(groups)
    .filter(([_, items]) => items.length > 0)
    .map(([type, items]) => ({
      label: groupMeta[type].label,
      icon: groupMeta[type].icon,
      items,
    }))
})

const graphEdges = computed(() => {
  void dragTick.value // trigger recompute on drag

  const edges = graphData.value.edges || []
  const nodes = graphNodes.value
  const nodeMap = new Map(nodes.map((n: any) => [n.id, n]))

  const filteredEdges = focusedNodes.value.size > 0
    ? edges.filter((e: any) => focusedNodes.value.has(e.source) && focusedNodes.value.has(e.target))
    : edges

  return filteredEdges.map((edge: any) => {
    const source = nodeMap.get(edge.source)
    const target = nodeMap.get(edge.target)
    if (!source || !target) return null

    const dx = target.x - source.x
    const dy = target.y - source.y
    const dist = Math.sqrt(dx * dx + dy * dy)
    const sx = source.x + (dx / dist) * (source.width / 2)
    const sy = source.y + (dy / dist) * (source.height / 2)
    const tx = target.x - (dx / dist) * (target.width / 2)
    const ty = target.y - (dy / dist) * (target.height / 2)

    const midX = (sx + tx) / 2
    const midY = (sy + ty) / 2

    return {
      id: `${edge.source}-${edge.target}`,
      path: `M${sx},${sy} Q${midX},${midY - 20} ${tx},${ty}`,
      color: getEdgeColor(edge.type),
      width: 1.5,
      label: edge.label || '',
      labelX: midX,
      labelY: midY - 25,
    }
  }).filter(Boolean)
})

function layoutNodes(nodes: any[], edges: any[]): LayoutNode[] {
  const nodeMap = new Map<string, LayoutNode>()
  const childMap = new Map<string, string[]>()

  // Initialize nodes
  nodes.forEach(node => {
    const layoutNode: LayoutNode = {
      ...node,
      x: 0,
      y: 0,
      width: getNodeWidth(node.type),
      height: getNodeHeight(node.type),
    }
    nodeMap.set(node.id, layoutNode)
  })

  // Build parent-child relationships
  edges.forEach(edge => {
    if (edge.type === 'ownership' || edge.type === 'selector') {
      if (!childMap.has(edge.source)) {
        childMap.set(edge.source, [])
      }
      childMap.get(edge.source)!.push(edge.target)
    }
  })

  // Layout based on mode
  if (layoutMode.value === 'hierarchical') {
    layoutHierarchical(nodeMap, childMap)
  } else if (layoutMode.value === 'force') {
    layoutForce(nodeMap, edges)
  } else {
    layoutRadial(nodeMap)
  }

  return Array.from(nodeMap.values())
}

function layoutHierarchical(nodeMap: Map<string, LayoutNode>, childMap: Map<string, string[]>) {
  const layers = [
    ['ingress'],
    ['service'],
    ['deployment', 'statefulset', 'daemonset'],
    ['pod'],
    ['configmap', 'secret', 'pvc', 'hpa'],
  ]

  const centerX = svgWidth.value / 2
  const layerHeight = svgHeight.value / (layers.length + 1)

  layers.forEach((types, layerIdx) => {
    const layerNodes = Array.from(nodeMap.values()).filter(n => types.includes(n.type))
    const totalWidth = layerNodes.length * 200
    const startX = centerX - totalWidth / 2 + 100

    layerNodes.forEach((node, idx) => {
      node.x = startX + idx * 200
      node.y = layerHeight * (layerIdx + 1)
    })
  })
}

function layoutForce(nodeMap: Map<string, LayoutNode>, edges: any[]) {
  const nodes = Array.from(nodeMap.values())
  const centerX = svgWidth.value / 2
  const centerY = svgHeight.value / 2

  // Initialize positions in a grid pattern based on type
  const typeOrder = ['ingress', 'service', 'deployment', 'statefulset', 'daemonset', 'pod', 'configmap', 'secret', 'pvc', 'hpa']
  const typeGroups = new Map<string, LayoutNode[]>()
  nodes.forEach(node => {
    if (!typeGroups.has(node.type)) typeGroups.set(node.type, [])
    typeGroups.get(node.type)!.push(node)
  })

  let rowY = 100
  typeOrder.forEach(type => {
    const group = typeGroups.get(type) || []
    const totalWidth = group.length * 200
    const startX = centerX - totalWidth / 2 + 100
    group.forEach((node, idx) => {
      node.x = startX + idx * 200
      node.y = rowY
    })
    if (group.length > 0) rowY += 150
  })

  // Force simulation with more iterations
  for (let i = 0; i < 100; i++) {
    // Repulsion between all nodes
    for (let a = 0; a < nodes.length; a++) {
      for (let b = a + 1; b < nodes.length; b++) {
        const dx = nodes[b].x - nodes[a].x
        const dy = nodes[b].y - nodes[a].y
        const dist = Math.max(1, Math.sqrt(dx * dx + dy * dy))
        const minDist = 180
        if (dist < minDist) {
          const force = (minDist - dist) * 0.3
          const fx = (dx / dist) * force
          const fy = (dy / dist) * force
          nodes[a].x -= fx
          nodes[a].y -= fy
          nodes[b].x += fx
          nodes[b].y += fy
        }
      }
    }

    // Attraction along edges
    edges.forEach(edge => {
      const source = nodeMap.get(edge.source)
      const target = nodeMap.get(edge.target)
      if (source && target) {
        const dx = target.x - source.x
        const dy = target.y - source.y
        const dist = Math.max(1, Math.sqrt(dx * dx + dy * dy))
        const idealDist = 200
        const force = (dist - idealDist) * 0.02
        const fx = (dx / dist) * force
        const fy = (dy / dist) * force
        source.x += fx
        source.y += fy
        target.x -= fx
        target.y -= fy
      }
    })

    // Center gravity
    nodes.forEach(node => {
      node.x += (centerX - node.x) * 0.005
      node.y += (centerY - node.y) * 0.005
    })
  }

  // Constrain to viewport
  nodes.forEach(node => {
    node.x = Math.max(150, Math.min(svgWidth.value - 150, node.x))
    node.y = Math.max(80, Math.min(svgHeight.value - 80, node.y))
  })
}

function layoutRadial(nodeMap: Map<string, LayoutNode>) {
  const nodes = Array.from(nodeMap.values())
  const centerX = svgWidth.value / 2
  const centerY = svgHeight.value / 2

  const typeOrder = ['ingress', 'service', 'deployment', 'statefulset', 'daemonset', 'pod', 'configmap', 'secret', 'pvc', 'hpa']
  const sortedNodes = nodes.sort((a, b) => typeOrder.indexOf(a.type) - typeOrder.indexOf(b.type))

  sortedNodes.forEach((node, idx) => {
    const angle = (2 * Math.PI * idx) / sortedNodes.length - Math.PI / 2
    const radius = 250
    node.x = centerX + radius * Math.cos(angle)
    node.y = centerY + radius * Math.sin(angle)
  })
}

function getNodeWidth(type: string): number {
  switch (type) {
    case 'deployment': case 'statefulset': case 'daemonset': return 200
    case 'service': case 'ingress': return 160
    case 'pod': return 120
    default: return 100
  }
}

function getNodeHeight(type: string): number {
  switch (type) {
    case 'deployment': case 'statefulset': case 'daemonset': return 100
    case 'service': case 'ingress': return 50
    case 'pod': return 40
    default: return 30
  }
}

function getNodeColor(node: any): string {
  switch (node.status) {
    case 'running': return '#1a3a2a'
    case 'warning': return '#3a2a1a'
    case 'error': return '#3a1a1a'
    default: return '#1a2332'
  }
}

function getNodeBorder(node: any): string {
  switch (node.status) {
    case 'running': return '#22c55e'
    case 'warning': return '#f59e0b'
    case 'error': return '#ef4444'
    default: return '#6b7280'
  }
}

function getPodColor(pod: any): string {
  switch (pod.status) {
    case 'running': return '#166534'
    case 'pending': return '#854d0e'
    case 'error': case 'failed': return '#991b1b'
    default: return '#374151'
  }
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'running': return '#4ade80'
    case 'warning': return '#fbbf24'
    case 'error': case 'failed': return '#f87171'
    case 'pending': return '#fbbf24'
    default: return '#9ca3af'
  }
}

function getEdgeColor(type: string): string {
  switch (type) {
    case 'selector': return '#3b82f6'
    case 'routing': return '#8b5cf6'
    case 'ownership': return '#22c55e'
    case 'volume': return '#f59e0b'
    case 'config': return '#6b7280'
    case 'autoscale': return '#ec4899'
    default: return '#6b7280'
  }
}

function getNodeIcon(type: string): string {
  switch (type) {
    case 'deployment': return '🚀'
    case 'statefulset': return '📦'
    case 'daemonset': return '🔄'
    case 'service': return '🌐'
    case 'ingress': return '🔗'
    case 'pod': return '🟢'
    case 'configmap': return '📋'
    case 'secret': return '🔑'
    case 'pvc': return '💾'
    case 'hpa': return '📈'
    default: return '📄'
  }
}

function getNodePods(node: any): any[] {
  const edges = graphData.value.edges || []
  const podIds = edges
    .filter((e: any) => e.source === node.id && e.type === 'ownership')
    .map((e: any) => e.target)
  return (graphData.value.nodes || []).filter((n: any) => podIds.includes(n.id))
}

// Pan and zoom handlers
function onWheel(e: WheelEvent) {
  const delta = e.deltaY > 0 ? -0.1 : 0.1
  scale.value = Math.max(0.2, Math.min(3, scale.value + delta))
}

function onSvgMouseDown(e: MouseEvent) {
  // Only start panning if clicking on the SVG background (not a node)
  if ((e.target as Element).closest('.node-group')) return
  isPanning = true
  panStartX = e.clientX
  panStartY = e.clientY
  panStartPanX = panX.value
  panStartPanY = panY.value
}

function onSvgMouseMove(e: MouseEvent) {
  if (!isPanning) return
  panX.value = panStartPanX + (e.clientX - panStartX) / scale.value
  panY.value = panStartPanY + (e.clientY - panStartY) / scale.value
}

function onSvgMouseUp() {
  isPanning = false
}

function setLayout(mode: string) {
  layoutMode.value = mode
}

function focusResource(resourceId: string) {
  if (!resourceId) {
    focusedNodes.value = new Set()
    return
  }
  // Find all related nodes recursively
  const edges = graphData.value.edges || []
  const relatedIds = new Set<string>()
  relatedIds.add(resourceId)

  // BFS to find all connected nodes
  let frontier = new Set<string>([resourceId])
  for (let depth = 0; depth < 5; depth++) {
    const nextFrontier = new Set<string>()
    edges.forEach((edge: any) => {
      if (frontier.has(edge.source) && !relatedIds.has(edge.target)) {
        nextFrontier.add(edge.target)
        relatedIds.add(edge.target)
      }
      if (frontier.has(edge.target) && !relatedIds.has(edge.source)) {
        nextFrontier.add(edge.source)
        relatedIds.add(edge.source)
      }
    })
    if (nextFrontier.size === 0) break
    frontier = nextFrontier
  }

  focusedNodes.value = relatedIds
}

function handleNodeClick(node: any) {
  if (didDrag) return
  const routeMap: Record<string, string> = {
    deployment: '/deployments',
    statefulset: '/statefulsets',
    daemonset: '/daemonsets',
    service: '/services',
    ingress: '/ingresses',
    pod: '/pods',
    configmap: '/configmaps',
    secret: '/secrets',
    pvc: '/persistentvolumeclaims',
    hpa: '/hpas',
  }
  const route = routeMap[node.type]
  if (route) {
    router.push({ path: route, query: { highlight: node.name, ns: node.namespace } })
  }
}

function handleNodeHover(node: any) {
  highlightedNode.value = node.id
  tooltip.value = {
    visible: true,
    x: node.x * scale.value + panX.value + 50,
    y: node.y * scale.value + panY.value - 50,
    node,
  }
}

function handleNodeLeave() {
  highlightedNode.value = ''
  tooltip.value.visible = false
}

let dragNode: any = null
let dragOffsetX = 0
let dragOffsetY = 0
let didDrag = false
let dragAnimFrame: number | null = null

function startDrag(event: MouseEvent, node: any) {
  event.stopPropagation()
  dragNode = node
  didDrag = false
  dragOffsetX = (event.clientX - panX.value) / scale.value - node.x
  dragOffsetY = (event.clientY - panY.value) / scale.value - node.y

  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
}

function onDrag(event: MouseEvent) {
  if (!dragNode) return
  didDrag = true
  const newX = (event.clientX - panX.value) / scale.value - dragOffsetX
  const newY = (event.clientY - panY.value) / scale.value - dragOffsetY

  if (dragAnimFrame) cancelAnimationFrame(dragAnimFrame)
  dragAnimFrame = requestAnimationFrame(() => {
    if (dragNode) {
      dragNode.x = newX
      dragNode.y = newY
      dragTick.value++
    }
  })
}

function stopDrag() {
  if (dragAnimFrame) cancelAnimationFrame(dragAnimFrame)
  dragAnimFrame = null
  dragNode = null
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}

const fetchGraph = async () => {
  try {
    const params = namespace.value ? { namespace: namespace.value } : {}
    const res: any = await api.get('/graph', { params })
    graphData.value = res.data || { nodes: [], edges: [] }
  } catch (e) {
    console.error(e)
  }
}

const fetchNs = async () => {
  try {
    const res: any = await api.get('/namespaces')
    nsList.value = res.data || []
  } catch (e) {
    console.error(e)
  }
}

onMounted(async () => {
  await nextTick()
  if (graphRef.value) {
    svgWidth.value = graphRef.value.clientWidth
    svgHeight.value = graphRef.value.clientHeight
  }
  fetchNs()
  fetchGraph()
})
</script>

<style scoped>
.graph-page {
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

.ns-selector :deep(.el-select) {
  width: 200px;
}

.ns-selector :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 8px;
  box-shadow: none;
}

.ns-selector :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.ns-selector :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.ns-selector :deep(.el-input__inner) {
  color: var(--text-primary);
}

.ns-selector :deep(.el-input__inner::placeholder) {
  color: var(--text-secondary);
}

.resource-selector {
  display: flex;
  align-items: center;
}

.resource-select {
  width: 300px;
}

.resource-select :deep(.el-input__wrapper) {
  background: rgba(30, 41, 59, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.1);
  box-shadow: none;
}

.resource-select :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.resource-select :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.resource-select :deep(.el-input__inner) {
  color: var(--text-primary);
}

.option-ns {
  float: right;
  color: var(--text-secondary);
  font-size: 12px;
}

.layout-switch {
  display: flex;
  gap: 4px;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 8px;
  padding: 4px;
}

.layout-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.layout-btn:hover {
  background: rgba(99, 102, 241, 0.15);
  color: var(--text-primary);
}

.layout-btn.active {
  background: rgba(99, 102, 241, 0.25);
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

.graph-container {
  flex: 1;
  background: rgba(30, 41, 59, 0.6);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.3);
  position: relative;
  overflow: hidden;
}

.graph-svg {
  width: 100%;
  height: 100%;
  cursor: grab;
}

.graph-svg:active {
  cursor: grabbing;
}

.node-group {
  cursor: pointer;
  transition: transform 0.1s ease;
}

.node-group:hover {
  filter: brightness(1.2);
}

.node-highlighted rect {
  stroke-width: 3;
}

.edge-path {
  transition: opacity 0.2s ease;
}

.edge-highlighted {
  opacity: 1;
  stroke-width: 2.5;
}

.edge-label {
  pointer-events: none;
}

.node-tooltip {
  position: absolute;
  background: rgba(30, 41, 59, 0.95);
  border: 1px solid rgba(99, 102, 241, 0.3);
  border-radius: 8px;
  padding: 12px;
  min-width: 180px;
  z-index: 100;
  backdrop-filter: blur(8px);
}

.tooltip-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
}

.tooltip-icon {
  font-size: 16px;
}

.tooltip-name {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 13px;
}

.tooltip-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.tooltip-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.tooltip-row span:first-child {
  color: var(--text-secondary);
}

.tooltip-row span:last-child {
  color: var(--text-primary);
}
</style>
