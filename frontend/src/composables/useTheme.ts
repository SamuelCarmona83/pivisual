import { ref, watchEffect } from 'vue'

const stored = localStorage.getItem('pi-theme')
const isDark = ref(stored === 'dark')

watchEffect(() => {
  document.documentElement.setAttribute('data-theme', isDark.value ? 'dark' : '')
  localStorage.setItem('pi-theme', isDark.value ? 'dark' : 'light')
})

export function useTheme() {
  function toggle() { isDark.value = !isDark.value }
  return { isDark, toggle }
}
