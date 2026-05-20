import { useState, useMemo, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useCache } from '@/hooks/useCache'
import { formatTime, dueFmt, relativeTime } from '@/lib/date'
import { useAppStore } from '@/stores/app'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import type { CalendarEvent, Todo, Memo } from '@/types/api'

interface DailyPageProps {
  onEditMemo: (id: string) => void
}

const pad2 = (n: number) => String(n).padStart(2, '0')

/** "YYYY-MM-DD" from a Date using its local components */
function dateInputValue(d: Date): string {
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}`
}

/** Parse a "YYYY-MM-DD" date-input value into a local-midnight Date */
function parseDateInput(value: string): Date {
  const [y, m, d] = value.split('-').map(Number)
  return new Date(y, m - 1, d, 0, 0, 0, 0)
}

export function DailyPage({ onEditMemo }: DailyPageProps) {
  const { t, i18n } = useTranslation()
  const { setPage } = useAppStore()
  const [selected, setSelected] = useState<Date>(() => {
    const n = new Date()
    return new Date(n.getFullYear(), n.getMonth(), n.getDate(), 0, 0, 0, 0)
  })

  const locale = i18n.language === 'ja' ? 'ja-JP' : 'en-US'

  // Local 00:00:00 of selected day -> from; next day local 00:00:00 -> to.
  // toISOString() converts each to UTC RFC3339 for the API.
  const { from, to } = useMemo(() => {
    const start = new Date(selected.getFullYear(), selected.getMonth(), selected.getDate(), 0, 0, 0, 0)
    const next = new Date(selected.getFullYear(), selected.getMonth(), selected.getDate() + 1, 0, 0, 0, 0)
    return { from: start.toISOString(), to: next.toISOString() }
  }, [selected])

  const fetchDaily = useCallback(() => api.daily(from, to).then(r => ({
    events: r.data.events || [],
    todos: r.data.todos || [],
    memos: r.data.memos || [],
  })), [from, to])
  const { data } = useCache(`daily-${from}`, fetchDaily)
  const events: CalendarEvent[] = data?.events || []
  const todos: Todo[] = data?.todos || []
  const memos: Memo[] = data?.memos || []

  function shiftDay(delta: number) {
    setSelected(prev => new Date(prev.getFullYear(), prev.getMonth(), prev.getDate() + delta, 0, 0, 0, 0))
  }

  function goToday() {
    const n = new Date()
    setSelected(new Date(n.getFullYear(), n.getMonth(), n.getDate(), 0, 0, 0, 0))
  }

  const isToday = useMemo(() => {
    const n = new Date()
    return selected.getFullYear() === n.getFullYear()
      && selected.getMonth() === n.getMonth()
      && selected.getDate() === n.getDate()
  }, [selected])

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-4">{t('daily.title')}</h1>

      <div className="mb-6 flex flex-wrap items-center gap-2">
        <Button variant="ghost" size="icon" className="h-9 w-9 shrink-0" onClick={() => shiftDay(-1)} aria-label={t('daily.prevDay')}>
          <ChevronLeft size={16} />
        </Button>
        <Button variant="ghost" size="icon" className="h-9 w-9 shrink-0" onClick={() => shiftDay(1)} aria-label={t('daily.nextDay')}>
          <ChevronRight size={16} />
        </Button>
        <Input
          type="date"
          value={dateInputValue(selected)}
          onChange={(e) => e.target.value && setSelected(parseDateInput(e.target.value))}
          className="h-9 w-auto"
        />
        <Button variant="outline" size="sm" className="h-9" disabled={isToday} onClick={goToday}>
          {t('daily.today')}
        </Button>
        <span className="ml-auto text-sm text-muted-foreground">
          {selected.toLocaleDateString(locale, { year: 'numeric', month: 'long', day: 'numeric', weekday: 'short' })}
        </span>
      </div>

      <div className="space-y-4">
        {/* Events */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">{t('daily.events')}</CardTitle>
          </CardHeader>
          <CardContent>
            {events.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t('daily.noEvents')}</p>
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
                    <span className="truncate">{e.title}</span>
                  </button>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Todos due on this day */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">
              {t('daily.todos')}
              <span className="ml-2 text-xs bg-primary text-primary-foreground rounded-full px-2 py-0.5">
                {todos.length}
              </span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            {todos.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t('daily.noTodos')}</p>
            ) : (
              <div className="space-y-1.5">
                {todos.map((td) => {
                  const df = dueFmt(td.due_date)
                  return (
                    <div
                      key={td.id}
                      className="flex items-center gap-2 text-sm rounded-md px-2 py-1 -mx-2 hover:bg-accent/50 transition-colors cursor-pointer"
                      onClick={() => setPage('todos')}
                    >
                      <span className={`w-4 h-4 rounded-full border shrink-0 ${td.status === 'done' ? 'bg-primary border-primary' : 'border-muted-foreground/40'}`} />
                      <span className={`flex-1 truncate ${td.status === 'done' ? 'line-through text-muted-foreground' : ''}`}>{td.title}</span>
                      <span className="text-xs shrink-0 text-muted-foreground">
                        {td.status === 'done' ? t('common.done') : t('common.open')}
                      </span>
                      {df && <span className={`text-xs shrink-0 ${df.className}`}>{df.text}</span>}
                    </div>
                  )
                })}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Memos created on this day */}
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">{t('daily.memos')}</CardTitle>
          </CardHeader>
          <CardContent>
            {memos.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t('daily.noMemos')}</p>
            ) : (
              <div className="space-y-1.5">
                {memos.map((m) => (
                  <button
                    key={m.id}
                    onClick={() => onEditMemo(m.id)}
                    className="flex items-center gap-2 text-sm w-full text-left rounded-md px-2 py-1 -mx-2 hover:bg-accent/50 transition-colors"
                  >
                    <span className="flex-1 truncate">{m.title || t('common.untitled')}</span>
                    <span className="text-[10px] uppercase tracking-wide rounded bg-muted px-1.5 py-0.5 text-muted-foreground shrink-0">
                      {m.type === 'table' ? t('nav.tables') : t('nav.memo')}
                    </span>
                    <span className="text-xs text-muted-foreground shrink-0">{relativeTime(m.created_at)}</span>
                  </button>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
