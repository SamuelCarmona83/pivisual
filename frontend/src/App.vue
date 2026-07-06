<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { PhList } from '@phosphor-icons/vue'
import { useSessions } from './composables/useSessions'
import AppSidebar from './components/AppSidebar.vue'
import AppDashboard from './components/AppDashboard.vue'
import ChatView from './components/ChatView.vue'

const { sessions, loading, error, selectedSession, loadSessions } = useSessions()
const sidebarOpen = ref(true)

onMounted(() => {
  loadSessions()
})

function toggleSidebar() {
  sidebarOpen.value = !sidebarOpen.value
}

// Ctrl+B / Cmd+B to toggle sidebar
function onKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'b') {
    e.preventDefault()
    toggleSidebar()
  }
}
</script>

<template>
  <div class="app-shell" @keydown="onKeydown">
    <Transition name="slide">
      <AppSidebar v-if="sidebarOpen" />
    </Transition>

    <main class="main-area">
      <button class="sidebar-toggle" @click="toggleSidebar" :title="sidebarOpen ? 'Cerrar sidebar' : 'Abrir sidebar'">
        <PhList :size="20" />
      </button>

      <ChatView v-if="selectedSession" />
      <AppDashboard v-else-if="sessions" />

      <div v-if="loading" class="loading-overlay">Cargando sesiones…</div>
      <div v-if="error" class="error-banner">{{ error }}</div>
    </main>
  </div>
</template>

<style scoped>
.app-shell {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--canvas);
}
.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
  min-width: 0;
}
.sidebar-toggle {
  position: absolute;
  top: 14px;
  left: 12px;
  z-index: 10;
  background: rgba(255,255,255,.85);
  backdrop-filter: blur(8px);
  border: 1px solid var(--hairline);
  border-radius: 8px;
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--muted);
  transition: color .15s, box-shadow .15s;
  box-shadow: 0 1px 3px rgba(0,0,0,.06);
}
.sidebar-toggle:hover { color: var(--ink); box-shadow: 0 2px 6px rgba(0,0,0,.1); }

.slide-enter-active, .slide-leave-active {
  transition: width .2s ease, min-width .2s ease, opacity .2s ease;
}
.slide-enter-from, .slide-leave-to {
  width: 0 !important;
  min-width: 0 !important;
  opacity: 0;
}

.loading-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--muted);
  font-size: 14px;
  background: rgba(250,249,245,.7);
  z-index: 20;
}
.error-banner {
  position: absolute;
  bottom: 16px;
  left: 50%;
  transform: translateX(-50%);
  background: #e06c75;
  color: #fff;
  padding: 8px 20px;
  border-radius: 8px;
  font-size: 13px;
  z-index: 20;
}
</style>
