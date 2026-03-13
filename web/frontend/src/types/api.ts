export interface Memo {
  id: string
  title: string
  content: string
  type: 'markdown' | 'table'
  tags: Tag[]
  created_at: string
  updated_at: string
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
  type: 'memo' | 'todo' | 'event'
  id: string
  title: string
  snippet: string
  updated_at: string
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

export interface ListResponse<T> {
  data: T[]
  total: number
  limit: number
  offset: number
}

export interface SingleResponse<T> {
  data: T
}
