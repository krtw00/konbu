import { useState, useMemo, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useCache, invalidateCache } from '@/hooks/useCache'
import { dateKey, formatTime, localToISO, isoToLocal } from '@/lib/date'
import { getHolidays } from '@/lib/holidays'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ChevronLeft, ChevronRight, Repeat, Plus } from 'lucide-react'
import type { CalendarEvent } from '@/types/api'

type ViewMode = 'month' | 'week' | 'list'

const DAYS_SUN = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa']
const DAYS_MON = ['Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su']
const MAX_EVENTS = 3
const EVENT_COLORS = ['bg-blue-500/20 text-blue-700 dark:text-blue-300', 'bg-green-500/20 text-green-700 dark:text-green-300', 'bg-purple-500/20 text-purple-700 dark:text-purple-300', 'bg-orange-500/20 text-orange-700 dark:text-orange-300']
const RECURRENCE_OPTIONS = [
  { value: '', label: 'なし' },
  { value: 'daily', label: '毎日' },
  { value: 'weekly', label: '毎週' },
  { value: 'monthly', label: '毎月' },
  { value: 'yearly', label: '毎年' },
]
const HOURS = Array.from({ length: 24 }, (_, i) => i)

function isMobile(): boolean {
  return typeof window !== 'undefined' && window.innerWidth < 768
}

function getWeekStart(date: Date, firstDayOfWeek: number): Date {
  const d = new Date(date)
  const day = d.getDay()
  const diff = firstDayOfWeek === 1
    ? (day === 0 ? 6 : day - 1)
    : day
  d.setDate(d.getDate() - diff)
  d.setHours(0, 0, 0, 0)
  return d
}

function getWeekDays(weekStart: Date): Date[] {
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(weekStart)
    d.setDate(d.getDate() + i)
    return d
  })
}

export function CalendarPage() {
  const { t, i18n } = useTranslation()
  const [viewMode, setViewMode] = useState<ViewMode>(() => isMobile() ? 'list' : 'month')
  const [year, setYear] = useState(() => new Date().getFullYear())
  const [month, setMonth] = useState(() => new Date().getMonth())
  const [weekStart, setWeekStart] = useState<Date>(() => getWeekStart(new Date(), 0))
  const fetchCalendar = useCallback(() => Promise.all([api.listEvents(100), api.getSettings().catch(() => null)]).then(([r, sRes]) => ({
    events: r.data || [] as CalendarEvent[],
    firstDay: sRes?.data?.first_day_of_week ?? 0,
  })), [])
  const { data: calData } = useCache('calendar', fetchCalendar)
  const events = calData?.events || []
  const [selectedDay, setSelectedDay] = useState<[number, number, number] | null>(null)
  const [editingEvent, setEditingEvent] = useState<CalendarEvent | null>(null)
  const [newRecurrence, setNewRecurrence] = useState('')
  const [firstDayOfWeek] = useState(calData?.firstDay ?? 0)
  const [listNewEventDate, setListNewEventDate] = useState<string | null>(null)

  useEffect(() => {
    setWeekStart(getWeekStart(new Date(), firstDayOfWeek))
  }, [firstDayOfWeek])

  const holidays = useMemo(() => getHolidays(year), [year])

  // Expand recurring events for the visible range
  const expandedEvents = useMemo(() => {
    const result: CalendarEvent[] = []
    let rangeStart: Date
    let rangeEnd: Date

    if (viewMode === 'week') {
      rangeStart = new Date(weekStart)
      rangeEnd = new Date(weekStart)
      rangeEnd.setDate(rangeEnd.getDate() + 6)
      rangeEnd.setHours(23, 59, 59, 999)
    } else if (viewMode === 'list') {
      rangeStart = new Date()
      rangeStart.setHours(0, 0, 0, 0)
      rangeEnd = new Date()
      rangeEnd.setDate(rangeEnd.getDate() + 90)
    } else {
      rangeStart = new Date(year, month, 1)
      rangeEnd = new Date(year, month + 1, 0)
    }

    for (const ev of events) {
      const start = new Date(ev.start_at)

      if (!ev.recurrence_rule) {
        result.push(ev)
        continue
      }

      const rule = ev.recurrence_rule
      const endDate = ev.recurrence_end ? new Date(ev.recurrence_end) : new Date(rangeEnd.getFullYear() + 1, 0, 1)
      const cur = new Date(start)

      for (let i = 0; i < 366; i++) {
        if (cur > endDate || cur > rangeEnd) break
        if (cur >= rangeStart) {
          result.push({
            ...ev,
            start_at: cur.toISOString(),
            end_at: ev.end_at ? new Date(cur.getTime() + (new Date(ev.end_at).getTime() - start.getTime())).toISOString() : null,
          })
        }
        if (rule === 'daily') cur.setDate(cur.getDate() + 1)
        else if (rule === 'weekly') cur.setDate(cur.getDate() + 7)
        else if (rule === 'monthly') cur.setMonth(cur.getMonth() + 1)
        else if (rule === 'yearly') cur.setFullYear(cur.getFullYear() + 1)
        else break
      }
    }
    return result
  }, [events, year, month, viewMode, weekStart])

  function eventsForDay(dk: string) {
    return expandedEvents.filter((e) => {
      const d = new Date(e.start_at)
      return dateKey(d.getFullYear(), d.getMonth(), d.getDate()) === dk
    })
  }

  function prevMonth() {
    if (month === 0) { setMonth(11); setYear(year - 1) }
    else setMonth(month - 1)
  }

  function nextMonth() {
    if (month === 11) { setMonth(0); setYear(year + 1) }
    else setMonth(month + 1)
  }

  function goToday() {
    const n = new Date()
    setYear(n.getFullYear())
    setMonth(n.getMonth())
    setWeekStart(getWeekStart(n, firstDayOfWeek))
  }

  function prevWeek() {
    const d = new Date(weekStart)
    d.setDate(d.getDate() - 7)
    setWeekStart(d)
  }

  function nextWeek() {
    const d = new Date(weekStart)
    d.setDate(d.getDate() + 7)
    setWeekStart(d)
  }

  const daysHeader = firstDayOfWeek === 1 ? DAYS_MON : DAYS_SUN
  const firstDay = new Date(year, month, 1)
  const lastDate = new Date(year, month + 1, 0).getDate()
  const rawDow = firstDay.getDay()
  const startPad = firstDayOfWeek === 1 ? (rawDow === 0 ? 6 : rawDow - 1) : rawDow
  const prevLastDate = new Date(year, month, 0).getDate()
  const today = new Date()
  const totalCells = startPad + lastDate
  const remainder = (7 - totalCells % 7) % 7
  const locale = i18n.language === 'ja' ? 'ja-JP' : 'en-US'

  function headerDayColor(idx: number): string {
    const dow = firstDayOfWeek === 1 ? ((idx + 1) % 7) : idx
    if (dow === 0) return 'text-red-500'
    if (dow === 6) return 'text-blue-500'
    return ''
  }

  function dayColor(dayOfWeek: number): string {
    if (dayOfWeek === 0) return 'text-red-500'
    if (dayOfWeek === 6) return 'text-blue-500'
    return ''
  }

  async function handleNewEvent(_dk: string) {
    const title = (document.getElementById('new-ev-title') as HTMLInputElement)?.value.trim()
    if (!title) return
    const startAt = (document.getElementById('new-ev-start') as HTMLInputElement)?.value
    const endAt = (document.getElementById('new-ev-end') as HTMLInputElement)?.value
    const desc = (document.getElementById('new-ev-desc') as HTMLTextAreaElement)?.value
    await api.createEvent({
      title,
      description: desc || '',
      start_at: localToISO(startAt),
      end_at: endAt ? localToISO(endAt) : null,
      all_day: false,
      recurrence_rule: newRecurrence || null,
      recurrence_end: null,
      tags: [],
    })
    setNewRecurrence('')
    invalidateCache('calendar', 'home')
  }

  async function handleNewEventList(_targetDk: string) {
    const title = (document.getElementById('new-ev-title-list') as HTMLInputElement)?.value.trim()
    if (!title) return
    const startAt = (document.getElementById('new-ev-start-list') as HTMLInputElement)?.value
    const endAt = (document.getElementById('new-ev-end-list') as HTMLInputElement)?.value
    const desc = (document.getElementById('new-ev-desc-list') as HTMLTextAreaElement)?.value
    await api.createEvent({
      title,
      description: desc || '',
      start_at: localToISO(startAt),
      end_at: endAt ? localToISO(endAt) : null,
      all_day: false,
      recurrence_rule: null,
      recurrence_end: null,
      tags: [],
    })
    setListNewEventDate(null)
    invalidateCache('calendar', 'home')
  }

  async function saveEvent() {
    if (!editingEvent) return
    await api.updateEvent(editingEvent.id, {
      title: editingEvent.title,
      description: editingEvent.description,
      start_at: editingEvent.start_at,
      end_at: editingEvent.end_at,
      all_day: editingEvent.all_day,
      recurrence_rule: editingEvent.recurrence_rule || null,
      recurrence_end: editingEvent.recurrence_end || null,
      tags: editingEvent.tags?.map((tag) => tag.name) || [],
    })
    setEditingEvent(null)
    invalidateCache('calendar', 'home')
  }

  async function deleteEvent(id: string) {
    if (!confirm(t('calendar.confirmDelete'))) return
    await api.deleteEvent(id)
    setEditingEvent(null)
    invalidateCache('calendar', 'home')
  }

  const dk = selectedDay ? dateKey(selectedDay[0], selectedDay[1], selectedDay[2]) : null

  // Week view data
  const weekDays = useMemo(() => getWeekDays(weekStart), [weekStart])

  const weekLabel = useMemo(() => {
    const end = new Date(weekStart)
    end.setDate(end.getDate() + 6)
    const startStr = weekStart.toLocaleDateString(locale, { month: 'short', day: 'numeric' })
    const endStr = end.toLocaleDateString(locale, { month: 'short', day: 'numeric' })
    return `${startStr} - ${endStr}`
  }, [weekStart, locale])

  // List view data
  const listGroups = useMemo(() => {
    if (viewMode !== 'list') return []
    const todayStart = new Date()
    todayStart.setHours(0, 0, 0, 0)

    const futureEvents = expandedEvents
      .filter((ev) => new Date(ev.start_at) >= todayStart)
      .sort((a, b) => new Date(a.start_at).getTime() - new Date(b.start_at).getTime())

    const groups: { date: string; dk: string; events: CalendarEvent[] }[] = []
    for (const ev of futureEvents) {
      const d = new Date(ev.start_at)
      const evDk = dateKey(d.getFullYear(), d.getMonth(), d.getDate())
      const dateStr = d.toLocaleDateString(locale, { year: 'numeric', month: 'long', day: 'numeric', weekday: 'short' })
      const existing = groups.find((g) => g.dk === evDk)
      if (existing) {
        existing.events.push(ev)
      } else {
        groups.push({ date: dateStr, dk: evDk, events: [ev] })
      }
    }
    return groups
  }, [expandedEvents, viewMode, locale])

  // Editing event panel (shared between views)
  function renderEditPanel(onClose: () => void) {
    if (!editingEvent) return null
    return (
      <div className="space-y-2">
        <Input
          value={editingEvent.title}
          onChange={(e) => setEditingEvent({ ...editingEvent, title: e.target.value })}
          className="font-medium"
        />
        <div>
          <label className="text-xs text-muted-foreground">{t('calendar.start')}</label>
          <Input
            type="datetime-local"
            value={editingEvent.start_at ? isoToLocal(editingEvent.start_at) : ''}
            onChange={(e) => setEditingEvent({ ...editingEvent, start_at: localToISO(e.target.value) })}
            className="mt-1"
          />
        </div>
        <div>
          <label className="text-xs text-muted-foreground">{t('calendar.end')}</label>
          <Input
            type="datetime-local"
            value={editingEvent.end_at ? isoToLocal(editingEvent.end_at) : ''}
            onChange={(e) => setEditingEvent({ ...editingEvent, end_at: e.target.value ? localToISO(e.target.value) : null })}
            className="mt-1"
          />
        </div>
        <div>
          <label className="text-xs text-muted-foreground">{t('calendar.description')}</label>
          <Textarea
            value={editingEvent.description || ''}
            onChange={(e) => setEditingEvent({ ...editingEvent, description: e.target.value })}
            className="mt-1"
          />
        </div>
        <div>
          <label className="text-xs text-muted-foreground">{t('calendar.recurrence')}</label>
          <select
            value={editingEvent.recurrence_rule || ''}
            onChange={(e) => setEditingEvent({ ...editingEvent, recurrence_rule: e.target.value || null })}
            className="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
          >
            {RECURRENCE_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>{o.label}</option>
            ))}
          </select>
        </div>
        <div className="flex gap-2">
          <Button variant="destructive" size="sm" onClick={() => deleteEvent(editingEvent.id)}>{t('common.delete')}</Button>
          <div className="flex-1" />
          <Button variant="ghost" size="sm" onClick={onClose}>{t('common.cancel')}</Button>
          <Button size="sm" onClick={saveEvent}>{t('common.save')}</Button>
        </div>
      </div>
    )
  }

  // View mode buttons
  function renderViewModeButtons() {
    return (
      <div className="flex rounded-lg border border-border overflow-hidden">
        {(['month', 'week', 'list'] as ViewMode[]).map((mode) => (
          <button
            key={mode}
            onClick={() => setViewMode(mode)}
            className={`px-3 py-1.5 text-xs font-medium transition-colors ${
              viewMode === mode
                ? 'bg-primary text-primary-foreground'
                : 'bg-background hover:bg-accent/50 text-muted-foreground'
            }`}
          >
            {t(`calendar.${mode}`)}
          </button>
        ))}
      </div>
    )
  }

  // Navigation header
  function renderNavigation() {
    if (viewMode === 'month') {
      return (
        <div className="flex items-center justify-between mb-3">
          <span className="font-medium">{year}/{String(month + 1).padStart(2, '0')}</span>
          <div className="flex items-center gap-2">
            {renderViewModeButtons()}
            <div className="flex gap-1">
              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={prevMonth}>
                <ChevronLeft size={16} />
              </Button>
              <Button variant="ghost" size="sm" className="h-8" onClick={goToday}>{t('calendar.today')}</Button>
              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={nextMonth}>
                <ChevronRight size={16} />
              </Button>
            </div>
          </div>
        </div>
      )
    }
    if (viewMode === 'week') {
      return (
        <div className="flex items-center justify-between mb-3">
          <span className="font-medium">{weekLabel}</span>
          <div className="flex items-center gap-2">
            {renderViewModeButtons()}
            <div className="flex gap-1">
              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={prevWeek}>
                <ChevronLeft size={16} />
              </Button>
              <Button variant="ghost" size="sm" className="h-8" onClick={goToday}>{t('calendar.thisWeek')}</Button>
              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={nextWeek}>
                <ChevronRight size={16} />
              </Button>
            </div>
          </div>
        </div>
      )
    }
    // list
    return (
      <div className="flex items-center justify-between mb-3">
        <span className="font-medium">{t('calendar.upcoming')}</span>
        <div className="flex items-center gap-2">
          {renderViewModeButtons()}
          <Button variant="ghost" size="sm" className="h-8" onClick={goToday}>{t('calendar.today')}</Button>
        </div>
      </div>
    )
  }

  // Week view
  function renderWeekView() {
    const todayDk = dateKey(today.getFullYear(), today.getMonth(), today.getDate())

    // Separate all-day vs timed events per day
    const allDayByDay: Record<string, CalendarEvent[]> = {}
    const timedByDay: Record<string, CalendarEvent[]> = {}
    for (const d of weekDays) {
      const wdk = dateKey(d.getFullYear(), d.getMonth(), d.getDate())
      const dayEvs = eventsForDay(wdk)
      allDayByDay[wdk] = dayEvs.filter((ev) => ev.all_day)
      timedByDay[wdk] = dayEvs.filter((ev) => !ev.all_day)
    }

    const hasAllDay = weekDays.some((d) => {
      const wdk = dateKey(d.getFullYear(), d.getMonth(), d.getDate())
      return allDayByDay[wdk].length > 0
    })

    return (
      <div className="flex-1 flex flex-col min-h-0">
        <div className="flex-1 flex flex-col min-h-0 overflow-auto">
          {/* Day headers */}
          <div className="grid grid-cols-[40px_repeat(7,1fr)] gap-px bg-border rounded-t-lg overflow-hidden sticky top-0 z-10">
            <div className="bg-muted/50" />
            {weekDays.map((d, i) => {
              const wdk = dateKey(d.getFullYear(), d.getMonth(), d.getDate())
              const isToday = wdk === todayDk
              const dow = d.getDay()
              return (
                <div key={i} className={`bg-muted/50 text-center py-1.5 ${headerDayColor(firstDayOfWeek === 1 ? ((i + 1) % 7 === 0 ? 6 : i) : i)}`}>
                  <div className="text-xs font-medium">{daysHeader[i]}</div>
                  <div className={`text-lg inline-flex items-center justify-center w-8 h-8 rounded-full ${
                    isToday ? 'bg-primary text-primary-foreground font-medium' : dayColor(dow)
                  }`}>
                    {d.getDate()}
                  </div>
                </div>
              )
            })}
          </div>

          {/* All-day events row */}
          {hasAllDay && (
            <div className="grid grid-cols-[60px_repeat(7,1fr)] gap-px bg-border">
              <div className="bg-background p-1 text-xs text-muted-foreground text-right pr-2 py-1">{t('common.allDay')}</div>
              {weekDays.map((d, i) => {
                const wdk = dateKey(d.getFullYear(), d.getMonth(), d.getDate())
                const allDay = allDayByDay[wdk]
                return (
                  <div key={i} className="bg-background p-1 min-h-8">
                    {allDay.map((ev, idx) => (
                      <div
                        key={`${ev.id}-${idx}`}
                        onClick={() => setEditingEvent(ev)}
                        className={`text-xs px-1 py-0.5 rounded truncate cursor-pointer mb-0.5 flex items-center gap-0.5 ${EVENT_COLORS[idx % 4]}`}
                      >
                        {ev.recurrence_rule && <Repeat size={8} className="shrink-0" />}
                        {ev.title}
                      </div>
                    ))}
                  </div>
                )
              })}
            </div>
          )}

          {/* Time grid */}
          <div className="grid grid-cols-[40px_repeat(7,1fr)] gap-px bg-border rounded-b-lg overflow-hidden" style={{ gridTemplateRows: `repeat(24, minmax(28px, 1fr))` }}>
            {HOURS.map((hour) => (
              <div key={hour} className="contents">
                <div className="bg-background text-[10px] text-muted-foreground text-right pr-1 flex items-start justify-end border-t border-border/50">
                  {String(hour).padStart(2, '0')}
                </div>
                {weekDays.map((d, i) => {
                  const wdk = dateKey(d.getFullYear(), d.getMonth(), d.getDate())
                  const isToday = wdk === todayDk
                  const hourEvents = (timedByDay[wdk] || []).filter((ev) => {
                    const h = new Date(ev.start_at).getHours()
                    return h === hour
                  })
                  return (
                    <div
                      key={i}
                      className={`bg-background border-t border-border/50 px-0.5 overflow-hidden ${
                        isToday ? 'bg-primary/5' : ''
                      }`}
                    >
                      {hourEvents.map((ev, idx) => (
                        <div
                          key={`${ev.id}-${idx}`}
                          onClick={() => setEditingEvent(ev)}
                          className={`text-xs px-1 py-0.5 rounded truncate cursor-pointer mb-0.5 flex items-center gap-0.5 ${EVENT_COLORS[idx % 4]}`}
                        >
                          {ev.recurrence_rule && <Repeat size={8} className="shrink-0" />}
                          {formatTime(ev.start_at)} {ev.title}
                        </div>
                      ))}
                    </div>
                  )
                })}
              </div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  // List view
  function renderListView() {
    const todayDk = dateKey(today.getFullYear(), today.getMonth(), today.getDate())
    return (
      <div className="space-y-4">
        <div className="flex justify-end">
          <Button
            size="sm"
            onClick={() => setListNewEventDate(todayDk)}
          >
            <Plus size={14} className="mr-1" />
            {t('calendar.newEvent')}
          </Button>
        </div>

        {listGroups.length === 0 && (
          <p className="text-sm text-muted-foreground py-8 text-center">{t('calendar.noEvents')}</p>
        )}

        {listGroups.map((group) => {
          const isToday = group.dk === todayDk
          return (
            <div key={group.dk} className="border border-border rounded-lg overflow-hidden">
              <div className={`px-3 py-2 text-sm font-medium ${isToday ? 'bg-primary/10 text-primary' : 'bg-muted/50'}`}>
                {group.date}
                {isToday && <span className="ml-2 text-xs">({t('calendar.today')})</span>}
              </div>
              <div className="divide-y divide-border">
                {group.events.map((ev, idx) => (
                  <div
                    key={`${ev.id}-${idx}`}
                    onClick={() => setEditingEvent(ev)}
                    className="px-3 py-2.5 hover:bg-accent/30 cursor-pointer transition-colors"
                  >
                    <div className="flex items-center gap-2">
                      <div className={`w-1 h-8 rounded-full ${EVENT_COLORS[idx % 4].split(' ')[0]}`} />
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-1.5">
                          {ev.recurrence_rule && <Repeat size={12} className="text-muted-foreground shrink-0" />}
                          <span className="font-medium text-sm truncate">{ev.title}</span>
                        </div>
                        <div className="text-xs text-muted-foreground mt-0.5">
                          {ev.all_day
                            ? t('common.allDay')
                            : `${formatTime(ev.start_at)}${ev.end_at ? ' - ' + formatTime(ev.end_at) : ''}`
                          }
                        </div>
                        {ev.description && (
                          <div className="text-xs text-muted-foreground mt-1 truncate">{ev.description}</div>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )
        })}

        {/* New event form for list view */}
        {listNewEventDate && (
          <div className="border border-border rounded-lg p-3 space-y-2">
            <div className="flex items-center justify-between">
              <div className="text-xs text-muted-foreground font-medium">{t('calendar.newEvent')}</div>
              <button className="text-muted-foreground hover:text-foreground" onClick={() => setListNewEventDate(null)}>x</button>
            </div>
            <Input id="new-ev-title-list" placeholder={t('calendar.eventTitle')} />
            <div className="flex gap-2">
              <Input id="new-ev-start-list" type="datetime-local" defaultValue={listNewEventDate + 'T09:00'} />
              <Input id="new-ev-end-list" type="datetime-local" defaultValue={listNewEventDate + 'T10:00'} />
            </div>
            <Textarea id="new-ev-desc-list" placeholder={t('calendar.descriptionPlaceholder')} />
            <Button size="sm" className="w-full" onClick={() => handleNewEventList(listNewEventDate)}>{t('common.add')}</Button>
          </div>
        )}
      </div>
    )
  }

  // Edit event overlay (for week/list views)
  function renderEditOverlay() {
    if (!editingEvent) return null
    if (viewMode === 'month') return null // month view has its own panel
    return (
      <>
        {/* Desktop: side panel */}
        <div className="hidden md:block fixed right-4 top-20 w-80 bg-background border border-border rounded-lg shadow-lg p-4 z-40 space-y-3">
          <div className="flex items-center justify-between mb-2">
            <span className="font-medium text-sm">{editingEvent.title}</span>
            <button className="text-muted-foreground hover:text-foreground" onClick={() => setEditingEvent(null)}>x</button>
          </div>
          {renderEditPanel(() => setEditingEvent(null))}
        </div>
        {/* Mobile: bottom sheet */}
        <div className="md:hidden fixed inset-0 z-50 flex flex-col justify-end">
          <div className="absolute inset-0 bg-black/30" onClick={() => setEditingEvent(null)} />
          <div className="relative bg-background rounded-t-2xl border-t border-border p-4 space-y-3 max-h-[80vh] overflow-auto">
            <div className="flex items-center justify-between">
              <span className="font-medium text-sm">{editingEvent.title}</span>
              <button className="text-muted-foreground hover:text-foreground p-1" onClick={() => setEditingEvent(null)}>x</button>
            </div>
            {renderEditPanel(() => setEditingEvent(null))}
          </div>
        </div>
      </>
    )
  }

  const weekRows = Math.ceil((startPad + lastDate + remainder) / 7)

  return (
    <div className="h-full flex flex-col">
      <h1 className="text-lg font-semibold mb-2">{t('calendar.title')}</h1>

      {viewMode === 'month' && (
        <div className="flex gap-4 flex-1 min-h-0">
          <div className="flex-1 flex flex-col min-h-0">
            {renderNavigation()}

            <div className={`grid grid-cols-7 gap-px bg-border rounded-lg overflow-hidden flex-1`} style={{ gridTemplateRows: `auto repeat(${weekRows}, 1fr)` }}>
              {daysHeader.map((d, i) => (
                <div key={d} className={`bg-muted/50 text-center text-xs font-medium py-1 ${headerDayColor(i)}`}>
                  {d}
                </div>
              ))}
              {Array.from({ length: startPad }, (_, i) => (
                <div key={`p${i}`} className="bg-background px-1 py-0.5 text-muted-foreground/40 overflow-hidden">
                  <span className="text-xs leading-none">{prevLastDate - startPad + 1 + i}</span>
                </div>
              ))}
              {Array.from({ length: lastDate }, (_, i) => {
                const d = i + 1
                const isToday = d === today.getDate() && month === today.getMonth() && year === today.getFullYear()
                const currentDk = dateKey(year, month, d)
                const dayEvents = eventsForDay(currentDk)
                const isSelected = selectedDay && selectedDay[0] === year && selectedDay[1] === month && selectedDay[2] === d
                const dow = new Date(year, month, d).getDay()
                const holidayKey = `${year}-${String(month + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`
                const holiday = holidays.get(holidayKey)

                return (
                  <div
                    key={d}
                    onClick={() => setSelectedDay([year, month, d])}
                    className={`bg-background px-1 py-0.5 cursor-pointer hover:bg-accent/30 transition-colors overflow-hidden ${
                      isSelected ? 'ring-2 ring-primary ring-inset' : ''
                    }`}
                  >
                    <div className="flex items-center gap-0.5">
                      <span className={`text-xs inline-flex items-center justify-center w-5 h-5 rounded-full leading-none ${
                        isToday ? 'bg-primary text-primary-foreground font-medium' : holiday ? 'text-red-500' : dayColor(dow)
                      }`}>
                        {d}
                      </span>
                      {holiday && (
                        <span className="text-[9px] text-red-500 truncate">{holiday}</span>
                      )}
                    </div>
                    <div className="space-y-px">
                      {dayEvents.slice(0, MAX_EVENTS).map((ev, idx) => (
                        <div
                          key={`${ev.id}-${idx}`}
                          onClick={(e) => { e.stopPropagation(); setEditingEvent(ev); setSelectedDay([year, month, d]) }}
                          className={`text-xs px-1 py-0.5 rounded truncate cursor-pointer flex items-center gap-0.5 ${EVENT_COLORS[idx % 4]}`}
                        >
                          {ev.recurrence_rule && <Repeat size={8} className="shrink-0" />}
                          {ev.all_day ? ev.title : `${formatTime(ev.start_at)} ${ev.title}`}
                        </div>
                      ))}
                      {dayEvents.length > MAX_EVENTS && (
                        <div className="text-xs text-muted-foreground px-1">{t('calendar.more', { count: dayEvents.length - MAX_EVENTS })}</div>
                      )}
                    </div>
                  </div>
                )
              })}
              {Array.from({ length: remainder }, (_, i) => (
                <div key={`n${i}`} className="bg-background px-1 py-0.5 text-muted-foreground/40 overflow-hidden">
                  <span className="text-xs leading-none">{i + 1}</span>
                </div>
              ))}
            </div>
          </div>

          {selectedDay && (
            <>
            {/* Desktop: side panel */}
            <div className="hidden md:block w-72 shrink-0 border-l border-border pl-4 space-y-3">
              <div className="flex items-center justify-between">
                <span className="font-medium text-sm">
                  {new Date(selectedDay[0], selectedDay[1], selectedDay[2]).toLocaleDateString(locale, {
                    month: 'long', day: 'numeric', weekday: 'short',
                  })}
                </span>
                <button className="text-muted-foreground hover:text-foreground" onClick={() => { setSelectedDay(null); setEditingEvent(null) }}>
                  x
                </button>
              </div>

              {editingEvent ? (
                <div className="space-y-2">
                  <Input
                    value={editingEvent.title}
                    onChange={(e) => setEditingEvent({ ...editingEvent, title: e.target.value })}
                    className="font-medium"
                  />
                  <div>
                    <label className="text-xs text-muted-foreground">{t('calendar.start')}</label>
                    <Input
                      type="datetime-local"
                      value={editingEvent.start_at ? isoToLocal(editingEvent.start_at) : ''}
                      onChange={(e) => setEditingEvent({ ...editingEvent, start_at: localToISO(e.target.value) })}
                      className="mt-1"
                    />
                  </div>
                  <div>
                    <label className="text-xs text-muted-foreground">{t('calendar.end')}</label>
                    <Input
                      type="datetime-local"
                      value={editingEvent.end_at ? isoToLocal(editingEvent.end_at) : ''}
                      onChange={(e) => setEditingEvent({ ...editingEvent, end_at: e.target.value ? localToISO(e.target.value) : null })}
                      className="mt-1"
                    />
                  </div>
                  <div>
                    <label className="text-xs text-muted-foreground">{t('calendar.description')}</label>
                    <Textarea
                      value={editingEvent.description || ''}
                      onChange={(e) => setEditingEvent({ ...editingEvent, description: e.target.value })}
                      className="mt-1"
                    />
                  </div>
                  <div>
                    <label className="text-xs text-muted-foreground">{t('calendar.recurrence')}</label>
                    <select
                      value={editingEvent.recurrence_rule || ''}
                      onChange={(e) => setEditingEvent({ ...editingEvent, recurrence_rule: e.target.value || null })}
                      className="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                    >
                      {RECURRENCE_OPTIONS.map((o) => (
                        <option key={o.value} value={o.value}>{o.label}</option>
                      ))}
                    </select>
                  </div>
                  <div className="flex gap-2">
                    <Button variant="destructive" size="sm" onClick={() => deleteEvent(editingEvent.id)}>{t('common.delete')}</Button>
                    <div className="flex-1" />
                    <Button size="sm" onClick={saveEvent}>{t('common.save')}</Button>
                  </div>
                </div>
              ) : (
                <>
                  <div className="space-y-1.5">
                    {dk && eventsForDay(dk).length === 0 && (
                      <p className="text-sm text-muted-foreground">{t('calendar.noEvents')}</p>
                    )}
                    {dk && eventsForDay(dk).map((ev, idx) => (
                      <div
                        key={`${ev.id}-${idx}`}
                        onClick={() => setEditingEvent(ev)}
                        className="text-sm p-2 rounded hover:bg-accent/50 cursor-pointer"
                      >
                        <div className="flex items-center gap-1 text-xs text-muted-foreground">
                          {ev.recurrence_rule && <Repeat size={10} />}
                          {ev.all_day ? t('common.allDay') : `${formatTime(ev.start_at)}${ev.end_at ? ' – ' + formatTime(ev.end_at) : ''}`}
                        </div>
                        <div>{ev.title}</div>
                      </div>
                    ))}
                  </div>
                  <div className="border-t border-border pt-3 space-y-2">
                    <div className="text-xs text-muted-foreground font-medium">{t('calendar.newEvent')}</div>
                    <Input id="new-ev-title" placeholder={t('calendar.eventTitle')} />
                    <div className="flex gap-2">
                      <Input id="new-ev-start" type="datetime-local" defaultValue={dk + 'T09:00'} />
                      <Input id="new-ev-end" type="datetime-local" defaultValue={dk + 'T10:00'} />
                    </div>
                    <Textarea id="new-ev-desc" placeholder={t('calendar.descriptionPlaceholder')} />
                    <div>
                      <label className="text-xs text-muted-foreground">{t('calendar.recurrence')}</label>
                      <select
                        value={newRecurrence}
                        onChange={(e) => setNewRecurrence(e.target.value)}
                        className="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      >
                        {RECURRENCE_OPTIONS.map((o) => (
                          <option key={o.value} value={o.value}>{o.label}</option>
                        ))}
                      </select>
                    </div>
                    <Button size="sm" className="w-full" onClick={() => dk && handleNewEvent(dk)}>{t('common.add')}</Button>
                  </div>
                </>
              )}
            </div>
            {/* Mobile: bottom sheet */}
            <div className="md:hidden fixed inset-0 z-50 flex flex-col justify-end">
              <div className="absolute inset-0 bg-black/30" onClick={() => { setSelectedDay(null); setEditingEvent(null) }} />
              <div className="relative bg-background rounded-t-2xl border-t border-border p-4 space-y-3 max-h-[80vh] overflow-auto">
                <div className="flex items-center justify-between">
                  <span className="font-medium text-sm">
                    {new Date(selectedDay[0], selectedDay[1], selectedDay[2]).toLocaleDateString(locale, {
                      month: 'long', day: 'numeric', weekday: 'short',
                    })}
                  </span>
                  <button className="text-muted-foreground hover:text-foreground p-1" onClick={() => { setSelectedDay(null); setEditingEvent(null) }}>
                    x
                  </button>
                </div>

                {editingEvent ? (
                  <div className="space-y-2">
                    <Input
                      value={editingEvent.title}
                      onChange={(e) => setEditingEvent({ ...editingEvent, title: e.target.value })}
                      className="font-medium"
                    />
                    <div>
                      <label className="text-xs text-muted-foreground">{t('calendar.start')}</label>
                      <Input
                        type="datetime-local"
                        value={editingEvent.start_at ? isoToLocal(editingEvent.start_at) : ''}
                        onChange={(e) => setEditingEvent({ ...editingEvent, start_at: localToISO(e.target.value) })}
                        className="mt-1"
                      />
                    </div>
                    <div>
                      <label className="text-xs text-muted-foreground">{t('calendar.end')}</label>
                      <Input
                        type="datetime-local"
                        value={editingEvent.end_at ? isoToLocal(editingEvent.end_at) : ''}
                        onChange={(e) => setEditingEvent({ ...editingEvent, end_at: e.target.value ? localToISO(e.target.value) : null })}
                        className="mt-1"
                      />
                    </div>
                    <div>
                      <label className="text-xs text-muted-foreground">{t('calendar.description')}</label>
                      <Textarea
                        value={editingEvent.description || ''}
                        onChange={(e) => setEditingEvent({ ...editingEvent, description: e.target.value })}
                        className="mt-1"
                      />
                    </div>
                    <div className="flex gap-2">
                      <Button variant="destructive" size="sm" onClick={() => deleteEvent(editingEvent.id)}>{t('common.delete')}</Button>
                      <div className="flex-1" />
                      <Button size="sm" onClick={saveEvent}>{t('common.save')}</Button>
                    </div>
                  </div>
                ) : (
                  <>
                    <div className="space-y-1.5">
                      {dk && eventsForDay(dk).length === 0 && (
                        <p className="text-sm text-muted-foreground">{t('calendar.noEvents')}</p>
                      )}
                      {dk && eventsForDay(dk).map((ev, idx) => (
                        <div
                          key={`m-${ev.id}-${idx}`}
                          onClick={() => setEditingEvent(ev)}
                          className="text-sm p-2 rounded hover:bg-accent/50 cursor-pointer"
                        >
                          <div className="flex items-center gap-1 text-xs text-muted-foreground">
                            {ev.recurrence_rule && <Repeat size={10} />}
                            {ev.all_day ? t('common.allDay') : `${formatTime(ev.start_at)}${ev.end_at ? ' – ' + formatTime(ev.end_at) : ''}`}
                          </div>
                          <div>{ev.title}</div>
                        </div>
                      ))}
                    </div>
                    <div className="border-t border-border pt-3 space-y-2">
                      <div className="text-xs text-muted-foreground font-medium">{t('calendar.newEvent')}</div>
                      <Input id="new-ev-title-m" placeholder={t('calendar.eventTitle')} />
                      <div className="flex flex-col gap-2">
                        <Input id="new-ev-start-m" type="datetime-local" defaultValue={dk + 'T09:00'} />
                        <Input id="new-ev-end-m" type="datetime-local" defaultValue={dk + 'T10:00'} />
                      </div>
                      <Textarea id="new-ev-desc-m" placeholder={t('calendar.descriptionPlaceholder')} />
                      <Button size="sm" className="w-full" onClick={() => dk && handleNewEvent(dk)}>{t('common.add')}</Button>
                    </div>
                  </>
                )}
              </div>
            </div>
            </>
          )}
        </div>
      )}

      {viewMode === 'week' && (
        <div className="flex-1 min-h-0 flex flex-col">
          {renderNavigation()}
          <div className="flex-1 min-h-0 overflow-auto">
            {renderWeekView()}
          </div>
          {renderEditOverlay()}
        </div>
      )}

      {viewMode === 'list' && (
        <div className="flex-1 min-h-0 flex flex-col">
          {renderNavigation()}
          <div className="flex-1 min-h-0 overflow-auto">
            {renderListView()}
          </div>
          {renderEditOverlay()}
        </div>
      )}
    </div>
  )
}
