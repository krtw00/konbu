const BASE = '/api/v1'

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const opts: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json' },
  }
  if (body) opts.body = JSON.stringify(body)
  const res = await fetch(BASE + path, opts)
  if (res.status === 204) return null as T
  const data = await res.json()
  if (!res.ok) throw new Error(data.error?.message || 'Request failed')
  return data
}

export const api = {
  // Memos
  listMemos: (limit = 100) => request<{ data: import('@/types/api').Memo[] }>('GET', `/memos?limit=${limit}`),
  getMemo: (id: string) => request<{ data: import('@/types/api').Memo }>('GET', `/memos/${id}`),
  createMemo: (body: { title: string; type: string; content: string; tags: string[] }) =>
    request<{ data: import('@/types/api').Memo }>('POST', '/memos', body),
  updateMemo: (id: string, body: { title: string; content: string; tags: string[] }) =>
    request<{ data: import('@/types/api').Memo }>('PUT', `/memos/${id}`, body),
  deleteMemo: (id: string) => request<null>('DELETE', `/memos/${id}`),

  // Todos
  listTodos: (limit = 100) => request<{ data: import('@/types/api').Todo[] }>('GET', `/todos?limit=${limit}`),
  getTodo: (id: string) => request<{ data: import('@/types/api').Todo }>('GET', `/todos/${id}`),
  createTodo: (body: { title: string; due_date?: string | null; tags: string[] }) =>
    request<{ data: import('@/types/api').Todo }>('POST', '/todos', body),
  updateTodo: (id: string, body: { title: string; description: string; status: string; due_date: string | null; tags: string[] }) =>
    request<{ data: import('@/types/api').Todo }>('PUT', `/todos/${id}`, body),
  doneTodo: (id: string) => request<null>('PATCH', `/todos/${id}/done`),
  reopenTodo: (id: string) => request<null>('PATCH', `/todos/${id}/reopen`),
  deleteTodo: (id: string) => request<null>('DELETE', `/todos/${id}`),

  // Events
  listEvents: (limit = 100) => request<{ data: import('@/types/api').CalendarEvent[] }>('GET', `/events?limit=${limit}&sort=start_at:asc`),
  getEvent: (id: string) => request<{ data: import('@/types/api').CalendarEvent }>('GET', `/events/${id}`),
  createEvent: (body: { title: string; description: string; start_at: string; end_at: string | null; all_day: boolean; recurrence_rule?: string | null; recurrence_end?: string | null; tags: string[] }) =>
    request<{ data: import('@/types/api').CalendarEvent }>('POST', '/events', body),
  updateEvent: (id: string, body: { title: string; description: string; start_at: string; end_at: string | null; all_day: boolean; recurrence_rule?: string | null; recurrence_end?: string | null; tags: string[] }) =>
    request<{ data: import('@/types/api').CalendarEvent }>('PUT', `/events/${id}`, body),
  deleteEvent: (id: string) => request<null>('DELETE', `/events/${id}`),

  // Tools
  listTools: () => request<{ data: import('@/types/api').Tool[] }>('GET', '/tools'),
  createTool: (body: { name: string; url: string; category?: string }) =>
    request<{ data: import('@/types/api').Tool }>('POST', '/tools', body),
  updateTool: (id: string, body: { name: string; url: string; category?: string }) =>
    request<{ data: import('@/types/api').Tool }>('PUT', `/tools/${id}`, body),
  deleteTool: (id: string) => request<null>('DELETE', `/tools/${id}`),
  fetchFavicon: (url: string) => request<{ data: { icon: string } }>('GET', `/tools/favicon?url=${encodeURIComponent(url)}`),
  healthCheckTools: () => request<{ data: { id: string; url: string; alive: boolean; status: number }[] }>('POST', '/tools/health-check'),

  // Tags
  listTags: () => request<{ data: import('@/types/api').Tag[] }>('GET', '/tags'),

  // Search
  search: (q: string, limit = 20) => request<{ data: import('@/types/api').SearchResult[] }>('GET', `/search?q=${encodeURIComponent(q)}&limit=${limit}`),

  // Auth
  setupStatus: () => request<{ data: { needs_setup: boolean; user_count: number; open_registration: boolean } }>('GET', '/auth/setup-status'),
  providers: () => request<{ data: { google: boolean } }>('GET', '/auth/providers'),
  register: (body: { email: string; password: string; name: string }) =>
    request<{ data: import('@/types/api').User }>('POST', '/auth/register', body),
  login: (body: { email: string; password: string }) =>
    request<{ data: import('@/types/api').User }>('POST', '/auth/login', body),
  logout: () => request<null>('POST', '/auth/logout'),
  getMe: () => request<{ data: import('@/types/api').User }>('GET', '/auth/me'),
  updateMe: (body: { name: string }) =>
    request<{ data: import('@/types/api').User }>('PUT', '/auth/me', body),
  changePassword: (body: { old_password: string; new_password: string }) =>
    request<null>('POST', '/auth/change-password', body),
  deleteAccount: (body: { password: string }) =>
    request<null>('POST', '/auth/delete-account', body),
  getSettings: () => request<{ data: import('@/types/api').UserSettings }>('GET', '/auth/settings'),
  updateSettings: (body: import('@/types/api').UserSettings) =>
    request<{ data: import('@/types/api').UserSettings }>('PUT', '/auth/settings', body),

  // API Keys
  listApiKeys: () => request<{ data: import('@/types/api').ApiKey[] }>('GET', '/api-keys'),
  createApiKey: (body: { name: string }) =>
    request<{ data: import('@/types/api').ApiKey }>('POST', '/api-keys', body),
  deleteApiKey: (id: string) => request<null>('DELETE', `/api-keys/${id}`),

  // Chat
  listChatSessions: () => request<{ data: import('@/types/api').ChatSession[] }>('GET', '/chat/sessions'),
  createChatSession: () => request<{ data: import('@/types/api').ChatSession }>('POST', '/chat/sessions'),
  getChatSession: (id: string) => request<{ data: import('@/types/api').ChatSessionDetail }>('GET', `/chat/sessions/${id}`),
  updateChatSession: (id: string, body: { title: string }) => request<null>('PUT', `/chat/sessions/${id}`, body),
  deleteChatSession: (id: string) => request<null>('DELETE', `/chat/sessions/${id}`),
  getChatConfig: () => request<{ data: import('@/types/api').AIChatConfig }>('GET', '/chat/config'),
  updateChatConfig: (body: import('@/types/api').AIChatConfig) => request<{ data: import('@/types/api').AIChatConfig }>('PUT', '/chat/config', body),
}
