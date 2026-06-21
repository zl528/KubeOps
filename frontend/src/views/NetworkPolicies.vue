<template>
  <div class="networkpolicies-page">
    <div class="page-header-gradient">
      <div class="header-left">
        <h1 class="page-title">NetworkPolicies</h1>
        <span class="page-subtitle">管理集群中的网络策略</span>
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
        :data="policies"
        v-loading="loading"
        :header-cell-style="headerCellStyle"
        :cell-style="cellStyle"
        :row-class-name="rowClassName"
        class="custom-table"
        :empty-text="'暂无 NetworkPolicy 数据'"
      >
        <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="cell-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="150">
          <template #default="{ row }">
            <span class="cell-ns">{{ row.namespace }}</span>
          </template>
        </el-table-column>
        <el-table-column label="Pod 选择器" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="(v, k) in (row.podSelector || {})" :key="k" size="small" style="margin: 2px">
              {{ k }}={{ v }}
            </el-tag>
            <span v-if="!row.podSelector || Object.keys(row.podSelector).length === 0" class="cell-dim">全部 Pod</span>
          </template>
        </el-table-column>
        <el-table-column label="Ingress 规则" width="120">
          <template #default="{ row }">
            <span class="cell-metric">{{ (row.ingress || []).length }} 条</span>
          </template>
        </el-table-column>
        <el-table-column label="Egress 规则" width="120">
          <template #default="{ row }">
            <span class="cell-metric">{{ (row.egress || []).length }} 条</span>
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
      :title="isEdit ? '编辑 NetworkPolicy' : '创建 NetworkPolicy'"
      width="780px"
      :close-on-click-modal="false"
      class="np-dialog"
    >
      <el-form :model="form" label-position="top" class="np-form">
        <div class="form-row">
          <el-form-item label="名称" class="form-item-half">
            <el-input v-model="form.name" placeholder="network-policy-name" :disabled="isEdit" />
          </el-form-item>
          <el-form-item label="命名空间" class="form-item-half">
            <el-select v-model="form.namespace" placeholder="选择命名空间" :disabled="isEdit">
              <el-option v-for="ns in nsList" :key="ns.name" :label="ns.name" :value="ns.name" />
            </el-select>
          </el-form-item>
        </div>

        <el-divider content-position="left">策略类型 (Policy Types)</el-divider>

        <el-form-item label="策略类型">
          <el-select v-model="form.policyTypes" placeholder="选择策略类型" multiple>
            <el-option label="Ingress" value="Ingress" />
            <el-option label="Egress" value="Egress" />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">标签 (Labels)</el-divider>

        <div v-for="(label, idx) in form.labels" :key="'label-'+idx" class="kv-row">
          <el-input v-model="label.key" placeholder="Key" class="kv-field" />
          <el-input v-model="label.value" placeholder="Value" class="kv-field" />
          <button v-if="form.labels.length > 1" class="btn-remove-small" @click="form.labels.splice(idx, 1)">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <button type="button" class="btn-add-row" @click="form.labels.push({ key: '', value: '' })">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加标签
        </button>

        <el-divider content-position="left">Pod 选择器</el-divider>

        <div v-for="(sel, idx) in form.podSelector" :key="idx" class="kv-row">
          <el-input v-model="sel.key" placeholder="Key" class="kv-field" />
          <el-input v-model="sel.value" placeholder="Value" class="kv-field" />
          <button v-if="form.podSelector.length > 1" class="btn-remove-small" @click="form.podSelector.splice(idx, 1)">
            <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <button type="button" class="btn-add-row" @click="form.podSelector.push({ key: '', value: '' })">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加标签
        </button>

        <el-divider content-position="left">Ingress 规则</el-divider>

        <div v-for="(rule, rIdx) in form.ingress" :key="rIdx" class="rule-section">
          <div class="rule-header">
            <span class="rule-label">Ingress #{{ rIdx + 1 }}</span>
            <button type="button" class="btn-remove-small" @click="form.ingress.splice(rIdx, 1)">
              <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>

          <div class="sub-section">
            <span class="sub-label">来源 (from)</span>
            <div v-for="(from, fIdx) in rule.from" :key="fIdx" class="peer-row">
              <el-input v-model="from.namespaceSelector" placeholder="Namespace Selector (key=value)" class="peer-field" />
              <el-input v-model="from.podSelector" placeholder="Pod Selector (key=value)" class="peer-field" />
              <el-input v-model="from.ipBlock" placeholder="IPBlock CIDR (可选)" class="peer-field" />
              <button v-if="rule.from.length > 1" class="btn-remove-small" @click="rule.from.splice(fIdx, 1)">
                <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <button type="button" class="btn-add-row-sm" @click="rule.from.push({ namespaceSelector: '', podSelector: '', ipBlock: '' })">+ 添加来源</button>
          </div>

          <div class="sub-section">
            <span class="sub-label">端口 (ports)</span>
            <div v-for="(port, pIdx) in rule.ports" :key="pIdx" class="port-row">
              <el-input v-model="port.port" placeholder="端口" class="port-field" />
              <el-select v-model="port.protocol" placeholder="协议" class="port-proto">
                <el-option label="TCP" value="TCP" />
                <el-option label="UDP" value="UDP" />
                <el-option label="SCTP" value="SCTP" />
              </el-select>
              <button v-if="rule.ports.length > 1" class="btn-remove-small" @click="rule.ports.splice(pIdx, 1)">
                <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <button type="button" class="btn-add-row-sm" @click="rule.ports.push({ port: '', protocol: 'TCP' })">+ 添加端口</button>
          </div>
        </div>
        <button type="button" class="btn-add-rule" @click="addIngressRule">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加 Ingress 规则
        </button>

        <el-divider content-position="left">Egress 规则</el-divider>

        <div v-for="(rule, rIdx) in form.egress" :key="rIdx" class="rule-section">
          <div class="rule-header">
            <span class="rule-label">Egress #{{ rIdx + 1 }}</span>
            <button type="button" class="btn-remove-small" @click="form.egress.splice(rIdx, 1)">
              <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>

          <div class="sub-section">
            <span class="sub-label">目标 (to)</span>
            <div v-for="(to, tIdx) in rule.to" :key="tIdx" class="peer-row">
              <el-input v-model="to.namespaceSelector" placeholder="Namespace Selector (key=value)" class="peer-field" />
              <el-input v-model="to.podSelector" placeholder="Pod Selector (key=value)" class="peer-field" />
              <el-input v-model="to.ipBlock" placeholder="IPBlock CIDR (可选)" class="peer-field" />
              <button v-if="rule.to.length > 1" class="btn-remove-small" @click="rule.to.splice(tIdx, 1)">
                <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <button type="button" class="btn-add-row-sm" @click="rule.to.push({ namespaceSelector: '', podSelector: '', ipBlock: '' })">+ 添加目标</button>
          </div>

          <div class="sub-section">
            <span class="sub-label">端口 (ports)</span>
            <div v-for="(port, pIdx) in rule.ports" :key="pIdx" class="port-row">
              <el-input v-model="port.port" placeholder="端口" class="port-field" />
              <el-select v-model="port.protocol" placeholder="协议" class="port-proto">
                <el-option label="TCP" value="TCP" />
                <el-option label="UDP" value="UDP" />
                <el-option label="SCTP" value="SCTP" />
              </el-select>
              <button v-if="rule.ports.length > 1" class="btn-remove-small" @click="rule.ports.splice(pIdx, 1)">
                <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
            <button type="button" class="btn-add-row-sm" @click="rule.ports.push({ port: '', protocol: 'TCP' })">+ 添加端口</button>
          </div>
        </div>
        <button type="button" class="btn-add-rule" @click="addEgressRule">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          添加 Egress 规则
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
import api from '../api'

interface SelectorItem {
  key: string
  value: string
}

interface PeerForm {
  namespaceSelector: string
  podSelector: string
  ipBlock: string
}

interface PortForm {
  port: string
  protocol: string
}

interface RuleForm {
  from: PeerForm[]
  to: PeerForm[]
  ports: PortForm[]
}

interface KeyValueItem {
  key: string
  value: string
}

const policies = ref<any[]>([])
const nsList = ref<any[]>([])
const { namespace } = useGlobalNamespace()
const loading = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)

const form = ref({
  name: '',
  namespace: '',
  podSelector: [{ key: '', value: '' }] as SelectorItem[],
  ingress: [] as RuleForm[],
  egress: [] as RuleForm[],
  labels: [{ key: '', value: '' }] as KeyValueItem[],
  policyTypes: [] as string[],
})

const resetForm = () => {
  form.value = {
    name: '',
    namespace: namespace.value || '',
    podSelector: [{ key: '', value: '' }],
    ingress: [],
    egress: [],
    labels: [{ key: '', value: '' }],
    policyTypes: [],
  }
}

const parseLabelMap = (map: Record<string, string> | undefined): SelectorItem[] => {
  if (!map || Object.keys(map).length === 0) return [{ key: '', value: '' }]
  return Object.entries(map).map(([k, v]) => ({ key: k, value: v }))
}

const parseSelectorString = (map: Record<string, string> | undefined): string => {
  if (!map || Object.keys(map).length === 0) return ''
  return Object.entries(map).map(([k, v]) => `${k}=${v}`).join(',')
}

const parseSelectorForm = (str: string): Record<string, string> => {
  const result: Record<string, string> = {}
  if (!str) return result
  str.split(',').forEach(part => {
    const [k, ...rest] = part.trim().split('=')
    if (k && rest.length > 0) {
      result[k.trim()] = rest.join('=').trim()
    }
  })
  return result
}

const openCreateDialog = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = async (row: any) => {
  isEdit.value = true
  try {
    const res: any = await api.get('/networkpolicies/get', {
      params: { namespace: row.namespace, name: row.name },
    })
    const detail = res.data
    form.value = {
      name: detail.name,
      namespace: detail.namespace,
      podSelector: parseLabelMap(detail.podSelector),
      ingress: (detail.ingress || []).map((r: any) => ({
        from: (r.from || []).map((f: any) => ({
          namespaceSelector: parseSelectorString(f.namespaceSelector),
          podSelector: parseSelectorString(f.podSelector),
          ipBlock: f.ipBlock || '',
        })),
        to: [],
        ports: (r.ports || []).map((p: any) => ({
          port: p.port || '',
          protocol: p.protocol || 'TCP',
        })),
      })),
      egress: (detail.egress || []).map((r: any) => ({
        from: [],
        to: (r.to || []).map((t: any) => ({
          namespaceSelector: parseSelectorString(t.namespaceSelector),
          podSelector: parseSelectorString(t.podSelector),
          ipBlock: t.ipBlock || '',
        })),
        ports: (r.ports || []).map((p: any) => ({
          port: p.port || '',
          protocol: p.protocol || 'TCP',
        })),
      })),
      labels: parseLabelMapForForm(detail.labels),
      policyTypes: detail.policyTypes || [],
    }
    dialogVisible.value = true
  } catch (e) {
    ElMessage.error('获取 NetworkPolicy 详情失败')
  }
}

const addIngressRule = () => {
  form.value.ingress.push({
    from: [{ namespaceSelector: '', podSelector: '', ipBlock: '' }],
    to: [],
    ports: [{ port: '', protocol: 'TCP' }],
  })
}

const addEgressRule = () => {
  form.value.egress.push({
    from: [],
    to: [{ namespaceSelector: '', podSelector: '', ipBlock: '' }],
    ports: [{ port: '', protocol: 'TCP' }],
  })
}

const buildPodSelector = (items: SelectorItem[]): Record<string, string> => {
  const result: Record<string, string> = {}
  items.forEach(s => {
    if (s.key && s.value) {
      result[s.key] = s.value
    }
  })
  return result
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

const parseLabelMapForForm = (map: Record<string, string> | undefined): KeyValueItem[] => {
  if (!map || Object.keys(map).length === 0) return [{ key: '', value: '' }]
  return Object.entries(map).map(([k, v]) => ({ key: k, value: v }))
}

const handleSubmit = async () => {
  if (!form.value.name) {
    ElMessage.warning('请输入名称')
    return
  }
  if (!form.value.namespace) {
    ElMessage.warning('请选择命名空间')
    return
  }

  submitting.value = true
  try {
    const payload: any = {
      name: form.value.name,
      namespace: form.value.namespace,
      podSelector: buildPodSelector(form.value.podSelector),
      labels: buildLabelMap(form.value.labels),
      policyTypes: form.value.policyTypes.length > 0 ? form.value.policyTypes : undefined,
      ingress: form.value.ingress.map(r => ({
        from: r.from.map(f => ({
          namespaceSelector: parseSelectorString(f.namespaceSelector) ? parseSelectorForm(f.namespaceSelector) : undefined,
          podSelector: parseSelectorString(f.podSelector) ? parseSelectorForm(f.podSelector) : undefined,
          ipBlock: f.ipBlock || undefined,
        })),
        ports: r.ports.filter(p => p.port).map(p => ({
          port: p.port,
          protocol: p.protocol,
        })),
      })),
      egress: form.value.egress.map(r => ({
        to: r.to.map(t => ({
          namespaceSelector: parseSelectorString(t.namespaceSelector) ? parseSelectorForm(t.namespaceSelector) : undefined,
          podSelector: parseSelectorString(t.podSelector) ? parseSelectorForm(t.podSelector) : undefined,
          ipBlock: t.ipBlock || undefined,
        })),
        ports: r.ports.filter(p => p.port).map(p => ({
          port: p.port,
          protocol: p.protocol,
        })),
      })),
    }

    await api.post('/networkpolicies/create', payload)
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
    const res: any = await api.get('/networkpolicies', { params })
    policies.value = res.data || []
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
    await ElMessageBox.confirm(`确定要删除 NetworkPolicy ${row.name} 吗？`, '确认', { type: 'warning' })
    await api.delete(`/networkpolicies/delete?namespace=${row.namespace}&name=${row.name}`)
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
.networkpolicies-page {
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

.cell-metric {
  font-variant-numeric: tabular-nums;
  font-weight: 500;
}

.cell-dim {
  font-size: 12px;
  color: #64748b;
  font-style: italic;
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

/* Dialog styles */
.networkpolicies-page :deep(.el-overlay) {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
}

.np-dialog :deep(.el-dialog) {
  background: #1e293b !important;
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 16px !important;
}

.np-dialog :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(148, 163, 184, 0.1);
  padding: 20px 24px;
  margin: 0;
}

.np-dialog :deep(.el-dialog__title) {
  color: #f1f5f9;
  font-size: 18px;
  font-weight: 600;
}

.np-dialog :deep(.el-dialog__body) {
  padding: 24px;
  color: #e2e8f0;
  max-height: 60vh;
  overflow-y: auto;
}

.np-dialog :deep(.el-dialog__footer) {
  border-top: 1px solid rgba(148, 163, 184, 0.1);
  padding: 16px 24px;
}

.np-form :deep(.el-form-item__label) {
  color: #94a3b8 !important;
  font-size: 13px;
}

.np-form :deep(.el-input__wrapper) {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 8px;
  box-shadow: none;
}

.np-form :deep(.el-input__wrapper:hover) {
  border-color: rgba(99, 102, 241, 0.4);
}

.np-form :deep(.el-input__wrapper.is-focus) {
  border-color: #6366f1;
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.np-form :deep(.el-input__inner) {
  color: #f1f5f9;
}

.np-form :deep(.el-input__inner::placeholder) {
  color: #64748b;
}

.np-form :deep(.el-select) {
  width: 100%;
}

.np-form :deep(.el-select .el-input__wrapper) {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.15);
  border-radius: 8px;
  box-shadow: none;
}

.np-form :deep(.el-divider) {
  border-color: rgba(148, 163, 184, 0.1);
  margin: 20px 0;
}

.np-form :deep(.el-divider__text) {
  background: #1e293b;
  color: #94a3b8;
  font-size: 13px;
  font-weight: 500;
}

.form-row {
  display: flex;
  gap: 16px;
}

.form-item-half {
  flex: 1;
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

.rule-section {
  background: rgba(15, 23, 42, 0.5);
  border: 1px solid rgba(148, 163, 184, 0.08);
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.rule-label {
  font-size: 14px;
  font-weight: 600;
  color: #e2e8f0;
}

.sub-section {
  margin-bottom: 12px;
}

.sub-label {
  font-size: 12px;
  font-weight: 500;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  display: block;
  margin-bottom: 8px;
}

.peer-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.peer-field {
  flex: 1;
}

.port-row {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
  align-items: center;
}

.port-field {
  flex: 1;
}

.port-proto {
  width: 120px;
  flex: none;
}

.btn-remove-small {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 6px;
  background: rgba(239, 68, 68, 0.12);
  color: #f87171;
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
}

.btn-remove-small:hover {
  background: rgba(239, 68, 68, 0.25);
}

.btn-add-row,
.btn-add-row-sm,
.btn-add-rule {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px dashed rgba(99, 102, 241, 0.3);
  border-radius: 6px;
  background: transparent;
  color: #818cf8;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-add-row {
  margin-bottom: 16px;
}

.btn-add-row-sm {
  margin-bottom: 8px;
}

.btn-add-rule {
  display: flex;
  justify-content: center;
  gap: 8px;
  width: 100%;
  padding: 10px;
  border-radius: 8px;
  font-size: 13px;
  margin-top: 8px;
}

.btn-add-row:hover,
.btn-add-row-sm:hover,
.btn-add-rule:hover {
  border-color: rgba(99, 102, 241, 0.5);
  background: rgba(99, 102, 241, 0.08);
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
