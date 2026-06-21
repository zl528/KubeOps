import { onMounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'

export function useHighlightRow() {
  const route = useRoute()

  const scrollToRow = () => {
    const name = route.query.highlight as string
    if (!name) return

    nextTick(() => {
      setTimeout(() => {
        const rows = document.querySelectorAll('.el-table__body tr')
        for (const row of rows) {
          const nameCell = row.querySelector('.cell-name')
          if (nameCell && nameCell.textContent?.trim() === name) {
            row.scrollIntoView({ behavior: 'smooth', block: 'center' })
            row.classList.add('highlighted-row')
            setTimeout(() => row.classList.remove('highlighted-row'), 3000)
            break
          }
        }
      }, 500)
    })
  }

  onMounted(scrollToRow)
  watch(() => route.query.highlight, scrollToRow)
}
