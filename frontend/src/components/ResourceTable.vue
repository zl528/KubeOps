<template>
  <el-card shadow="hover" class="resource-table-card">
    <template #header>
      <div class="card-header">
        <el-input
          v-model="searchText"
          placeholder="搜索..."
          clearable
          :prefix-icon="Search"
          style="width: 260px"
        />
        <el-button type="primary" @click="$emit('refresh')" :loading="loading">刷新</el-button>
      </div>
    </template>
    <el-table
      :data="filteredData"
      v-loading="loading"
      stripe
      @row-click="(row: any) => $emit('row-click', row)"
    >
      <el-table-column
        v-for="col in columns"
        :key="col.prop"
        :prop="col.prop"
        :label="col.label"
        :width="col.width"
        :min-width="col.minWidth"
        :show-overflow-tooltip="col.tooltip !== false"
      >
        <template v-if="col.slot" #default="{ row }">
          <slot :name="col.slot" :row="row" :prop="col.prop" />
        </template>
      </el-table-column>
      <el-table-column v-if="actions.length" label="操作" :width="actions.length * 70 + 20" fixed="right">
        <template #default="{ row }">
          <el-button
            v-for="action in actions"
            :key="action.name"
            :type="action.type || 'info'"
            link
            size="small"
            :loading="row[`_loading_${action.name}`]"
            @click.stop="$emit('action', { name: action.name, row })"
          >
            {{ action.label }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Search } from '@element-plus/icons-vue'

export interface Column {
  prop: string
  label: string
  width?: number
  minWidth?: number
  slot?: string
  tooltip?: boolean
}

export interface Action {
  name: string
  label: string
  type?: 'primary' | 'warning' | 'info' | 'success' | 'danger'
}

const props = withDefaults(defineProps<{
  columns: Column[]
  actions?: Action[]
  data?: any[]
  loading?: boolean
}>(), {
  actions: () => [],
  data: () => [],
  loading: false
})

const emit = defineEmits<{
  (e: 'refresh'): void
  (e: 'action', payload: { action: string; row: any }): void
  (e: 'row-click', row: any): void
}>()

const searchText = ref('')

const filteredData = computed(() => {
  if (!searchText.value) return props.data
  const q = searchText.value.toLowerCase()
  return props.data.filter(row =>
    props.columns.some(col => {
      const val = row[col.prop]
      return val != null && String(val).toLowerCase().includes(q)
    })
  )
})
</script>

<style scoped>
.resource-table-card :deep(.card-header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
</style>
