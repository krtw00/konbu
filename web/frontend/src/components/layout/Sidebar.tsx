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
  const collapsed = currentPage === 'memo-edit'

  return (
    <nav className={`hidden md:flex flex-col border-r border-border bg-sidebar h-screen sticky top-0 transition-all ${
      collapsed ? 'w-12' : 'w-52'
    }`}>
      <div className={`flex items-center gap-2 py-4 ${collapsed ? 'justify-center px-0' : 'px-4'}`}>
        <img className="w-6 h-6" src="/favicon.svg" alt="K" />
        {!collapsed && <span className="font-semibold text-sm text-sidebar-foreground">konbu</span>}
      </div>
      <div className={`flex-1 flex flex-col gap-0.5 ${collapsed ? 'px-1' : 'px-2'}`}>
        {navItems.map(({ page, icon: Icon, labelKey }) => (
          <button
            key={page}
            onClick={() => setPage(page)}
            title={collapsed ? t(labelKey) : undefined}
            className={`flex items-center rounded-md text-sm transition-colors ${
              collapsed ? 'justify-center p-2' : 'gap-2.5 px-3 py-2'
            } ${
              currentPage === page || (currentPage === 'memo-edit' && page === 'memos')
                ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
                : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50'
            }`}
          >
            <Icon size={18} />
            {!collapsed && <span>{t(labelKey)}</span>}
          </button>
        ))}
      </div>
      <div className={`pb-4 flex flex-col gap-0.5 ${collapsed ? 'px-1' : 'px-2'}`}>
        <button
          onClick={() => setCommandOpen(true)}
          title={collapsed ? t('common.search') : undefined}
          className={`flex items-center rounded-md text-sm text-sidebar-foreground/70 hover:bg-sidebar-accent/50 w-full ${
            collapsed ? 'justify-center p-2' : 'gap-2.5 px-3 py-2'
          }`}
        >
          <Search size={18} />
          {!collapsed && (
            <>
              <span>{t('common.search')}</span>
              <kbd className="ml-auto text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded">
                {navigator.platform.includes('Mac') ? '\u2318' : 'Ctrl'}K
              </kbd>
            </>
          )}
        </button>
        <button
          onClick={() => setPage('settings')}
          title={collapsed ? t('nav.settings') : undefined}
          className={`flex items-center rounded-md text-sm transition-colors w-full ${
            collapsed ? 'justify-center p-2' : 'gap-2.5 px-3 py-2'
          } ${
            currentPage === 'settings'
              ? 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
              : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50'
          }`}
        >
          <Settings size={18} />
          {!collapsed && <span>{t('nav.settings')}</span>}
        </button>
        <button
          onClick={logout}
          title={collapsed ? t('settings.logout') : undefined}
          className={`flex items-center rounded-md text-sm text-sidebar-foreground/70 hover:bg-sidebar-accent/50 w-full ${
            collapsed ? 'justify-center p-2' : 'gap-2.5 px-3 py-2'
          }`}
        >
          <LogOut size={18} />
          {!collapsed && <span>{t('settings.logout')}</span>}
        </button>
      </div>
    </nav>
  )
}
