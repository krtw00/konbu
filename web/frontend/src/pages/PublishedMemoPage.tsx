import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { renderMarkdown } from '@/lib/markdown'
import type { PublishedMemoView } from '@/types/api'

interface PublishedMemoPageProps {
  slug: string
}

function formatDateTime(value?: string | null) {
  if (!value) return ''
  return new Date(value).toLocaleString()
}

function upsertMeta(selector: string, attrs: Record<string, string>) {
  let element = document.head.querySelector<HTMLMetaElement>(selector)
  if (!element) {
    element = document.createElement('meta')
    document.head.appendChild(element)
  }
  for (const [key, value] of Object.entries(attrs)) {
    element.setAttribute(key, value)
  }
  return element
}

export function PublishedMemoPage({ slug }: PublishedMemoPageProps) {
  const { t } = useTranslation()
  const [view, setView] = useState<PublishedMemoView | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    api.getPublishedMemoView(slug)
      .then((res) => {
        if (!cancelled) setView(res.data)
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : 'Not found')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => { cancelled = true }
  }, [slug])

  useEffect(() => {
    if (!view) return

    const previousTitle = document.title
    const descriptionMeta = document.head.querySelector<HTMLMetaElement>('meta[name="description"]')
    const previousDescription = descriptionMeta?.getAttribute('content') ?? null
    const title = view.publish.title || view.memo.title || 'konbu'
    const description = view.publish.description || ''

    document.title = `${title} | konbu`
    if (description) {
      upsertMeta('meta[name="description"]', { name: 'description', content: description })
      upsertMeta('meta[property="og:title"]', { property: 'og:title', content: title })
      upsertMeta('meta[property="og:description"]', { property: 'og:description', content: description })
    }

    return () => {
      document.title = previousTitle
      if (descriptionMeta) {
        if (previousDescription === null) {
          descriptionMeta.remove()
        } else {
          descriptionMeta.setAttribute('content', previousDescription)
        }
      }
    }
  }, [view])

  if (loading) {
    return <div className="min-h-screen bg-background" />
  }

  if (error || !view) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center p-6">
        <div className="max-w-md text-center space-y-2">
          <h1 className="text-2xl font-semibold">{t('publicShare.unavailable')}</h1>
          <p className="text-sm text-muted-foreground">{error || t('publicShare.unavailableDescription')}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <main className="max-w-4xl mx-auto px-4 py-10 md:px-6">
        <div className="mb-8">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">konbu</p>
        </div>

        <section className="space-y-6">
          <div className="space-y-2">
            <h1 className="text-3xl font-semibold">{view.publish.title || view.memo.title || t('common.untitled')}</h1>
            {view.publish.description && (
              <p className="text-sm text-muted-foreground">{view.publish.description}</p>
            )}
            <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
              {view.memo.tags?.map((tag) => <span key={tag.id}>#{tag.name}</span>)}
              <span>{formatDateTime(view.memo.updated_at)}</span>
            </div>
          </div>

          {view.memo.type === 'table' ? (
            <div className="overflow-auto rounded-xl border border-border">
              <table className="w-full text-sm">
                <thead className="bg-muted/50">
                  <tr>
                    {view.memo.table_columns?.map((column) => (
                      <th key={column.id} className="px-4 py-3 text-left font-medium">{column.name}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {view.memo.rows?.map((row) => (
                    <tr key={row.id} className="border-t border-border">
                      {view.memo.table_columns?.map((column) => {
                        const value = row.row_data?.[column.id] || ''
                        return <td key={column.id} className="px-4 py-3 align-top">{value}</td>
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <article
              className="prose prose-neutral max-w-none dark:prose-invert"
              dangerouslySetInnerHTML={{ __html: renderMarkdown(view.memo.content || '') }}
            />
          )}
        </section>
      </main>
    </div>
  )
}
