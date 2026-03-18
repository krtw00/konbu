import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'
import { sectionColors, sectionBgColors, sectionBorderColors } from '@/lib/colors'
import { Home, FileText, Table2, CheckSquare, Calendar, Monitor, MessageCircle, Search, Settings, LogOut, Megaphone } from 'lucide-react'

const navItems = [
  { page: 'home' as const, icon: Home, labelKey: 'nav.home' },
  { page: 'memos' as const, icon: FileText, labelKey: 'nav.memos' },
  { page: 'tables' as const, icon: Table2, labelKey: 'nav.tables' },
  { page: 'todos' as const, icon: CheckSquare, labelKey: 'nav.todo' },
  { page: 'calendar' as const, icon: Calendar, labelKey: 'nav.calendar' },
  { page: 'tools' as const, icon: Monitor, labelKey: 'nav.tools' },
  { page: 'search' as const, icon: Search, labelKey: 'nav.search' },
  { page: 'chat' as const, icon: MessageCircle, labelKey: 'nav.chat' },
]

export function Sidebar() {
  const { t } = useTranslation()
  const { currentPage, setPage, setCommandOpen, logout, theme } = useAppStore()
  const collapsed = currentPage === 'memo-edit'
  const isColorful = theme === 'colorful' || theme === 'colorful-dark'

  return (
    <nav className={`hidden md:flex flex-col border-r border-border bg-sidebar h-screen sticky top-0 transition-all ${
      collapsed ? 'w-12' : 'w-52'
    }`}>
      <div className={`flex items-center gap-2 py-4 ${collapsed ? 'justify-center px-0' : 'px-4'}`}>
        <img className="w-6 h-6" src="/favicon.svg" alt="K" />
        {!collapsed && <span className="font-semibold text-sm text-sidebar-foreground">konbu</span>}
      </div>
      <div className={`flex-1 flex flex-col gap-0.5 ${collapsed ? 'px-1' : 'px-2'}`}>
        {navItems.map(({ page, icon: Icon, labelKey }) => {
          const isActive = currentPage === page || (currentPage === 'memo-edit' && page === 'memos') || (currentPage === 'table-edit' && page === 'tables')
          const iconColor = sectionColors[page] || 'text-muted-foreground'
          const activeClass = isActive
            ? isColorful
              ? `${sectionBgColors[page] || ''} ${sectionBorderColors[page] || 'border-transparent'} border-l-3 text-sidebar-accent-foreground font-medium`
              : 'bg-sidebar-accent text-sidebar-accent-foreground font-medium'
            : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50'
          return (
            <button
              key={page}
              onClick={() => setPage(page)}
              title={collapsed ? t(labelKey) : undefined}
              className={`flex items-center rounded-md text-sm transition-colors ${
                collapsed ? 'justify-center p-2' : 'gap-2.5 px-3 py-2'
              } ${activeClass}`}
            >
              <Icon size={18} className={iconColor} />
              {!collapsed && <span>{t(labelKey)}</span>}
            </button>
          )
        })}
      </div>
      <div className={`pb-4 flex flex-col gap-0.5 ${collapsed ? 'px-1' : 'px-2'}`}>
        <a
          href="/feedback"
          title={collapsed ? t('feedback.link') : undefined}
          className={`mb-2 flex items-center rounded-xl border border-primary/20 bg-primary/8 text-sm text-sidebar-foreground transition-colors hover:bg-primary/12 ${
            collapsed ? 'justify-center p-2' : 'gap-2.5 px-3 py-3'
          }`}
        >
          <Megaphone size={18} className="text-primary" />
          {!collapsed && (
            <div className="min-w-0">
              <div className="font-medium">{t('feedback.sidebarTitle')}</div>
              <div className="text-xs text-sidebar-foreground/70">{t('feedback.sidebarDescription')}</div>
            </div>
          )}
        </a>
        <button
          onClick={() => setCommandOpen(true)}
          title={collapsed ? t('common.search') : undefined}
          className={`flex items-center rounded-md text-sm text-sidebar-foreground/70 hover:bg-sidebar-accent/50 w-full ${
            collapsed ? 'justify-center p-2' : 'gap-2.5 px-3 py-2'
          }`}
        >
          <Search size={18} className="text-muted-foreground" />
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
          <Settings size={18} className={sectionColors.settings} />
          {!collapsed && <span>{t('nav.settings')}</span>}
        </button>
        <button
          onClick={logout}
          title={collapsed ? t('settings.logout') : undefined}
          className={`flex items-center rounded-md text-sm text-sidebar-foreground/70 hover:bg-sidebar-accent/50 w-full ${
            collapsed ? 'justify-center p-2' : 'gap-2.5 px-3 py-2'
          }`}
        >
          <LogOut size={18} className="text-muted-foreground" />
          {!collapsed && <span>{t('settings.logout')}</span>}
        </button>
      </div>
    </nav>
  )
}
