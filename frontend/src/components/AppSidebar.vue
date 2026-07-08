<script setup lang="ts">
import { ref, computed } from 'vue'
import { PhStar } from '@phosphor-icons/vue'
import { useSessions } from '../composables/useSessions'

const { sessions, groupedSessions, selectedSession, togglePin, isPinned, selectSession } = useSessions()
const search = ref('')
const projectFilter = ref('')

const projects = computed(() => {
  const set = new Set(sessions.value.map(s => s.project))
  return [...set].sort()
})

function formatDate(ts: string): string {
  const d = new Date(ts)
  return d.toLocaleDateString('es-ES', { day: '2-digit', month: 'short' }) +
    ' ' + d.toLocaleTimeString('es-ES', { hour: '2-digit', minute: '2-digit' })
}

function filterSessions(group: any) {
  if (!search.value && !projectFilter.value) return group.sessions
  const q = search.value.toLowerCase()
  return group.sessions.filter((s: any) => {
    const text = (s.name || s.first_user || '').toLowerCase()
    return (!q || text.includes(q)) && (!projectFilter.value || s.project === projectFilter.value)
  })
}

function isEmpty(group: any) {
  return filterSessions(group).length === 0
}

function allEmpty() {
  return groupedSessions.value.every(isEmpty)
}
</script>

<template>
  <aside class="sidebar" :class="{ collapsed: false }">
    <div class="sidebar-header">
      <div class="si-logo">
        <span class="si-logo-icon">π</span>
        <span class="si-logo-text">Sessions</span>
      </div>
    </div>

    <div class="sidebar-filters">
      <input v-model="search" type="text" class="si-search" placeholder="Buscar sesiones..." />
      <select v-model="projectFilter" class="si-search si-select">
        <option value="">Todos los proyectos</option>
        <option v-for="p in projects" :key="p" :value="p">{{ p }}</option>
      </select>
    </div>

    <div class="sidebar-list">
      <template v-for="group in groupedSessions" :key="group.label">
        <div v-if="!isEmpty(group)" class="session-group">
          <div class="session-group-label">{{ group.label }}</div>
          <div
            v-for="s in filterSessions(group)"
            :key="s.id"
            class="si-item"
            :class="{ active: selectedSession?.id === s.id }"
            @click="selectSession(s)"
          >
            <div class="si-text">
              <div class="si-title">
                <span v-if="s.parent_session" class="si-fork" title="Fork de otra sesión">↳</span>
                {{ s.name || s.first_user || '(vacío)' }}
              </div>
              <div class="si-meta">
                <span>{{ formatDate(s.timestamp) }}</span>
                <span class="si-id-mono">{{ s.id.slice(0, 8) }}</span>
                <span>{{ s.provider }}/{{ s.model }}</span>
              </div>
            </div>
            <button
              class="si-pin"
              :class="{ pinned: isPinned(s.id) }"
              @click.stop="togglePin(s.id)"
              :title="isPinned(s.id) ? 'Quitar pin' : 'Fijar'"
            >
              <PhStar :weight="isPinned(s.id) ? 'fill' : 'regular'" :size="14" />
            </button>
          </div>
        </div>
      </template>
      <div v-if="allEmpty()" class="sidebar-empty">
        Sin resultados
      </div>
    </div>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 320px;
  min-width: 320px;
  height: 100vh;
  background: var(--surface-dark);
  color: var(--on-surface-dark);
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--hairline);
  overflow: hidden;
}
.sidebar-header {
  padding: 16px 20px 12px;
  flex-shrink: 0;
}
.si-logo {
  display: flex;
  align-items: center;
  gap: 8px;
}
.si-logo-icon {
  font-family: var(--font-display);
  font-size: 22px;
  color: var(--primary);
}
.si-logo-text {
  font-family: var(--font-display);
  font-size: 18px;
  letter-spacing: -0.2px;
}
.sidebar-filters {
  padding: 0 16px 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex-shrink: 0;
}
.si-search {
  width: 100%;
  background: rgba(255,255,255,.06);
  border: 1px solid rgba(255,255,255,.08);
  color: var(--on-surface-dark);
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 13px;
  font-family: var(--font-body);
  outline: none;
  box-sizing: border-box;
}
.si-search:focus {
  border-color: var(--primary);
  background: rgba(255,255,255,.1);
}
.si-search::placeholder {
  color: rgba(255,255,255,.25);
}
.si-select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='none'%3E%3Cpath d='M3 4.5l3 3 3-3' stroke='%23ffffff44' stroke-width='1.5'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  padding-right: 30px;
}
.si-select option {
  background: var(--surface-dark);
  color: var(--on-surface-dark);
}
.sidebar-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 12px 12px;
  scrollbar-width: thin;
  scrollbar-color: rgba(255,255,255,.08) transparent;
}
.sidebar-list::-webkit-scrollbar { width: 5px }
.sidebar-list::-webkit-scrollbar-track { background: transparent }
.sidebar-list::-webkit-scrollbar-thumb { background: rgba(255,255,255,.08); border-radius: 3px }
.sidebar-list::-webkit-scrollbar-thumb:hover { background: rgba(255,255,255,.15) }
.session-group {
  margin-bottom: 4px;
}
.session-group-label {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 1.2px;
  color: rgba(255,255,255,.35);
  padding: 12px 8px 4px;
  font-weight: 500;
}
.si-item {
  display: flex;
  align-items: flex-start;
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  gap: 8px;
  transition: background .12s;
}
.si-item:hover {
  background: rgba(255,255,255,.05);
}
.si-item.active {
  background: rgba(204,120,92,.15);
}
.si-text {
  flex: 1;
  min-width: 0;
}
.si-title {
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.si-fork {
  color: var(--primary);
  font-size: 14px;
  margin-right: 2px;
}
.si-meta {
  font-size: 11px;
  color: rgba(255,255,255,.4);
  margin-top: 2px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.si-id-mono {
  font-family: var(--font-mono);
  font-size: 10px;
  opacity: .6;
}
.si-pin {
  background: none;
  border: none;
  color: rgba(255,255,255,.2);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  flex-shrink: 0;
  transition: color .15s;
  display: flex;
}
.si-pin:hover {
  color: rgba(255,255,255,.5);
}
.si-pin.pinned {
  color: var(--primary);
}
.sidebar-empty {
  text-align: center;
  color: rgba(255,255,255,.25);
  font-size: 13px;
  padding: 40px 0;
}
</style>
