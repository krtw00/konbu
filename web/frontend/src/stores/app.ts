import { create } from 'zustand'
import type { User, CalendarEvent, Todo } from '@/types/api'
import { api } from '@/lib/api'
import { prefetchCache } from '@/hooks/useCache'

const DEFAULT_ORDER = ['schedule', 'todos', 'memos']

function prefetchHomeData() {
  prefetchCache('home', () => Promise.all([
    api.listEvents(10),
    api.listTodos(100),
    api.listMemos(6),
    api.getSettings().catch(() => null),
  ]).then(([evR, tdR, mmR, sR]) => {
    const today = new Date().toDateString()
    return {
      events: (evR.data || []).filter((e: CalendarEvent) => new Date(e.start_at).toDateString() === today),
      todos: (tdR.data || []).filter((t: Todo) => t.status === 'open'),
      memos: mmR.data || [],
      widgetOrder: sR?.data?.widget_order || DEFAULT_ORDER,
    }
  }))
  prefetchCache('memos', () => Promise.all([api.listMemos(), api.listTags()]).then(([r, t]) => ({
    memos: r.data || [],
    tags: (t.data || []).map((tag: { name: string }) => tag.name),
  })))
  prefetchCache('todos', () => Promise.all([api.listTodos(), api.listTags()]).then(([r, tRes]) => ({
    todos: r.data || [],
    tags: (tRes.data || []).map((tag: { name: string }) => tag.name),
  })))
}

type Page = 'home' | 'memos' | 'memo-edit' | 'todos' | 'calendar' | 'tools' | 'chat' | 'settings'

interface AppState {
  currentPage: Page
  commandOpen: boolean
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  needsSetup: boolean
  openRegistration: boolean
  googleAuth: boolean
  setPage: (page: Page) => void
  setCommandOpen: (open: boolean) => void
  setUser: (user: User) => void
  clearUser: () => void
  checkAuth: () => Promise<void>
  logout: () => Promise<void>
}

export const useAppStore = create<AppState>((set) => ({
  currentPage: 'home',
  commandOpen: false,
  user: null,
  isAuthenticated: false,
  isLoading: true,
  needsSetup: false,
  openRegistration: false,
  googleAuth: false,
  setPage: (page) => set({ currentPage: page }),
  setCommandOpen: (open) => set({ commandOpen: open }),
  setUser: (user) => set({ user, isAuthenticated: true }),
  clearUser: () => set({ user: null, isAuthenticated: false }),
  checkAuth: async () => {
    try {
      const [setup, providers, me] = await Promise.all([
        api.setupStatus(),
        api.providers().catch(() => ({ data: { google: false } })),
        api.getMe().catch(() => null),
      ])
      if (setup.data.needs_setup) {
        set({ needsSetup: true, isLoading: false, openRegistration: setup.data.open_registration, googleAuth: providers.data.google })
        return
      }
      if (me) {
        set({ user: me.data, isAuthenticated: true, isLoading: false, needsSetup: false, openRegistration: setup.data.open_registration, googleAuth: providers.data.google })
        prefetchHomeData()
      } else {
        set({ user: null, isAuthenticated: false, isLoading: false, openRegistration: setup.data.open_registration, googleAuth: providers.data.google })
      }
    } catch {
      set({ user: null, isAuthenticated: false, isLoading: false })
    }
  },
  logout: async () => {
    try {
      await api.logout()
    } catch {
      // ignore
    }
    set({ user: null, isAuthenticated: false, currentPage: 'home' })
  },
}))
