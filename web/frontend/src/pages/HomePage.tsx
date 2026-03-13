import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { relativeTime, formatTime, dueFmt } from '@/lib/date'
import { useAppStore } from '@/stores/app'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Memo, Todo, CalendarEvent } from '@/types/api'

export function HomePage() {
  const { t, i18n } = useTranslation()
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [todos, setTodos] = useState<Todo[]>([])
  const [memos, setMemos] = useState<Memo[]>([])
  const setPage = useAppStore((s) => s.setPage)

  useEffect(() => {
    loadHome()
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

  const locale = i18n.language === 'ja' ? 'ja-JP' : 'en-US'

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">
        {new Date().toLocaleDateString(locale, {
          year: 'numeric', month: 'long', day: 'numeric', weekday: 'short',
        })}
      </h1>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">{t('home.todaySchedule')}</CardTitle>
          </CardHeader>
          <CardContent>
            {events.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t('home.noEventsToday')}</p>
            ) : (
              <div className="space-y-2">
                {events.map((e) => (
                  <div key={e.id} className="flex gap-3 text-sm">
                    <span className="text-muted-foreground w-14 shrink-0">
                      {e.all_day ? t('common.allDay') : formatTime(e.start_at)}
                    </span>
                    <span>{e.title}</span>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">
              {t('nav.todo')}
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
                    <div key={t_.id} className="flex items-center gap-2 text-sm">
                      <button
                        className="w-4 h-4 rounded-full border border-muted-foreground/40 hover:border-primary shrink-0"
                        onClick={async () => {
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

        <Card className="md:col-span-2">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">{t('home.recentMemos')}</CardTitle>
          </CardHeader>
          <CardContent>
            {memos.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t('home.noMemos')}</p>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                {memos.map((m) => (
                  <button
                    key={m.id}
                    onClick={() => setPage('memos')}
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
      </div>
    </div>
  )
}
