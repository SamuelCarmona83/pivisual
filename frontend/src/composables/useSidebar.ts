import { ref } from 'vue'

const sidebarOpen = ref(true)

export function useSidebar() {
  return { sidebarOpen }
}
