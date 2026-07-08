<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { PhStar } from '@phosphor-icons/vue'
import { useSessions } from '../composables/useSessions'

const { sessions, groupedSessions, selectedSession, togglePin, isPinned, selectSession } = useSessions()
const search = ref('')
const projectFilter = ref('')
const collapsedGroups = ref(new Set<string>())
const focusedIdx = ref(-1)
const tooltip = ref<{ x: number; y: number; session: any } | null>(null)
const sidebarWidth = ref(Number(localStorage.getItem('pi-sidebar-width') || 320))

const projects = computed(() => {
  const set = new Set(sessions.value.map(s => s.project))
  return [...set].sort()
})

function shortProject(path: string): string {
  return path
    .replace(/^~\/[^/]+\/Documents\/Repositories\//, '')
    .replace(/^~\/[^/]+\//, '~/') || path
}

function formatDate(ts: string): string {
  const d = new Date(ts)
  return d.toLocaleDateString('es-ES', { day: '2-digit', month: 'short' }) +
    ' ' + d.toLocaleTimeString('es-ES', { hour: '2-digit', minute: '2-digit' })
}

function formatFullDate(ts: string): string {
  const d = new Date(ts)
  return d.toLocaleDateString('es-ES', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' }) +
    ' ' + d.toLocaleTimeString('es-ES', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

function formatTokens(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(0) + 'k'
  return String(n)
}

function highlightText(text: string): string {
  if (!search.value) return text
  const q = search.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return text.replace(new RegExp(`(${q})`, 'gi'), '<mark>$1</mark>')
}

function toggleGroup(label: string) {
  const next = new Set(collapsedGroups.value)
  if (next.has(label)) next.delete(label)
  else next.add(label)
  collapsedGroups.value = next
}

function isCollapsed(label: string): boolean {
  return collapsedGroups.value.has(label)
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

// Flatten visible sessions for keyboard nav
const flatSessions = computed(() => {
  const result: { session: any; groupLabel: string }[] = []
  for (const group of groupedSessions.value) {
    if (isCollapsed(group.label)) continue
    for (const s of filterSessions(group)) {
      result.push({ session: s, groupLabel: group.label })
    }
  }
  return result
})

function onKeydown(e: KeyboardEvent) {
  const len = flatSessions.value.length
  if (!len) return
  if (e.key === 'ArrowDown') { e.preventDefault(); focusedIdx.value = Math.min(focusedIdx.value + 1, len - 1) }
  else if (e.key === 'ArrowUp') { e.preventDefault(); focusedIdx.value = Math.max(focusedIdx.value - 1, 0) }
  else if (e.key === 'Enter' && focusedIdx.value >= 0) {
    const s = flatSessions.value[focusedIdx.value]
    if (s) selectSession(s.session)
  }
  else if (e.key === ' ' && focusedIdx.value >= 0) {
    e.preventDefault()
    const s = flatSessions.value[focusedIdx.value]
    if (s) togglePin(s.session.id)
  }
}

function showTooltip(e: MouseEvent, s: any) {
  tooltip.value = { x: e.clientX + 10, y: e.clientY + 10, session: s }
}
function hideTooltip() { tooltip.value = null }

// Resize
let resizing = false
function startResize(e: MouseEvent) {
  resizing = true
  document.addEventListener('mousemove', onResize)
  document.addEventListener('mouseup', stopResize)
  e.preventDefault()
}
function onResize(e: MouseEvent) {
  if (!resizing) return
  sidebarWidth.value = Math.max(200, Math.min(500, e.clientX))
}
function stopResize() {
  resizing = false
  document.removeEventListener('mousemove', onResize)
  document.removeEventListener('mouseup', stopResize)
  localStorage.setItem('pi-sidebar-width', String(sidebarWidth.value))
}
onUnmounted(() => { document.removeEventListener('mousemove', onResize); document.removeEventListener('mouseup', stopResize) })
</script>

<template>
  <aside class="sidebar" :style="{ width: sidebarWidth + 'px', minWidth: sidebarWidth + 'px' }" @keydown="onKeydown" tabindex="0">
    <div class="sidebar-header">
      <div class="si-logo">
        <span class="si-logo-icon">π</span>
        <span class="si-logo-text">Sessions</span>
      </div>
      <span class="si-total">{{ sessions.length }}</span>
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
          <div class="session-group-label" @click="toggleGroup(group.label)">
            <span class="sgl-arrow">{{ isCollapsed(group.label) ? '▸' : '▾' }}</span>
            {{ group.label }}
            <span class="sgl-count">{{ filterSessions(group).length }}</span>
          </div>
          <template v-if="!isCollapsed(group.label)">
            <div
              v-for="(s, si) in filterSessions(group)"
              :key="s.id"
              class="si-item"
              :class="{
                active: selectedSession?.id === s.id,
                focused: flatSessions.findIndex(f => f.session.id === s.id) === focusedIdx
              }"
              @click="selectSession(s)"
              @mouseenter="showTooltip($event, s)"
              @mouseleave="hideTooltip"
            >
              <div class="si-text">
                <div class="si-title" v-html="highlightText(s.name || s.first_user || '(vacío)')" />
                <div class="si-meta">
                  <span>{{ formatDate(s.timestamp) }}</span>
                  <span v-if="s.total_cost" class="si-cost">${{ s.total_cost.toFixed(2) }}</span>
                  <span v-if="s.total_tokens" class="si-tokens">{{ formatTokens(s.total_tokens) }} tk</span>
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
          </template>
        </div>
      </template>
      <div v-if="allEmpty()" class="sidebar-empty">Sin resultados</div>
    </div>

    <!-- Resize handle -->
    <div class="sidebar-resize" @mousedown="startResize" />

    <!-- Tooltip -->
    <Teleport to="body">
      <div v-if="tooltip" class="si-tooltip" :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }">
        <div class="si-tt-title" :title="tooltip.session.name || tooltip.session.first_user">{{ tooltip.session.name || tooltip.session.first_user }}</div>
        <div class="si-tt-row" :title="tooltip.session.project">📁 {{ shortProject(tooltip.session.project) }}</div>
        <div class="si-tt-row">🕐 {{ formatFullDate(tooltip.session.timestamp) }}</div>
        <div class="si-tt-row">💬 {{ tooltip.session.user_msgs }} user · {{ tooltip.session.assistant_msgs }} assistant</div>
        <div class="si-tt-row">🤖 {{ tooltip.session.provider }}/{{ tooltip.session.model }}</div>
        <div class="si-tt-row" v-if="tooltip.session.total_cost">💰 ${{ tooltip.session.total_cost.toFixed(4) }}</div>
      </div>
    </Teleport>
  </aside>
</template>

<style scoped>
.sidebar {
  height: 100vh;
  background: var(--surface-dark);
  color: var(--on-surface-dark);
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--hairline);
  overflow: hidden;
  position: relative;
  outline: none;
}
.sidebar-header {
  padding: 16px 20px 12px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.si-logo { display: flex; align-items: center; gap: 8px; }
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
.si-total {
  font-size: 12px;
  color: rgba(255,255,255,.3);
  font-family: var(--font-mono);
  background: rgba(255,255,255,.06);
  padding: 2px 8px;
  border-radius: 10px;
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
.si-search:focus { border-color: var(--primary); background: rgba(255,255,255,.1); }
.si-search::placeholder { color: rgba(255,255,255,.25); }
.si-select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='none'%3E%3Cpath d='M3 4.5l3 3 3-3' stroke='%23ffffff44' stroke-width='1.5'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  padding-right: 30px;
}
.si-select option { background: var(--surface-dark); color: var(--on-surface-dark); }

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

.session-group { margin-bottom: 4px; }
.session-group-label {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 1.2px;
  color: rgba(255,255,255,.35);
  padding: 12px 8px 4px;
  font-weight: 500;
  cursor: pointer;
  user-select: none;
  display: flex;
  align-items: center;
  gap: 4px;
  transition: color .15s;
}
.session-group-label:hover { color: rgba(255,255,255,.55); }
.sgl-arrow { font-size: 9px; width: 10px; }
.sgl-count {
  font-size: 10px;
  color: rgba(255,255,255,.2);
  font-weight: 400;
  margin-left: auto;
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
.si-item:hover { background: rgba(255,255,255,.05); }
.si-item.active { background: rgba(204,120,92,.15); }
.si-item.focused { background: rgba(255,255,255,.08); }
.si-text { flex: 1; min-width: 0; }
.si-title {
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.si-title :deep(mark) {
  background: rgba(204,120,92,.3);
  color: var(--on-surface-dark);
  border-radius: 2px;
  padding: 0 1px;
}
.si-fork { color: var(--primary); font-size: 14px; margin-right: 2px; }
.si-meta {
  font-size: 11px;
  color: rgba(255,255,255,.4);
  margin-top: 2px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}
.si-cost { color: rgba(255,255,255,.5); }
.si-tokens { color: rgba(255,255,255,.35); }
.si-id-mono { font-family: var(--font-mono); font-size: 10px; opacity: .6; }
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
.si-pin:hover { color: rgba(255,255,255,.5); }
.si-pin.pinned { color: var(--primary); }
.sidebar-empty { text-align: center; color: rgba(255,255,255,.25); font-size: 13px; padding: 40px 0; }

/* Resize handle */
.sidebar-resize {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  cursor: col-resize;
  z-index: 5;
}
.sidebar-resize:hover { background: rgba(204,120,92,.3); }

/* Tooltip */
.si-tooltip {
  position: fixed;
  z-index: 1000;
  background: var(--surface-dark);
  border: 1px solid rgba(255,255,255,.1);
  border-radius: 10px;
  padding: 12px 16px;
  max-width: 300px;
  pointer-events: none;
  box-shadow: 0 8px 24px rgba(0,0,0,.4);
}
.si-tt-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--on-surface-dark);
  margin-bottom: 6px;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 260px;
}
.si-tt-row {
  font-size: 11px;
  color: rgba(255,255,255,.5);
  margin: 2px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
