import { useEffect, useState, useCallback } from 'react'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Plus, X } from 'lucide-react'
import type { Tool } from '@/types/api'

export function ToolsPage() {
  const [tools, setTools] = useState<Tool[]>([])
  const [dialogOpen, setDialogOpen] = useState(false)
  const [newName, setNewName] = useState('')
  const [newUrl, setNewUrl] = useState('')
  const [newIcon, setNewIcon] = useState('')

  const load = useCallback(async () => {
    const r = await api.listTools()
    setTools(r.data || [])
  }, [])

  useEffect(() => { load() }, [load])

  async function createTool() {
    await api.createTool({ name: newName, url: newUrl, icon: newIcon })
    setDialogOpen(false)
    setNewName('')
    setNewUrl('')
    setNewIcon('')
    load()
  }

  async function deleteTool(id: string, name: string) {
    if (!confirm(`Delete "${name}"?`)) return
    await api.deleteTool(id)
    load()
  }

  const ICON_COLORS = [
    'bg-blue-500/20 text-blue-600',
    'bg-green-500/20 text-green-600',
    'bg-purple-500/20 text-purple-600',
    'bg-orange-500/20 text-orange-600',
    'bg-pink-500/20 text-pink-600',
    'bg-cyan-500/20 text-cyan-600',
  ]

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-semibold">Tools</h1>
        <Button size="sm" onClick={() => setDialogOpen(true)}>
          <Plus size={16} className="mr-1" /> New
        </Button>
      </div>

      {tools.length === 0 ? (
        <p className="text-sm text-muted-foreground py-8 text-center">No tools yet</p>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
          {tools.map((t, i) => (
            <a
              key={t.id}
              href={t.url}
              target="_blank"
              rel="noopener noreferrer"
              className="group relative flex flex-col items-center gap-2 p-4 rounded-xl border border-border hover:bg-accent/50 transition-colors"
            >
              {t.icon ? (
                <img
                  src={t.icon}
                  alt={t.name}
                  className="w-10 h-10 rounded-lg object-contain"
                  onError={(e) => {
                    const el = e.target as HTMLImageElement
                    el.style.display = 'none'
                    el.nextElementSibling?.classList.remove('hidden')
                  }}
                />
              ) : null}
              <div className={`w-10 h-10 rounded-lg flex items-center justify-center text-lg font-semibold ${ICON_COLORS[i % 6]} ${t.icon ? 'hidden' : ''}`}>
                {(t.name || '?')[0].toUpperCase()}
              </div>
              <span className="text-sm font-medium text-center truncate w-full">{t.name}</span>
              <button
                onClick={(e) => { e.preventDefault(); e.stopPropagation(); deleteTool(t.id, t.name) }}
                className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-opacity"
              >
                <X size={14} />
              </button>
            </a>
          ))}
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New Tool</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium">Name</label>
              <Input value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="Name" className="mt-1" />
            </div>
            <div>
              <label className="text-sm font-medium">URL</label>
              <Input value={newUrl} onChange={(e) => setNewUrl(e.target.value)} placeholder="https://..." className="mt-1" />
            </div>
            <div>
              <label className="text-sm font-medium">Icon</label>
              <Input value={newIcon} onChange={(e) => setNewIcon(e.target.value)} placeholder="Emoji or letter" className="mt-1" />
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setDialogOpen(false)}>Cancel</Button>
            <Button onClick={createTool}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
