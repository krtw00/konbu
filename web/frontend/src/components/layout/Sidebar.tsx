import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'
import { Home, FileText, CheckSquare, Calendar, Monitor, Search, Settings, LogOut } from 'lucide-react'

const navItems = [
  { page: 'home' as const, icon: Home, labelKey: 'nav.home' },
  { page: 'memos' as const, icon: FileText, labelKey: 'nav.memos' },
  { page: 'todos' as const, icon: CheckSquare, labelKey: 'nav.todo' },
  { page: 'calendar' as const, icon: Calendar, labelKey: 'nav.calendar' },
  { page: 'tools' as const, icon: Monitor, labelKey: 'nav.tools' },
]

export function Sidebar() {
  const { t } = useTranslation()
  const { currentPage, setPage, setCommandOpen, logout } = useAppStore()

  return (
    <nav className="hidden md:flex flex-col w-52 border-r border-border bg-sidebar h-screen sticky top-0">
      <div className="flex items-center gap-2 px-4 py-4">
        <img className="w-6 h-6" src="/static/favicon.svg" alt="K" />
        <span className="font-semibold text-sm text-sidebar-foreground">konbu</span>
      </div>
      <div className="flex-1 flex flex-col gap-0.5 px-2">
        {navItems.map(({ page, icon: Icon, labelKey }) => (
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
            <span>{t(labelKey)}</span>
          </button>
        ))}
      </div>
      <div className="px-2 pb-4 flex flex-col gap-0.5">
        <button
          onClick={() => setCommandOpen(true)}
          className="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm text-sidebar-foreground/70 hover:bg-sidebar-accent/50 w-full"
        >
          <Search size={18} />
          <span>{t('common.search')}</span>
          <kbd className="ml-auto text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
            {navigator.platform.includes('Mac') ? '\u2318' : 'Ctrl'}K
          </kbd>
        </button>
        <button
          onClick={() => setPage('settings')}
          className={`flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors ${
            currentPage === 'settings'
              ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
              : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50'
          } w-full`}
        >
          <Settings size={18} />
          <span>{t('nav.settings')}</span>
        </button>
        <button
          onClick={logout}
          className="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm text-sidebar-foreground/70 hover:bg-sidebar-accent/50 w-full"
        >
          <LogOut size={18} />
          <span>{t('settings.logout')}</span>
        </button>
      </div>
    </nav>
  )
}
