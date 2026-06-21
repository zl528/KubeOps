<template>
  <div class="ingresses-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">Ingresses</h1>
        <span class="page-subtitle">管理集群中的入站规则</span>
      </div>
      <div class="header-actions">
        <div class="ns-selector">
          <el-select v-model="namespace" placeholder="选择命名空间" clearable @change="fetchData">
            <el-option label="全部命名空间" value="" />
            <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </div>
        <button type="button" class="btn-gradient btn-create" @click="openCreateDialog">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          <span>创建</span>
        </button>
        <button type="button" class="btn-gradient btn-refresh" @click="fetchData">
          <el-icon :size="16"><Refresh /></el-icon>
          <span>刷新</span>
        </button>
      </div>
    </div>

    <div class="glass-table-container">
      <el-table
        :data="ingresses"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 Ingress 数据'"
      >
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="cell-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140">
          <template #default="{ row }">
            <span class="cell-ns">{{ row.namespace }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="className" label="Ingress Class" width="150">
          <template #default="{ row }">
            <span class="cell-class">{{ row.className }}</span>
          </template>
        </el-table-column>
        <el-table-column label="Hosts" min-width="250">
          <template #default="{ row }">
            <div class="host-tags">
              <span v-for="host in (row.hosts || [])" :key="host" class="host-tag">{{ host }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="tls" label="TLS" width="80">
          <template #default="{ row }">
            <span class="cell-status" :class="row.tls ? 'status-ok' : 'status-info'">{{ row.tls ? '是' : '否' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="存活时间" width="100" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <div class="action-cell">
              <button type="button" class="action-btn action-edit" @click="openEditDialog(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                编辑
              </button>
              <button type="button" class="action-btn action-delete" @click="handleDelete(row)">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                删除
              </button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑 Ingress' : '创建 Ingress'"
      width="780px"
      :close-on-click-modal="false"
      class="ingress-dialog"
    >
      <el-form :model="form" label-position="top" class="ingress-form">
        <div class="form-row">
          <el-form-item label="名称" class="form-item-half">
            <el-input v-model="form.name" placeholder="ingress-name" :disabled="isEdit" />
          </el-form-item>
          <el-form-item label="命名空间" class="form-item-half">
            <el-select v-model="form.namespace" placeholder="选择命名空间" :disabled="isEdit">
              <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
            </el-select>
          </el-form-item>
        </div>

        <div class="form-row">
          <el-form-item label="Ingress Class" class="form-item-half">
            <el-input v-model="form.className" placeholder="nginx (可选)" />
          </el-form-item>
          <el-form-item label="" class="form-item-half form-spacer">
          </el-form-item>
        </div>

        <el-divider content-position="left">标签 (Labels)</el-divider>

        <div v-for="(label, idx) in form.labels" :key="'label-'+idx" class="kv-row">
          <el-input v-model="label.key" placeholder="Key" class="kv-field" />
          <el-input v-model="label.value" placeholder="Value" class="kv-field" />
          <button v-if="form.labels.length > 1" class="btn-remove-small" @click="form.labels.splice(idx, 1)">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <button type="button" class="btn-add-path" @click="form.labels.push({ key: '', value: '' })">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加标签
        </button>

        <el-divider content-position="left">注解 (Annotations)</el-divider>

        <div v-for="(ann, idx) in form.annotations" :key="'ann-'+idx" class="kv-row">
          <el-input v-model="ann.key" placeholder="Key" class="kv-field" />
          <el-input v-model="ann.value" placeholder="Value" class="kv-field" />
          <button v-if="form.annotations.length > 1" class="btn-remove-small" @click="form.annotations.splice(idx, 1)">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <button type="button" class="btn-add-path" @click="form.annotations.push({ key: '', value: '' })">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加注解
        </button>
        <div class="annotation-presets">
          <span class="preset-label">常用:</span>
          <button type="button" class="preset-btn" @click="addPresetAnnotation('nginx.ingress.kubernetes.io/rewrite-target', '/')">rewrite-target</button>
          <button type="button" class="preset-btn" @click="addPresetAnnotation('nginx.ingress.kubernetes.io/ssl-redirect', 'true')">ssl-redirect</button>
          <button type="button" class="preset-btn" @click="addPresetAnnotation('nginx.ingress.kubernetes.io/proxy-body-size', '50m')">proxy-body-size</button>
          <button type="button" class="preset-btn" @click="addPresetAnnotation('nginx.ingress.kubernetes.io/cors-allow-origin', '*')">cors-allow-origin</button>
        </div>

        <el-divider content-position="left">TLS 配置</el-divider>

        <el-form-item label="启用 TLS">
          <el-switch v-model="form.tlsEnabled" active-color="#6366f1" />
        </el-form-item>

        <el-form-item v-if="form.tlsEnabled" label="TLS Secret 名称">
          <el-input v-model="form.tlsSecretName" placeholder="my-tls-secret" />
        </el-form-item>

        <el-divider content-position="left">默认后端 (Default Backend)</el-divider>

        <el-form-item label="启用默认后端">
          <el-switch v-model="form.defaultBackendEnabled" active-color="#6366f1" />
        </el-form-item>

        <div v-if="form.defaultBackendEnabled" class="form-row">
          <el-form-item label="服务名称" class="form-item-half">
            <el-input v-model="form.defaultBackendService" placeholder="default-backend-service" />
          </el-form-item>
          <el-form-item label="服务端口" class="form-item-half">
            <el-input-number v-model="form.defaultBackendPort" :min="1" :max="65535" placeholder="80" />
          </el-form-item>
        </div>

        <el-divider content-position="left">规则配置</el-divider>

        <div v-for="(host, hIndex) in form.hosts" :key="hIndex" class="host-section">
          <div class="host-header">
            <span class="host-label">主机 #{{ hIndex + 1 }}</span>
            <button
              v-if="form.hosts.length > 1"
              class="btn-remove"
              @click="removeHost(hIndex)"
            >
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>

          <el-form-item label="主机名">
            <el-input v-model="host.host" placeholder="example.com" />
          </el-form-item>

          <div v-for="(path, pIndex) in host.paths" :key="pIndex" class="path-section">
            <div class="path-header">
              <span class="path-label">路径 #{{ pIndex + 1 }}</span>
              <button
                v-if="host.paths.length > 1"
                class="btn-remove-small"
                @click="removePath(hIndex, pIndex)"
              >
                <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>

            <div class="path-row">
              <el-form-item label="路径" class="path-field">
                <el-input v-model="path.path" placeholder="/" />
              </el-form-item>
              <el-form-item label="类型" class="path-field-small">
                <el-select v-model="path.pathType" placeholder="Prefix">
                  <el-option label="Prefix" value="Prefix" />
                  <el-option label="Exact" value="Exact" />
                  <el-option label="ImplementationSpecific" value="ImplementationSpecific" />
                </el-select>
              </el-form-item>
            </div>

            <div class="path-row">
              <el-form-item label="服务名称" class="path-field">
                <el-input v-model="path.serviceName" placeholder="my-service" />
              </el-form-item>
              <el-form-item label="服务端口" class="path-field-small">
                <el-input-number v-model="path.servicePort" :min="1" :max="65535" placeholder="80" />
              </el-form-item>
            </div>
          </div>

          <button type="button" class="btn-add-path" @click="addPath(hIndex)">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
            添加路径
          </button>
        </div>

        <button type="button" class="btn-add-host" @click="addHost">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加主机
        </button>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <button type="button" class="btn-cancel" @click="dialogVisible = false">取消</button>
          <button type="button" class="btn-confirm" @click="handleSubmit" :loading="submitting">
            {{ isEdit ? '保存' : '创建' }}
          </button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useGlobalNamespace } from '../store/namespace'
import { useHighlightRow } from '../composables/useHighlightRow'
import api from '../api'

interface IngressPathForm {
  path: string
  pathType: string
  serviceName: string
  servicePort: number
}

interface IngressHostForm {
  host: string
  paths: IngressPathForm[]
}

interface KeyValueItem {
  key: string
  value: string
}

useHighlightRow()
const ingresses = ref<any[]>([])
useHighlightRow()
const nsList = ref<any[]>([])
const { namespace } = useGlobalNamespace()
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)

const defaultPath = (): IngressPathForm => ({
  path: '/',
  pathType: 'Prefix',
  serviceName: '',
  servicePort: 80,
})

const defaultHost = (): IngressHostForm => ({
  host: '',
  paths: [defaultPath()],
})

const form = ref({
  name: '',
  namespace: '',
  className: '',
  tlsEnabled: false,
  tlsSecretName: '',
  defaultBackendEnabled: false,
  defaultBackendService: '',
  defaultBackendPort: 80,
  labels: [{ key: '', value: '' }] as KeyValueItem[],
  annotations: [{ key: '', value: '' }] as KeyValueItem[],
  hosts: [defaultHost()] as IngressHostForm[],
})

const resetForm = () => {
  form.value = {
    name: '',
    namespace: namespace.value || '',
    className: '',
    tlsEnabled: false,
    tlsSecretName: '',
    defaultBackendEnabled: false,
    defaultBackendService: '',
    defaultBackendPort: 80,
    labels: [{ key: '', value: '' }],
    annotations: [{ key: '', value: '' }],
    hosts: [defaultHost()],
  }
}

const openCreateDialog = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = async (row: any) => {
  isEdit.value = true
  try {
    const res: any = await api.get('/ingresses/get', {
      params: { namespace: row.namespace, name: row.name },
    })
    const detail = res.data
    form.value = {
      name: detail.name,
      namespace: detail.namespace,
      className: detail.className || '',
      tlsEnabled: (detail.tls || []).length > 0,
      tlsSecretName: (detail.tls || [])[0]?.secretName || '',
      defaultBackendEnabled: !!detail.defaultBackend,
      defaultBackendService: detail.defaultBackend?.serviceName || '',
      defaultBackendPort: detail.defaultBackend?.servicePort || 80,
      labels: parseLabelMap(detail.labels),
      annotations: parseLabelMap(detail.annotations),
      hosts: (detail.hosts || []).map((h: any) => ({
        host: h.host,
        paths: (h.paths || []).map((p: any) => ({
          path: p.path || '/',
          pathType: p.pathType || 'Prefix',
          serviceName: p.serviceName || '',
          servicePort: p.servicePort || 80,
        })),
      })),
    }
    dialogVisible.value = true
  } catch (e) {
    ElMessage.error('获取 Ingress 详情失败')
  }
}

const addHost = () => {
  form.value.hosts.push(defaultHost())
}

const removeHost = (hIndex: number) => {
  form.value.hosts.splice(hIndex, 1)
}

const addPath = (hIndex: number) => {
  form.value.hosts[hIndex].paths.push(defaultPath())
}

const removePath = (hIndex: number, pIndex: number) => {
  form.value.hosts[hIndex].paths.splice(pIndex, 1)
}

const buildLabelMap = (items: KeyValueItem[]): Record<string, string> => {
  const result: Record<string, string> = {}
  items.forEach(s => {
    if (s.key && s.value) {
      result[s.key] = s.value
    }
  })
  return result
}

const parseLabelMap = (map: Record<string, string> | undefined): KeyValueItem[] => {
  if (!map || Object.keys(map).length === 0) return [{ key: '', value: '' }]
  return Object.entries(map).map(([k, v]) => ({ key: k, value: v }))
}

const addPresetAnnotation = (key: string, value: string) => {
  const emptyIdx = form.value.annotations.findIndex(a => !a.key && !a.value)
  if (emptyIdx >= 0) {
    form.value.annotations[emptyIdx] = { key, value }
  } else {
    form.value.annotations.push({ key, value })
  }
}

const handleSubmit = async () => {
  if (!form.value.name) {
    ElMessage.warning('请输入 Ingress 名称')
    return
  }
  if (!form.value.namespace) {
    ElMessage.warning('请选择命名空间')
    return
  }

  for (const host of form.value.hosts) {
    if (!host.host) {
      ElMessage.warning('请输入主机名')
      return
    }
    for (const path of host.paths) {
      if (!path.serviceName) {
        ElMessage.warning('请输入服务名称')
        return
      }
    }
  }

  submitting.value = true
  try {
    const payload: any = {
      name: form.value.name,
      namespace: form.value.namespace,
      className: form.value.className,
      labels: buildLabelMap(form.value.labels),
      annotations: buildLabelMap(form.value.annotations),
      hosts: form.value.hosts,
      tls: [],
    }

    if (form.value.tlsEnabled && form.value.tlsSecretName) {
      const allHosts = form.value.hosts.map((h) => h.host).filter(Boolean)
      payload.tls = [{ hosts: allHosts, secretName: form.value.tlsSecretName }]
    }

    if (form.value.defaultBackendEnabled && form.value.defaultBackendService) {
      payload.defaultBackend = {
        serviceName: form.value.defaultBackendService,
        servicePort: form.value.defaultBackendPort,
      }
    }

    await api.post('/ingresses/create', payload)
    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    dialogVisible.value = false
    fetchData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.message || (isEdit.value ? '更新失败' : '创建失败'))
  } finally {
    submitting.value = false
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const params = namespace.value ? { namespace: namespace.value } : {}
    const res: any = await api.get('/ingresses', { params })
    ingresses.value = res.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
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

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(`确定要删除 Ingress ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.delete(`/ingresses/delete?namespace=${row.namespace}&name=${row.name}`)
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

onMounted(() => { fetchNs(); fetchData() })
</script>

<style scoped>
.ingresses-page {
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

.cell-ns {
  font-size: 12px;
  color: var(--text-secondary);
  background: rgba(51, 65, 85, 0.5);
  padding: 2px 8px;
  border-radius: 4px;
}

.cell-class {
  font-family: monospace;
  font-size: 13px;
  color: #e2e8f0;
}

.host-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.host-tag {
  font-size: 12px;
  color: #e2e8f0;
  background: rgba(99, 102, 241, 0.12);
  padding: 2px 8px;
  border-radius: 4px;
  border: 1px solid rgba(99, 102, 241, 0.2);
}

.cell-status {
  font-size: 12px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 4px;
}

.status-ok {
  background: rgba(34, 197, 94, 0.12);
  color: #4ade80;
}

.status-info {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
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

.action-delete {
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
}

.action-delete:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.25);
}

.action-edit {
  background: rgba(99, 102, 241, 0.12);
  color: #818cf8;
}

.action-edit:hover:not(:disabled) {
  background: rgba(99, 102, 241, 0.25);
}

.glass-table-container :deep(.el-loading-mask) {
  background: rgba(15, 23, 42, 0.7);
  backdrop-filter: blur(4px);
}

.glass-table-container :deep(.el-loading-spinner .circular) {
  stroke: var(--primary);
}

.ingresses-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.ingresses-page :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 16px !important;
}

.ingresses-page :deep(.el-message-box) {
  background: #1e293b;
  border: 1px solid rgba(148, 163, 184, 0.1);
  border-radius: 12px;
}

.ingresses-page :deep(.el-message-box__title) {
  color: var(--text-primary);
}

.ingresses-page :deep(.el-message-box__content) {
  color: var(--text-secondary);
}

.ingresses-page :deep(.el-message-box__btns .el-button--primary) {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-color: #6366f1;
}

.ingress-dialog :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 16px !important;
}

.ingress-dialog :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  padding: 20px 24px;
  margin: 0;
}

.ingress-dialog :deep(.el-dialog__title) {
  color: #f1f5f9;
  font-size: 18px;
  font-weight: 600;
}

.ingress-dialog :deep(.el-dialog__body) {
  padding: 24px;
  color: #e2e8f0;
}

.ingress-dialog :deep(.el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.1);
  padding: 16px 24px;
}

.ingress-form :deep(.el-form-item__label) {
  color: #94a3b8 !important;
  font-size: 13px;
}

.ingress-form :deep(.el-input__wrapper) {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 8px;
  box-shadow: none;
}

.ingress-form :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.ingress-form :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.ingress-form :deep(.el-input__inner) {
  color: #f1f5f9;
}

.ingress-form :deep(.el-input__inner::placeholder) {
  color: #64748b;
}

.ingress-form :deep(.el-select) {
  width: 100%;
}

.ingress-form :deep(.el-select .el-input__wrapper) {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 8px;
  box-shadow: none;
}

.ingress-form :deep(.el-divider) {
  border-color: rgba(148, 163, 184, 0.1);
  margin: 20px 0;
}

.ingress-form :deep(.el-divider__text) {
  background: #1e293b;
  color: #94a3b8;
  font-size: 13px;
  font-weight: 500;
}

.ingress-form :deep(.el-input-number) {
  width: 100%;
}

.ingress-form :deep(.el-input-number .el-input__wrapper) {
  padding-left: 8px;
  padding-right: 40px;
}

.ingress-form :deep(.el-switch) {
  --el-switch-on-color: #6366f1;
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-item-half {
  flex: 1;
}

.form-spacer {
  visibility: hidden;
}

.host-section {
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 16px;
}

.host-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.host-label {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
}

.path-section {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.06);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
}

.path-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.path-label {
  font-size: 12px;
  font-weight: 500;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.path-row {
  display: flex;
  gap: 12px;
}

.path-field {
  flex: 1;
}

.path-field-small {
  width: 140px;
  flex: none;
}

.btn-remove,
.btn-remove-small {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
}

.btn-remove:hover,
.btn-remove-small:hover {
  background: rgba(239, 68, 68, 0.25);
}

.btn-remove-small {
  width: 24px;
  height: 24px;
}

.btn-add-path {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px dashed rgba(99, 102, 241, 0.3);
  border-radius: 8px;
  background: transparent;
  color: #818cf8;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-add-path:hover {
  border-color: rgba(99, 102, 241, 0.5);
  background: rgba(99, 102, 241, 0.08);
}

.btn-add-host {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  padding: 12px;
  border: 1px dashed rgba(99, 102, 241, 0.3);
  border-radius: 10px;
  background: transparent;
  color: #818cf8;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
  margin-top: 8px;
}

.btn-add-host:hover {
  border-color: rgba(99, 102, 241, 0.5);
  background: rgba(99, 102, 241, 0.08);
}

.kv-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.kv-field {
  flex: 1;
}

.annotation-presets {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.preset-label {
  font-size: 12px;
  color: #64748b;
}

.preset-btn {
  padding: 4px 10px;
  border: 1px solid rgba(99, 102, 241, 0.2);
  border-radius: 6px;
  background: rgba(99, 102, 241, 0.08);
  color: #818cf8;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.preset-btn:hover {
  border-color: rgba(99, 102, 241, 0.4);
  background: rgba(99, 102, 241, 0.15);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.btn-cancel {
  padding: 10px 20px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 8px;
  background: transparent;
  color: #94a3b8;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-cancel:hover {
  border-color: rgba(148, 163, 184, 0.4);
  color: #e2e8f0;
}

.btn-confirm {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
}

.btn-confirm:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(99, 102, 241, 0.4);
}

.btn-confirm:active {
  transform: translateY(0);
}

.btn-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}
</style>
