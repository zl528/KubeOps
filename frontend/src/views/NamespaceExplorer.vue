<template>
  <div class="mindmap-container">
    <div class="toolbar">
      <div class="cluster-selector" @click="showClusterMenu = !showClusterMenu">
        <span class="cluster-icon">🌳</span>
        <span class="cluster-name">{{ activeCluster || '选择集群' }}</span>
        <span class="arrow">▼</span>
      </div>
      <div class="cluster-menu" v-if="showClusterMenu">
        <div v-for="c in clusters" :key="c.name" class="cluster-item" :class="{ active: c.name === activeCluster }" @click="switchCluster(c.name)">{{ c.name }}</div>
      </div>
      <div class="zoom-controls">
        <button type="button" class="zoom-btn" @click="zoomIn">+</button>
        <span class="zoom-level">{{ Math.round(scale * 100) }}%</span>
        <button type="button" class="zoom-btn" @click="zoomOut">−</button>
        <button type="button" class="zoom-btn" @click="resetView">⟲</button>
      </div>
    </div>

    <div class="mindmap-main" ref="mindmapRef" @wheel.prevent="onWheel" @mousedown="onMouseDown" @mousemove="onMouseMove" @mouseup="onMouseUp" @mouseleave="onMouseUp">
      <svg :width="svgWidth" :height="svgHeight" class="mindmap-svg">
        <defs>
          <filter id="glow"><feGaussianBlur stdDeviation="2" result="b"/><feMerge><feMergeNode in="b"/><feMergeNode in="SourceGraphic"/></feMerge></filter>
        </defs>
        <g :transform="`translate(${panX}, ${panY}) scale(${scale})`">
          <!-- 所有连线 -->
          <template v-for="line in allLines" :key="line.id">
            <path :d="line.path" :stroke="line.color" :stroke-width="line.width" fill="none" opacity="0.6" class="animated-path"/>
          </template>

          <!-- 中心节点 -->
          <g class="tree-node center-node" @click="collapseAll" :transform="`translate(${treeLayout.center.x}, ${treeLayout.center.y})`">
            <rect x="-50" y="-22" width="100" height="44" rx="22" fill="#1a1a2e" stroke="#67C23A" stroke-width="3" filter="url(#glow)"/>
            <text x="0" y="6" text-anchor="middle" fill="#67C23A" font-size="13" font-weight="bold">🌳 {{ activeCluster || 'cluster' }}</text>
          </g>

          <!-- 命名空间节点 -->
          <template v-for="ns in treeLayout.namespaces" :key="ns.name">
            <path :d="ns.connectPath" :stroke="ns.color" stroke-width="2" fill="none" opacity="0.5" class="animated-path"/>
            <g class="tree-node ns-node" @click.stop="toggleNamespace(ns.name)" :transform="`translate(${ns.x}, ${ns.y})`">
              <rect :x="-ns.width/2" y="-16" :width="ns.width" height="32" rx="16" :fill="ns.expanded ? ns.color : '#1a1a2e'" :stroke="ns.color" stroke-width="2"/>
              <text x="0" y="5" text-anchor="middle" :fill="ns.expanded ? '#fff' : ns.color" font-size="11" font-weight="500">{{ ns.name }}</text>
              <text :x="ns.width/2 + 12" y="5" fill="#666" font-size="10">{{ ns.expanded ? '−' : '+' }}</text>
            </g>

            <!-- 资源类型节点 -->
            <template v-if="ns.expanded">
              <template v-for="res in ns.resources" :key="res.type">
                <path :d="res.connectPath" :stroke="res.color" stroke-width="1.5" fill="none" opacity="0.5" class="animated-path"/>
                <g class="tree-node res-node" @click.stop="toggleResource(ns.name, res.type)" :transform="`translate(${res.x}, ${res.y})`">
                  <rect x="-55" y="-14" width="110" height="28" rx="14" :fill="res.expanded ? res.color : '#1a1a2e'" :stroke="res.color" stroke-width="2"/>
                  <text x="0" y="5" text-anchor="middle" :fill="res.expanded ? '#fff' : res.color" font-size="10">{{ res.label }} ({{ res.count }})</text>
                </g>

                <!-- 具体资源节点 -->
                <template v-if="res.expanded && res.items.length > 0">
                  <template v-for="(item, ii) in res.items" :key="item.name">
                    <path :d="item.connectPath" :stroke="res.color" stroke-width="1" fill="none" opacity="0.3" class="animated-path"/>
                    <g class="tree-node item-node" @click.stop="handleItemClick(res.type, item)" :transform="`translate(${item.x}, ${item.y})`">
                      <rect x="-70" y="-11" width="140" height="22" rx="11" :fill="isItemExpanded(res.type, item.name) ? res.color + '30' : '#2d2d5e'" :stroke="res.color" stroke-width="1"/>
                      <text x="-60" y="4" text-anchor="start" fill="#e0e0e0" font-size="9">{{ truncate(item.name, 16) }}</text>
                      <text x="60" y="4" text-anchor="end" fill="#666" font-size="8">{{ isItemExpanded(res.type, item.name) ? '−' : '+' }}</text>
                      <title>{{ item.name }} ({{ item.status || item.phase || '' }})</title>
                    </g>

                    <!-- 展开的操作节点 -->
                    <template v-if="isItemExpanded(res.type, item.name)">
                      <!-- Pod 操作按钮 -->
                      <template v-if="res.type === 'pods' && !getPanel(res.type, item.name)">
                        <g class="tree-node action-node" @click.stop="openTerminal(item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y - 20})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#67C23A" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#67C23A" font-size="9">⌨ 终端</text>
                        </g>
                        <g class="tree-node action-node" @click.stop="openLogs(item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#409EFF" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#409EFF" font-size="9">📋 日志</text>
                        </g>
                        <g class="tree-node action-node" @click.stop="deleteResource('pods', item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y + 20})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#F56C6C" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#F56C6C" font-size="9">🗑 删除</text>
                        </g>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y - 20" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y + 20" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                      </template>

                      <!-- 日志面板 -->
                      <template v-if="res.type === 'pods' && getPanel(res.type, item.name) === 'logs'">
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 140" :y2="item.y" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <g :transform="`translate(${item.x + 140 + getPanelState(res.type, item.name).x}, ${item.y - getPanelState(res.type, item.name).h/2 + getPanelState(res.type, item.name).y})`">
                          <rect x="0" y="0" :width="getPanelState(res.type, item.name).w" :height="getPanelState(res.type, item.name).h" rx="8" fill="#0d0d1a" stroke="#409EFF" stroke-width="1"/>
                          <rect x="0" y="0" :width="getPanelState(res.type, item.name).w" height="28" rx="8" fill="#1a1a2e"/>
                          <text x="12" y="18" fill="#409EFF" font-size="11" font-weight="bold">📋 {{ item.name }} - 日志</text>
                          <g class="window-btn" @click.stop="toggleMaximize(res.type, item.name)" :transform="`translate(${getPanelState(res.type, item.name).w - 55}, 14)`">
                            <rect x="-10" y="-8" width="16" height="16" rx="3" fill="none" stroke="#666" stroke-width="1"/>
                            <rect x="-7" y="-5" width="10" height="10" rx="1" fill="none" stroke="#666" stroke-width="1"/>
                          </g>
                          <g class="window-btn close-btn" @click.stop="closePanel(res.type, item.name)" :transform="`translate(${getPanelState(res.type, item.name).w - 22}, 14)`">
                            <text x="0" y="5" fill="#666" font-size="16" text-anchor="middle">×</text>
                          </g>
                          <foreignObject x="8" y="34" :width="getPanelState(res.type, item.name).w - 16" :height="getPanelState(res.type, item.name).h - 42">
                            <div xmlns="http://www.w3.org/1999/xhtml" style="overflow-y:auto;height:100%;padding:4px;font-family:monospace;font-size:11px;color:#90EE90;white-space:pre-wrap;background:#0a0a15;border-radius:4px;">{{ logLoading ? '加载中...' : logContent }}</div>
                          </foreignObject>
                          <g class="resize-handle" @mousedown.stop="startResize($event, res.type, item.name)" :transform="`translate(${getPanelState(res.type, item.name).w - 10}, ${getPanelState(res.type, item.name).h - 10})`">
                            <rect x="0" y="0" width="10" height="10" fill="transparent" style="cursor:nwse-resize"/>
                            <path d="M10,0 L0,10 M10,4 L4,10 M10,8 L8,10" stroke="#666" stroke-width="1"/>
                          </g>
                        </g>
                      </template>

                      <!-- 终端面板 -->
                      <template v-if="res.type === 'pods' && getPanel(res.type, item.name) === 'terminal'">
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 140" :y2="item.y" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <g :transform="`translate(${item.x + 140 + getPanelState(res.type, item.name).x}, ${item.y - getPanelState(res.type, item.name).h/2 + getPanelState(res.type, item.name).y})`">
                          <rect x="0" y="0" :width="getPanelState(res.type, item.name).w" :height="getPanelState(res.type, item.name).h" rx="8" fill="#0d0d1a" stroke="#67C23A" stroke-width="1"/>
                          <rect x="0" y="0" :width="getPanelState(res.type, item.name).w" height="28" rx="8" fill="#1a1a2e"/>
                          <text x="12" y="18" fill="#67C23A" font-size="11" font-weight="bold">⌨ {{ item.name }} - 终端</text>
                          <!-- Shell 选择器 -->
                          <foreignObject :x="getPanelState(res.type, item.name).w - 260" y="4" width="120" height="22">
                            <select xmlns="http://www.w3.org/1999/xhtml" :value="terminalShell" @change="changeShell(item.ns, item.name, $event.target.value)" style="width:100%;height:100%;background:#0d0d1a;border:1px solid #3d3d6e;color:#90EE90;font-size:10px;border-radius:4px;padding:0 4px;outline:none;cursor:pointer;">
                              <option value="/bin/sh">/bin/sh</option>
                              <option value="/bin/bash">/bin/bash</option>
                              <option value="/usr/bin/sh">/usr/bin/sh</option>
                              <option value="/usr/bin/bash">/usr/bin/bash</option>
                              <option value="/bin/zsh">/bin/zsh</option>
                              <option value="/usr/bin/zsh">/usr/bin/zsh</option>
                            </select>
                          </foreignObject>
                          <!-- 重连按钮 -->
                          <g class="window-btn" @click.stop="changeShell(item.ns, item.name, terminalShell)" :transform="`translate(${getPanelState(res.type, item.name).w - 130}, 14)`">
                            <rect x="-12" y="-8" width="24" height="16" rx="3" fill="none" stroke="#67C23A" stroke-width="1"/>
                            <text x="0" y="4" text-anchor="middle" fill="#67C23A" font-size="9">重连</text>
                          </g>
                          <circle :cx="getPanelState(res.type, item.name).w - 95" cy="14" r="5" :fill="terminalConnected ? '#67C23A' : '#F56C6C'"/>
                          <g class="window-btn" @click.stop="toggleMaximize(res.type, item.name)" :transform="`translate(${getPanelState(res.type, item.name).w - 45}, 14)`">
                            <rect x="-10" y="-8" width="16" height="16" rx="3" fill="none" stroke="#666" stroke-width="1"/>
                            <rect x="-7" y="-5" width="10" height="10" rx="1" fill="none" stroke="#666" stroke-width="1"/>
                          </g>
                          <g class="window-btn close-btn" @click.stop="closePanel(res.type, item.name)" :transform="`translate(${getPanelState(res.type, item.name).w - 22}, 14)`">
                            <text x="0" y="5" fill="#666" font-size="16" text-anchor="middle">×</text>
                          </g>
                          <foreignObject x="8" y="34" :width="getPanelState(res.type, item.name).w - 16" :height="getPanelState(res.type, item.name).h - 70" style="pointer-events: auto;">
                            <div xmlns="http://www.w3.org/1999/xhtml" style="overflow-y:auto;height:100%;padding:4px;font-family:monospace;font-size:11px;color:#90EE90;white-space:pre-wrap;background:#0a0a15;border-radius:4px;pointer-events:auto;cursor:text;" @mousedown.stop @click.stop>{{ terminalOutput || '等待连接...' }}</div>
                          </foreignObject>
                          <foreignObject x="8" :y="getPanelState(res.type, item.name).h - 32" :width="getPanelState(res.type, item.name).w - 45" height="28" style="pointer-events: auto;">
                            <input xmlns="http://www.w3.org/1999/xhtml" v-model="terminalInput" @keyup.enter="sendTerminalInput" @mousedown.stop @click.stop placeholder="输入命令..." style="width:100%;height:100%;background:#1a1a2e;border:1px solid #3d3d6e;color:#e0e0e0;padding:0 8px;font-family:monospace;font-size:11px;border-radius:4px;outline:none;pointer-events:auto;cursor:text;"/>
                          </foreignObject>
                          <g class="tree-node" @click.stop="sendTerminalInput" :transform="`translate(${getPanelState(res.type, item.name).w - 25}, ${getPanelState(res.type, item.name).h - 18})`">
                            <rect x="-12" y="-10" width="24" height="20" rx="4" fill="#67C23A"/>
                            <text x="0" y="4" text-anchor="middle" fill="#fff" font-size="10">↵</text>
                          </g>
                          <g class="resize-handle" @mousedown.stop="startResize($event, res.type, item.name)" :transform="`translate(${getPanelState(res.type, item.name).w - 10}, ${getPanelState(res.type, item.name).h - 10})`">
                            <rect x="0" y="0" width="10" height="10" fill="transparent" style="cursor:nwse-resize"/>
                            <path d="M10,0 L0,10 M10,4 L4,10 M10,8 L8,10" stroke="#666" stroke-width="1"/>
                          </g>
                        </g>
                      </template>

                      <!-- Deployment 操作 -->
                      <template v-if="res.type === 'deployments' && !getPanel(res.type, item.name)">
                        <g class="tree-node action-node" @click.stop="openDeployDetail(item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y - 20})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#409EFF" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#409EFF" font-size="9">📊 详情</text>
                        </g>
                        <g class="tree-node action-node" @click.stop="scaleDeploy(item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#E6A23C" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#E6A23C" font-size="9">📐 扩缩容</text>
                        </g>
                        <g class="tree-node action-node" @click.stop="restartDeploy(item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y + 20})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#67C23A" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#67C23A" font-size="9">🔄 重启</text>
                        </g>
                        <g class="tree-node action-node" @click.stop="deleteResource('deployments', item.ns, item.name)" :transform="`translate(${item.x + 260}, ${item.y - 10})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#F56C6C" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#F56C6C" font-size="9">🗑 删除</text>
                        </g>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y - 20" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y + 20" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 225" :y2="item.y - 10" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                      </template>

                      <!-- Deployment 详情面板 -->
                      <template v-if="res.type === 'deployments' && getPanel(res.type, item.name) === 'detail'">
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 140" :y2="item.y" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <g :transform="`translate(${item.x + 140}, ${item.y - 80})`">
                          <rect x="0" y="0" width="360" height="160" rx="8" fill="#0d0d1a" stroke="#409EFF" stroke-width="1"/>
                          <rect x="0" y="0" width="360" height="28" rx="8" fill="#1a1a2e"/>
                          <text x="12" y="18" fill="#409EFF" font-size="11" font-weight="bold">📊 {{ item.name }}</text>
                          <text x="348" y="18" fill="#666" font-size="14" text-anchor="end" class="close-btn" @click.stop="closePanel(res.type, item.name)">×</text>
                          <foreignObject x="12" y="36" width="336" height="118">
                            <div xmlns="http://www.w3.org/1999/xhtml" style="color:#ccc;font-size:11px;padding:8px;">
                              <div v-if="getPanelData(res.type, item.name)?.loading">加载中...</div>
                              <div v-else>
                                <div>副本数: {{ getPanelData(res.type, item.name)?.data?.replicas || '-' }}</div>
                                <div>可用: {{ getPanelData(res.type, item.name)?.data?.availableReplicas || '-' }}</div>
                                <div>镜像: {{ getPanelData(res.type, item.name)?.data?.image || '-' }}</div>
                              </div>
                            </div>
                          </foreignObject>
                        </g>
                      </template>

                      <!-- ConfigMap/Secret 操作 -->
                      <template v-if="(res.type === 'configmaps' || res.type === 'secrets') && !getPanel(res.type, item.name)">
                        <g class="tree-node action-node" @click.stop="openConfigDetail(res.type, item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y - 10})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#409EFF" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#409EFF" font-size="9">👁 查看</text>
                        </g>
                        <g class="tree-node action-node" @click.stop="deleteResource(res.type, item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y + 10})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#F56C6C" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#F56C6C" font-size="9">🗑 删除</text>
                        </g>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y - 10" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y + 10" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                      </template>

                      <!-- ConfigMap/Secret 详情面板 -->
                      <template v-if="(res.type === 'configmaps' || res.type === 'secrets') && getPanel(res.type, item.name) === 'detail'">
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 140" :y2="item.y" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <g :transform="`translate(${item.x + 140}, ${item.y - 80})`">
                          <rect x="0" y="0" width="400" height="160" rx="8" fill="#0d0d1a" stroke="#409EFF" stroke-width="1"/>
                          <rect x="0" y="0" width="400" height="28" rx="8" fill="#1a1a2e"/>
                          <text x="12" y="18" fill="#409EFF" font-size="11" font-weight="bold">👁 {{ item.name }}</text>
                          <text x="388" y="18" fill="#666" font-size="14" text-anchor="end" class="close-btn" @click.stop="closePanel(res.type, item.name)">×</text>
                          <foreignObject x="12" y="36" width="376" height="118">
                            <div xmlns="http://www.w3.org/1999/xhtml" style="overflow-y:auto;height:100%;padding:8px;font-family:monospace;font-size:10px;color:#90EE90;white-space:pre-wrap;background:#0a0a15;border-radius:4px;">
                              {{ getPanelData(res.type, item.name)?.loading ? '加载中...' : JSON.stringify(getPanelData(res.type, item.name)?.data, null, 2) }}
                            </div>
                          </foreignObject>
                        </g>
                      </template>

                      <!-- Service/Ingress 操作 -->
                      <template v-if="(res.type === 'services' || res.type === 'ingresses') && !getPanel(res.type, item.name)">
                        <g class="tree-node action-node" @click.stop="deleteResource(res.type, item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#F56C6C" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#F56C6C" font-size="9">🗑 删除</text>
                        </g>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                      </template>

                      <!-- StatefulSet/DaemonSet 操作 -->
                      <template v-if="(res.type === 'statefulsets' || res.type === 'daemonsets') && !getPanel(res.type, item.name)">
                        <g class="tree-node action-node" @click.stop="restartWorkload(res.type, item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y - 10})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#67C23A" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#67C23A" font-size="9">🔄 重启</text>
                        </g>
                        <g class="tree-node action-node" @click.stop="deleteResource(res.type, item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y + 10})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#F56C6C" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#F56C6C" font-size="9">🗑 删除</text>
                        </g>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y - 10" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y + 10" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                      </template>

                      <!-- CronJob 操作 -->
                      <template v-if="res.type === 'cronjobs' && !getPanel(res.type, item.name)">
                        <g class="tree-node action-node" @click.stop="toggleCronJob(item.ns, item.name, !item.suspend)" :transform="`translate(${item.x + 180}, ${item.y - 10})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" :stroke="item.suspend ? '#67C23A' : '#E6A23C'" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" :fill="item.suspend ? '#67C23A' : '#E6A23C'" font-size="9">{{ item.suspend ? '▶ 恢复' : '⏸ 暂停' }}</text>
                        </g>
                        <g class="tree-node action-node" @click.stop="deleteResource(res.type, item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y + 10})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#F56C6C" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#F56C6C" font-size="9">🗑 删除</text>
                        </g>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y - 10" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y + 10" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                      </template>

                      <!-- Role/ServiceAccount 操作 -->
                      <template v-if="(res.type === 'roles' || res.type === 'serviceaccounts') && !getPanel(res.type, item.name)">
                        <g class="tree-node action-node" @click.stop="deleteRbac(res.type, item.ns, item.name)" :transform="`translate(${item.x + 180}, ${item.y})`">
                          <rect x="-35" y="-10" width="70" height="20" rx="10" fill="#2d2d5e" stroke="#F56C6C" stroke-width="1"/>
                          <text x="0" y="4" text-anchor="middle" fill="#F56C6C" font-size="9">🗑 删除</text>
                        </g>
                        <line :x1="item.x + 70" :y1="item.y" :x2="item.x + 145" :y2="item.y" :stroke="res.color" stroke-width="1" opacity="0.3"/>
                      </template>
                    </template>
                  </template>
                </template>
              </template>
            </template>
          </template>
        </g>
      </svg>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api'

const router = useRouter()
const mindmapRef = ref<HTMLElement>()

const svgWidth = ref(1400)
const svgHeight = ref(800)

const scale = ref(1)
const panX = ref(50)
const panY = ref(0)
const isDragging = ref(false)
const dragStartX = ref(0)
const dragStartY = ref(0)
const panStartX = ref(0)
const panStartY = ref(0)

const activeCluster = ref('')
const clusters = ref<any[]>([])
const showClusterMenu = ref(false)
const namespaces = ref<any[]>([])

const nsColors = ['#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399', '#B37FEB', '#36CFC9', '#FF85C0']
const resourceDefs = [
  { type: 'pods', label: 'Pod 容器组', color: '#67C23A', key: 'podCount' },
  { type: 'deployments', label: 'Deploy 部署', color: '#409EFF', key: 'deployCount' },
  { type: 'statefulsets', label: 'STS 有状态集', color: '#36CFC9', key: 'stsCount' },
  { type: 'daemonsets', label: 'DS 守护进程', color: '#FF85C0', key: 'dsCount' },
  { type: 'cronjobs', label: 'CronJob 定时任务', color: '#FFA940', key: 'cronjobCount' },
  { type: 'services', label: 'Service 服务', color: '#E6A23C', key: 'serviceCount' },
  { type: 'ingresses', label: 'Ingress 入口', color: '#B37FEB', key: 'ingCount' },
  { type: 'configmaps', label: 'CM 配置映射', color: '#909399', key: 'cmCount' },
  { type: 'secrets', label: 'Secret 密钥', color: '#F56C6C', key: 'secretCount' },
  { type: 'roles', label: 'Role 角色', color: '#722ED1', key: 'roleCount' },
  { type: 'serviceaccounts', label: 'SA 服务账号', color: '#13C2C2', key: 'saCount' },
]

interface ExpandedState {
  namespaces: Record<string, boolean>
  resources: Record<string, Record<string, boolean>>
  items: Record<string, any[]>
}
const expanded = reactive<ExpandedState>({
  namespaces: {},
  resources: {},
  items: {},
})

const horizontalGap = 220
const verticalGap = 28

function buildConnectPath(x1: number, y1: number, x2: number, y2: number): string {
  const midX = (x1 + x2) / 2
  return `M${x1},${y1} C${midX},${y1} ${midX},${y2} ${x2},${y2}`
}

const treeLayout = computed(() => {
  const centerX = svgWidth.value / 2
  const centerY = svgHeight.value / 2

  const nsList = namespaces.value
  const nsCount = nsList.length

  // 计算每个命名空间需要的垂直空间
  const nsHeights = nsList.map((ns) => {
    const isExpanded = !!expanded.namespaces[ns.name]
    if (!isExpanded) return verticalGap

    let totalHeight = 0
    resourceDefs.forEach(rd => {
      const resExpanded = !!expanded.resources[ns.name]?.[rd.type]
      if (!resExpanded) {
        totalHeight += verticalGap
      } else {
        const items = expanded.items[rd.type] || []
        totalHeight += Math.max(verticalGap, items.length * verticalGap)
      }
    })
    return Math.max(verticalGap, totalHeight)
  })

  // 分配左右：偶数索引放右边，奇数索引放左边
  const rightIndices: number[] = []
  const leftIndices: number[] = []
  nsList.forEach((_, i) => {
    if (i % 2 === 0) rightIndices.push(i)
    else leftIndices.push(i)
  })

  // 计算右侧布局
  const rightTotalHeight = rightIndices.reduce((sum, i) => sum + nsHeights[i], 0) + Math.max(0, rightIndices.length - 1) * 20
  let rightCurrentY = centerY - rightTotalHeight / 2

  // 计算左侧布局
  const leftTotalHeight = leftIndices.reduce((sum, i) => sum + nsHeights[i], 0) + Math.max(0, leftIndices.length - 1) * 20
  let leftCurrentY = centerY - leftTotalHeight / 2

  const nsNodes = nsList.map((ns, i) => {
    const nsHeight = nsHeights[i]
    const isRight = i % 2 === 0
    const isExpanded = !!expanded.namespaces[ns.name]
    const color = nsColors[i % nsColors.length]
    const nameWidth = Math.max(80, ns.name.length * 9 + 40)

    let nsY: number
    if (isRight) {
      nsY = rightCurrentY + nsHeight / 2
      rightCurrentY += nsHeight + 20
    } else {
      nsY = leftCurrentY + nsHeight / 2
      leftCurrentY += nsHeight + 20
    }

    const nsX = isRight ? centerX + horizontalGap : centerX - horizontalGap

    const resources: any[] = []
    if (isExpanded) {
      let resY = nsY - nsHeight / 2 + verticalGap / 2

      resourceDefs.forEach(rd => {
        const count = ns[rd.key] || 0
        const resExpanded = !!expanded.resources[ns.name]?.[rd.type]
        const items = resExpanded ? (expanded.items[rd.type] || []) : []
        const resX = isRight ? nsX + horizontalGap : nsX - horizontalGap

        let resHeight: number
        if (!resExpanded) {
          resHeight = verticalGap
        } else {
          resHeight = Math.max(verticalGap, items.length * verticalGap)
        }

        const resYCenter = resY + resHeight / 2

        const resourceItems = items.map((item: any, ii: number) => {
          const itemY = resY + ii * verticalGap + verticalGap / 2
          return {
            name: item.name,
            ns: ns.name,
            status: item.status || item.phase || '',
            x: isRight ? resX + horizontalGap : resX - horizontalGap,
            y: itemY,
            connectPath: isRight
              ? buildConnectPath(resX + 55, resYCenter, resX + horizontalGap - 70, itemY)
              : buildConnectPath(resX - 55, resYCenter, resX - horizontalGap + 70, itemY),
          }
        })

        resources.push({
          type: rd.type,
          label: rd.label,
          color: rd.color,
          count,
          x: resX,
          y: resYCenter,
          expanded: resExpanded,
          items: resourceItems,
          connectPath: isRight
            ? buildConnectPath(nsX + nameWidth / 2, nsY, resX - 55, resYCenter)
            : buildConnectPath(nsX - nameWidth / 2, nsY, resX + 55, resYCenter),
        })

        resY += resHeight
      })
    }

    return {
      name: ns.name,
      x: nsX,
      y: nsY,
      width: nameWidth,
      color,
      expanded: isExpanded,
      resources,
      connectPath: isRight
        ? buildConnectPath(centerX + 50, centerY, nsX - nameWidth / 2, nsY)
        : buildConnectPath(centerX - 50, centerY, nsX + nameWidth / 2, nsY),
    }
  })

  return {
    center: { x: centerX, y: centerY },
    namespaces: nsNodes,
  }
})

const allLines = computed(() => {
  const lines: any[] = []
  treeLayout.value.namespaces.forEach(ns => {
    lines.push({ id: 'c-' + ns.name, path: ns.connectPath, color: ns.color, width: 2 })
    ns.resources.forEach(res => {
      lines.push({ id: 'r-' + ns.name + '-' + res.type, path: res.connectPath, color: res.color, width: 1.5 })
      res.items.forEach((item: any) => {
        lines.push({ id: 'i-' + item.name, path: item.connectPath, color: res.color, width: 1 })
      })
    })
  })
  return lines
})

const truncate = (s: string, len: number) => s.length > len ? s.slice(0, len) + '..' : s

// 缩放
const onWheel = (e: WheelEvent) => {
  const delta = e.deltaY > 0 ? -0.1 : 0.1
  scale.value = Math.max(0.2, Math.min(3, scale.value + delta))
}
const zoomIn = () => { scale.value = Math.min(3, scale.value + 0.2) }
const zoomOut = () => { scale.value = Math.max(0.2, scale.value - 0.2) }
const resetView = () => { scale.value = 1; panX.value = 50; panY.value = 0 }

// 拖拽
const onMouseDown = (e: MouseEvent) => {
  isDragging.value = true
  dragStartX.value = e.clientX
  dragStartY.value = e.clientY
  panStartX.value = panX.value
  panStartY.value = panY.value
}
const onMouseMove = (e: MouseEvent) => {
  if (!isDragging.value) return
  panX.value = panStartX.value + (e.clientX - dragStartX.value) / scale.value
  panY.value = panStartY.value + (e.clientY - dragStartY.value) / scale.value
}
const onMouseUp = () => { isDragging.value = false }

// 展开/收起
const toggleNamespace = async (name: string) => {
  expanded.namespaces[name] = !expanded.namespaces[name]
  if (expanded.namespaces[name]) {
    await fetchResourcesForNs(name)
  } else {
    delete expanded.resources[name]
  }
}

const toggleResource = async (ns: string, type: string) => {
  if (!expanded.resources[ns]) expanded.resources[ns] = {}
  expanded.resources[ns][type] = !expanded.resources[ns][type]
  if (expanded.resources[ns][type]) {
    await fetchItems(type, ns)
  }
}

const collapseAll = () => {
  Object.keys(expanded.namespaces).forEach(k => delete expanded.namespaces[k])
  Object.keys(expanded.resources).forEach(k => delete expanded.resources[k])
  Object.keys(expanded.items).forEach(k => delete expanded.items[k])
}

const fetchResourcesForNs = async (ns: string) => {
  try {
    const [pods, deploys, stss, dss, crons, svcs, ings, cms, secrets, roles, sas] = await Promise.all([
      api.get('/pods', { params: { namespace: ns } }).catch(() => ({ data: [] })),
      api.get('/deployments', { params: { namespace: ns } }).catch(() => ({ data: [] })),
      api.get('/statefulsets', { params: { namespace: ns } }).catch(() => ({ data: [] })),
      api.get('/daemonsets', { params: { namespace: ns } }).catch(() => ({ data: [] })),
      api.get('/cronjobs', { params: { namespace: ns } }).catch(() => ({ data: [] })),
      api.get('/services', { params: { namespace: ns } }).catch(() => ({ data: [] })),
      api.get('/ingresses', { params: { namespace: ns } }).catch(() => ({ data: [] })),
      api.get('/configmaps', { params: { namespace: ns } }).catch(() => ({ data: [] })),
      api.get('/secrets', { params: { namespace: ns } }).catch(() => ({ data: [] })),
      api.get('/roles', { params: { namespace: ns } }).catch(() => ({ data: [] })),
      api.get('/serviceaccounts', { params: { namespace: ns } }).catch(() => ({ data: [] })),
    ])
    const idx = namespaces.value.findIndex(n => n.name === ns)
    if (idx >= 0) {
      namespaces.value[idx].podCount = (pods.data || []).length
      namespaces.value[idx].deployCount = (deploys.data || []).length
      namespaces.value[idx].stsCount = (stss.data || []).length
      namespaces.value[idx].dsCount = (dss.data || []).length
      namespaces.value[idx].cronjobCount = (crons.data || []).length
      namespaces.value[idx].serviceCount = (svcs.data || []).length
      namespaces.value[idx].ingCount = (ings.data || []).length
      namespaces.value[idx].cmCount = (cms.data || []).length
      namespaces.value[idx].secretCount = (secrets.data || []).length
      namespaces.value[idx].roleCount = (roles.data || []).length
      namespaces.value[idx].saCount = (sas.data || []).length
    }
  } catch (e) { console.error(e) }
}

const fetchItems = async (type: string, ns: string) => {
  try {
    const res = await api.get(`/${type}`, { params: { namespace: ns } })
    expanded.items[type] = res.data || []
  } catch (e) { console.error(e) }
}

const expandedItems = reactive<Record<string, boolean>>({})
const activePanel = reactive<Record<string, string>>({})
const panelData = reactive<Record<string, any>>({})
const logContent = ref('')
const logLoading = ref(false)
const terminalInput = ref('')
const terminalOutput = ref('')
const terminalConnected = ref(false)
const terminalShell = ref('/bin/sh')
let terminalWs: WebSocket | null = null
const formData = reactive<Record<string, any>>({})
const formVisible = reactive<Record<string, boolean>>({})
const saving = ref(false)

// 面板尺寸和状态
const panelState = reactive<Record<string, { w: number; h: number; x: number; y: number; maximized: boolean }>>({})

const getPanelState = (type: string, name: string) => {
  const key = `${type}-${name}`
  if (!panelState[key]) {
    panelState[key] = { w: 500, h: 240, x: 0, y: 0, maximized: false }
  }
  return panelState[key]
}

const toggleMaximize = (type: string, name: string) => {
  const key = `${type}-${name}`
  const s = getPanelState(type, name)
  s.maximized = !s.maximized
  if (s.maximized) {
    s.w = 900
    s.h = 500
    s.x = 100
    s.y = 50
  } else {
    s.w = 500
    s.h = 240
    s.x = 0
    s.y = 0
  }
}

let resizeTarget = ''
let resizeStartX = 0
let resizeStartY = 0
let resizeOrigW = 0
let resizeOrigH = 0

const startResize = (e: MouseEvent, type: string, name: string) => {
  e.stopPropagation()
  resizeTarget = `${type}-${name}`
  resizeStartX = e.clientX
  resizeStartY = e.clientY
  const s = getPanelState(type, name)
  resizeOrigW = s.w
  resizeOrigH = s.h
  document.addEventListener('mousemove', onResize)
  document.addEventListener('mouseup', stopResize)
}

const onResize = (e: MouseEvent) => {
  if (!resizeTarget) return
  const s = panelState[resizeTarget]
  if (s) {
    s.w = Math.max(300, resizeOrigW + (e.clientX - resizeStartX))
    s.h = Math.max(150, resizeOrigH + (e.clientY - resizeStartY))
  }
}

const stopResize = () => {
  resizeTarget = ''
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
}

const toggleItem = (type: string, item: any) => {
  const key = `${type}-${item.name}`
  expandedItems[key] = !expandedItems[key]
  if (!expandedItems[key]) {
    delete activePanel[key]
    delete panelData[key]
    if (terminalWs) { terminalWs.close(); terminalWs = null }
  }
}

const isItemExpanded = (type: string, name: string) => !!expandedItems[`${type}-${name}`]
const getPanel = (type: string, name: string) => activePanel[`${type}-${name}`]
const getPanelData = (type: string, name: string) => panelData[`${type}-${name}`]

const handleItemClick = (type: string, item: any) => {
  toggleItem(type, item)
}

const withNs = (item: any, ns: string) => ({ ...item, ns })

const openPanel = (type: string, name: string, panel: string) => {
  const key = `${type}-${name}`
  activePanel[key] = panel
}

const changeShell = (ns: string, pod: string, shell: string) => {
  terminalShell.value = shell
  terminalOutput.value += `\r\n[切换到 ${shell}...]\r\n`
  terminalConnected.value = false
  connectTerminalWs(ns, pod)
}

const closePanel = (type: string, name: string) => {
  const key = `${type}-${name}`
  delete activePanel[key]
  delete panelData[key]
  if (terminalWs) { terminalWs.close(); terminalWs = null }
}

// 日志 - 复用现有 API
const openLogs = async (ns: string, pod: string) => {
  openPanel('pods', pod, 'logs')
  logLoading.value = true
  logContent.value = ''
  try {
    const res: any = await api.get('/pods/logs', { params: { namespace: ns, pod, tailLines: 100 } })
    logContent.value = res.data || '暂无日志'
  } catch (e) { logContent.value = '获取日志失败' }
  logLoading.value = false
}

// 终端 - 复用现有 WebSocket
const openTerminal = (ns: string, pod: string) => {
  openPanel('pods', pod, 'terminal')
  terminalOutput.value = ''
  terminalInput.value = ''
  terminalConnected.value = false
  connectTerminalWs(ns, pod)
}

const connectTerminalWs = (ns: string, pod: string) => {
  if (terminalWs) {
    terminalWs.onclose = null
    terminalWs.close()
    terminalWs = null
  }
  terminalConnected.value = false
  terminalOutput.value += `\r\n连接中... (${terminalShell.value})\r\n`
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsHost = window.location.hostname
  const token = localStorage.getItem('token') || ''
  const ws = new WebSocket(`${protocol}//${wsHost}:8080/api/pods/exec/ws?token=${token}`)
  terminalWs = ws
  ws.onopen = () => {
    if (ws !== terminalWs) return
    terminalOutput.value += `WebSocket 已连接，正在执行 exec...\r\n`
    ws.send(JSON.stringify({ namespace: ns, pod, container: '', command: [terminalShell.value, '-il'], columns: 80, rows: 24 }))
  }
  ws.onmessage = (e) => {
    if (ws !== terminalWs) return
    const processData = (data) => {
      terminalOutput.value += data
      if (!terminalConnected.value && data.length > 0) {
        terminalConnected.value = true
      }
    }
    if (e.data instanceof Blob) {
      e.data.text().then(processData)
    } else if (typeof e.data === 'string') {
      processData(e.data)
    } else {
      processData(new TextDecoder().decode(e.data))
    }
  }
  ws.onerror = () => {
    if (ws !== terminalWs) return
    terminalConnected.value = false
  }
  ws.onclose = () => {
    if (ws !== terminalWs) return
    terminalConnected.value = false
  }
}

const sendTerminalInput = () => {
  if (terminalWs && terminalConnected.value && terminalInput.value) {
    terminalWs.send(JSON.stringify({ type: 'input', data: terminalInput.value + '\n' }))
    terminalInput.value = ''
  }
}

// 通用删除 - 复用现有 API
const deleteResource = async (type: string, ns: string, name: string) => {
  const endpoint = type === 'persistentvolumeclaims' ? 'persistentvolumeclaims' : type
  try {
    await ElMessageBox.confirm(`确定要删除 ${name} 吗？`, '确认', { type: 'error' })
    await api.delete(`/${endpoint}/delete?namespace=${ns}&name=${name}`)
    ElMessage.success('已删除')
    fetchResourcesForNs(ns)
    toggleItem(type, { name })
  } catch (e) { if (e !== 'cancel') ElMessage.error('删除失败') }
}

// Deployment - 复用现有 API
const openDeployDetail = async (ns: string, name: string) => {
  openPanel('deployments', name, 'detail')
  const key = `deployments-${name}`
  panelData[key] = { ns, loading: true, data: null }
  try {
    const res: any = await api.get('/deployments/get', { params: { namespace: ns, name } })
    panelData[key].data = res.data || {}
  } catch (e) { panelData[key].data = {} }
  panelData[key].loading = false
}

const scaleDeploy = async (ns: string, name: string) => {
  const key = `deployments-${name}`
  const replicas = prompt('输入副本数:', panelData[key]?.data?.replicas || '1')
  if (replicas) {
    try {
      await api.post('/deployments/scale', { namespace: ns, name, replicas: parseInt(replicas) })
      ElMessage.success('扩缩容成功')
      openDeployDetail(ns, name)
      fetchResourcesForNs(ns)
    } catch (e) { ElMessage.error('扩缩容失败') }
  }
}

const restartDeploy = async (ns: string, name: string) => {
  try {
    await api.post('/deployments/restart', { namespace: ns, name })
    ElMessage.success('重启成功')
    fetchResourcesForNs(ns)
  } catch (e) { ElMessage.error('重启失败') }
}

// ConfigMap/Secret - 复用现有 API
const openConfigDetail = async (type: string, ns: string, name: string) => {
  openPanel(type, name, 'detail')
  const key = `${type}-${name}`
  panelData[key] = { ns, loading: true, data: null }
  try {
    const res: any = await api.get(`/${type}/get`, { params: { namespace: ns, name } })
    panelData[key].data = res.data || {}
  } catch (e) { panelData[key].data = {} }
  panelData[key].loading = false
}

const openConfigEdit = async (type: string, ns: string, name: string) => {
  openPanel(type, name, 'edit')
  const key = `${type}-${name}`
  panelData[key] = { ns, loading: true, data: null }
  try {
    const res: any = await api.get(`/${type}/get`, { params: { namespace: ns, name } })
    const data = res.data || {}
    panelData[key].data = { name, namespace: ns, items: Object.entries(data).map(([k, v]) => ({ key: k, value: String(v) })) }
  } catch (e) { panelData[key].data = { name, namespace: ns, items: [{ key: '', value: '' }] } }
  panelData[key].loading = false
}

const saveConfigItem = async (type: string) => {
  const key = `${type}-${formData.editName}`
  const data = panelData[key]?.data
  if (!data) return
  saving.value = true
  try {
    const items: Record<string, string> = {}
    data.items.forEach((i: any) => { if (i.key) items[i.key] = i.value })
    await api.post(`/${type}/update`, { namespace: data.namespace, name: data.name, data: items })
    ElMessage.success('保存成功')
    closePanel(type, data.name)
    fetchResourcesForNs(data.namespace)
  } catch (e) { ElMessage.error('保存失败') }
  saving.value = false
}

// 创建资源
const openCreateForm = (type: string, ns: string) => {
  const key = `create-${type}-${ns}`
  activePanel[key] = 'create'
  formVisible[key] = true
  if (type === 'configmaps' || type === 'secrets') {
    formData[key] = { namespace: ns, name: '', items: [{ key: '', value: '' }] }
  } else if (type === 'services') {
    formData[key] = { namespace: ns, name: '', type: 'ClusterIP', ports: [{ port: 80, targetPort: 80, protocol: 'TCP' }], selectors: [{ key: '', value: '' }] }
  } else if (type === 'namespaces') {
    formData[key] = { name: '', labels: [{ key: '', value: '' }] }
  }
}

const saveCreateForm = async (type: string, ns: string) => {
  const key = `create-${type}-${ns}`
  const data = formData[key]
  if (!data) return
  saving.value = true
  try {
    if (type === 'configmaps') {
      const items: Record<string, string> = {}
      data.items.forEach((i: any) => { if (i.key) items[i.key] = i.value })
      await api.post('/configmaps/create', { namespace: ns, name: data.name, data: items })
    } else if (type === 'secrets') {
      const items: Record<string, string> = {}
      data.items.forEach((i: any) => { if (i.key) items[i.key] = i.value })
      await api.post('/secrets/create', { namespace: ns, name: data.name, type: 'Opaque', data: items })
    } else if (type === 'services') {
      const selectors: Record<string, string> = {}
      data.selectors.forEach((s: any) => { if (s.key) selectors[s.key] = s.value })
      await api.post('/services/create', { namespace: ns, name: data.name, type: data.type, ports: data.ports, selectors })
    } else if (type === 'namespaces') {
      const labels: Record<string, string> = {}
      data.labels.forEach((l: any) => { if (l.key) labels[l.key] = l.value })
      await api.post('/namespaces/create', { name: data.name, labels })
    }
    ElMessage.success('创建成功')
    formVisible[key] = false
    delete activePanel[key]
    fetchResourcesForNs(ns)
    if (type === 'namespaces') fetchNamespaces()
  } catch (e) { ElMessage.error('创建失败') }
  saving.value = false
}

// Deployment 回滚
const rollbackDeploy = async (ns: string, name: string) => {
  try {
    await api.post('/deployments/rollback', { namespace: ns, name })
    ElMessage.success('回滚成功')
    fetchResourcesForNs(ns)
  } catch (e) { ElMessage.error('回滚失败') }
}

// CronJob 暂停/恢复
const toggleCronJob = async (ns: string, name: string, suspend: boolean) => {
  try {
    await api.post(`/cronjobs/${suspend ? 'resume' : 'suspend'}?namespace=${ns}&name=${name}`)
    ElMessage.success(suspend ? '已暂停' : '已恢复')
    fetchResourcesForNs(ns)
  } catch (e) { ElMessage.error('操作失败') }
}

// StatefulSet/DaemonSet 重启
const restartWorkload = async (type: string, ns: string, name: string) => {
  try {
    await api.post(`/${type}/restart`, { namespace: ns, name })
    ElMessage.success('重启成功')
    fetchResourcesForNs(ns)
  } catch (e) { ElMessage.error('重启失败') }
}

// RBAC 删除
const deleteRbac = async (type: string, ns: string, name: string) => {
  try {
    await ElMessageBox.confirm(`确定要删除 ${name} 吗？`, '确认', { type: 'error' })
    const endpoint = type.replace('cluster', 'cluster')
    if (ns) {
      await api.delete(`/${endpoint}/delete?namespace=${ns}&name=${name}`)
    } else {
      await api.delete(`/${endpoint}/delete?name=${name}`)
    }
    ElMessage.success('已删除')
    fetchResourcesForNs(ns)
    toggleItem(type, { name })
  } catch (e) { if (e !== 'cancel') ElMessage.error('删除失败') }
}

// PV/PVC/StorageClass 删除
const deleteStorage = async (type: string, ns: string, name: string) => {
  try {
    await ElMessageBox.confirm(`确定要删除 ${name} 吗？`, '确认', { type: 'error' })
    if (ns) {
      await api.delete(`/${type}/delete?namespace=${ns}&name=${name}`)
    } else {
      await api.delete(`/${type}/delete?name=${name}`)
    }
    ElMessage.success('已删除')
    fetchResourcesForNs(ns)
    toggleItem(type, { name })
  } catch (e) { if (e !== 'cancel') ElMessage.error('删除失败') }
}

// 事件查看
const openEvents = async (ns: string) => {
  openPanel('events', ns, 'list')
  const key = `events-${ns}`
  panelData[key] = { ns, loading: true, data: [] }
  try {
    const res: any = await api.get('/events', { params: { namespace: ns } })
    panelData[key].data = res.data || []
  } catch (e) { panelData[key].data = [] }
  panelData[key].loading = false
}

const fetchNamespaces = async () => {
  try {
    const res: any = await api.get('/namespaces')
    namespaces.value = (res.data || []).map((ns: any) => ({
      ...ns, podCount: 0, deployCount: 0, serviceCount: 0, cmCount: 0, secretCount: 0, ingCount: 0,
    }))
  } catch (e) { console.error(e) }
}

const switchCluster = async (name: string) => {
  try {
    await api.post('/clusters/switch', { name })
    activeCluster.value = name
    showClusterMenu.value = false
    collapseAll()
    fetchNamespaces()
  } catch (e) { ElMessage.error('切换失败') }
}

onMounted(async () => {
  if (mindmapRef.value) {
    svgWidth.value = mindmapRef.value.clientWidth
    svgHeight.value = mindmapRef.value.clientHeight
  }
  try {
    const res: any = await api.get('/clusters')
    if (res.clusters) {
      clusters.value = res.clusters
      const active = res.clusters.find((c: any) => c.active)
      if (active) activeCluster.value = active.name
    }
  } catch {}
  await fetchNamespaces()
})
</script>

<style scoped>
.mindmap-container {
  height: calc(100vh - 100px);
  display: flex;
  flex-direction: column;
  background: #0f0f23;
  border-radius: 12px;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 20px;
  background: #1a1a2e;
  border-bottom: 1px solid #2d2d5e;
  z-index: 10;
}

.cluster-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: #2d2d5e;
  border-radius: 20px;
  cursor: pointer;
}

.cluster-selector:hover { background: #3d3d6e; }
.cluster-icon { font-size: 16px; }
.cluster-name { color: #67C23A; font-size: 13px; font-weight: 600; }
.arrow { color: #666; font-size: 10px; }

.cluster-menu {
  position: absolute;
  top: 55px;
  left: 20px;
  background: #2d2d5e;
  border-radius: 8px;
  padding: 8px;
  min-width: 150px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
  z-index: 100;
}

.cluster-item {
  padding: 8px 12px;
  color: #ccc;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
}

.cluster-item:hover { background: #3d3d6e; }
.cluster-item.active { color: #67C23A; background: #67C23A20; }

.zoom-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.zoom-btn {
  width: 32px;
  height: 32px;
  border: 1px solid #3d3d6e;
  background: #2d2d5e;
  color: #ccc;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.zoom-btn:hover { background: #3d3d6e; color: #fff; }
.zoom-level { color: #999; font-size: 12px; min-width: 40px; text-align: center; }

.mindmap-main {
  flex: 1;
  cursor: grab;
}

.mindmap-main:active { cursor: grabbing; }
.mindmap-svg { width: 100%; height: 100%; }

.tree-node { cursor: pointer; transition: transform 0.4s ease; }
.center-node:hover rect { filter: brightness(1.2); }
.ns-node:hover rect { filter: brightness(1.3); }
.res-node:hover rect { filter: brightness(1.3); }
.item-node:hover rect { fill: #4d4d7e; }

.animated-path {
  transition: d 0.4s ease;
}

.window-btn {
  cursor: pointer;
  opacity: 0.6;
  transition: opacity 0.2s;
}
.window-btn:hover {
  opacity: 1;
}
.window-btn:hover text {
  color: #fff !important;
}
.window-btn:hover rect {
  stroke: #fff !important;
}
.close-btn:hover text {
  color: #F56C6C !important;
}
.resize-handle {
  cursor: nwse-resize;
  opacity: 0.5;
}
.resize-handle:hover {
  opacity: 1;
}
</style>
