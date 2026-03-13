import { useEffect, useState, useCallback, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { dateKey, formatTime } from '@/lib/date'
import { getHolidays } from '@/lib/holidays'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ChevronLeft, ChevronRight, Repeat } from 'lucide-react'
import type { CalendarEvent } from '@/types/api'

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

export function CalendarPage() {
  const { t, i18n } = useTranslation()
  const [year, setYear] = useState(() => new Date().getFullYear())
  const [month, setMonth] = useState(() => new Date().getMonth())
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [selectedDay, setSelectedDay] = useState<[number, number, number] | null>(null)
  const [editingEvent, setEditingEvent] = useState<CalendarEvent | null>(null)
  const [newRecurrence, setNewRecurrence] = useState('')
  const [firstDayOfWeek, setFirstDayOfWeek] = useState(0)

  const holidays = useMemo(() => getHolidays(year), [year])

  const load = useCallback(async () => {
    const [r, sRes] = await Promise.all([api.listEvents(100), api.getSettings().catch(() => null)])
    setEvents(r.data || [])
    if (sRes?.data?.first_day_of_week !== undefined) {
      setFirstDayOfWeek(sRes.data.first_day_of_week)
    }
  }, [])

  useEffect(() => { load() }, [load])

  // Expand recurring events for the visible month
  const expandedEvents = useMemo(() => {
    const result: CalendarEvent[] = []
    const monthStart = new Date(year, month, 1)
    const monthEnd = new Date(year, month + 1, 0)

    for (const ev of events) {
      const start = new Date(ev.start_at)

      if (!ev.recurrence_rule) {
        result.push(ev)
        continue
      }

      const rule = ev.recurrence_rule
      const endDate = ev.recurrence_end ? new Date(ev.recurrence_end) : new Date(year + 1, 0, 1)
      const cur = new Date(start)

      for (let i = 0; i < 366; i++) {
        if (cur > endDate || cur > monthEnd) break
        if (cur >= monthStart) {
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
  }, [events, year, month])

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
  }

  const daysHeader = firstDayOfWeek === 1 ? DAYS_MON : DAYS_SUN
  const firstDay = new Date(year, month, 1)
  const lastDate = new Date(year, month + 1, 0).getDate()
  const rawDow = firstDay.getDay() // 0=Sunday
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
      start_at: new Date(startAt).toISOString(),
      end_at: endAt ? new Date(endAt).toISOString() : null,
      all_day: false,
      recurrence_rule: newRecurrence || null,
      recurrence_end: null,
      tags: [],
    })
    setNewRecurrence('')
    load()
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
    load()
  }

  async function deleteEvent(id: string) {
    if (!confirm(t('calendar.confirmDelete'))) return
    await api.deleteEvent(id)
    setEditingEvent(null)
    load()
  }

  const dk = selectedDay ? dateKey(selectedDay[0], selectedDay[1], selectedDay[2]) : null

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-4">{t('calendar.title')}</h1>

      <div className="flex gap-4">
        <div className="flex-1">
          <div className="flex items-center justify-between mb-3">
            <span className="font-medium">{year}/{String(month + 1).padStart(2, '0')}</span>
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

          <div className="grid grid-cols-7 gap-px bg-border rounded-lg overflow-hidden">
            {daysHeader.map((d, i) => (
              <div key={d} className={`bg-muted/50 text-center text-xs font-medium py-1.5 ${headerDayColor(i)}`}>
                {d}
              </div>
            ))}
            {Array.from({ length: startPad }, (_, i) => (
              <div key={`p${i}`} className="bg-background p-1 min-h-20 text-muted-foreground/40">
                <span className="text-xs">{prevLastDate - startPad + 1 + i}</span>
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
                  className={`bg-background p-1 min-h-20 cursor-pointer hover:bg-accent/30 transition-colors ${
                    isSelected ? 'ring-2 ring-primary ring-inset' : ''
                  }`}
                >
                  <div className="flex items-center gap-0.5">
                    <span className={`text-xs inline-flex items-center justify-center w-6 h-6 rounded-full ${
                      isToday ? 'bg-primary text-primary-foreground font-medium' : holiday ? 'text-red-500' : dayColor(dow)
                    }`}>
                      {d}
                    </span>
                    {holiday && (
                      <span className="text-[9px] text-red-500 truncate">{holiday}</span>
                    )}
                  </div>
                  <div className="mt-0.5 space-y-0.5">
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
              <div key={`n${i}`} className="bg-background p-1 min-h-20 text-muted-foreground/40">
                <span className="text-xs">{i + 1}</span>
              </div>
            ))}
          </div>
        </div>

        {selectedDay && (
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
                    value={editingEvent.start_at ? new Date(editingEvent.start_at).toISOString().slice(0, 16) : ''}
                    onChange={(e) => setEditingEvent({ ...editingEvent, start_at: new Date(e.target.value).toISOString() })}
                    className="mt-1"
                  />
                </div>
                <div>
                  <label className="text-xs text-muted-foreground">{t('calendar.end')}</label>
                  <Input
                    type="datetime-local"
                    value={editingEvent.end_at ? new Date(editingEvent.end_at).toISOString().slice(0, 16) : ''}
                    onChange={(e) => setEditingEvent({ ...editingEvent, end_at: e.target.value ? new Date(e.target.value).toISOString() : null })}
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
                  <label className="text-xs text-muted-foreground">繰り返し</label>
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
                    <label className="text-xs text-muted-foreground">繰り返し</label>
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
        )}
      </div>
    </div>
  )
}
