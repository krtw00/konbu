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
  createEvent: (body: { title: string; description: string; start_at: string; end_at: string | null; all_day: boolean; tags: string[] }) =>
    request<{ data: import('@/types/api').CalendarEvent }>('POST', '/events', body),
  updateEvent: (id: string, body: { title: string; description: string; start_at: string; end_at: string | null; all_day: boolean; tags: string[] }) =>
    request<{ data: import('@/types/api').CalendarEvent }>('PUT', `/events/${id}`, body),
  deleteEvent: (id: string) => request<null>('DELETE', `/events/${id}`),

  // Tools
  listTools: () => request<{ data: import('@/types/api').Tool[] }>('GET', '/tools'),
  createTool: (body: { name: string; url: string; icon: string }) =>
    request<{ data: import('@/types/api').Tool }>('POST', '/tools', body),
  deleteTool: (id: string) => request<null>('DELETE', `/tools/${id}`),

  // Tags
  listTags: () => request<{ data: import('@/types/api').Tag[] }>('GET', '/tags'),
}
