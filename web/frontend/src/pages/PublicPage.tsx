import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { dateKey, formatTime } from '@/lib/date'
import { renderMarkdown } from '@/lib/markdown'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import type { PublicShareView } from '@/types/api'

interface PublicPageProps {
  token: string
}

function formatDateTime(value?: string | null) {
  if (!value) return ''
  return new Date(value).toLocaleString()
}

type CalendarViewMode = 'month' | 'week' | 'list'
type PublicCalendar = NonNullable<PublicShareView['calendar']>

function localeFor(language: string) {
  return language.startsWith('ja') ? 'ja-JP' : 'en-US'
}

function getWeekStart(date: Date, weekStartsOnMonday: boolean) {
  const d = new Date(date)
  const day = d.getDay()
  const diff = weekStartsOnMonday ? (day === 0 ? 6 : day - 1) : day
  d.setDate(d.getDate() - diff)
  d.setHours(0, 0, 0, 0)
  return d
}

function addDays(date: Date, days: number) {
  const d = new Date(date)
  d.setDate(d.getDate() + days)
  return d
}

function PublicCalendarSection({ calendar }: { calendar: PublicCalendar }) {
  const { t, i18n } = useTranslation()
  const locale = localeFor(i18n.language)
  const weekStartsOnMonday = i18n.language.startsWith('ja')
  const events = useMemo(
    () => [...calendar.events].sort((a, b) => new Date(a.start_at).getTime() - new Date(b.start_at).getTime()),
    [calendar.events],
  )
  const initialDate = useMemo(() => {
    const firstEvent = events[0]?.start_at
    return firstEvent ? new Date(firstEvent) : new Date()
  }, [events])
  const [mode, setMode] = useState<CalendarViewMode>(() =>
    typeof window !== 'undefined' && window.innerWidth < 768 ? 'list' : 'month',
  )
  const [focusDate, setFocusDate] = useState<Date>(initialDate)

  useEffect(() => {
    setFocusDate(initialDate)
  }, [initialDate])

  const eventsByDay = useMemo(() => {
    const grouped = new Map<string, typeof events>()
    for (const event of events) {
      const d = new Date(event.start_at)
      const key = dateKey(d.getFullYear(), d.getMonth(), d.getDate())
      const existing = grouped.get(key)
      if (existing) {
        existing.push(event)
      } else {
        grouped.set(key, [event])
      }
    }
    return grouped
  }, [events])

  const monthDays = useMemo(() => {
    const start = new Date(focusDate.getFullYear(), focusDate.getMonth(), 1)
    const last = new Date(focusDate.getFullYear(), focusDate.getMonth() + 1, 0)
    const firstGridDate = getWeekStart(start, weekStartsOnMonday)
    const total = Math.ceil((start.getDay() + last.getDate()) / 7) * 7
    return Array.from({ length: total }, (_, index) => addDays(firstGridDate, index))
  }, [focusDate, weekStartsOnMonday])

  const weekDays = useMemo(() => {
    const start = getWeekStart(focusDate, weekStartsOnMonday)
    return Array.from({ length: 7 }, (_, index) => addDays(start, index))
  }, [focusDate, weekStartsOnMonday])

  const listGroups = useMemo(() => {
    const groups = new Map<string, typeof events>()
    for (const event of events) {
      const key = new Date(event.start_at).toLocaleDateString(locale, {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        weekday: 'short',
      })
      const existing = groups.get(key)
      if (existing) {
        existing.push(event)
      } else {
        groups.set(key, [event])
      }
    }
    return [...groups.entries()]
  }, [events, locale])

  function move(offset: number) {
    setFocusDate((current) => {
      if (mode === 'month') {
        return new Date(current.getFullYear(), current.getMonth() + offset, 1)
      }
      return addDays(current, offset * 7)
    })
  }

  const monthLabel = focusDate.toLocaleDateString(locale, { year: 'numeric', month: 'long' })
  const weekLabel = `${weekDays[0].toLocaleDateString(locale, { month: 'short', day: 'numeric' })} - ${weekDays[6].toLocaleDateString(locale, { month: 'short', day: 'numeric' })}`
  const weekdayLabels = weekDays.map((day) => day.toLocaleDateString(locale, { weekday: 'short' }))

  return (
    <section className="space-y-6">
      <div className="space-y-2">
        <div className="flex items-center gap-3">
          <span className="w-3 h-3 rounded-full" style={{ backgroundColor: calendar.color }} />
          <h1 className="text-3xl font-semibold">{calendar.name}</h1>
        </div>
      </div>

      <Tabs value={mode} onValueChange={(value) => setMode(value as CalendarViewMode)}>
        <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <TabsList>
            <TabsTrigger value="month">{t('calendar.month')}</TabsTrigger>
            <TabsTrigger value="week">{t('calendar.week')}</TabsTrigger>
            <TabsTrigger value="list">{t('calendar.list')}</TabsTrigger>
          </TabsList>
          {mode !== 'list' && (
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium min-w-40">{mode === 'month' ? monthLabel : weekLabel}</span>
              <Button variant="ghost" size="icon-sm" onClick={() => move(-1)}>
                <ChevronLeft size={14} />
              </Button>
              <Button variant="ghost" size="sm" onClick={() => setFocusDate(initialDate)}>{t('calendar.today')}</Button>
              <Button variant="ghost" size="icon-sm" onClick={() => move(1)}>
                <ChevronRight size={14} />
              </Button>
            </div>
          )}
        </div>

        <TabsContent value="month" className="space-y-3">
          <div className="grid grid-cols-7 gap-px rounded-xl border border-border bg-border overflow-hidden">
            {weekdayLabels.map((label) => (
              <div key={label} className="bg-muted/60 px-2 py-2 text-center text-xs font-medium">
                {label}
              </div>
            ))}
            {monthDays.map((day) => {
              const key = dateKey(day.getFullYear(), day.getMonth(), day.getDate())
              const dayEvents = eventsByDay.get(key) ?? []
              const isCurrentMonth = day.getMonth() === focusDate.getMonth()
              return (
                <div key={key} className={`min-h-28 bg-background p-2 ${isCurrentMonth ? '' : 'text-muted-foreground/50'}`}>
                  <div className="mb-2 text-xs font-medium">{day.getDate()}</div>
                  <div className="space-y-1">
                    {dayEvents.slice(0, 3).map((event) => (
                      <div key={event.id} className="rounded-md bg-accent/60 px-2 py-1 text-[11px]">
                        <div className="truncate font-medium">{event.title}</div>
                        <div className="truncate text-muted-foreground">
                          {event.all_day ? t('common.allDay') : formatTime(event.start_at)}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )
            })}
          </div>
        </TabsContent>

        <TabsContent value="week" className="space-y-3">
          <div className="grid gap-3 md:grid-cols-7">
            {weekDays.map((day) => {
              const key = dateKey(day.getFullYear(), day.getMonth(), day.getDate())
              const dayEvents = eventsByDay.get(key) ?? []
              return (
                <div key={key} className="rounded-xl border border-border bg-background p-3">
                  <div className="mb-3">
                    <div className="text-xs text-muted-foreground">{day.toLocaleDateString(locale, { weekday: 'short' })}</div>
                    <div className="font-medium">{day.toLocaleDateString(locale, { month: 'short', day: 'numeric' })}</div>
                  </div>
                  <div className="space-y-2">
                    {dayEvents.length === 0 ? (
                      <p className="text-xs text-muted-foreground">{t('calendar.noEvents')}</p>
                    ) : (
                      dayEvents.map((event) => (
                        <div key={event.id} className="rounded-lg border border-border p-2">
                          <div className="text-sm font-medium">{event.title}</div>
                          <div className="text-xs text-muted-foreground">
                            {event.all_day ? t('common.allDay') : formatTime(event.start_at)}
                            {event.end_at ? ` - ${formatTime(event.end_at)}` : ''}
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        </TabsContent>

        <TabsContent value="list" className="space-y-4">
          {listGroups.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t('calendar.noEvents')}</p>
          ) : (
            listGroups.map(([label, dayEvents]) => (
              <div key={label} className="space-y-2">
                <h2 className="text-sm font-medium text-muted-foreground">{label}</h2>
                <div className="space-y-3">
                  {dayEvents.map((event) => (
                    <div key={event.id} className="rounded-xl border border-border p-4 space-y-1">
                      <div className="font-medium">{event.title}</div>
                      <div className="text-xs text-muted-foreground">
                        {event.all_day ? t('common.allDay') : formatDateTime(event.start_at)}
                        {event.end_at ? ` - ${formatDateTime(event.end_at)}` : ''}
                      </div>
                      {event.description && <p className="text-sm whitespace-pre-wrap">{event.description}</p>}
                    </div>
                  ))}
                </div>
              </div>
            ))
          )}
        </TabsContent>
      </Tabs>
    </section>
  )
}

export function PublicPage({ token }: PublicPageProps) {
  const { t } = useTranslation()
  const [view, setView] = useState<PublicShareView | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    api.getPublicShareView(token)
      .then((res) => {
        if (!cancelled) setView(res.data)
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : 'Not found')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => { cancelled = true }
  }, [token])

  useEffect(() => {
    if (!view) return
    const prevTitle = document.title
    const title = view.memo?.title
      || view.todo?.title
      || view.tool?.name
      || view.event?.title
      || view.calendar?.name
      || 'konbu'
    document.title = `${title} | konbu`
    return () => {
      document.title = prevTitle
    }
  }, [view])

  if (loading) {
    return <div className="min-h-screen bg-background" />
  }

  if (error || !view) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-6">
        <div className="max-w-md text-center space-y-2">
          <h1 className="text-2xl font-semibold">{t('publicShare.unavailable')}</h1>
          <p className="text-sm text-muted-foreground">{error || t('publicShare.unavailableDescription')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <main className="max-w-4xl mx-auto px-4 py-10 md:px-6">
        <div className="mb-8">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">konbu</p>
        </div>

        {view.memo && (
          <section className="space-y-6">
            <div className="space-y-2">
              <h1 className="text-3xl font-semibold">{view.memo.title || t('common.untitled')}</h1>
              <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                {view.memo.tags?.map((tag) => <span key={tag.id}>#{tag.name}</span>)}
                <span>{formatDateTime(view.memo.updated_at)}</span>
              </div>
            </div>
            {view.memo.type === 'table' ? (
              <div className="overflow-auto rounded-xl border border-border">
                <table className="w-full text-sm">
                  <thead className="bg-muted/50">
                    <tr>
                      {view.memo.table_columns?.map((column) => (
                        <th key={column.id} className="px-4 py-3 text-left font-medium">{column.name}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {view.memo.rows?.map((row) => (
                      <tr key={row.id} className="border-t border-border">
                        {view.memo?.table_columns?.map((column) => {
                          const value = row.row_data?.[column.id] || ''
                          return <td key={column.id} className="px-4 py-3 align-top">{value}</td>
                        })}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <article
                className="prose prose-neutral max-w-none dark:prose-invert"
                dangerouslySetInnerHTML={{ __html: renderMarkdown(view.memo.content || '') }}
              />
            )}
          </section>
        )}

        {view.todo && (
          <section className="space-y-5">
            <div className="space-y-2">
              <h1 className="text-3xl font-semibold">{view.todo.title}</h1>
              <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                <span>{view.todo.status === 'done' ? t('common.done') : t('common.open')}</span>
                {view.todo.due_date && <span>{view.todo.due_date}</span>}
                <span>{formatDateTime(view.todo.updated_at)}</span>
              </div>
            </div>
            {view.todo.description && (
              <div className="rounded-xl border border-border p-4 whitespace-pre-wrap">{view.todo.description}</div>
            )}
            {view.todo.tags?.length ? (
              <div className="flex flex-wrap gap-2">
                {view.todo.tags.map((tag) => <span key={tag.id} className="text-xs text-muted-foreground">#{tag.name}</span>)}
              </div>
            ) : null}
          </section>
        )}

        {view.tool && (
          <section className="space-y-5">
            <div className="space-y-2">
              <h1 className="text-3xl font-semibold">{view.tool.name}</h1>
              <p className="text-sm text-muted-foreground">{view.tool.category || t('tools.title')}</p>
            </div>
            <a
              href={view.tool.url}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center rounded-lg border border-border px-4 py-2 text-sm hover:bg-accent/50"
            >
              {view.tool.url}
            </a>
          </section>
        )}

        {view.event && (
          <section className="space-y-5">
            <div className="space-y-2">
              <h1 className="text-3xl font-semibold">{view.event.title}</h1>
              <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                <span>{view.event.all_day ? t('common.allDay') : formatDateTime(view.event.start_at)}</span>
                {view.event.end_at && <span>{formatDateTime(view.event.end_at)}</span>}
                <span>{formatDateTime(view.event.updated_at)}</span>
              </div>
            </div>
            {view.event.description && (
              <div className="rounded-xl border border-border p-4 whitespace-pre-wrap">{view.event.description}</div>
            )}
            {view.event.tags?.length ? (
              <div className="flex flex-wrap gap-2">
                {view.event.tags.map((tag) => <span key={tag.id} className="text-xs text-muted-foreground">#{tag.name}</span>)}
              </div>
            ) : null}
          </section>
        )}

        {view.calendar && <PublicCalendarSection calendar={view.calendar} />}
      </main>
    </div>
  )
}
