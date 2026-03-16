import { useState, useEffect, useCallback, useRef } from 'react'

const cache = new Map<string, { data: unknown; ts: number }>()
const listeners = new Map<string, Set<() => void>>()

export function useCache<T>(key: string, fetcher: () => Promise<T>, maxAge = 300000): {
  data: T | null
  loading: boolean
  refresh: () => Promise<void>
} {
  const cached = cache.get(key)
  const [data, setData] = useState<T | null>((cached?.data as T) ?? null)
  const [loading, setLoading] = useState(!cached)
  const mountedRef = useRef(true)
  const fetcherRef = useRef(fetcher)
  fetcherRef.current = fetcher

  const refresh = useCallback(async () => {
    try {
      const result = await fetcherRef.current()
      cache.set(key, { data: result, ts: Date.now() })
      if (mountedRef.current) {
        setData(result)
        setLoading(false)
      }
    } catch {
      if (mountedRef.current) setLoading(false)
    }
  }, [key])

  // Subscribe to invalidation events
  useEffect(() => {
    if (!listeners.has(key)) listeners.set(key, new Set())
    const set = listeners.get(key)!
    set.add(refresh)
    return () => { set.delete(refresh) }
  }, [key, refresh])

  // Initial fetch if stale
  useEffect(() => {
    mountedRef.current = true
    const entry = cache.get(key)
    const stale = !entry || Date.now() - entry.ts > maxAge
    if (stale) {
      refresh()
    } else {
      setData(entry!.data as T)
      setLoading(false)
    }
    return () => { mountedRef.current = false }
  // Only run on mount/key change, not on every refresh reference
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key, maxAge])

  return { data, loading, refresh }
}

export function prefetchCache<T>(key: string, fetcher: () => Promise<T>) {
  fetcher().then(data => cache.set(key, { data, ts: Date.now() })).catch(() => {})
}

/**
 * Invalidate cache entries and notify all subscribers to refetch.
 * Call this after create/update/delete operations.
 */
export function invalidateCache(...keys: string[]) {
  for (const key of keys) {
    cache.delete(key)
    const set = listeners.get(key)
    if (set) {
      for (const fn of set) fn()
    }
  }
}

/**
 * Invalidate all cache entries.
 */
export function invalidateAll() {
  const allKeys = [...cache.keys()]
  cache.clear()
  for (const key of allKeys) {
    const set = listeners.get(key)
    if (set) {
      for (const fn of set) fn()
    }
  }
}
