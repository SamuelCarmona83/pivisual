<script setup lang="ts">
import { computed, watch, nextTick, ref } from 'vue'
import { PhChats } from '@phosphor-icons/vue'
import { useSessions } from '../composables/useSessions'
import ChatMessage from './ChatMessage.vue'
import type { SessionEntry, ChatMessage as ChatMsg } from '../types'

const { selectedSession, sessionLines } = useSessions()

function copySessionId() {
  if (!selectedSession.value) return
  navigator.clipboard.writeText('pi --session ' + selectedSession.value.id).catch(() => {})
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

  // Second pass: build message list
  const result: RenderedMessage[] = []
  for (const entry of entries) {
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
      <span class="ch-title" :title="selectedSession?.name || selectedSession?.first_user || 'Seleccioná una sesión'">
        {{ selectedSession?.name || selectedSession?.first_user || 'Seleccioná una sesión' }}
      </span>
      <span v-if="selectedSession" class="ch-id" :title="'pi --session ' + selectedSession.id" @click="copySessionId">
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
          <summary>🔨 {{ msg.systemText }}</summary>
          <ChatMessage
            v-for="(step, si) in (msg.content as any[])"
            :key="si"
            role="assistant"
            :content="step.content"
            :tool-results="step.toolResults"
          />
        </details>

        <ChatMessage
          v-else-if="msg.role !== 'system'"
          :role="msg.role"
          :content="msg.content"
          :usage="msg.usage || null"
          :tool-results="msg.toolResults"
        />
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
  gap: 12px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--hairline);
  background: var(--canvas);
  flex-shrink: 0;
}
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
