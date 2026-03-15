import { create } from 'zustand'
import { api } from '@/lib/api'
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

      const res = await fetch(`/api/v1/chat/sessions/${sessionId}/messages`, {
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

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          if (line.startsWith('event: ')) {
            const eventType = line.slice(7).trim()
            if (eventType === 'done') {
              // handled below
            }
          } else if (line.startsWith('data: ')) {
            const data = line.slice(6)
            try {
              const parsed = JSON.parse(data)
              if (parsed.content !== undefined) {
                fullContent += parsed.content
                set({ streamingContent: fullContent })
              }
              if (parsed.tool_name) {
                set({ toolStatus: parsed.tool_name })
              }
              if (parsed.tool_result !== undefined) {
                set({ toolStatus: null })
              }
              if (parsed.done) {
                // final
              }
            } catch { /* not json, ignore */ }
          }
        }
      }

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
      await loadSessions()
    } catch {
      set({ isStreaming: false, streamingContent: '', toolStatus: null })
    }
  },
}))
