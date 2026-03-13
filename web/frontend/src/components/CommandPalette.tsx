import { useEffect, useState, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useAppStore } from '@/stores/app'
import { CommandDialog, CommandInput, CommandList, CommandItem, CommandGroup, CommandEmpty } from '@/components/ui/command'
import { FileText, CheckSquare, Calendar } from 'lucide-react'
import type { SearchResult } from '@/types/api'

interface CommandPaletteProps {
  onOpenMemo: (id: string) => void
}

export function CommandPalette({ onOpenMemo }: CommandPaletteProps) {
  const { t } = useTranslation()
  const { commandOpen, setCommandOpen, setPage } = useAppStore()
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [fallbackItems, setFallbackItems] = useState<SearchResult[]>([])
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Load fallback items when opening
  useEffect(() => {
    if (!commandOpen) {
      setQuery('')
      setResults([])
      return
    }
    async function load() {
      const [mm, td, ev] = await Promise.all([
        api.listMemos(30),
        api.listTodos(30),
        api.listEvents(20),
      ])
      const items: SearchResult[] = [
        ...(mm.data || []).map((m) => ({ type: 'memo' as const, id: m.id, title: m.title || t('common.untitled'), snippet: '', updated_at: m.updated_at })),
        ...(td.data || []).map((todo) => ({ type: 'todo' as const, id: todo.id, title: todo.title, snippet: '', updated_at: todo.updated_at })),
        ...(ev.data || []).map((e) => ({ type: 'event' as const, id: e.id, title: e.title, snippet: '', updated_at: e.updated_at })),
      ]
      setFallbackItems(items)
    }
    load()
  }, [commandOpen, t])

  const doSearch = useCallback(async (q: string) => {
    if (q.length < 2) {
      setResults([])
      return
    }
    try {
      const r = await api.search(q, 20)
      setResults(r.data || [])
    } catch {
      setResults([])
    }
  }, [])

  function handleQueryChange(val: string) {
    setQuery(val)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => doSearch(val), 300)
  }

  useEffect(() => {
    function handler(e: KeyboardEvent) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setCommandOpen(!commandOpen)
      }
    }
    document.addEventListener('keydown', handler)
    return () => document.removeEventListener('keydown', handler)
  }, [commandOpen, setCommandOpen])

  function handleSelect(item: SearchResult) {
    setCommandOpen(false)
    switch (item.type) {
      case 'memo':
        onOpenMemo(item.id)
        break
      case 'todo':
        setPage('todos')
        break
      case 'event':
        setPage('calendar')
        break
    }
  }

  const icons = {
    memo: FileText,
    todo: CheckSquare,
    event: Calendar,
  }

  const typeLabels: Record<string, string> = {
    memo: t('command.memo'),
    todo: t('command.todo'),
    event: t('command.event'),
  }

  const displayItems = query.length >= 2 ? results : fallbackItems

  return (
    <CommandDialog open={commandOpen} onOpenChange={setCommandOpen}>
      <CommandInput
        placeholder={t('command.searchPlaceholder')}
        value={query}
        onValueChange={handleQueryChange}
      />
      <CommandList>
        <CommandEmpty>{t('command.noResults')}</CommandEmpty>
        <CommandGroup>
          {displayItems.map((item) => {
            const Icon = icons[item.type]
            return (
              <CommandItem key={`${item.type}-${item.id}`} onSelect={() => handleSelect(item)}>
                <Icon size={14} className="mr-2 text-muted-foreground" />
                <span className="text-xs text-muted-foreground mr-2 w-10">{typeLabels[item.type]}</span>
                <span className="flex-1 truncate">{item.title}</span>
                {item.snippet && <span className="text-xs text-muted-foreground truncate ml-2 max-w-40">{item.snippet}</span>}
              </CommandItem>
            )
          })}
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  )
}
