import { useState, useEffect, useCallback, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { useAppStore } from '@/stores/app'
import { relativeTime } from '@/lib/date'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { FileText, CheckSquare, Calendar, Monitor, Search, ChevronLeft, ChevronRight, SlidersHorizontal, Lightbulb, X } from 'lucide-react'
import { sectionColors } from '@/lib/colors'
import type { SearchResult, SearchResponse, Tag } from '@/types/api'

interface SearchPageProps {
  onOpenMemo: (id: string) => void
}

const typeIcons: Record<string, typeof FileText> = {
  memo: FileText,
  todo: CheckSquare,
  event: Calendar,
  tool: Monitor,
}

const LIMIT = 20

export function SearchPage({ onOpenMemo }: SearchPageProps) {
  const { t } = useTranslation()
  const { searchQuery, setSearchQuery, setPage } = useAppStore()
  const [query, setQuery] = useState(searchQuery)
  const [results, setResults] = useState<SearchResult[]>([])
  const [suggestions, setSuggestions] = useState<SearchResult[]>([])
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(false)
  const [typeFilter, setTypeFilter] = useState<string[]>([])
  const [tagFilter, setTagFilter] = useState('')
  const [allTags, setAllTags] = useState<string[]>([])
  const [showFilters, setShowFilters] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    api.listTags().then(r => {
      setAllTags((r.data || []).map((t: Tag) => t.name))
    }).catch(() => {})
  }, [])

  const doSearch = useCallback(async (q: string, off: number, types: string[], tag: string) => {
    if (q.length < 2) {
      setResults([])
      setSuggestions([])
      setTotal(0)
      return
    }
    setLoading(true)
    try {
      const r: SearchResponse = await api.searchAdvanced({
        q,
        limit: LIMIT,
        offset: off,
        type: types.length > 0 ? types.join(',') : undefined,
        tag: tag || undefined,
      })
      setResults(r.data || [])
      setSuggestions(r.suggestions || [])
      setTotal(r.total || 0)
    } catch {
      setResults([])
      setSuggestions([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (searchQuery && searchQuery.length >= 2) {
      setQuery(searchQuery)
      doSearch(searchQuery, 0, typeFilter, tagFilter)
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  function handleInputChange(val: string) {
    setQuery(val)
    setOffset(0)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      setSearchQuery(val)
      doSearch(val, 0, typeFilter, tagFilter)
    }, 300)
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (debounceRef.current) clearTimeout(debounceRef.current)
    setSearchQuery(query)
    setOffset(0)
    doSearch(query, 0, typeFilter, tagFilter)
  }

  function toggleType(type: string) {
    const next = typeFilter.includes(type)
      ? typeFilter.filter(t => t !== type)
      : [...typeFilter, type]
    setTypeFilter(next)
    setOffset(0)
    doSearch(query, 0, next, tagFilter)
  }

  function handleTagFilter(tag: string) {
    const next = tagFilter === tag ? '' : tag
    setTagFilter(next)
    setOffset(0)
    doSearch(query, 0, typeFilter, next)
  }

  function handlePageChange(newOffset: number) {
    setOffset(newOffset)
    doSearch(query, newOffset, typeFilter, tagFilter)
  }

  function handleSelect(item: SearchResult) {
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

  const typeOptions = [
    { key: 'memo', label: t('command.memo'), icon: FileText },
    { key: 'todo', label: t('command.todo'), icon: CheckSquare },
    { key: 'event', label: t('command.event'), icon: Calendar },
    { key: 'tool', label: t('command.tool'), icon: Monitor },
  ]

  const totalPages = Math.ceil(total / LIMIT)
  const currentPage = Math.floor(offset / LIMIT) + 1

  return (
    <div className="mx-auto max-w-4xl space-y-4 p-4 md:p-6">
      {/* Search bar */}
      <form onSubmit={handleSubmit} className="flex gap-2">
        <div className="relative flex-1">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={query}
            onChange={(e) => handleInputChange(e.target.value)}
            placeholder={t('search.placeholder')}
            className="pl-9"
            autoFocus
          />
        </div>
        <Button
          type="button"
          variant="outline"
          size="icon"
          onClick={() => setShowFilters(!showFilters)}
          className={showFilters ? 'bg-accent' : ''}
        >
          <SlidersHorizontal size={16} />
        </Button>
      </form>

      {/* Filters */}
      {showFilters && (
        <div className="space-y-3 rounded-lg border p-3">
          <div className="flex flex-wrap gap-2">
            <span className="text-sm text-muted-foreground mr-1">{t('search.filterType')}:</span>
            {typeOptions.map(({ key, label, icon: Icon }) => (
              <Button
                key={key}
                variant={typeFilter.includes(key) ? 'default' : 'outline'}
                size="sm"
                onClick={() => toggleType(key)}
                className="h-7 text-xs"
              >
                <Icon size={12} className="mr-1" />
                {label}
              </Button>
            ))}
          </div>
          {allTags.length > 0 && (
            <div className="flex flex-wrap gap-1.5">
              <span className="text-sm text-muted-foreground mr-1">{t('search.filterTag')}:</span>
              {allTags.map(tag => (
                <Badge
                  key={tag}
                  variant={tagFilter === tag ? 'default' : 'outline'}
                  className="cursor-pointer text-xs"
                  onClick={() => handleTagFilter(tag)}
                >
                  {tag}
                  {tagFilter === tag && <X size={10} className="ml-1" />}
                </Badge>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Results count */}
      {query.length >= 2 && !loading && (
        <p className="text-sm text-muted-foreground">
          {t('search.resultCount', { count: total })}
        </p>
      )}

      {/* Results list */}
      {loading ? (
        <div className="py-12 text-center text-muted-foreground">{t('search.loading')}</div>
      ) : (
        <div className="space-y-1">
          {results.map(item => {
            const Icon = typeIcons[item.type] || FileText
            return (
              <button
                key={`${item.type}-${item.id}`}
                onClick={() => handleSelect(item)}
                className="flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-left transition-colors hover:bg-accent"
              >
                <Icon size={16} className={`shrink-0 ${sectionColors[item.type] || 'text-muted-foreground'}`} />
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-muted-foreground w-12">{t(`command.${item.type}`) || item.type}</span>
                    <span className="truncate font-medium">{item.title}</span>
                  </div>
                  {item.snippet && item.type !== 'tool' && (
                    <p className="mt-0.5 truncate text-xs text-muted-foreground">{item.snippet}</p>
                  )}
                  {item.type === 'tool' && item.snippet && (
                    <p className="mt-0.5 truncate text-xs text-blue-500">{item.snippet}</p>
                  )}
                </div>
                {item.tags?.length > 0 && (
                  <div className="hidden shrink-0 gap-1 sm:flex">
                    {item.tags.slice(0, 3).map(tag => (
                      <Badge key={tag} variant="secondary" className="text-xs">{tag}</Badge>
                    ))}
                  </div>
                )}
                <span className="shrink-0 text-xs text-muted-foreground">{relativeTime(item.updated_at)}</span>
              </button>
            )
          })}

          {results.length === 0 && query.length >= 2 && !loading && (
            <div className="py-12 text-center text-muted-foreground">{t('search.noResults')}</div>
          )}
        </div>
      )}

      {/* Suggestions ("Did you mean?") */}
      {suggestions.length > 0 && (
        <div className="mt-4 rounded-lg border p-3">
          <div className="mb-2 flex items-center gap-1.5 text-sm text-muted-foreground">
            <Lightbulb size={14} />
            {t('search.didYouMean')}
          </div>
          <div className="space-y-1">
            {suggestions.map(item => {
              const Icon = typeIcons[item.type] || FileText
              return (
                <button
                  key={`sug-${item.type}-${item.id}`}
                  onClick={() => handleSelect(item)}
                  className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors hover:bg-accent"
                >
                  <Icon size={14} className={sectionColors[item.type] || 'text-muted-foreground'} />
                  <span className="text-xs text-muted-foreground w-10">{t(`command.${item.type}`) || item.type}</span>
                  <span className="truncate">{item.title}</span>
                </button>
              )
            })}
          </div>
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 pt-2">
          <Button
            variant="outline"
            size="sm"
            disabled={offset === 0}
            onClick={() => handlePageChange(offset - LIMIT)}
          >
            <ChevronLeft size={14} />
          </Button>
          <span className="text-sm text-muted-foreground">
            {currentPage} / {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={offset + LIMIT >= total}
            onClick={() => handlePageChange(offset + LIMIT)}
          >
            <ChevronRight size={14} />
          </Button>
        </div>
      )}
    </div>
  )
}
