import { useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useCache, invalidateCache } from '@/hooks/useCache'
import { relativeTime } from '@/lib/date'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Plus, MoreHorizontal, Table2 } from 'lucide-react'
import type { Memo } from '@/types/api'

interface TablesPageProps {
  onEditTable: (id: string) => void
}

export function TablesPage({ onEditTable }: TablesPageProps) {
  const { t } = useTranslation()
  const fetchTables = useCallback(() => api.listMemos(500).then(r => ({
    tables: (r.data || []).filter((m: Memo) => m.type === 'table'),
  })), [])
  const { data } = useCache('tables', fetchTables)
  const tables = data?.tables || []
  const [creating, setCreating] = useState(false)

  async function createTable() {
    if (creating) return
    setCreating(true)
    try {
      const r = await api.createMemo({ title: '', type: 'table', content: '', tags: [] })
      invalidateCache('tables', 'memos')
      onEditTable(r.data.id)
    } finally {
      setCreating(false)
    }
  }

  async function deleteTable(id: string, title: string) {
    if (!confirm(t('tables.confirmDelete', { title: title || t('common.untitled') }))) return
    await api.deleteMemo(id)
    invalidateCache('tables', 'memos', 'home')
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-semibold">{t('tables.title')}</h1>
        <Button size="sm" onClick={createTable} disabled={creating}>
          <Plus size={16} className="mr-1" /> {t('common.new')}
        </Button>
      </div>

      {tables.length === 0 ? (
        <p className="text-sm text-muted-foreground py-8 text-center">{t('tables.noTables')}</p>
      ) : (
        <div className="space-y-1">
          {tables.map((m) => (
            <div
              key={m.id}
              onClick={() => onEditTable(m.id)}
              className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-accent/50 cursor-pointer group"
            >
              <Table2 size={16} className="text-blue-500 shrink-0" />
              <div className="flex-1 min-w-0">
                <div className="text-sm font-medium truncate">{m.title || t('common.untitled')}</div>
                {m.row_count !== undefined && (
                  <div className="text-xs text-muted-foreground">{t('table.rowCount', { count: m.row_count })}</div>
                )}
              </div>
              <span className="text-xs text-muted-foreground shrink-0">{relativeTime(m.updated_at)}</span>
              <DropdownMenu>
                <DropdownMenuTrigger onClick={(e) => e.stopPropagation()}>
                  <Button variant="ghost" size="icon" className="h-7 w-7 opacity-100 md:opacity-0 md:group-hover:opacity-100">
                    <MoreHorizontal size={14} />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onEditTable(m.id) }}>
                    {t('common.edit')}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    className="text-destructive"
                    onClick={(e) => { e.stopPropagation(); deleteTable(m.id, m.title) }}
                  >
                    {t('common.delete')}
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
