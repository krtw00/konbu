import { create } from 'zustand'
import { api, apiFetch } from '@/lib/api'
import type { ChatSession, ChatMessage } from '@/types/api'

interface ChatState {
  isOpen: boolean
  sessions: ChatSession[]
  currentSessionId: string | null
  messages: ChatMessage[]
  isStreaming: boolean
  streamingContent: string
  toolStatus: string | null
  toggle: () => void
  close: () => void
  loadSessions: () => Promise<void>
  selectSession: (id: string) => Promise<void>
  sendMessage: (content: string) => Promise<void>
  newSession: () => void
  deleteSession: (id: string) => Promise<void>
}

export const useChatStore = create<ChatState>((set, get) => ({
  isOpen: false,
  sessions: [],
  currentSessionId: null,
  messages: [],
  isStreaming: false,
  streamingContent: '',
  toolStatus: null,

  toggle: () => set(s => ({ isOpen: !s.isOpen })),
  close: () => set({ isOpen: false }),

  loadSessions: async () => {
    try {
      const r = await api.listChatSessions()
      set({ sessions: r.data || [] })
    } catch { /* ignore */ }
  },

  selectSession: async (id: string) => {
    set({ currentSessionId: id, messages: [] })
    try {
      const r = await api.getChatSession(id)
      set({ messages: r.data?.messages || [] })
    } catch { /* ignore */ }
  },

  newSession: () => {
    set({ currentSessionId: null, messages: [] })
  },

  deleteSession: async (id: string) => {
    await api.deleteChatSession(id)
    const { currentSessionId, loadSessions } = get()
    if (currentSessionId === id) {
      set({ currentSessionId: null, messages: [] })
    }
    await loadSessions()
  },

  sendMessage: async (content: string) => {
    const { currentSessionId, messages, loadSessions } = get()

    const userMsg: ChatMessage = {
      id: crypto.randomUUID(),
      role: 'user',
      content,
      created_at: new Date().toISOString(),
    }
    set({ messages: [...messages, userMsg], isStreaming: true, streamingContent: '', toolStatus: null })

    try {
      let sessionId = currentSessionId
      if (!sessionId) {
        const r = await api.createChatSession()
        sessionId = r.data.id
        set({ currentSessionId: sessionId })
      }

      const res = await apiFetch(`/chat/sessions/${sessionId}/messages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content }),
      })

      if (!res.ok || !res.body) {
        const err = await res.json().catch(() => ({ error: { message: 'Unknown error' } }))
        const errMsg: ChatMessage = {
          id: crypto.randomUUID(),
          role: 'assistant',
          content: err.error?.message || 'Error',
          created_at: new Date().toISOString(),
        }
        set(s => ({ messages: [...s.messages, errMsg], isStreaming: false }))
        return
      }

      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      let fullContent = ''
      let currentEvent = ''
      let errorOccurred = false

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          if (line.startsWith('event: ')) {
            currentEvent = line.slice(7).trim()
          } else if (line.startsWith('data: ')) {
            const data = line.slice(6)
            try {
              const parsed = JSON.parse(data)

              if (currentEvent === 'error') {
                const errMsg: ChatMessage = {
                  id: crypto.randomUUID(),
                  role: 'assistant',
                  content: parsed.message || 'Error',
                  created_at: new Date().toISOString(),
                }
                set(s => ({ messages: [...s.messages, errMsg], isStreaming: false, streamingContent: '', toolStatus: null }))
                errorOccurred = true
              } else if (currentEvent === 'text_delta') {
                if (parsed.content !== undefined) {
                  fullContent += parsed.content
                  set({ streamingContent: fullContent })
                }
              } else if (currentEvent === 'tool_call') {
                if (parsed.tool_name) set({ toolStatus: parsed.tool_name })
              } else if (currentEvent === 'tool_result') {
                set({ toolStatus: null })
              }
            } catch { /* not json, ignore */ }
            currentEvent = ''
          }
        }
      }

      if (errorOccurred) return

      const assistantMsg: ChatMessage = {
        id: crypto.randomUUID(),
        role: 'assistant',
        content: fullContent,
        created_at: new Date().toISOString(),
      }
      set(s => ({
        messages: [...s.messages, assistantMsg],
        isStreaming: false,
        streamingContent: '',
        toolStatus: null,
      }))
      // Reload messages from DB to get the full history including tool results
      const sessionId2 = get().currentSessionId
      if (sessionId2) {
        try {
          const r = await api.getChatSession(sessionId2)
          set({ messages: r.data?.messages || [] })
        } catch { /* ignore */ }
      }
      await loadSessions()
    } catch {
      set({ isStreaming: false, streamingContent: '', toolStatus: null })
    }
  },
}))
