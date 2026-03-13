import { useEffect, useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { relativeTime } from '@/lib/date'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Plus, MoreHorizontal } from 'lucide-react'
import type { Memo, Tag } from '@/types/api'

interface MemosPageProps {
  onEditMemo: (id: string) => void
}

export function MemosPage({ onEditMemo }: MemosPageProps) {
  const { t } = useTranslation()
  const [memos, setMemos] = useState<Memo[]>([])
  const [, setAllTags] = useState<string[]>([])
  const [tagFilter, setTagFilter] = useState<string | null>(null)

  const loadMemos = useCallback(async () => {
    const [r, t] = await Promise.all([api.listMemos(), api.listTags()])
    setMemos(r.data || [])
    setAllTags((t.data || []).map((tag: Tag) => tag.name))
  }, [])

  useEffect(() => {
    loadMemos()
  }, [loadMemos])

  const filtered = tagFilter
    ? memos.filter((m) => m.tags?.some((t) => t.name === tagFilter))
    : memos

  const tagCounts: Record<string, number> = {}
  memos.forEach((m) => m.tags?.forEach((t) => {
    tagCounts[t.name] = (tagCounts[t.name] || 0) + 1
  }))

  async function createMemo() {
    const r = await api.createMemo({ title: '', type: 'markdown', content: '', tags: [] })
    onEditMemo(r.data.id)
  }

  async function deleteMemo(id: string, title: string) {
    if (!confirm(t('memos.confirmDelete', { title: title || t('common.untitled') }))) return
    await api.deleteMemo(id)
    loadMemos()
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-semibold">{t('memos.title')}</h1>
        <Button size="sm" onClick={createMemo}>
          <Plus size={16} className="mr-1" /> {t('common.new')}
        </Button>
      </div>

      <div className="flex gap-4">
        {Object.keys(tagCounts).length > 0 && (
          <div className="hidden md:block w-40 shrink-0 space-y-0.5">
            <button
              onClick={() => setTagFilter(null)}
              className={`w-full text-left text-sm px-2 py-1 rounded ${
                tagFilter === null ? 'bg-accent font-medium' : 'hover:bg-accent/50'
              }`}
            >
              {t('common.all')} <span className="text-muted-foreground ml-1">{memos.length}</span>
            </button>
            {Object.entries(tagCounts)
              .sort((a, b) => b[1] - a[1])
              .map(([name, count]) => (
                <button
                  key={name}
                  onClick={() => setTagFilter(name)}
                  className={`w-full text-left text-sm px-2 py-1 rounded ${
                    tagFilter === name ? 'bg-accent font-medium' : 'hover:bg-accent/50'
                  }`}
                >
                  {name} <span className="text-muted-foreground ml-1">{count}</span>
                </button>
              ))}
          </div>
        )}

        <div className="flex-1 space-y-1">
          {filtered.length === 0 ? (
            <p className="text-sm text-muted-foreground py-8 text-center">{t('memos.noMemos')}</p>
          ) : (
            filtered.map((m) => (
              <div
                key={m.id}
                onClick={() => onEditMemo(m.id)}
                className="flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-accent/50 cursor-pointer group"
              >
                <span className={`text-xs font-mono px-1.5 py-0.5 rounded ${
                  m.type === 'table' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300' : 'bg-muted text-muted-foreground'
                }`}>
                  {m.type === 'table' ? 'TBL' : 'MD'}
                </span>
                <div className="flex-1 min-w-0">
                  <div className="text-sm font-medium truncate">{m.title || t('common.untitled')}</div>
                  {m.content && (
                    <div className="text-xs text-muted-foreground truncate">{m.content}</div>
                  )}
                </div>
                <div className="flex items-center gap-1.5 shrink-0">
                  {m.tags?.map((tag) => (
                    <Badge key={tag.id} variant="secondary" className="text-xs">
                      {tag.name}
                    </Badge>
                  ))}
                  <span className="text-xs text-muted-foreground">{relativeTime(m.updated_at)}</span>
                </div>
                <DropdownMenu>
                  <DropdownMenuTrigger onClick={(e) => e.stopPropagation()}>
                    <Button variant="ghost" size="icon" className="h-7 w-7 opacity-0 group-hover:opacity-100">
                      <MoreHorizontal size={14} />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onEditMemo(m.id) }}>
                      {t('common.edit')}
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      className="text-destructive"
                      onClick={(e) => { e.stopPropagation(); deleteMemo(m.id, m.title) }}
                    >
                      {t('common.delete')}
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
