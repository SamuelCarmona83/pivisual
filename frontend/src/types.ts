export interface SessionSummary {
  id: string
  file: string
  project: string
  timestamp: string
  name: string | null
  first_user: string
  model: string
  provider: string
  user_msgs: number
  assistant_msgs: number
  total_tokens: number
  total_cost: number
  parent_session?: string
}

export interface SessionEntry {
  type: string
  id?: string
  parentId?: string | null
  timestamp?: string
  message?: ChatMessage
  [key: string]: any
}

export interface ChatMessage {
  role: 'user' | 'assistant' | 'toolResult'
  content: string | ContentBlock[]
  model?: string
  provider?: string
  usage?: MessageUsage
  toolCallId?: string
  [key: string]: any
}

export interface ContentBlock {
  type: 'text' | 'toolCall' | 'thinking'
  text?: string
  id?: string
  name?: string
  input?: Record<string, any>
  thinking?: string
  signature?: string
}

export interface MessageUsage {
  input: number
  output: number
  cost: {
    total: number
    [key: string]: number
  }
}
