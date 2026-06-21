import { defineStore } from 'pinia'
import api from '../api'

interface ClusterOverview {
  totalNodes: number
  readyNodes: number
  namespaces: number
  totalPods: number
  runningPods: number
  pendingPods: number
  failedPods: number
  cpuUsage: { used: string; total: string; percentage: string }
  memoryUsage: { used: string; total: string; percentage: string }
}

export const useClusterStore = defineStore('cluster', {
  state: () => ({
    overview: null as ClusterOverview | null,
    loading: false
  }),
  actions: {
    async fetchOverview() {
      this.loading = true
      try {
        const res: any = await api.get('/cluster/overview')
        this.overview = res.data
      } catch (e) {
        console.error('Failed to fetch overview:', e)
      } finally {
        this.loading = false
      }
    }
  }
})
