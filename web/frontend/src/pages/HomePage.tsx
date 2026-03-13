import { useEffect, useState, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { relativeTime, formatTime, dueFmt } from '@/lib/date'
import { useAppStore } from '@/stores/app'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { DndContext, closestCenter, PointerSensor, useSensor, useSensors } from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy, useSortable, arrayMove } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { GripVertical } from 'lucide-react'
import type { Memo, Todo, CalendarEvent } from '@/types/api'
import type { DragEndEvent } from '@dnd-kit/core'

interface HomePageProps {
  onEditMemo: (id: string) => void
}

const DEFAULT_ORDER = ['schedule', 'todos', 'memos']

function SortableWidget({ id, children }: { id: string; children: React.ReactNode }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id })
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }
  return (
    <div ref={setNodeRef} style={style} className="relative group">
      <div
        {...attributes}
        {...listeners}
        className="absolute left-0 top-0 bottom-0 w-6 flex items-center justify-center cursor-grab opacity-0 group-hover:opacity-50 hover:!opacity-100 z-10"
      >
        <GripVertical size={14} className="text-muted-foreground" />
      </div>
      {children}
    </div>
  )
}

export function HomePage({ onEditMemo }: HomePageProps) {
  const { t, i18n } = useTranslation()
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [todos, setTodos] = useState<Todo[]>([])
  const [memos, setMemos] = useState<Memo[]>([])
  const [widgetOrder, setWidgetOrder] = useState<string[]>(DEFAULT_ORDER)
  const setPage = useAppStore((s) => s.setPage)

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } })
  )

  useEffect(() => {
    loadHome()
    loadSettings()
  }, [])

  async function loadHome() {
    const [evR, tdR, mmR] = await Promise.all([
      api.listEvents(10),
      api.listTodos(100),
      api.listMemos(6),
    ])
    const today = new Date().toDateString()
    setEvents((evR.data || []).filter((e) => new Date(e.start_at).toDateString() === today))
    setTodos((tdR.data || []).filter((t) => t.status === 'open'))
    setMemos(mmR.data || [])
  }

  async function loadSettings() {
    try {
      const r = await api.getSettings()
      if (r.data?.widget_order?.length) {
        setWidgetOrder(r.data.widget_order)
      }
    } catch {
      // ignore
    }
  }

  const handleDragEnd = useCallback(async (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    const oldIndex = widgetOrder.indexOf(active.id as string)
    const newIndex = widgetOrder.indexOf(over.id as string)
    const newOrder = arrayMove(widgetOrder, oldIndex, newIndex)
    setWidgetOrder(newOrder)
    try {
      const current = await api.getSettings()
      await api.updateSettings({ ...current.data, widget_order: newOrder })
    } catch {
      // ignore
    }
  }, [widgetOrder])

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
                        loadHome()
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
                    <div className="text-xs text-muted-foreground mt-1 line-clamp-2">{m.content}</div>
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

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">
        {new Date().toLocaleDateString(locale, {
          year: 'numeric', month: 'long', day: 'numeric', weekday: 'short',
        })}
      </h1>
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <SortableContext items={widgetOrder} strategy={verticalListSortingStrategy}>
          <div className="space-y-4">
            {widgetOrder.map((id) => (
              <SortableWidget key={id} id={id}>
                {widgets[id]}
              </SortableWidget>
            ))}
          </div>
        </SortableContext>
      </DndContext>
    </div>
  )
}
