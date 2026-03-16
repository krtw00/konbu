import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useCache } from '@/hooks/useCache'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Plus, X, Pencil, Heart, HeartCrack, Loader2 } from 'lucide-react'
import type { Tool } from '@/types/api'

interface HealthResult {
  id: string
  url: string
  alive: boolean
  status: number
}

export function ToolsPage() {
  const { t } = useTranslation()
  const fetchTools = () => api.listTools().then(r => r.data || [] as Tool[])
  const { data: tools_, refresh: load } = useCache('tools', fetchTools)
  const tools = tools_ || []
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingTool, setEditingTool] = useState<Tool | null>(null)
  const [formName, setFormName] = useState('')
  const [formUrl, setFormUrl] = useState('')
  const [formCategory, setFormCategory] = useState('')
  const [faviconPreview, setFaviconPreview] = useState('')
  const [faviconLoading, setFaviconLoading] = useState(false)
  const [healthMap, setHealthMap] = useState<Map<string, HealthResult>>(new Map())
  const [healthLoading, setHealthLoading] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout>>(undefined)
  const dragItem = useRef<string | null>(null)
  const dragOverItem = useRef<string | null>(null)

  function handleDragStart(id: string) {
    dragItem.current = id
  }

  function handleDragOver(e: React.DragEvent, id: string) {
    e.preventDefault()
    dragOverItem.current = id
  }

  async function handleDrop() {
    if (!dragItem.current || !dragOverItem.current || dragItem.current === dragOverItem.current) return
    const allIds = tools.map(t => t.id)
    const fromIdx = allIds.indexOf(dragItem.current)
    const toIdx = allIds.indexOf(dragOverItem.current)
    if (fromIdx === -1 || toIdx === -1) return
    const reordered = [...allIds]
    reordered.splice(fromIdx, 1)
    reordered.splice(toIdx, 0, dragItem.current)
    dragItem.current = null
    dragOverItem.current = null
    await api.reorderTools(reordered)
    load()
  }

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current)
    if (!formUrl || !formUrl.startsWith('http')) {
      setFaviconPreview('')
      return
    }
    setFaviconLoading(true)
    debounceRef.current = setTimeout(async () => {
      try {
        const r = await api.fetchFavicon(formUrl)
        setFaviconPreview(r.data?.icon || '')
      } catch {
        setFaviconPreview('')
      }
      setFaviconLoading(false)
    }, 500)
    return () => { if (debounceRef.current) clearTimeout(debounceRef.current) }
  }, [formUrl])

  function openCreateDialog() {
    setEditingTool(null)
    setFormName('')
    setFormUrl('')
    setFormCategory('')
    setFaviconPreview('')
    setDialogOpen(true)
  }

  function openEditDialog(tool: Tool) {
    setEditingTool(tool)
    setFormName(tool.name)
    setFormUrl(tool.url)
    setFormCategory(tool.category || '')
    setFaviconPreview(tool.icon || '')
    setDialogOpen(true)
  }

  async function saveTool() {
    const body = { name: formName, url: formUrl, category: formCategory || undefined }
    if (editingTool) {
      await api.updateTool(editingTool.id, body)
    } else {
      await api.createTool(body)
    }
    setDialogOpen(false)
    load()
  }

  async function deleteTool(id: string, name: string) {
    if (!confirm(t('tools.confirmDelete', { name }))) return
    await api.deleteTool(id)
    load()
  }

  async function runHealthCheck() {
    setHealthLoading(true)
    try {
      const r = await api.healthCheckTools()
      const map = new Map<string, HealthResult>()
      for (const h of r.data || []) {
        map.set(h.id, h)
      }
      setHealthMap(map)
    } catch {
      // ignore
    }
    setHealthLoading(false)
  }

  // Group tools by category
  const grouped = new Map<string, Tool[]>()
  for (const tool of tools) {
    const cat = tool.category || ''
    if (!grouped.has(cat)) grouped.set(cat, [])
    grouped.get(cat)!.push(tool)
  }
  const categories = [...grouped.keys()].sort((a, b) => {
    if (!a) return 1
    if (!b) return -1
    return a.localeCompare(b)
  })

  const ICON_COLORS = [
    'bg-blue-500/20 text-blue-600',
    'bg-green-500/20 text-green-600',
    'bg-purple-500/20 text-purple-600',
    'bg-orange-500/20 text-orange-600',
    'bg-pink-500/20 text-pink-600',
    'bg-cyan-500/20 text-cyan-600',
  ]

  function healthIcon(toolId: string) {
    const h = healthMap.get(toolId)
    if (!h) return null
    return h.alive
      ? <Heart size={12} className="text-green-500 fill-green-500" />
      : <HeartCrack size={12} className="text-red-500" />
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-semibold">{t('tools.title')}</h1>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={runHealthCheck} disabled={healthLoading}>
            {healthLoading ? <Loader2 size={14} className="mr-1 animate-spin" /> : <Heart size={14} className="mr-1" />}
            {t('tools.healthCheck')}
          </Button>
          <Button size="sm" onClick={openCreateDialog}>
            <Plus size={16} className="mr-1" /> {t('common.new')}
          </Button>
        </div>
      </div>

      {tools.length === 0 ? (
        <p className="text-sm text-muted-foreground py-8 text-center">{t('tools.noTools')}</p>
      ) : (
        <div className="space-y-6">
          {categories.map((cat) => (
            <div key={cat}>
              {cat && <h2 className="text-sm font-medium text-muted-foreground mb-2">{cat}</h2>}
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
                {grouped.get(cat)!.map((tool, i) => (
                  <a
                    key={tool.id}
                    href={tool.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    draggable
                    onDragStart={() => handleDragStart(tool.id)}
                    onDragOver={(e) => handleDragOver(e, tool.id)}
                    onDrop={handleDrop}
                    className="group relative flex flex-col items-center gap-2 p-4 rounded-xl border border-border hover:bg-accent/50 transition-colors"
                  >
                    <div className="absolute top-1 left-1">
                      {healthIcon(tool.id)}
                    </div>
                    {tool.icon ? (
                      <img
                        src={tool.icon}
                        alt={tool.name}
                        className="w-10 h-10 rounded-lg object-contain"
                        onError={(e) => {
                          const el = e.target as HTMLImageElement
                          el.style.display = 'none'
                          el.nextElementSibling?.classList.remove('hidden')
                        }}
                      />
                    ) : null}
                    <div className={`w-10 h-10 rounded-lg flex items-center justify-center text-lg font-semibold ${ICON_COLORS[i % 6]} ${tool.icon ? 'hidden' : ''}`}>
                      {(tool.name || '?')[0].toUpperCase()}
                    </div>
                    <span className="text-sm font-medium text-center truncate w-full">{tool.name}</span>
                    <div className="absolute top-1 right-1 flex gap-0.5 opacity-100 md:opacity-0 md:group-hover:opacity-100 transition-opacity">
                      <button
                        onClick={(e) => { e.preventDefault(); e.stopPropagation(); openEditDialog(tool) }}
                        className="text-muted-foreground hover:text-foreground"
                      >
                        <Pencil size={12} />
                      </button>
                      <button
                        onClick={(e) => { e.preventDefault(); e.stopPropagation(); deleteTool(tool.id, tool.name) }}
                        className="text-muted-foreground hover:text-destructive"
                      >
                        <X size={14} />
                      </button>
                    </div>
                  </a>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingTool ? t('tools.editTool') : t('tools.newTool')}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium">{t('tools.url')}</label>
              <div className="flex items-center gap-2 mt-1">
                <div className="w-8 h-8 shrink-0 flex items-center justify-center">
                  {faviconLoading ? (
                    <Loader2 size={16} className="animate-spin text-muted-foreground" />
                  ) : faviconPreview ? (
                    <img src={faviconPreview} alt="" className="w-6 h-6 rounded object-contain" />
                  ) : (
                    <div className="w-6 h-6 rounded bg-muted" />
                  )}
                </div>
                <Input value={formUrl} onChange={(e) => setFormUrl(e.target.value)} placeholder="https://..." className="flex-1" />
              </div>
            </div>
            <div>
              <label className="text-sm font-medium">{t('tools.name')}</label>
              <Input value={formName} onChange={(e) => setFormName(e.target.value)} placeholder={t('tools.name')} className="mt-1" />
            </div>
            <div>
              <label className="text-sm font-medium">{t('tools.category')}</label>
              <Input value={formCategory} onChange={(e) => setFormCategory(e.target.value)} placeholder={t('tools.categoryPlaceholder')} className="mt-1" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setDialogOpen(false)}>{t('common.cancel')}</Button>
            <Button onClick={saveTool}>{editingTool ? t('common.save') : t('common.create')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
