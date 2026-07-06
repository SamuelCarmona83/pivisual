<script setup lang="ts">
import { ref } from 'vue'
import { PhStar } from '@phosphor-icons/vue'
import { useSessions } from '../composables/useSessions'

const { groupedSessions, selectedSession, togglePin, isPinned, selectSession } = useSessions()
const search = ref('')
const projectFilter = ref('')

function formatDate(ts: string): string {
  const d = new Date(ts)
  return d.toLocaleDateString('es-ES', { day: '2-digit', month: 'short' }) +
    ' ' + d.toLocaleTimeString('es-ES', { hour: '2-digit', minute: '2-digit' })
}

function filterSessions(group: any) {
  if (!search.value && !projectFilter.value) return group.sessions
  const q = search.value.toLowerCase()
  const pf = projectFilter.value.toLowerCase()
  return group.sessions.filter((s: any) => {
    const text = (s.name || s.first_user || '').toLowerCase()
    const proj = s.project.toLowerCase()
    return (!q || text.includes(q)) && (!pf || proj.includes(pf))
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
      <input v-model="projectFilter" type="text" class="si-search" placeholder="Filtrar proyecto..." />
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
              <div class="si-title">{{ s.name || s.first_user || '(vacío)' }}</div>
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
.sidebar-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 12px 12px;
}
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
