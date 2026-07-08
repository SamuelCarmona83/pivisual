<script setup lang="ts">
import { computed } from 'vue'
import { PhWrench, PhBrain, PhUser } from '@phosphor-icons/vue'
import { marked } from 'marked'
import type { ContentBlock, MessageUsage } from '../types'

interface ToolGroup { name: string; calls: ContentBlock[] }

const props = defineProps<{
  role: string
  content: string | ContentBlock[]
  usage?: MessageUsage | null
  toolResults: Record<string, string>
}>()

function esc(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}

function toolSummary(tc: any): string {
  const args = tc.arguments || tc.input || {}
  return args.command || args.path || ''
}

function renderToolResult(text: string): string {
  if (!text) return text
  text = text.replace(/\x1b\[[0-9;]*[a-zA-Z]/g, '')
  text = text.replace(/\r[^\n]*\r/g, '\n').replace(/\r/g, '')
  const isDiff = /^diff --git/m.test(text) || /^@@ /m.test(text)
  if (!isDiff) return esc(text)
  return text.split('\n').map(line => {
    if (/^diff --git/.test(line) || /^(index |--- |\+\+\+ )/.test(line) || /^(new|old|similarity|rename|copy|deleted) /.test(line))
      return `<span class="diff-meta">${esc(line)}</span>`
    if (/^@@/.test(line)) return `<span class="diff-hunk">${esc(line)}</span>`
    if (/^\+/.test(line)) return `<span class="diff-add">${esc(line)}</span>`
    if (/^-/.test(line)) return `<span class="diff-del">${esc(line)}</span>`
    return esc(line)
  }).join('\n')
}


const parsedContent = computed(() => {
  if (typeof props.content === 'string') return { text: props.content, thinking: null as string|null, toolGroups: [] as ToolGroup[] }
  const textParts: string[] = []
  const thinkingPieces: string[] = []
  const rawCalls: ContentBlock[] = []
  for (const block of props.content) {
    if (block.type === 'text' && block.text) textParts.push(block.text)
    else if (block.type === 'thinking' && block.thinking) thinkingPieces.push(block.thinking)
    else if (block.type === 'toolCall') rawCalls.push(block)
  }
  // Group consecutive tool calls by name
  const toolGroups: ToolGroup[] = []
  for (const tc of rawCalls) {
    const last = toolGroups[toolGroups.length - 1]
    if (last && last.name === tc.name) {
      last.calls.push(tc)
    } else {
      toolGroups.push({ name: tc.name || 'tool', calls: [tc] })
    }
  }
  return { text: textParts.join(''), thinking: thinkingPieces.join('\n'), toolGroups }
})

const renderedText = computed(() => {
  return marked.parse(parsedContent.value.text) as string
})

const costStr = computed(() => {
  if (!props.usage?.cost?.total || props.usage.cost.total <= 0) return ''
  return `$${props.usage.cost.total.toFixed(4)}`
})

const tkStr = computed(() => {
  if (!props.usage || (!props.usage.input && !props.usage.output)) return ''
  return `${props.usage.input}+${props.usage.output || 0} tk`
})
</script>

<template>
  <div v-if="role === 'system'" class="sys-event">
    <slot />
  </div>

  <div v-else class="msg-row" :class="role">
    <div v-if="role === 'user'" class="avatar">
      <PhUser :size="20" />
    </div>
    <div class="msg-body">
      <div v-if="parsedContent.text" class="bubble" v-html="renderedText" />

      <!-- Thinking block -->
      <details v-if="parsedContent.thinking" class="think-block">
        <summary><PhBrain :size="14" /> Thinking</summary>
        <div class="think-content">{{ parsedContent.thinking }}</div>
      </details>

      <!-- Tool calls -->
      <template v-if="parsedContent.toolGroups.length">
        <details v-for="(group, gi) in parsedContent.toolGroups" :key="gi" class="tool-block">
          <summary>
            <PhWrench :size="14" />
            <span class="tool-name">{{ group.calls.length > 1 ? group.calls.length + '× ' : '' }}{{ group.name }}</span>
            <span class="tool-arg" v-if="group.calls.length === 1">{{ toolSummary(group.calls[0]) }}</span>
          </summary>
          <template v-if="group.calls.length === 1">
            <div class="tool-result" v-html="renderToolResult(toolResults[group.calls[0].id || ''] || '')" />
          </template>
          <template v-else>
            <details v-for="tc in group.calls" :key="tc.id" class="tool-sub">
              <summary class="tool-sub-summary">
                <span class="tool-arg">{{ toolSummary(tc) || tc.id?.slice(0, 8) }}</span>
              </summary>
              <div class="tool-result" v-html="renderToolResult(toolResults[tc.id || ''] || '')" />
            </details>
          </template>
        </details>
      </template>

      <div v-if="tkStr || costStr" class="msg-footer">
        <span v-if="tkStr">{{ tkStr }}</span>
        <span v-if="costStr">{{ costStr }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.msg-row {
  display: flex;
  gap: 10px;
  padding: 8px 20px;
  max-width: 900px;
  margin: 0 auto;
}
.msg-row.user {
  flex-direction: row-reverse;
}
.avatar {
  width: 32px;
  height: 32px;
  border-radius: 32px;
  background: var(--primary-disabled);
  color: var(--on-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}
.msg-body {
  flex: 1;
  min-width: 0;
}
.msg-row.assistant .avatar { display: none; }

.bubble {
  background: var(--surface-card);
  border-radius: 12px;
  padding: 10px 16px;
  font-size: 14px;
  line-height: 1.6;
  word-break: break-word;
}
.msg-row.user .bubble {
  background: var(--primary);
  color: var(--on-primary);
  border-radius: 12px 12px 4px 12px;
}

/* Markdown inside bubbles */
.bubble :deep(h1), .bubble :deep(h2), .bubble :deep(h3) {
  font-family: var(--font-display);
  font-weight: 400;
  letter-spacing: -0.2px;
  margin: 16px 0 8px;
}
.bubble :deep(h1) { font-size: 22px; }
.bubble :deep(h2) { font-size: 18px; }
.bubble :deep(h3) { font-size: 16px; }
.msg-row.user .bubble :deep(h1),
.msg-row.user .bubble :deep(h2),
.msg-row.user .bubble :deep(h3) { color: var(--on-primary); }

.bubble :deep(ul), .bubble :deep(ol) { padding-left: 20px; margin: 8px 0; }
.bubble :deep(li) { margin: 2px 0; }
.bubble :deep(blockquote) {
  border-left: 3px solid var(--primary);
  margin: 8px 0;
  padding: 4px 0 4px 12px;
  color: var(--muted);
  font-style: italic;
}
.msg-row.user .bubble :deep(blockquote) {
  border-left-color: rgba(255,255,255,.4);
  color: rgba(255,255,255,.8);
}
.bubble :deep(a) { color: var(--primary); }
.msg-row.user .bubble :deep(a) { color: var(--on-primary); }
.bubble :deep(table) { border-collapse: collapse; width: 100%; margin: 8px 0; font-size: 13px; }
.bubble :deep(th), .bubble :deep(td) { border: 1px solid var(--hairline); padding: 6px 10px; text-align: left; }
.bubble :deep(th) { background: var(--surface-card); font-weight: 500; }
.bubble :deep(hr) { border: none; border-top: 1px solid var(--hairline); margin: 12px 0; }
.bubble :deep(img) { max-width: 100%; border-radius: 4px; }
.bubble :deep(p) { margin: 0 0 8px; }
.bubble :deep(p:last-child), .bubble :deep(p:only-child) { margin-bottom: 0; }

.bubble :deep(code) {
  font-family: var(--font-mono);
  font-size: 13px;
  padding: 2px 5px;
  border-radius: 4px;
}
.msg-row.user .bubble :deep(code) { background: rgba(255,255,255,.15); }
.msg-row:not(.user) .bubble :deep(code) { background: rgba(0,0,0,.06); color: var(--primary-active); }

.bubble :deep(pre) {
  font-family: var(--font-mono);
  font-size: 13px;
  overflow-x: auto;
  margin: 8px 0;
  white-space: pre;
  padding: 12px;
  border-radius: 6px;
  line-height: 1.6;
}
.msg-row.user .bubble :deep(pre) { background: rgba(0,0,0,.2); color: rgba(255,255,255,.85); }
.msg-row:not(.user) .bubble :deep(pre) { background: var(--surface-dark-soft); color: var(--on-dark-soft); }

/* Tool blocks */
.tool-block { margin: 8px 0; }
.think-block { margin: 8px 0; }
.think-block summary {
  font-size: 12px; color: var(--muted); cursor: pointer;
  padding: 6px 10px; background: rgba(0,0,0,.04); border-radius: 6px;
  display: inline-flex; align-items: center; gap: 6px; user-select: none;
}
.msg-row.user .think-block summary { background: rgba(255,255,255,.1); color: rgba(255,255,255,.85); }
.think-content {
  font-size: 13px; color: var(--muted); padding: 8px 10px;
  white-space: pre-wrap; font-style: italic; line-height: 1.5;
}
.msg-row.user .think-content { color: rgba(255,255,255,.7); }
.tool-block { margin: 8px 0; }
.tool-block summary {
  font-size: 12px;
  color: var(--muted);
  cursor: pointer;
  padding: 6px 10px;
  background: rgba(0,0,0,.04);
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  user-select: none;
}
.msg-row.user .tool-block summary { background: rgba(255,255,255,.1); color: rgba(255,255,255,.85); }
.tool-name { font-weight: 500 }
.tool-arg {
  font-family: var(--font-mono); font-size: 11px; opacity: .7;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 400px;
}
.tool-sub { margin: 2px 0 }
.tool-sub-summary {
  font-size: 11px; color: var(--muted-soft); cursor: pointer;
  padding: 3px 8px; font-family: var(--font-mono);
  list-style: none;
}
.tool-sub-summary::-webkit-details-marker { display: none }
.tool-sub-summary::before { content: '▸ '; font-size: 10px; color: var(--muted-soft) }
.tool-sub[open] .tool-sub-summary::before { content: '▾ ' }
.msg-row.user .tool-sub-summary { color: rgba(255,255,255,.6) }
.tool-result {
  font-size: 13px;
  padding: 8px 10px;
  margin-top: 4px;
  background: var(--surface-dark-soft);
  border-radius: 6px;
  white-space: pre-wrap;
  max-height: 260px;
  overflow-y: auto;
  font-family: var(--font-mono);
  line-height: 1.6;
  color: var(--on-dark-soft);
  tab-size: 4;
}
.msg-row.user .tool-result { background: rgba(0,0,0,.15); color: rgba(255,255,255,.7); }

/* Diff highlights */
.tool-result :deep(.diff-add) { color: #7eb77f; display: block; }
.tool-result :deep(.diff-del) { color: #e06c75; display: block; }
.tool-result :deep(.diff-hunk) { color: #61afef; display: block; font-weight: 500; }
.tool-result :deep(.diff-meta) { color: #c678dd; display: block; font-weight: 500; }
.msg-row.user .tool-result :deep(.diff-add) { color: #a3d9a5; }
.msg-row.user .tool-result :deep(.diff-del) { color: #f5a0a5; }
.msg-row.user .tool-result :deep(.diff-hunk) { color: #96cbfa; }
.msg-row.user .tool-result :deep(.diff-meta) { color: #d9a8ef; }

/* System events */
.sys-event {
  text-align: center;
  padding: 16px 20px 8px;
  font-size: 11px;
  color: var(--muted-soft);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.msg-footer {
  font-size: 11px;
  color: var(--muted);
  margin-top: 4px;
  display: flex;
  gap: 8px;
}
.msg-row.user .msg-footer { text-align: right; }
</style>
