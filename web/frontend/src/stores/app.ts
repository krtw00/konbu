import { create } from 'zustand'
import type { User } from '@/types/api'
import { api } from '@/lib/api'

type Page = 'home' | 'memos' | 'memo-edit' | 'todos' | 'calendar' | 'tools' | 'settings'

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
      const setup = await api.setupStatus()
      if (setup.data.needs_setup) {
        set({ needsSetup: true, isLoading: false })
        return
      }
      set({ openRegistration: setup.data.open_registration })
      try {
        const providers = await api.providers()
        set({ googleAuth: providers.data.google })
      } catch { /* ignore */ }
      const me = await api.getMe()
      set({ user: me.data, isAuthenticated: true, isLoading: false, needsSetup: false })
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
