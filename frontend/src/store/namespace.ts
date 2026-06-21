import { ref, watch } from 'vue'

const STORAGE_KEY = 'kubeops-namespace'

// 全局命名空间状态
const globalNamespace = ref(localStorage.getItem(STORAGE_KEY) || '')

// 监听变化并持久化
watch(globalNamespace, (val) => {
  if (val) {
    localStorage.setItem(STORAGE_KEY, val)
  } else {
    localStorage.removeItem(STORAGE_KEY)
  }
})

export function useGlobalNamespace() {
  return {
    namespace: globalNamespace,
    setNamespace: (val: string) => {
      globalNamespace.value = val
    }
  }
}
