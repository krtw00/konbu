import { useState, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useCache, invalidateCache } from '@/hooks/useCache'
import { relativeTime, formatTime, dueFmt } from '@/lib/date'
import { stripMarkdown } from '@/lib/markdown'
import { useAppStore } from '@/stores/app'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { GripVertical, Search } from 'lucide-react'
import type { Todo, CalendarEvent } from '@/types/api'

interface HomePageProps {
  onEditMemo: (id: string) => void
}

const DEFAULT_ORDER = ['schedule', 'todos', 'memos']

export function HomePage({ onEditMemo }: HomePageProps) {
  const { t, i18n } = useTranslation()
  const fetchHome = useCallback(() => Promise.all([
    api.listEvents(10),
    api.listTodos(100),
    api.listMemos(6),
    api.getSettings().catch(() => null),
  ]).then(([evR, tdR, mmR, sR]) => {
    const today = new Date().toDateString()
    return {
      events: (evR.data || []).filter((e: CalendarEvent) => new Date(e.start_at).toDateString() === today),
      todos: (tdR.data || []).filter((t: Todo) => t.status === 'open'),
      memos: mmR.data || [],
      widgetOrder: sR?.data?.widget_order || DEFAULT_ORDER,
    }
  }), [])
  const { data: homeData } = useCache('home', fetchHome)
  const events = homeData?.events || []
  const todos = homeData?.todos || []
  const memos = homeData?.memos || []
  const [widgetOrder, setWidgetOrder] = useState<string[]>(homeData?.widgetOrder || DEFAULT_ORDER)
  const { setPage, setSearchQuery } = useAppStore()
  const dragItem = useRef<number | null>(null)
  const dragOver = useRef<number | null>(null)

  function handleDragStart(idx: number) {
    dragItem.current = idx
  }

  function handleDragEnter(idx: number) {
    dragOver.current = idx
  }

  async function handleDragEnd() {
    if (dragItem.current === null || dragOver.current === null || dragItem.current === dragOver.current) {
      dragItem.current = null
      dragOver.current = null
      return
    }
    const newOrder = [...widgetOrder]
    const [removed] = newOrder.splice(dragItem.current, 1)
    newOrder.splice(dragOver.current, 0, removed)
    setWidgetOrder(newOrder)
    dragItem.current = null
    dragOver.current = null
    try {
      const current = await api.getSettings()
      await api.updateSettings({ ...current.data, widget_order: newOrder })
    } catch {
      // ignore
    }
  }

  const locale = i18n.language === 'ja' ? 'ja-JP' : 'en-US'

  const widgets: Record<string, React.ReactNode> = {
    schedule: (
      <Card>
        <CardHeader className="pb-2 cursor-pointer hover:opacity-70" onClick={() => setPage('calendar')}>
          <CardTitle className="text-sm font-medium">{t('home.todaySchedule')} →</CardTitle>
        </CardHeader>
        <CardContent>
          {events.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t('home.noEventsToday')}</p>
          ) : (
            <div className="space-y-2">
              {events.map((e) => (
                <button
                  key={e.id}
                  onClick={() => setPage('calendar')}
                  className="flex gap-3 text-sm w-full text-left rounded-md px-2 py-1 -mx-2 hover:bg-accent/50 transition-colors"
                >
                  <span className="text-muted-foreground w-14 shrink-0">
                    {e.all_day ? t('common.allDay') : formatTime(e.start_at)}
                  </span>
                  <span>{e.title}</span>
                </button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    ),
    todos: (
      <Card>
        <CardHeader className="pb-2 cursor-pointer hover:opacity-70" onClick={() => setPage('todos')}>
          <CardTitle className="text-sm font-medium">
            {t('nav.todo')} →
            <span className="ml-2 text-xs bg-primary text-primary-foreground rounded-full px-2 py-0.5">
              {todos.length}
            </span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {todos.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t('home.allDone')}</p>
          ) : (
            <div className="space-y-1.5">
              {todos.slice(0, 8).map((t_) => {
                const df = dueFmt(t_.due_date)
                return (
                  <div key={t_.id} className="flex items-center gap-2 text-sm rounded-md px-2 py-1 -mx-2 hover:bg-accent/50 transition-colors cursor-pointer" onClick={() => setPage('todos')}>
                    <button
                      className="w-4 h-4 rounded-full border border-muted-foreground/40 hover:border-primary shrink-0"
                      onClick={async (e) => {
                        e.stopPropagation()
                        await api.doneTodo(t_.id)
                        invalidateCache('home', 'todos')
                      }}
                    />
                    <span className="flex-1 truncate">{t_.title}</span>
                    {df && (
                      <span className={`text-xs shrink-0 ${df.className}`}>{df.text}</span>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </CardContent>
      </Card>
    ),
    memos: (
      <Card>
        <CardHeader className="pb-2 cursor-pointer hover:opacity-70" onClick={() => setPage('memos')}>
          <CardTitle className="text-sm font-medium">{t('home.recentMemos')} →</CardTitle>
        </CardHeader>
        <CardContent>
          {memos.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t('home.noMemos')}</p>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {memos.map((m) => (
                <button
                  key={m.id}
                  onClick={() => onEditMemo(m.id)}
                  className="text-left p-3 rounded-lg border border-border hover:bg-accent/50 transition-colors"
                >
                  <div className="font-medium text-sm truncate">{m.title || t('common.untitled')}</div>
                  {m.content && (
                    <div className="text-xs text-muted-foreground mt-1 line-clamp-2">{stripMarkdown(m.content)}</div>
                  )}
                  <div className="text-xs text-muted-foreground mt-2">{relativeTime(m.updated_at)}</div>
                </button>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    ),
  }

  function handleSearchSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const val = (e.currentTarget.elements.namedItem('homeSearch') as HTMLInputElement)?.value?.trim()
    if (val) {
      setSearchQuery(val)
      setPage('search')
    }
  }

  return (
    <div>
      <form onSubmit={handleSearchSubmit} className="mb-4">
        <div className="relative">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            name="homeSearch"
            placeholder={t('search.placeholder')}
            className="pl-9"
          />
        </div>
      </form>
      <h1 className="text-2xl font-semibold mb-6">
        {new Date().toLocaleDateString(locale, {
          year: 'numeric', month: 'long', day: 'numeric', weekday: 'short',
        })}
      </h1>
      <div className="space-y-4">
        {widgetOrder.map((id, idx) => (
          <div
            key={id}
            draggable
            onDragStart={() => handleDragStart(idx)}
            onDragEnter={() => handleDragEnter(idx)}
            onDragEnd={handleDragEnd}
            onDragOver={(e) => e.preventDefault()}
            className="relative group"
          >
            <div className="absolute left-0 top-0 bottom-0 w-6 flex items-center justify-center cursor-grab opacity-0 group-hover:opacity-50 hover:!opacity-100 z-10">
              <GripVertical size={14} className="text-muted-foreground" />
            </div>
            {widgets[id]}
          </div>
        ))}
      </div>
    </div>
  )
}
