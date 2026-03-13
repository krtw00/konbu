import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'
import { Home, FileText, CheckSquare, Calendar, Settings } from 'lucide-react'

const navItems = [
  { page: 'home' as const, icon: Home, labelKey: 'nav.home' },
  { page: 'memos' as const, icon: FileText, labelKey: 'nav.memo' },
  { page: 'todos' as const, icon: CheckSquare, labelKey: 'nav.todo' },
  { page: 'calendar' as const, icon: Calendar, labelKey: 'nav.cal' },
  { page: 'settings' as const, icon: Settings, labelKey: 'nav.settings' },
]

export function BottomNav() {
  const { t } = useTranslation()
  const { currentPage, setPage } = useAppStore()

  return (
    <nav className="md:hidden fixed bottom-0 left-0 right-0 flex border-t border-border bg-background z-50">
      {navItems.map(({ page, icon: Icon, labelKey }) => (
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
          <span>{t(labelKey)}</span>
        </button>
      ))}
    </nav>
  )
}
