import { ref, computed } from 'vue'
import type { SessionSummary } from '../types'

const sessions = ref<SessionSummary[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const selectedSession = ref<SessionSummary | null>(null)
const sessionLines = ref<string[]>([])

// Pins from localStorage
const pinnedIds = ref<Set<string>>(loadPins())

function loadPins(): Set<string> {
  try {
    const raw = localStorage.getItem('pi-sessions-pinned')
    if (raw) return new Set(JSON.parse(raw))
  } catch { /* empty */ }
  return new Set()
}

function savePins() {
  localStorage.setItem('pi-sessions-pinned', JSON.stringify([...pinnedIds.value]))
}

function togglePin(id: string) {
  if (pinnedIds.value.has(id)) {
    pinnedIds.value.delete(id)
  } else {
    pinnedIds.value = new Set([...pinnedIds.value, id])
  }
  savePins()
}

function isPinned(id: string): boolean {
  return pinnedIds.value.has(id)
}

// Grouped sessions for sidebar
interface SessionGroup {
  label: string
  sessions: SessionSummary[]
}

const groupedSessions = computed<SessionGroup[]>(() => {
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const yesterday = new Date(today.getTime() - 86400000)
  const weekStart = new Date(today.getTime() - 7 * 86400000)

  const pinned: SessionSummary[] = []
  const todaySessions: SessionSummary[] = []
  const yesterdaySessions: SessionSummary[] = []
  const weekSessions: SessionSummary[] = []
  const older: SessionSummary[] = []

  for (const s of sessions.value) {
    if (isPinned(s.id)) {
      pinned.push(s)
      continue
    }
    const d = new Date(s.timestamp)
    if (d >= today) todaySessions.push(s)
    else if (d >= yesterday) yesterdaySessions.push(s)
    else if (d >= weekStart) weekSessions.push(s)
    else older.push(s)
  }

  pinned.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())

  const groups: SessionGroup[] = []
  if (pinned.length) groups.push({ label: '📌 Fijadas', sessions: pinned })
  if (todaySessions.length) groups.push({ label: 'Hoy', sessions: todaySessions })
  if (yesterdaySessions.length) groups.push({ label: 'Ayer', sessions: yesterdaySessions })
  if (weekSessions.length) groups.push({ label: 'Esta semana', sessions: weekSessions })
  if (older.length) groups.push({ label: 'Anterior', sessions: older })
  return groups
})

async function loadSessions() {
  loading.value = true
  error.value = null
  try {
    const res = await fetch('/api/sessions')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    sessions.value = await res.json()
  } catch (e: any) {
    error.value = e.message
    sessions.value = []
  } finally {
    loading.value = false
  }
}

async function selectSession(s: SessionSummary) {
  selectedSession.value = s
  sessionLines.value = []
  try {
    const res = await fetch(`/api/session?file=${encodeURIComponent(s.file)}`)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    sessionLines.value = await res.json()
  } catch (e: any) {
    error.value = e.message
    sessionLines.value = []
  }
}

export function useSessions() {
  return {
    sessions, sessionLines,
    loading,
    error,
    selectedSession,
    groupedSessions,
    pinnedIds,
    togglePin,
    isPinned,
    loadSessions,
    selectSession,
  }
}
