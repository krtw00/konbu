import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Link2, Copy, Trash2, ExternalLink } from 'lucide-react'
import { api } from '@/lib/api'
import { appURL } from '@/lib/runtime'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import type { PublicResourceType, PublicShare } from '@/types/api'

interface PublicShareDialogProps {
  resourceType: PublicResourceType
  resourceId: string
  disabled?: boolean
  triggerLabel?: string
  iconOnly?: boolean
  size?: 'sm' | 'default'
  variant?: 'default' | 'outline' | 'ghost'
}

export function PublicShareDialog({
  resourceType,
  resourceId,
  disabled,
  triggerLabel,
  iconOnly = false,
  size = 'sm',
  variant = 'outline',
}: PublicShareDialogProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [share, setShare] = useState<PublicShare | null>(null)
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState('')

  const publicURL = useMemo(() => {
    if (!share) return ''
    return appURL(`/public/${share.token}`)
  }, [share])

  useEffect(() => {
    if (!open || !resourceId) return
    let cancelled = false
    setLoading(true)
    setMessage('')
    api.getPublicShare(resourceType, resourceId)
      .then((res) => {
        if (!cancelled) setShare(res.data)
      })
      .catch((error) => {
        if (!cancelled) {
          setShare(null)
          setMessage(error instanceof Error ? error.message : t('publicShare.loadError'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => { cancelled = true }
  }, [open, resourceId, resourceType, t])

  async function handleCreate() {
    setLoading(true)
    try {
      const res = await api.createPublicShare(resourceType, resourceId)
      setShare(res.data)
      setMessage('')
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t('publicShare.publishError'))
    } finally {
      setLoading(false)
    }
  }

  async function handleDelete() {
    setLoading(true)
    try {
      await api.deletePublicShare(resourceType, resourceId)
      setShare(null)
      setMessage('')
    } catch (error) {
      setMessage(error instanceof Error ? error.message : t('publicShare.deleteError'))
    } finally {
      setLoading(false)
    }
  }

  async function handleCopy() {
    if (!publicURL) return
    await navigator.clipboard.writeText(publicURL)
    setMessage(t('publicShare.copied'))
    setTimeout(() => setMessage(''), 2000)
  }

  function handleOpenPublicPage() {
    if (!publicURL) return
    window.open(publicURL, '_blank', 'noopener,noreferrer')
  }

  return (
    <>
      <Button type="button" size={size} variant={variant} disabled={disabled} onClick={() => setOpen(true)}>
        <Link2 size={14} className={iconOnly ? '' : 'mr-1'} />
        {!iconOnly && (triggerLabel || t('publicShare.button'))}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{t('publicShare.title')}</DialogTitle>
            <DialogDescription>{t('publicShare.description')}</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            {share ? (
              <>
                <label className="text-sm font-medium">{t('publicShare.url')}</label>
                <div className="space-y-2">
                  <div className="max-w-full rounded-md bg-muted px-3 py-2 text-xs leading-relaxed font-mono break-all">
                    {publicURL}
                  </div>
                  <div className="flex flex-wrap items-center justify-end gap-2">
                    <Button size="sm" variant="outline" onClick={handleCopy}>
                      <Copy size={12} className="mr-1" />
                      {t('common.copy')}
                    </Button>
                    <Button size="sm" variant="outline" onClick={handleOpenPublicPage}>
                      <ExternalLink size={12} className="mr-1" />
                      {t('common.open')}
                    </Button>
                    <Button size="sm" variant="ghost" className="text-destructive" onClick={handleDelete}>
                      <Trash2 size={12} className="mr-1" />
                      {t('common.delete')}
                    </Button>
                  </div>
                </div>
              </>
            ) : (
              <div className="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
                {t('publicShare.notPublished')}
              </div>
            )}
            {message && <p className="text-xs text-muted-foreground">{message}</p>}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setOpen(false)}>{t('common.cancel')}</Button>
            {!share && <Button onClick={handleCreate} disabled={loading}>{t('publicShare.publish')}</Button>}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
