import { useState, useEffect, useCallback, useRef } from 'react'

const cache = new Map<string, { data: unknown; ts: number }>()

export function useCache<T>(key: string, fetcher: () => Promise<T>, maxAge = 30000): {
  data: T | null
  loading: boolean
  refresh: () => Promise<void>
} {
  const cached = cache.get(key)
  const [data, setData] = useState<T | null>((cached?.data as T) ?? null)
  const [loading, setLoading] = useState(cached ? false : true)
  const mountedRef = useRef(true)

  const refresh = useCallback(async () => {
    try {
      const result = await fetcher()
      cache.set(key, { data: result, ts: Date.now() })
      if (mountedRef.current) {
        setData(result)
        setLoading(false)
      }
    } catch {
      if (mountedRef.current) setLoading(false)
    }
  }, [key, fetcher])

  useEffect(() => {
    mountedRef.current = true
    const stale = !cached || Date.now() - cached.ts > maxAge
    if (stale) refresh()
    else setLoading(false)
    return () => { mountedRef.current = false }
  }, [key, maxAge, refresh, cached])

  return { data, loading, refresh }
}

export function prefetchCache<T>(key: string, fetcher: () => Promise<T>) {
  fetcher().then(data => cache.set(key, { data, ts: Date.now() })).catch(() => {})
}

export function invalidateCache(prefix?: string) {
  if (prefix) {
    for (const key of cache.keys()) {
      if (key.startsWith(prefix)) cache.delete(key)
    }
  } else {
    cache.clear()
  }
}
