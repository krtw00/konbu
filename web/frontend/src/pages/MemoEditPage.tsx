import { useEffect, useState, useRef, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import Editor from '@monaco-editor/react'
import type * as Monaco from 'monaco-editor'
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'

if (!window.MonacoEnvironment) {
  window.MonacoEnvironment = {
    getWorker: () => new editorWorker(),
  }
}
import { renderMarkdown } from '@/lib/markdown'
import { registerMarkdownFeatures } from '@/lib/monaco-markdown'
import { ArrowLeft, Tag, Trash2, Eye, EyeOff, Bold, Italic, Strikethrough, Code, Link, List, ListOrdered, CheckSquare, Heading1, Heading2, Heading3, Quote, Minus, Table, ImageIcon } from 'lucide-react'
import type { Memo } from '@/types/api'

interface MemoEditPageProps {
  memoId: string
  onClose: () => void
}

export function MemoEditPage({ memoId, onClose }: MemoEditPageProps) {
  const { t } = useTranslation()
  const [memo, setMemo] = useState<Memo | null>(null)
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [allTags, setAllTags] = useState<string[]>([])
  const [status, setStatus] = useState('')
  const [showPreview, setShowPreview] = useState(false)
  const editorRef = useRef<Monaco.editor.IStandaloneCodeEditor | null>(null)
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    async function load() {
      const [r, tRes] = await Promise.all([api.getMemo(memoId), api.listTags()])
      const m = r.data
      setMemo(m)
      setTitle(m.title)
      setContent(m.content)
      setTags(m.tags?.map((tag) => tag.name) || [])
      setAllTags((tRes.data || []).map((tag) => tag.name))
    }
    load()
  }, [memoId])

  const save = useCallback(async (ti?: string, c?: string, tg?: string[]) => {
    const saveTitle = ti ?? title
    const saveContent = c ?? content
    const saveTags = tg ?? tags
    try {
      await api.updateMemo(memoId, { title: saveTitle, content: saveContent, tags: saveTags })
      setStatus(t('memoEdit.saved'))
      setTimeout(() => setStatus(''), 2000)
    } catch {
      setStatus(t('memoEdit.errorSaving'))
    }
  }, [memoId, title, content, tags, t])

  function scheduleAutosave(newTitle?: string, newContent?: string) {
    setStatus(t('memoEdit.editing'))
    if (saveTimer.current) clearTimeout(saveTimer.current)
    saveTimer.current = setTimeout(() => save(newTitle, newContent), 1500)
  }

  function handleTitleChange(val: string) {
    setTitle(val)
    scheduleAutosave(val, undefined)
  }

  function handleContentChange(val: string | undefined) {
    const v = val ?? ''
    setContent(v)
    scheduleAutosave(undefined, v)
  }

  function toggleTag(name: string) {
    const next = tags.includes(name) ? tags.filter((n) => n !== name) : [...tags, name]
    setTags(next)
    save(undefined, undefined, next)
  }

  async function handleDelete() {
    if (!confirm(t('memoEdit.confirmDelete'))) return
    await api.deleteMemo(memoId)
    onClose()
  }

  const [uploading, setUploading] = useState(false)

  async function handleImageUpload(file: File) {
    if (!file.type.startsWith('image/')) return
    if (file.size > 5 * 1024 * 1024) {
      setStatus(t('attachment.tooLarge'))
      return
    }
    setUploading(true)
    setStatus(t('attachment.uploading'))
    try {
      const res = await api.uploadAttachment(file)
      insertText(`![${file.name}](${res.data.url})`)
    } catch (e) {
      const msg = e instanceof Error ? e.message : t('attachment.error')
      setStatus(msg)
    } finally {
      setUploading(false)
      setTimeout(() => setStatus(''), 3000)
    }
  }

  function insertText(text: string) {
    const ed = editorRef.current
    if (!ed) return
    const sel = ed.getSelection()
    if (!sel) return
    ed.executeEdits('toolbar', [{ range: sel, text }])
    ed.focus()
  }

  function wrapSelection(wrapper: string) {
    const ed = editorRef.current
    if (!ed) return
    const sel = ed.getSelection()
    if (!sel) return
    const model = ed.getModel()
    if (!model) return
    const selected = model.getValueInRange(sel)
    if (selected.startsWith(wrapper) && selected.endsWith(wrapper)) {
      ed.executeEdits('unwrap', [{ range: sel, text: selected.slice(wrapper.length, -wrapper.length) }])
    } else {
      ed.executeEdits('wrap', [{ range: sel, text: `${wrapper}${selected}${wrapper}` }])
    }
    ed.focus()
  }

  function insertLinePrefix(prefix: string) {
    const ed = editorRef.current
    if (!ed) return
    const pos = ed.getPosition()
    if (!pos) return
    const model = ed.getModel()
    if (!model) return
    const line = model.getLineContent(pos.lineNumber)
    ed.executeEdits('prefix', [{
      range: { startLineNumber: pos.lineNumber, startColumn: 1, endLineNumber: pos.lineNumber, endColumn: line.length + 1 },
      text: `${prefix}${line}`,
    }])
    ed.focus()
  }

  useEffect(() => {
    return () => {
      if (saveTimer.current) clearTimeout(saveTimer.current)
    }
  }, [])

  // Handle memo link clicks
  function handlePreviewClick(e: React.MouseEvent) {
    const target = e.target as HTMLElement
    const link = target.closest('[data-memo-link]') as HTMLElement | null
    if (link) {
      e.preventDefault()
      const memoName = link.getAttribute('data-memo-link')
      if (memoName) {
        // Find memo by title and navigate
        api.listMemos(100).then((r) => {
          const found = (r.data || []).find((m) => m.title === memoName)
          if (found) {
            onClose()
            setTimeout(() => {
              // Re-open with found memo
              window.dispatchEvent(new CustomEvent('open-memo', { detail: found.id }))
            }, 50)
          }
        })
      }
    }
  }

  if (!memo) return null

  const lineCount = content.split('\n').length
  const charCount = content.length

  return (
    <div className="relative flex flex-col h-full overflow-hidden">
      {/* Compact toolbar */}
      <div className="flex items-center gap-1 px-2 py-0.5 border-b border-border shrink-0" style={{ minHeight: 32 }}>
        <Button variant="ghost" size="sm" className="h-6 px-1.5 text-xs" onClick={() => { save(); onClose() }}>
          <ArrowLeft size={14} />
        </Button>
        <input
          type="text"
          value={title}
          onChange={(e) => handleTitleChange(e.target.value)}
          placeholder={t('common.untitled')}
          className="flex-1 bg-transparent text-xs font-medium outline-none min-w-0"
        />
        {tags.map((name) => (
          <Badge key={name} variant="secondary" className="text-[10px] h-5 px-1.5 shrink-0">
            {name}
            <button className="ml-0.5 hover:text-destructive" onClick={() => toggleTag(name)}>×</button>
          </Badge>
        ))}
        <span className="text-[10px] text-muted-foreground shrink-0">{status}</span>
        <DropdownMenu>
          <DropdownMenuTrigger>
            <Button variant="ghost" size="sm" className="h-6 px-1.5">
              <Tag size={12} />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {allTags.map((name) => (
              <DropdownMenuItem key={name} onClick={() => toggleTag(name)}>
                {tags.includes(name) ? '✓ ' : ''}{name}
              </DropdownMenuItem>
            ))}
            {allTags.length === 0 && (
              <DropdownMenuItem disabled>{t('common.noTags')}</DropdownMenuItem>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
        <Button variant="ghost" size="sm" className="h-6 px-1.5 text-destructive" onClick={handleDelete}>
          <Trash2 size={12} />
        </Button>
      </div>

      {/* Format toolbar */}
      <div className="flex items-center gap-0.5 px-1.5 py-0.5 border-b border-border bg-muted/50 shrink-0 overflow-x-auto" style={{ minHeight: 30 }}>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="Bold" onClick={() => wrapSelection('**')}><Bold size={13} /></Button>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="Italic" onClick={() => wrapSelection('*')}><Italic size={13} /></Button>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="Strikethrough" onClick={() => wrapSelection('~~')}><Strikethrough size={13} /></Button>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="Code" onClick={() => wrapSelection('`')}><Code size={13} /></Button>
        <div className="w-px h-4 bg-border mx-0.5" />
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="H1" onClick={() => insertLinePrefix('# ')}><Heading1 size={13} /></Button>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="H2" onClick={() => insertLinePrefix('## ')}><Heading2 size={13} /></Button>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="H3" onClick={() => insertLinePrefix('### ')}><Heading3 size={13} /></Button>
        <div className="w-px h-4 bg-border mx-0.5" />
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="Bullet list" onClick={() => insertLinePrefix('- ')}><List size={13} /></Button>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="Numbered list" onClick={() => insertLinePrefix('1. ')}><ListOrdered size={13} /></Button>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="Task list" onClick={() => insertLinePrefix('- [ ] ')}><CheckSquare size={13} /></Button>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="Blockquote" onClick={() => insertLinePrefix('> ')}><Quote size={13} /></Button>
        <div className="w-px h-4 bg-border mx-0.5" />
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="Link" onClick={() => insertText('[text](url)')}><Link size={13} /></Button>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title={t('attachment.upload')} onClick={() => {
          const input = document.createElement('input')
          input.type = 'file'
          input.accept = 'image/*'
          input.onchange = () => { if (input.files?.[0]) handleImageUpload(input.files[0]) }
          input.click()
        }} disabled={uploading}><ImageIcon size={13} /></Button>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="Table" onClick={() => insertText('| Header | Header |\n| --- | --- |\n| Cell | Cell |')}><Table size={13} /></Button>
        <Button variant="ghost" size="sm" className="h-6 w-6 p-0" title="Horizontal rule" onClick={() => insertText('\n---\n')}><Minus size={13} /></Button>
        <div className="flex-1" />
        <Button
          variant={showPreview ? 'default' : 'outline'}
          size="sm"
          className="h-6 px-2.5 text-[11px] gap-1.5 shrink-0"
          onClick={() => setShowPreview((v) => !v)}
        >
          {showPreview ? <EyeOff size={14} /> : <Eye size={14} />}
          {showPreview ? 'Preview OFF' : 'Preview'}
        </Button>
      </div>

      {/* Editor + Preview — full remaining space */}
      <div className="flex-1 flex overflow-hidden min-h-0">
        <div className="flex-1 min-w-0">
          <Editor
            value={content}
            onChange={handleContentChange}
            language="markdown"
            theme="light"
            onMount={(editor: Monaco.editor.IStandaloneCodeEditor, monaco: typeof Monaco) => {
              editorRef.current = editor
              registerMarkdownFeatures(monaco, editor)

              const domNode = editor.getDomNode()
              if (domNode) {
                domNode.addEventListener('dragover', (e) => { e.preventDefault() })
                domNode.addEventListener('drop', (e) => {
                  e.preventDefault()
                  const files = e.dataTransfer?.files
                  if (files?.length) {
                    for (const f of Array.from(files)) {
                      if (f.type.startsWith('image/')) handleImageUpload(f)
                    }
                  }
                })
                domNode.addEventListener('paste', (e) => {
                  const items = e.clipboardData?.items
                  if (!items) return
                  for (const item of Array.from(items)) {
                    if (item.type.startsWith('image/')) {
                      e.preventDefault()
                      const file = item.getAsFile()
                      if (file) handleImageUpload(file)
                      break
                    }
                  }
                })
              }
            }}
            options={{
              minimap: { enabled: true, scale: 1 },
              wordWrap: 'on',
              fontSize: 15,
              lineHeight: 22,
              padding: { top: 4, bottom: 4 },
              scrollBeyondLastLine: false,
              renderWhitespace: 'none',
              automaticLayout: true,
              tabSize: 2,
              smoothScrolling: true,
              cursorBlinking: 'smooth',
              cursorSmoothCaretAnimation: 'on',
              lineNumbersMinChars: 3,
              glyphMargin: false,
              folding: true,
              quickSuggestions: true,
              suggestOnTriggerCharacters: true,
            }}
          />
        </div>

        {/* PC: side-by-side */}
        {showPreview && (
          <>
            <div className="w-px bg-border shrink-0 hidden md:block" />
            <div
              className="hidden md:block flex-1 min-w-0 overflow-auto px-3 py-2 md-preview"
              onClick={handlePreviewClick}
              dangerouslySetInnerHTML={{ __html: renderMarkdown(content) }}
            />
          </>
        )}
      </div>

      {/* Mobile: fullscreen overlay */}
      {showPreview && (
        <div className="md:hidden absolute inset-0 z-50 bg-background flex flex-col">
          <div className="flex items-center gap-2 px-3 py-2 border-b border-border shrink-0">
            <span className="text-sm font-medium flex-1">Preview</span>
            <Button variant="outline" size="sm" className="h-7 px-3 text-xs" onClick={() => setShowPreview(false)}>
              Close
            </Button>
          </div>
          <div
            className="flex-1 overflow-auto px-4 py-3 md-preview"
            onClick={handlePreviewClick}
            dangerouslySetInnerHTML={{ __html: renderMarkdown(content) }}
          />
        </div>
      )}

      {/* Status bar */}
      <div className="flex items-center gap-3 px-2 py-0.5 border-t border-border text-[10px] text-muted-foreground bg-muted shrink-0" style={{ minHeight: 22 }}>
        <span>Ln {lineCount}, Col 1</span>
        <span>{charCount} characters</span>
        <div className="flex-1" />
        <span>Markdown</span>
        <span>UTF-8</span>
      </div>
    </div>
  )
}
