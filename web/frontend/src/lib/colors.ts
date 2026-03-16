import { Home, FileText, CheckSquare, Calendar, Monitor, Search, MessageCircle, Settings } from 'lucide-react'

export const sectionColors: Record<string, string> = {
  home: 'text-orange-500',
  memos: 'text-blue-500',
  memo: 'text-blue-500',
  'memo-edit': 'text-blue-500',
  todos: 'text-emerald-500',
  todo: 'text-emerald-500',
  calendar: 'text-rose-500',
  event: 'text-rose-500',
  tools: 'text-violet-500',
  tool: 'text-violet-500',
  search: 'text-amber-500',
  chat: 'text-cyan-500',
  settings: 'text-zinc-400',
}

export const sectionBgColors: Record<string, string> = {
  home: 'bg-orange-500/10',
  memos: 'bg-blue-500/10',
  memo: 'bg-blue-500/10',
  'memo-edit': 'bg-blue-500/10',
  todos: 'bg-emerald-500/10',
  todo: 'bg-emerald-500/10',
  calendar: 'bg-rose-500/10',
  event: 'bg-rose-500/10',
  tools: 'bg-violet-500/10',
  tool: 'bg-violet-500/10',
  search: 'bg-amber-500/10',
  chat: 'bg-cyan-500/10',
  settings: 'bg-zinc-500/10',
}

export const sectionIcons: Record<string, typeof Home> = {
  home: Home,
  memos: FileText,
  memo: FileText,
  'memo-edit': FileText,
  todos: CheckSquare,
  todo: CheckSquare,
  calendar: Calendar,
  event: Calendar,
  tools: Monitor,
  tool: Monitor,
  search: Search,
  chat: MessageCircle,
  settings: Settings,
}
