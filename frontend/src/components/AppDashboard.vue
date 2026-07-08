<script setup lang="ts">
import { computed } from 'vue'
import { PhList } from '@phosphor-icons/vue'
import { useSessions } from '../composables/useSessions'
import { useSidebar } from '../composables/useSidebar'

const { sessions, selectSession } = useSessions()
const { sidebarOpen } = useSidebar()

function esc(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

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

const totalTokens = computed(() => sessions.value.reduce((a, s) => a + s.total_tokens, 0))
const totalCost = computed(() => sessions.value.reduce((a, s) => a + s.total_cost, 0))
const projectCount = computed(() => new Set(sessions.value.map(s => s.project)).size)

const topModel = computed(() => {
  const counts: Record<string, number> = {}
  for (const s of sessions.value) {
    const k = s.provider + '/' + s.model
    counts[k] = (counts[k] || 0) + 1
  }
  const entries = Object.entries(counts).sort((a, b) => b[1] - a[1])
  return entries[0] || ['—', 0]
})

const recentSessions = computed(() =>
  [...sessions.value]
    .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
    .slice(0, 6)
)
</script>

<template>
  <div class="dashboard">
    <button class="dash-toggle" @click="sidebarOpen = !sidebarOpen" :title="sidebarOpen ? 'Cerrar sidebar' : 'Abrir sidebar'">
      <PhList :size="18" />
    </button>
    <h2 class="dash-title">Pi Visual</h2>
    <p class="dash-subtitle">Session viewer for Pi Coding Agent</p>

    <div class="dash-stats">
      <div class="dash-stat">
        <div class="ds-value">{{ sessions.length }}</div>
        <div class="ds-label">Sesiones totales</div>
        <div class="ds-sub">{{ projectCount }} proyectos</div>
      </div>
      <div class="dash-stat">
        <div class="ds-value">{{ (totalTokens / 1000).toFixed(1) }}k</div>
        <div class="ds-label">Tokens totales</div>
        <div class="ds-sub">~{{ Math.round(totalTokens / (sessions.length || 1)).toLocaleString() }} / sesión</div>
      </div>
      <div class="dash-stat">
        <div class="ds-value">${{ totalCost.toFixed(2) }}</div>
        <div class="ds-label">Costo total</div>
        <div class="ds-sub">API usage acumulado</div>
      </div>
      <div class="dash-stat">
        <div class="ds-value ds-model">{{ topModel[0] }}</div>
        <div class="ds-label">Modelo más usado</div>
        <div class="ds-sub">{{ topModel[1] }} sesiones</div>
      </div>
    </div>

    <div v-if="recentSessions.length" class="dash-recent">
      <h3>Sesiones recientes</h3>
      <div class="dash-recent-grid">
        <div
          v-for="s in recentSessions"
          :key="s.id"
          class="dash-session-card"
          @click="selectSession(s)"
        >
          <div class="dsc-project">📁 {{ esc(shortProject(s.project)) }}</div>
          <div class="dsc-title">{{ s.name || s.first_user || '(vacío)' }}</div>
          <div class="dsc-meta">
            <span>{{ formatDate(s.timestamp) }}</span>
            <span>{{ s.provider }}/{{ s.model }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  max-width: 900px;
  margin: auto;
  padding: 48px 24px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 100%;
}
.dash-title {
  font-family: var(--font-display);
  font-size: 32px;
  font-weight: 400;
  letter-spacing: -0.3px;
  color: var(--ink);
  text-align: center;
  margin-bottom: 4px;
}

.dash-toggle {
  background: none;
  border: none;
  color: var(--muted);
  cursor: pointer;
  width: 30px;
  height: 30px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: absolute;
  top: 16px;
  left: 16px;
  transition: color .15s, background .15s;
}
.dash-toggle:hover { color: var(--ink); background: var(--hairline); }

.dash-title {
  font-family: var(--font-display);
  font-size: 32px;
  font-weight: 400;
  letter-spacing: -0.3px;
  color: var(--ink);
  text-align: center;
  margin-bottom: 4px;
}
.dash-subtitle {
  font-size: 14px;
  color: var(--muted);
  text-align: center;
  margin-bottom: 32px;
}
.dash-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 32px;
}
.dash-stat {
  background: var(--surface-card);
  border-radius: 12px;
  padding: 20px;
  text-align: center;
  border: 1px solid var(--hairline);
}
.ds-value {
  font-family: var(--font-display);
  font-size: 28px;
  font-weight: 400;
  letter-spacing: -0.3px;
  color: var(--ink);
  margin-bottom: 4px;
}
.ds-value.ds-model {
  font-size: 18px;
  font-family: var(--font-body);
  font-weight: 500;
}
.ds-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--muted);
  font-weight: 500;
}
.ds-sub {
  font-size: 12px;
  color: var(--muted-soft);
  margin-top: 2px;
}
.dash-recent h3 {
  font-family: var(--font-display);
  font-size: 20px;
  font-weight: 400;
  letter-spacing: -0.2px;
  color: var(--ink);
  margin-bottom: 16px;
}
.dash-recent-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 8px;
}
.dash-session-card {
  background: var(--canvas);
  border: 1px solid var(--hairline);
  border-radius: 10px;
  padding: 10px 14px;
  cursor: pointer;
  transition: border-color .15s, background .15s;
  overflow: hidden;
  min-width: 0;
}
.dash-session-card:hover {
  border-color: var(--primary);
  background: var(--surface-card);
}
.dsc-project {
  font-size: 11px;
  color: var(--muted-soft);
  margin-bottom: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.dsc-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.dsc-meta {
  font-size: 11px;
  color: var(--muted);
  margin-top: 4px;
  display: flex;
  gap: 8px;
  overflow: hidden;
}
@media (max-width: 700px) {
  .dash-stats { grid-template-columns: repeat(2, 1fr); }
}
</style>
