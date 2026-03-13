import { useEffect, useState, useCallback, useRef } from 'react'
import { api } from '@/lib/api'
import { dueFmt, dateDelta, nextMonday } from '@/lib/date'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Calendar as CalendarIcon, Check } from 'lucide-react'
import type { Todo, Tag } from '@/types/api'

export function TodosPage() {
  const [todos, setTodos] = useState<Todo[]>([])
  const [filter, setFilter] = useState<'open' | 'done' | 'all'>('open')
  const [, setAllTags] = useState<string[]>([])
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [selectedTodo, setSelectedTodo] = useState<Todo | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const load = useCallback(async () => {
    const [r, t] = await Promise.all([api.listTodos(), api.listTags()])
    setTodos(r.data || [])
    setAllTags((t.data || []).map((tag: Tag) => tag.name))
  }, [])

  useEffect(() => { load() }, [load])

  const filtered = filter === 'all'
    ? todos
    : todos.filter((t) => (filter === 'done' ? t.status === 'done' : t.status === 'open'))

  async function handleAdd(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key !== 'Enter') return
    let title = (e.target as HTMLInputElement).value.trim()
    if (!title) return
    let dueDate: string | null = null
    const shortcuts: Record<string, () => string> = {
      '!today': () => dateDelta(0),
      '!tomorrow': () => dateDelta(1),
      '!nextweek': () => nextMonday(),
    }
    for (const [kw, fn] of Object.entries(shortcuts)) {
      if (title.toLowerCase().includes(kw)) {
        dueDate = fn()
        title = title.replace(new RegExp(kw.replace('!', '\\!'), 'gi'), '').trim()
      }
    }
    const hashTags = [...title.matchAll(/#(\S+)/g)].map((m) => m[1])
    if (hashTags.length) title = title.replace(/#\S+/g, '').trim()
    if (!title) return

    ;(e.target as HTMLInputElement).value = ''
    await api.createTodo({ title, due_date: dueDate, tags: hashTags })
    load()
  }

  async function toggleDone(todo: Todo) {
    if (todo.status === 'done') await api.reopenTodo(todo.id)
    else await api.doneTodo(todo.id)
    load()
  }

  async function selectTodo(id: string) {
    setSelectedId(id)
    const r = await api.getTodo(id)
    setSelectedTodo(r.data)
  }

  async function saveDetail() {
    if (!selectedTodo) return
    await api.updateTodo(selectedTodo.id, {
      title: selectedTodo.title,
      description: selectedTodo.description,
      status: selectedTodo.status,
      due_date: selectedTodo.due_date,
      tags: selectedTodo.tags?.map((t) => t.name) || [],
    })
    load()
  }

  async function deleteDetail() {
    if (!selectedTodo || !confirm('Delete?')) return
    await api.deleteTodo(selectedTodo.id)
    setSelectedId(null)
    setSelectedTodo(null)
    load()
  }

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-4">ToDo</h1>

      <div className="flex gap-4">
        <div className="flex-1">
          {/* Add input */}
          <div className="flex items-center gap-2 mb-3 bg-accent/30 rounded-lg px-3 py-2">
            <span className="text-muted-foreground text-lg">+</span>
            <Input
              ref={inputRef}
              placeholder="Add a task... Enter to create, !today !tomorrow for date, #tag"
              className="border-0 bg-transparent shadow-none focus-visible:ring-0"
              onKeyDown={handleAdd}
            />
          </div>

          {/* Filter tabs */}
          <Tabs value={filter} onValueChange={(v) => setFilter(v as typeof filter)} className="mb-3">
            <TabsList>
              <TabsTrigger value="open">Open</TabsTrigger>
              <TabsTrigger value="done">Done</TabsTrigger>
              <TabsTrigger value="all">All</TabsTrigger>
            </TabsList>
          </Tabs>

          {/* Todo list */}
          <div className="space-y-0.5">
            {filtered.length === 0 ? (
              <p className="text-sm text-muted-foreground py-8 text-center">
                {filter === 'open' ? 'All done!' : 'No items'}
              </p>
            ) : (
              filtered.map((t) => {
                const df = dueFmt(t.due_date)
                return (
                  <div
                    key={t.id}
                    onClick={() => selectTodo(t.id)}
                    className={`flex items-center gap-3 px-3 py-2.5 rounded-lg cursor-pointer hover:bg-accent/50 ${
                      selectedId === t.id ? 'bg-accent' : ''
                    }`}
                  >
                    <button
                      onClick={(e) => { e.stopPropagation(); toggleDone(t) }}
                      className={`w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0 transition-colors ${
                        t.status === 'done'
                          ? 'bg-primary border-primary text-primary-foreground'
                          : 'border-muted-foreground/40 hover:border-primary'
                      }`}
                    >
                      {t.status === 'done' && <Check size={12} />}
                    </button>
                    <div className="flex-1 min-w-0">
                      <div className={`text-sm ${t.status === 'done' ? 'line-through text-muted-foreground' : ''}`}>
                        {t.title}
                      </div>
                      <div className="flex items-center gap-1.5 mt-0.5">
                        {t.tags?.map((tag) => (
                          <Badge key={tag.id} variant="secondary" className="text-xs py-0">
                            {tag.name}
                          </Badge>
                        ))}
                        {df && (
                          <span className={`text-xs flex items-center gap-1 ${df.className}`}>
                            <CalendarIcon size={10} /> {df.text}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                )
              })
            )}
          </div>
        </div>

        {/* Detail panel */}
        {selectedTodo && (
          <div className="hidden md:block w-72 shrink-0 border-l border-border pl-4 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-xs text-muted-foreground">
                {selectedTodo.status === 'done' ? 'Completed' : 'Open'}
              </span>
              <button className="text-muted-foreground hover:text-foreground" onClick={() => { setSelectedId(null); setSelectedTodo(null) }}>
                x
              </button>
            </div>
            <Input
              value={selectedTodo.title}
              onChange={(e) => setSelectedTodo({ ...selectedTodo, title: e.target.value })}
              onBlur={saveDetail}
              className="font-medium"
            />
            <div>
              <label className="text-xs text-muted-foreground">Notes</label>
              <Textarea
                value={selectedTodo.description || ''}
                onChange={(e) => setSelectedTodo({ ...selectedTodo, description: e.target.value })}
                onBlur={saveDetail}
                placeholder="Add notes..."
                className="mt-1"
              />
            </div>
            <div>
              <label className="text-xs text-muted-foreground">Due Date</label>
              <Input
                type="date"
                value={selectedTodo.due_date || ''}
                onChange={(e) => {
                  setSelectedTodo({ ...selectedTodo, due_date: e.target.value || null })
                  setTimeout(saveDetail, 0)
                }}
                className="mt-1"
              />
            </div>
            <div className="flex gap-2 pt-2">
              <Button variant="destructive" size="sm" onClick={deleteDetail}>Delete</Button>
              <div className="flex-1" />
              <Button size="sm" onClick={() => { saveDetail(); setSelectedId(null); setSelectedTodo(null) }}>Done</Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
