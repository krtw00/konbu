import { useEffect, useState, useCallback } from 'react'
import { api } from '@/lib/api'
import { dateKey, formatTime } from '@/lib/date'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import type { CalendarEvent } from '@/types/api'

const DAYS = ['Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su']
const MAX_EVENTS = 3
const EVENT_COLORS = ['bg-blue-500/20 text-blue-700 dark:text-blue-300', 'bg-green-500/20 text-green-700 dark:text-green-300', 'bg-purple-500/20 text-purple-700 dark:text-purple-300', 'bg-orange-500/20 text-orange-700 dark:text-orange-300']

export function CalendarPage() {
  const [year, setYear] = useState(() => new Date().getFullYear())
  const [month, setMonth] = useState(() => new Date().getMonth())
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [selectedDay, setSelectedDay] = useState<[number, number, number] | null>(null)
  const [editingEvent, setEditingEvent] = useState<CalendarEvent | null>(null)

  const load = useCallback(async () => {
    const r = await api.listEvents(100)
    setEvents(r.data || [])
  }, [])

  useEffect(() => { load() }, [load])

  function eventsForDay(dk: string) {
    return events.filter((e) => {
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

  const firstDay = new Date(year, month, 1)
  const lastDate = new Date(year, month + 1, 0).getDate()
  const startPad = (firstDay.getDay() + 6) % 7
  const prevLastDate = new Date(year, month, 0).getDate()
  const today = new Date()
  const totalCells = startPad + lastDate
  const remainder = (7 - totalCells % 7) % 7

  async function handleNewEvent(dk: string) {
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
      tags: [],
    })
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
      tags: editingEvent.tags?.map((t) => t.name) || [],
    })
    setEditingEvent(null)
    load()
  }

  async function deleteEvent(id: string) {
    if (!confirm('Delete?')) return
    await api.deleteEvent(id)
    setEditingEvent(null)
    load()
  }

  const dk = selectedDay ? dateKey(selectedDay[0], selectedDay[1], selectedDay[2]) : null

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-4">Calendar</h1>

      <div className="flex gap-4">
        <div className="flex-1">
          {/* Header */}
          <div className="flex items-center justify-between mb-3">
            <span className="font-medium">{year}/{String(month + 1).padStart(2, '0')}</span>
            <div className="flex gap-1">
              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={prevMonth}>
                <ChevronLeft size={16} />
              </Button>
              <Button variant="ghost" size="sm" className="h-8" onClick={goToday}>Today</Button>
              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={nextMonth}>
                <ChevronRight size={16} />
              </Button>
            </div>
          </div>

          {/* Grid */}
          <div className="grid grid-cols-7 gap-px bg-border rounded-lg overflow-hidden">
            {DAYS.map((d) => (
              <div key={d} className="bg-muted/50 text-center text-xs font-medium py-1.5 text-muted-foreground">
                {d}
              </div>
            ))}
            {/* Previous month padding */}
            {Array.from({ length: startPad }, (_, i) => (
              <div key={`p${i}`} className="bg-background p-1 min-h-20 text-muted-foreground/40">
                <span className="text-xs">{prevLastDate - startPad + 1 + i}</span>
              </div>
            ))}
            {/* Current month */}
            {Array.from({ length: lastDate }, (_, i) => {
              const d = i + 1
              const isToday = d === today.getDate() && month === today.getMonth() && year === today.getFullYear()
              const currentDk = dateKey(year, month, d)
              const dayEvents = eventsForDay(currentDk)
              const isSelected = selectedDay && selectedDay[0] === year && selectedDay[1] === month && selectedDay[2] === d

              return (
                <div
                  key={d}
                  onClick={() => setSelectedDay([year, month, d])}
                  className={`bg-background p-1 min-h-20 cursor-pointer hover:bg-accent/30 transition-colors ${
                    isSelected ? 'ring-2 ring-primary ring-inset' : ''
                  }`}
                >
                  <span className={`text-xs inline-flex items-center justify-center w-6 h-6 rounded-full ${
                    isToday ? 'bg-primary text-primary-foreground font-medium' : ''
                  }`}>
                    {d}
                  </span>
                  <div className="mt-0.5 space-y-0.5">
                    {dayEvents.slice(0, MAX_EVENTS).map((ev, idx) => (
                      <div
                        key={ev.id}
                        onClick={(e) => { e.stopPropagation(); setEditingEvent(ev); setSelectedDay([year, month, d]) }}
                        className={`text-xs px-1 py-0.5 rounded truncate cursor-pointer ${EVENT_COLORS[idx % 4]}`}
                      >
                        {ev.all_day ? ev.title : `${formatTime(ev.start_at)} ${ev.title}`}
                      </div>
                    ))}
                    {dayEvents.length > MAX_EVENTS && (
                      <div className="text-xs text-muted-foreground px-1">+{dayEvents.length - MAX_EVENTS} more</div>
                    )}
                  </div>
                </div>
              )
            })}
            {/* Next month padding */}
            {Array.from({ length: remainder }, (_, i) => (
              <div key={`n${i}`} className="bg-background p-1 min-h-20 text-muted-foreground/40">
                <span className="text-xs">{i + 1}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Day detail / Event edit panel */}
        {selectedDay && (
          <div className="hidden md:block w-72 shrink-0 border-l border-border pl-4 space-y-3">
            <div className="flex items-center justify-between">
              <span className="font-medium text-sm">
                {new Date(selectedDay[0], selectedDay[1], selectedDay[2]).toLocaleDateString('ja-JP', {
                  month: 'long', day: 'numeric', weekday: 'short',
                })}
              </span>
              <button className="text-muted-foreground hover:text-foreground" onClick={() => { setSelectedDay(null); setEditingEvent(null) }}>
                x
              </button>
            </div>

            {editingEvent ? (
              /* Edit event */
              <div className="space-y-2">
                <Input
                  value={editingEvent.title}
                  onChange={(e) => setEditingEvent({ ...editingEvent, title: e.target.value })}
                  className="font-medium"
                />
                <div>
                  <label className="text-xs text-muted-foreground">Start</label>
                  <Input
                    type="datetime-local"
                    value={editingEvent.start_at ? new Date(editingEvent.start_at).toISOString().slice(0, 16) : ''}
                    onChange={(e) => setEditingEvent({ ...editingEvent, start_at: new Date(e.target.value).toISOString() })}
                    className="mt-1"
                  />
                </div>
                <div>
                  <label className="text-xs text-muted-foreground">End</label>
                  <Input
                    type="datetime-local"
                    value={editingEvent.end_at ? new Date(editingEvent.end_at).toISOString().slice(0, 16) : ''}
                    onChange={(e) => setEditingEvent({ ...editingEvent, end_at: e.target.value ? new Date(e.target.value).toISOString() : null })}
                    className="mt-1"
                  />
                </div>
                <div>
                  <label className="text-xs text-muted-foreground">Description</label>
                  <Textarea
                    value={editingEvent.description || ''}
                    onChange={(e) => setEditingEvent({ ...editingEvent, description: e.target.value })}
                    className="mt-1"
                  />
                </div>
                <div className="flex gap-2">
                  <Button variant="destructive" size="sm" onClick={() => deleteEvent(editingEvent.id)}>Delete</Button>
                  <div className="flex-1" />
                  <Button size="sm" onClick={saveEvent}>Save</Button>
                </div>
              </div>
            ) : (
              /* Day events + new event form */
              <>
                <div className="space-y-1.5">
                  {dk && eventsForDay(dk).length === 0 && (
                    <p className="text-sm text-muted-foreground">No events</p>
                  )}
                  {dk && eventsForDay(dk).map((ev) => (
                    <div
                      key={ev.id}
                      onClick={() => setEditingEvent(ev)}
                      className="text-sm p-2 rounded hover:bg-accent/50 cursor-pointer"
                    >
                      <div className="text-xs text-muted-foreground">
                        {ev.all_day ? 'All day' : `${formatTime(ev.start_at)}${ev.end_at ? ' – ' + formatTime(ev.end_at) : ''}`}
                      </div>
                      <div>{ev.title}</div>
                    </div>
                  ))}
                </div>
                <div className="border-t border-border pt-3 space-y-2">
                  <div className="text-xs text-muted-foreground font-medium">New Event</div>
                  <Input id="new-ev-title" placeholder="Event title" />
                  <div className="flex gap-2">
                    <Input id="new-ev-start" type="datetime-local" defaultValue={dk + 'T09:00'} />
                    <Input id="new-ev-end" type="datetime-local" defaultValue={dk + 'T10:00'} />
                  </div>
                  <Textarea id="new-ev-desc" placeholder="Description (optional)" />
                  <Button size="sm" className="w-full" onClick={() => dk && handleNewEvent(dk)}>Add</Button>
                </div>
              </>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
