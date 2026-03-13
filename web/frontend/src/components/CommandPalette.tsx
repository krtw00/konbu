import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useAppStore } from '@/stores/app'
import { CommandDialog, CommandInput, CommandList, CommandItem, CommandGroup, CommandEmpty } from '@/components/ui/command'
import { FileText, CheckSquare, Calendar } from 'lucide-react'

interface CmdItem {
  type: 'memo' | 'todo' | 'event'
  title: string
  id: string
}

interface CommandPaletteProps {
  onOpenMemo: (id: string) => void
}

export function CommandPalette({ onOpenMemo }: CommandPaletteProps) {
  const { t } = useTranslation()
  const { commandOpen, setCommandOpen, setPage } = useAppStore()
  const [items, setItems] = useState<CmdItem[]>([])

  useEffect(() => {
    if (!commandOpen) return
    async function load() {
      const [mm, td, ev] = await Promise.all([
        api.listMemos(50),
        api.listTodos(100),
        api.listEvents(30),
      ])
      setItems([
        ...(mm.data || []).map((m) => ({ type: 'memo' as const, title: m.title || t('common.untitled'), id: m.id })),
        ...(td.data || []).map((todo) => ({ type: 'todo' as const, title: todo.title, id: todo.id })),
        ...(ev.data || []).map((e) => ({ type: 'event' as const, title: e.title, id: e.id })),
      ])
    }
    load()
  }, [commandOpen, t])

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

  function handleSelect(item: CmdItem) {
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

  return (
    <CommandDialog open={commandOpen} onOpenChange={setCommandOpen}>
      <CommandInput placeholder={t('command.searchPlaceholder')} />
      <CommandList>
        <CommandEmpty>{t('command.noResults')}</CommandEmpty>
        <CommandGroup>
          {items.map((item) => {
            const Icon = icons[item.type]
            return (
              <CommandItem key={`${item.type}-${item.id}`} onSelect={() => handleSelect(item)}>
                <Icon size={14} className="mr-2 text-muted-foreground" />
                <span className="text-xs text-muted-foreground mr-2 w-10">{typeLabels[item.type]}</span>
                {item.title}
              </CommandItem>
            )
          })}
        </CommandGroup>
      </CommandList>
    </CommandDialog>
  )
}
