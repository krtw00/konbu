import { useAppStore } from '@/stores/app'
import { Home, FileText, CheckSquare, Calendar, Monitor, Search } from 'lucide-react'

const navItems = [
  { page: 'home' as const, icon: Home, label: 'Home' },
  { page: 'memos' as const, icon: FileText, label: 'Memos' },
  { page: 'todos' as const, icon: CheckSquare, label: 'ToDo' },
  { page: 'calendar' as const, icon: Calendar, label: 'Calendar' },
  { page: 'tools' as const, icon: Monitor, label: 'Tools' },
]

export function Sidebar() {
  const { currentPage, setPage, setCommandOpen } = useAppStore()

  return (
    <nav className="hidden md:flex flex-col w-52 border-r border-border bg-sidebar h-screen sticky top-0">
      <div className="flex items-center gap-2 px-4 py-4">
        <img className="w-6 h-6" src="/static/favicon.svg" alt="K" />
        <span className="font-semibold text-sm text-sidebar-foreground">konbu</span>
      </div>
      <div className="flex-1 flex flex-col gap-0.5 px-2">
        {navItems.map(({ page, icon: Icon, label }) => (
          <button
            key={page}
            onClick={() => setPage(page)}
            className={`flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors ${
              currentPage === page || (currentPage === 'memo-edit' && page === 'memos')
                ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
                : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50'
            }`}
          >
            <Icon size={18} />
            <span>{label}</span>
          </button>
        ))}
      </div>
      <div className="px-2 pb-4">
        <button
          onClick={() => setCommandOpen(true)}
          className="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm text-sidebar-foreground/70 hover:bg-sidebar-accent/50 w-full"
        >
          <Search size={18} />
          <span>Search</span>
          <kbd className="ml-auto text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
            {navigator.platform.includes('Mac') ? '⌘' : 'Ctrl'}K
          </kbd>
        </button>
      </div>
    </nav>
  )
}
