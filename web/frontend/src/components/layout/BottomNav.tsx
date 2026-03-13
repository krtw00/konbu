import { useAppStore } from '@/stores/app'
import { Home, FileText, CheckSquare, Calendar, Monitor } from 'lucide-react'

const navItems = [
  { page: 'home' as const, icon: Home, label: 'Home' },
  { page: 'memos' as const, icon: FileText, label: 'Memo' },
  { page: 'todos' as const, icon: CheckSquare, label: 'ToDo' },
  { page: 'calendar' as const, icon: Calendar, label: 'Cal' },
  { page: 'tools' as const, icon: Monitor, label: 'Tools' },
]

export function BottomNav() {
  const { currentPage, setPage } = useAppStore()

  return (
    <nav className="md:hidden fixed bottom-0 left-0 right-0 flex border-t border-border bg-background z-50">
      {navItems.map(({ page, icon: Icon, label }) => (
        <button
          key={page}
          onClick={() => setPage(page)}
          className={`flex-1 flex flex-col items-center gap-0.5 py-2 text-xs transition-colors ${
            currentPage === page || (currentPage === 'memo-edit' && page === 'memos')
              ? 'text-primary'
              : 'text-muted-foreground'
          }`}
        >
          <Icon size={20} />
          <span>{label}</span>
        </button>
      ))}
    </nav>
  )
}
