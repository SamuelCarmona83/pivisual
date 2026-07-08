<script setup lang="ts">
import { computed, watch, nextTick, ref } from 'vue'
import { PhChats, PhGitFork, PhList, PhFileMd, PhSelectionAll, PhCpu } from '@phosphor-icons/vue'
import { useSidebar } from '../composables/useSidebar'
import { useSessions } from '../composables/useSessions'
import ChatMessage from './ChatMessage.vue'
import type { SessionEntry, ChatMessage as ChatMsg } from '../types'

const { sessions, selectedSession, sessionLines, selectSession: selectSessionFromComposable } = useSessions()
const { sidebarOpen } = useSidebar()

function copySessionId() {
  if (!selectedSession.value) return
  navigator.clipboard.writeText('pi --session ' + selectedSession.value.id).catch(() => {})
}

function openParentSession() {
  const parentFile = selectedSession.value?.parent_session
  if (!parentFile) return
  const parent = sessions.value.find(s => s.file === parentFile)
  if (parent) selectSessionFromComposable(parent)
}

// Message selection + markdown export
const selectionMode = ref(false)
const selectedIds = ref(new Set<string>())

function toggleSelectMode() {
  selectionMode.value = !selectionMode.value
  if (!selectionMode.value) selectedIds.value = new Set()
}

function toggleSelectMsg(id: string) {
  const next = new Set(selectedIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  selectedIds.value = next
}

function selectAll() {
  const ids = messages.value.map(m => m.id)
  selectedIds.value = new Set(ids)
}

function exportMarkdown() {
  const selected = messages.value.filter(m => selectedIds.value.has(m.id))
  let md = ''
  for (const msg of selected) {
    if (msg.role === 'user') {
      const text = typeof msg.content === 'string' ? msg.content : (Array.isArray(msg.content) ? msg.content.filter((b: any) => b.type === 'text').map((b: any) => b.text).join(' ') : '')
      md += `**You**: ${text}\n\n`
    } else if (msg.role === 'assistant') {
      const text = typeof msg.content === 'string' ? msg.content : (Array.isArray(msg.content) ? msg.content.filter((b: any) => b.type === 'text').map((b: any) => b.text).join(' ') : '')
      if (text) md += `**Assistant**: ${text}\n\n`
      // Tool calls
      if (Array.isArray(msg.content)) {
        for (const b of msg.content) {
          if (b.type === 'toolCall') {
            const result = msg.toolResults?.[b.id] || ''
            const cmd = (b.arguments as any)?.command || (b.arguments as any)?.path || ''
            md += `> \`${b.name} ${cmd}\`\n`
            if (result) md += `> \`\`\`\n> ${result.replace(/\n/g, '\n> ')}\n> \`\`\`\n`
            md += '\n'
          }
        }
      }
    }
  }
  navigator.clipboard.writeText(md).catch(() => {})
  selectionMode.value = false
  selectedIds.value = new Set()
}

interface RenderedMessage {
  role: string  // 'user' | 'assistant' | 'system' | 'work'
  content: string | any[]
  usage?: any
  toolResults: Record<string, string>
  systemText?: string
  id: string
}

const messages = computed<RenderedMessage[]>(() => {
  if (!sessionLines.value.length) return []

  // First pass: collect all tool results by call ID
  const toolResults: Record<string, string> = {}
  const entries: SessionEntry[] = []
  for (const line of sessionLines.value) {
    try {
      const entry: SessionEntry = JSON.parse(line)
      entries.push(entry)
      if (entry.type === 'message' && entry.message?.role === 'toolResult') {
        const content = entry.message.content
        const text = Array.isArray(content)
          ? content.map((c: any) => c.type === 'text' ? c.text || '' : '').join('\n')
          : String(content || '')
        toolResults[entry.message.toolCallId || ''] = text
      }
    } catch { /* skip malformed lines */ }
  }

  // Second pass: build message list, detect fork points
  const result: RenderedMessage[] = []
  let prevId: string | null = null
  for (const entry of entries) {
    // Detect fork: parentId jumps back to an earlier entry
    const entryId = entry.id || ''
    const parentId = entry.parentId || null
    if (prevId && parentId && parentId !== prevId) {
      // Find the forked entry to show context
      const forkedEntry = entries.find(e => e.id === parentId)
      let context = ''
      if (forkedEntry?.message) {
        const c = forkedEntry.message.content
        context = typeof c === 'string' ? c : (Array.isArray(c) ? c.filter(b => b.type === 'text').map(b => b.text).join(' ').slice(0, 100) : '')
      }
      result.push({
        role: 'fork',
        content: '',
        toolResults: {},
        systemText: context || 'punto anterior',
        id: entryId + '-fork',
      })
    }
    if (entryId) prevId = entryId

    if (entry.type === 'session_info' || entry.type === 'message') {
      const msg = entry.message as ChatMsg | undefined
      if (!msg) continue

      const role = msg.role
      if (role === 'user') {
        result.push({
          role: 'user',
          content: msg.content,
          usage: msg.usage,
          toolResults: {},
          id: entry.id || '',
        })
      } else if (role === 'assistant') {
        result.push({
          role: 'assistant',
          content: msg.content,
          usage: msg.usage,
          toolResults,
          id: entry.id || '',
        })
      }
    } else if (entry.type === 'model_change') {
      result.push({
        role: 'system',
        content: '',
        toolResults: {},
        systemText: `Modelo: ${entry.provider || '?'}/${entry.modelId || '?'}`,
        id: entry.id || '',
      })
    } else if (entry.type === 'thinking_level_change') {
      const levels: Record<string, string> = { low: 'Bajo', medium: 'Medio', high: 'Alto', ultra: 'Ultra' }
      result.push({
        role: 'system',
        content: '',
        toolResults: {},
        systemText: `Thinking: ${levels[entry.thinkingLevel as string] || entry.thinkingLevel}`,
        id: entry.id || '',
      })
    } else if (entry.type === 'compaction') {
      result.push({
        role: 'system',
        content: '',
        toolResults: {},
        systemText: 'Contexto compactado',
        id: entry.id || '',
      })
    } else if (entry.type === 'branch_summary') {
      result.push({
        role: 'system',
        content: '',
        toolResults: {},
        systemText: `↳ Fork: ${(entry.summary as string || '').slice(0, 80)}`,
        id: entry.id || '',
      })
    }
  }
  // Post-process: compact consecutive assistant-only (no text) messages into work blocks
  const compacted: RenderedMessage[] = []
  let workSteps: RenderedMessage[] = []
  
  function flushWork() {
    if (!workSteps.length) return
    if (workSteps.length === 1 && result.length && result[result.length-1].role === 'assistant') {
      // single work msg before final assistant: merge into a work-group with the next assistant
      // we'll handle it below
    }
    // Collect tool names and thinking presence
    let toolCount = 0
    let hasThinking = false
    const steps: any[] = []
    for (const w of workSteps) {
      const parsed = parseContentBlocks(w.content)
      toolCount += parsed.toolCount
      if (parsed.thinking) hasThinking = true
      steps.push({ id: w.id, content: w.content, toolResults: w.toolResults })
    }
    const label = []
    if (hasThinking) label.push('thinking')
    if (toolCount) label.push(toolCount + ' tools')
    compacted.push({
      role: 'work',
      content: steps,
      toolResults: {},
      systemText: label.join(', '),
      id: workSteps[0].id,
    })
    workSteps = []
  }

  /* rules: 

 ┌──────────────────────┬────────┬──────────────────────────────────────┐
 │ Content.             │ ¿Work? │ Result                                │
 ├──────────────────────┼────────┼──────────────────────────────────────┤
 │ {thinking, toolCall} │ ✅     │ compact                              │
 ├──────────────────────┼────────┼──────────────────────────────────────┤
 │ {} empty             │ ✅     │ compact                              │
 ├──────────────────────┼────────┼──────────────────────────────────────┤
 │ {text} solo          │ ❌     │ visible                              │
 ├──────────────────────┼────────┼──────────────────────────────────────┤
 │ {thinking, text}     │ ❌     │ visible (thinking colapse).          │
 ├──────────────────────┼────────┼──────────────────────────────────────┤
 │ {text, toolCall}     │ ✅     │ compact  (tiene tools)               │
 └──────────────────────┴────────┴──────────────────────────────────────┘

  {thinking, text} now shows as a normal message — the text is the real answer, the thinking is collapsed within the same message.
  Only messages with tool calls or empty are compacted.

  */
  function parseContentBlocks(content: string | any[]) {
    if (typeof content === 'string') return { toolCount: 0, thinking: false, isWork: false }
    if (!Array.isArray(content) || content.length === 0) return { toolCount: 0, thinking: false, isWork: true }
    let toolCount = 0, thinking = false, hasText = false
    for (const b of content) {
      if (b.type === 'toolCall') toolCount++
      else if (b.type === 'thinking') thinking = true
      else if (b.type === 'text' && (b.text || '').trim()) hasText = true
    }
    // Work = has tool calls OR empty. Text + thinking (no tools) = visible.
    return { toolCount, thinking, isWork: toolCount > 0 || !hasText }
  }

  for (const msg of result) {
    if (msg.role === 'assistant') {
      const parsed = parseContentBlocks(msg.content)
      if (parsed.isWork) {
        workSteps.push(msg)
        continue
      }
      // Assistant has no tools/thinking: flush pending work, then add this message
      flushWork()
      compacted.push(msg)
    } else {
      flushWork()
      compacted.push(msg)
    }
  }
  flushWork()

  return compacted
})

const chatBody = ref<HTMLElement | null>(null)
watch(messages, () => nextTick(() => {
  if (chatBody.value) chatBody.value.scrollTop = chatBody.value.scrollHeight
}))
</script>

<template>
  <div class="chat-view">
    <!-- Header -->
    <div class="chat-header">
      <button class="ch-toggle-sidebar" @click="sidebarOpen = !sidebarOpen" :title="sidebarOpen ? 'Cerrar sidebar' : 'Abrir sidebar'">
        <PhList :size="18" />
      </button>
      <button class="ch-select-btn" @click="toggleSelectMode" :title="selectionMode ? 'Salir de selección' : 'Seleccionar mensajes'">
        <span v-if="!selectionMode">
          <PhFileMd :size="16" />
        </span>
        <span v-else>
          <PhSelectionAll :size="16" />
        </span>
      </button>
      <span class="ch-title" :title="selectedSession?.name || selectedSession?.first_user || 'Seleccioná una sesión'">
        {{ selectedSession?.name || selectedSession?.first_user || 'Seleccioná una sesión' }}
      </span>
      <template v-if="selectionMode">
        <button class="ch-select-btn" @click="selectAll" title="Seleccionar todos">All</button>
        <button v-if="selectedIds.size" class="ch-export-btn" @click="exportMarkdown">
          Copy {{ selectedIds.size }} as Markdown
        </button>
      </template>
      <span v-if="selectedSession?.parent_session" class="ch-parent" @click="openParentSession" title="Abrir sesión padre">
        ↳ Fork
      </span>
      <span v-if="selectedSession && !selectionMode" class="ch-id" :title="'pi --session ' + selectedSession.id" @click="copySessionId">
        pi --session {{ selectedSession.id }}
      </span>
    </div>

    <!-- Empty state -->
    <div v-if="!selectedSession" class="chat-empty">
      <PhChats :size="48" weight="light" />
      <p>Seleccioná una sesión del sidebar para verla como chat</p>
    </div>

    <!-- Messages -->
    <div v-else class="chat-messages" ref="chatBody">
      <template v-for="msg in messages" :key="msg.id">
        <!-- Work block: compacted thinking + tools -->
        <details v-if="msg.role === 'work'" class="work-block">
          <summary><PhCpu :size="16" /> {{ msg.systemText }}</summary>
          <ChatMessage
            v-for="(step, si) in (msg.content as any[])"
            :key="si"
            role="assistant"
            :content="step.content"
            :tool-results="step.toolResults"
          />
        </details>

        <div v-else-if="msg.role === 'fork'" class="fork-divider">
          <PhGitFork :size="16" />
          <span>Fork desde:</span>
          <span class="fork-context">{{ msg.systemText }}</span>
        </div>
        <div v-else-if="msg.role !== 'system'" class="msg-row-wrapper" :class="{ selectable: selectionMode, selected: selectedIds.has(msg.id) }" @click="selectionMode && toggleSelectMsg(msg.id)">
          <div v-if="selectionMode" class="msg-check" @click.stop="toggleSelectMsg(msg.id)">
            <span v-if="selectedIds.has(msg.id)">✓</span>
          </div>
          <ChatMessage
            :role="msg.role"
            :content="msg.content"
            :usage="msg.usage || null"
            :tool-results="msg.toolResults"
          />
        </div>
        <div v-else class="sys-divider">
          {{ msg.systemText }}
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.chat-view {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  background: var(--canvas);
}
.chat-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--hairline);
  background: var(--canvas);
  flex-shrink: 0;
}

.ch-toggle-sidebar {
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
  flex-shrink: 0;
  transition: color .15s, background .15s;
}
.ch-toggle-sidebar:hover { color: var(--ink); background: var(--hairline); }
.ch-title {
  font-family: var(--font-display);
  font-size: 18px;
  font-weight: 400;
  letter-spacing: -0.2px;
  color: var(--ink);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-left: 32px;
  user-select:all; 
  cursor:text;
}
.ch-id {
  font-size: 12px;
  color: var(--muted-soft);
  font-family: var(--font-mono);
  user-select: all;
  cursor: pointer;
  transition: color .15s;
  white-space: nowrap;
}
.ch-id:hover { color: var(--primary); }

.ch-select-btn {
  background: none;
  border: 1px solid var(--hairline);
  border-radius: 6px;
  padding: 3px 8px;
  font-size: 12px;
  color: var(--muted);
  cursor: pointer;
  white-space: nowrap;
  transition: color .15s, border-color .15s;
}
.ch-select-btn:hover { color: var(--ink); border-color: var(--primary); }

.ch-export-btn {
  background: var(--primary);
  border: none;
  border-radius: 6px;
  padding: 4px 12px;
  font-size: 12px;
  color: var(--on-primary);
  cursor: pointer;
  white-space: nowrap;
  font-weight: 500;
  transition: background .15s;
}
.ch-export-btn:hover { background: var(--primary-active); }

.ch-parent {
  font-size: 12px;
  color: var(--primary);
  cursor: pointer;
  white-space: nowrap;
  padding: 2px 8px;
  border: 1px solid var(--hairline);
  border-radius: 6px;
  transition: background .15s;
}
.ch-parent:hover { background: rgba(204,120,92,.1); }
.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px 0;
}
.chat-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--muted-soft);
  gap: 12px;
  font-size: 14px;
}

.msg-row-wrapper {
  display: flex;
  align-items: flex-start;
  gap: 0;
}
.msg-row-wrapper.selectable { cursor: pointer }
.msg-row-wrapper.selectable:hover { background: rgba(204,120,92,.03) }
.msg-row-wrapper.selected { background: rgba(204,120,92,.06) }
.msg-row-wrapper :deep(.msg-row) { flex: 1 }

.msg-check {
  width: 22px;
  height: 22px;
  border: 2px solid var(--hairline);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin: 8px 0 0 12px;
  font-size: 12px;
  color: var(--primary);
  cursor: pointer;
  transition: border-color .15s, background .15s;
}
.selected .msg-check { border-color: var(--primary); background: rgba(204,120,92,.1) }

.fork-divider {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 20px;
  font-size: 12px;
  color: var(--primary);
  font-weight: 500;
  max-width: 900px;
  margin: 12px auto;
  border-top: 1px dashed var(--primary-disabled);
  border-bottom: 1px dashed var(--primary-disabled);
  background: rgba(204,120,92,.04);
}
.fork-context {
  color: var(--muted);
  font-weight: 400;
  font-style: italic;
}

.sys-divider {
  text-align: center;
  padding: 16px 20px 8px;
  font-size: 11px;
  color: var(--muted-soft);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.work-block {
  max-width: 900px;
  margin: 4px auto;
  padding: 0 20px;
}
.work-block summary {
  font-size: 12px;
  color: var(--muted);
  cursor: pointer;
  padding: 6px 12px;
  border: 1px dashed var(--hairline);
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  user-select: none;
  transition: border-color .15s;
}
.work-block summary:hover { border-color: var(--primary); color: var(--ink); }
.work-block[open] summary { border-style: solid; border-color: var(--hairline); }
</style>
