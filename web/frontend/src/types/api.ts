export interface Memo {
  id: string
  title: string
  content: string
  type: 'markdown' | 'table'
  table_columns?: TableColumn[]
  row_count?: number
  tags: Tag[]
  created_at: string
  updated_at: string
}

export interface TableColumn {
  id: string
  name: string
}

export interface MemoRow {
  id: string
  memo_id: string
  row_data: Record<string, string>
  sort_order: number
  created_at: string
}

export interface MemoRowsResponse {
  data: MemoRow[]
  total: number
  limit: number
  offset: number
}

export interface Todo {
  id: string
  title: string
  description: string
  status: 'open' | 'done'
  due_date: string | null
  tags: Tag[]
  created_at: string
  updated_at: string
}

export interface CalendarEvent {
  id: string
  title: string
  description: string
  start_at: string
  end_at: string | null
  all_day: boolean
  recurrence_rule: string | null
  recurrence_end: string | null
  tags: Tag[]
  created_at: string
  updated_at: string
}

export interface Tool {
  id: string
  name: string
  url: string
  icon: string
  category: string
  position: number
  created_at: string
  updated_at: string
}

export interface SearchResult {
  type: 'memo' | 'todo' | 'event' | 'tool'
  id: string
  title: string
  snippet: string
  tags: string[]
  updated_at: string
  similarity?: number
}

export interface SearchResponse {
  data: SearchResult[]
  total: number
  suggestions: SearchResult[]
}

export interface SearchParams {
  q: string
  limit?: number
  offset?: number
  type?: string
  tag?: string
  from?: string
  to?: string
}

export interface Tag {
  id: string
  name: string
}

export interface UserSettings {
  first_day_of_week?: number
  widget_order?: string[]
}

export interface User {
  id: string
  email: string
  name: string
  is_admin: boolean
  plan: string
  locale: string
  user_settings?: UserSettings
}

export interface ApiKey {
  id: string
  name: string
  key?: string
  created_at: string
  last_used_at: string | null
}

export interface ChatSession {
  id: string
  title: string
  created_at: string
  updated_at: string
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant' | 'system' | 'tool'
  content: string
  tool_calls?: { id: string; name: string; arguments: string }[]
  tool_call_id?: string
  created_at: string
}

export interface ChatSessionDetail extends ChatSession {
  messages: ChatMessage[]
}

export interface AIChatConfig {
  provider: string
  openai_key_masked?: string
  anthropic_key_masked?: string
  openai_key?: string
  anthropic_key?: string
}

export interface ListResponse<T> {
  data: T[]
  total: number
  limit: number
  offset: number
}

export interface SingleResponse<T> {
  data: T
}
