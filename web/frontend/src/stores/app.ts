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
