import { create } from 'zustand'

type Page = 'home' | 'memos' | 'memo-edit' | 'todos' | 'calendar' | 'tools'

interface AppState {
  currentPage: Page
  commandOpen: boolean
  setPage: (page: Page) => void
  setCommandOpen: (open: boolean) => void
}

export const useAppStore = create<AppState>((set) => ({
  currentPage: 'home',
  commandOpen: false,
  setPage: (page) => set({ currentPage: page }),
  setCommandOpen: (open) => set({ commandOpen: open }),
}))
