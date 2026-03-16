import { useEffect, useState, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useAppStore } from '@/stores/app'
import { Command, CommandInput, CommandList, CommandItem, CommandGroup, CommandEmpty } from '@/components/ui/command'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { FileText, CheckSquare, Calendar, Monitor } from 'lucide-react'
import { sectionColors } from '@/lib/colors'
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
        ...(mm.data || []).map((m) => ({ type: 'memo' as const, id: m.id, title: m.title || t('common.untitled'), snippet: '', tags: m.tags?.map(t => t.name) || [], updated_at: m.updated_at })),
        ...(td.data || []).map((todo) => ({ type: 'todo' as const, id: todo.id, title: todo.title, snippet: '', tags: todo.tags?.map(t => t.name) || [], updated_at: todo.updated_at })),
        ...(ev.data || []).map((e) => ({ type: 'event' as const, id: e.id, title: e.title, snippet: '', tags: e.tags?.map(t => t.name) || [], updated_at: e.updated_at })),
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
      case 'tool':
        if (item.snippet) window.open(item.snippet, '_blank')
        break
    }
  }

  const icons: Record<string, typeof FileText> = {
    memo: FileText,
    todo: CheckSquare,
    event: Calendar,
    tool: Monitor,
  }

  const typeLabels: Record<string, string> = {
    memo: t('command.memo'),
    todo: t('command.todo'),
    event: t('command.event'),
    tool: t('command.tool'),
  }

  const displayItems = query.length >= 2 ? results : fallbackItems

  return (
    <Dialog open={commandOpen} onOpenChange={setCommandOpen}>
      <DialogHeader className="sr-only">
        <DialogTitle>Search</DialogTitle>
        <DialogDescription>Search everything</DialogDescription>
      </DialogHeader>
      <DialogContent className="top-1/3 translate-y-0 overflow-hidden rounded-xl! p-0" showCloseButton={false}>
        <Command shouldFilter={false}>
          <CommandInput
            placeholder={t('command.searchPlaceholder')}
            value={query}
            onValueChange={handleQueryChange}
          />
          <CommandList>
            <CommandEmpty>{t('command.noResults')}</CommandEmpty>
            <CommandGroup>
              {displayItems.map((item) => {
                const Icon = icons[item.type] || FileText
                return (
                  <CommandItem key={`${item.type}-${item.id}`} onSelect={() => handleSelect(item)}>
                    <Icon size={14} className={`mr-2 ${sectionColors[item.type] || 'text-muted-foreground'}`} />
                    <span className="text-xs text-muted-foreground mr-2 w-10">{typeLabels[item.type] || item.type}</span>
                    <span className="flex-1 truncate">{item.title}</span>
                    {item.tags?.length > 0 && (
                      <span className="text-xs text-muted-foreground ml-1">
                        {item.tags.map(t => `#${t}`).join(' ')}
                      </span>
                    )}
                    {item.snippet && <span className="text-xs text-muted-foreground truncate ml-2 max-w-40">{item.snippet}</span>}
                  </CommandItem>
                )
              })}
            </CommandGroup>
          </CommandList>
        </Command>
      </DialogContent>
    </Dialog>
  )
}
