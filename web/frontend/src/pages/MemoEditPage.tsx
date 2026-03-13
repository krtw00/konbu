import { useEffect, useState, useRef, useCallback } from 'react'
import { api } from '@/lib/api'
import { useAppStore } from '@/stores/app'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import CodeMirror from '@uiw/react-codemirror'
import { markdown } from '@codemirror/lang-markdown'
import { oneDark } from '@codemirror/theme-one-dark'
import { marked } from 'marked'
import { ArrowLeft, Tag, Eye, Edit3, Trash2, Plus } from 'lucide-react'
import type { Memo } from '@/types/api'

interface MemoEditPageProps {
  memoId: string
  onClose: () => void
}

export function MemoEditPage({ memoId, onClose }: MemoEditPageProps) {
  const theme = useAppStore((s) => s.theme)
  const [memo, setMemo] = useState<Memo | null>(null)
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [allTags, setAllTags] = useState<string[]>([])
  const [preview, setPreview] = useState(false)
  const [status, setStatus] = useState('')
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    async function load() {
      const [r, t] = await Promise.all([api.getMemo(memoId), api.listTags()])
      const m = r.data
      setMemo(m)
      setTitle(m.title)
      setContent(m.content)
      setTags(m.tags?.map((t) => t.name) || [])
      setAllTags((t.data || []).map((tag) => tag.name))
    }
    load()
  }, [memoId])

  const save = useCallback(async (t?: string, c?: string, tg?: string[]) => {
    const saveTitle = t ?? title
    const saveContent = c ?? content
    const saveTags = tg ?? tags
    try {
      await api.updateMemo(memoId, { title: saveTitle, content: saveContent, tags: saveTags })
      setStatus('Saved')
      setTimeout(() => setStatus(''), 2000)
    } catch {
      setStatus('Error saving')
    }
  }, [memoId, title, content, tags])

  function scheduleAutosave(newTitle?: string, newContent?: string) {
    setStatus('Editing...')
    if (saveTimer.current) clearTimeout(saveTimer.current)
    saveTimer.current = setTimeout(() => save(newTitle, newContent), 1500)
  }

  function handleTitleChange(val: string) {
    setTitle(val)
    scheduleAutosave(val, undefined)
  }

  function handleContentChange(val: string) {
    setContent(val)
    scheduleAutosave(undefined, val)
  }

  function toggleTag(name: string) {
    const next = tags.includes(name) ? tags.filter((n) => n !== name) : [...tags, name]
    setTags(next)
    save(undefined, undefined, next)
  }

  async function handleDelete() {
    if (!confirm('Delete this memo?')) return
    await api.deleteMemo(memoId)
    onClose()
  }

  // Save on unmount
  useEffect(() => {
    return () => {
      if (saveTimer.current) clearTimeout(saveTimer.current)
    }
  }, [])

  const isDark = ['konbu', 'mocha', 'linear', 'nord'].includes(theme)

  if (!memo) return null

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-4 py-2 border-b border-border">
        <Button variant="ghost" size="sm" onClick={() => { save(); onClose() }}>
          <ArrowLeft size={16} className="mr-1" /> Memos
        </Button>
        <input
          type="text"
          value={title}
          onChange={(e) => handleTitleChange(e.target.value)}
          placeholder="Untitled"
          className="flex-1 bg-transparent text-lg font-medium outline-none"
        />
        <span className="text-xs text-muted-foreground">{status}</span>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="sm">
              <Tag size={14} className="mr-1" />
              Tags{tags.length > 0 && ` (${tags.length})`}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {allTags.map((name) => (
              <DropdownMenuItem key={name} onClick={() => toggleTag(name)}>
                {tags.includes(name) ? '✓ ' : ''}{name}
              </DropdownMenuItem>
            ))}
            {allTags.length === 0 && (
              <DropdownMenuItem disabled>No tags</DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
        <Button variant="ghost" size="sm" onClick={() => setPreview(!preview)}>
          {preview ? <Edit3 size={14} /> : <Eye size={14} />}
          <span className="ml-1">{preview ? 'Edit' : 'Preview'}</span>
        </Button>
        <Button variant="ghost" size="sm" className="text-destructive" onClick={handleDelete}>
          <Trash2 size={14} />
        </Button>
      </div>

      {/* Tags bar */}
      {tags.length > 0 && (
        <div className="flex items-center gap-1.5 px-4 py-1.5 border-b border-border">
          {tags.map((name) => (
            <Badge key={name} variant="secondary" className="text-xs">
              {name}
              <button className="ml-1 hover:text-destructive" onClick={() => toggleTag(name)}>
                x
              </button>
            </Badge>
          ))}
        </div>
      )}

      {/* Editor / Preview */}
      <div className="flex-1 overflow-hidden flex">
        {!preview ? (
          <div className="flex-1 overflow-auto">
            <CodeMirror
              value={content}
              onChange={handleContentChange}
              extensions={[markdown()]}
              theme={isDark ? oneDark : undefined}
              className="h-full"
              basicSetup={{
                lineNumbers: false,
                foldGutter: false,
              }}
            />
          </div>
        ) : (
          <div
            className="flex-1 overflow-auto p-6 prose prose-sm dark:prose-invert max-w-none"
            dangerouslySetInnerHTML={{ __html: marked.parse(content || '') as string }}
          />
        )}
      </div>
    </div>
  )
}
