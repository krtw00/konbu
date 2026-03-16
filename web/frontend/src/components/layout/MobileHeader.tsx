import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'
import { Menu, X, Home, FileText, CheckSquare, Calendar, Monitor, MessageCircle, Search, Settings, LogOut } from 'lucide-react'

const navItems = [
  { page: 'home' as const, icon: Home, labelKey: 'nav.home' },
  { page: 'memos' as const, icon: FileText, labelKey: 'nav.memos' },
  { page: 'todos' as const, icon: CheckSquare, labelKey: 'nav.todo' },
  { page: 'calendar' as const, icon: Calendar, labelKey: 'nav.calendar' },
  { page: 'tools' as const, icon: Monitor, labelKey: 'nav.tools' },
  { page: 'search' as const, icon: Search, labelKey: 'nav.search' },
  { page: 'chat' as const, icon: MessageCircle, labelKey: 'nav.chat' },
  { page: 'settings' as const, icon: Settings, labelKey: 'nav.settings' },
]

export function MobileHeader() {
  const { t } = useTranslation()
  const { currentPage, setPage, setCommandOpen, logout } = useAppStore()
  const [open, setOpen] = useState(false)

  function navigate(page: typeof navItems[number]['page']) {
    setPage(page)
    setOpen(false)
  }

  return (
    <div className="md:hidden">
      <header className="flex items-center justify-between px-4 py-3 border-b border-border bg-background sticky top-0 z-50">
        <div className="flex items-center gap-2">
          <img className="w-5 h-5" src="/favicon.svg" alt="K" />
          <span className="font-semibold text-sm">konbu</span>
        </div>
        <div className="flex items-center gap-1">
          <button
            onClick={() => setCommandOpen(true)}
            className="p-2 rounded-md text-muted-foreground hover:bg-accent"
          >
            <Search size={20} />
          </button>
          <button
            onClick={() => setOpen(!open)}
            className="p-2 rounded-md text-muted-foreground hover:bg-accent"
          >
            {open ? <X size={20} /> : <Menu size={20} />}
          </button>
        </div>
      </header>

      {open && (
        <>
          <div className="fixed inset-0 top-[53px] bg-black/40 z-40" onClick={() => setOpen(false)} />
          <nav className="fixed top-[53px] left-0 right-0 bg-background border-b border-border z-50 py-2">
            {navItems.map(({ page, icon: Icon, labelKey }) => (
              <button
                key={page}
                onClick={() => navigate(page)}
                className={`flex items-center gap-3 w-full px-5 py-3 text-sm transition-colors ${
                  currentPage === page || (currentPage === 'memo-edit' && page === 'memos')
                    ? 'text-primary font-medium bg-accent/50'
                    : 'text-foreground/80'
                }`}
              >
                <Icon size={18} />
                <span>{t(labelKey)}</span>
              </button>
            ))}
            <div className="border-t border-border my-1" />
            <button
              onClick={() => { logout(); setOpen(false) }}
              className="flex items-center gap-3 w-full px-5 py-3 text-sm text-foreground/80"
            >
              <LogOut size={18} />
              <span>{t('settings.logout')}</span>
            </button>
          </nav>
        </>
      )}
    </div>
  )
}
