import { create } from 'zustand'

type Page = 'home' | 'memos' | 'memo-edit' | 'todos' | 'calendar' | 'tools'

interface AppState {
  currentPage: Page
  theme: string
  commandOpen: boolean
  setPage: (page: Page) => void
  setTheme: (theme: string) => void
  setCommandOpen: (open: boolean) => void
}

export const useAppStore = create<AppState>((set) => ({
  currentPage: 'home',
  theme: localStorage.getItem('konbu-theme') || 'konbu',
  commandOpen: false,
  setPage: (page) => set({ currentPage: page }),
  setTheme: (theme) => {
    localStorage.setItem('konbu-theme', theme)
    set({ theme })
  },
  setCommandOpen: (open) => set({ commandOpen: open }),
}))
