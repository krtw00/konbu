import { lazy, type ComponentType } from 'react'

const RETRYABLE_IMPORT_ERRORS = [
  'chunkloaderror',
  'failed to fetch dynamically imported module',
  'importing a module script failed',
  'error loading dynamically imported module',
]

function isRetryableImportError(error: unknown): boolean {
  if (!(error instanceof Error)) return false
  const message = error.message.toLowerCase()
  return RETRYABLE_IMPORT_ERRORS.some((fragment) => message.includes(fragment))
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function lazyWithRetry<T extends ComponentType<any>>(
  importer: () => Promise<{ default: T }>,
  key: string,
) {
  return lazy(async (): Promise<{ default: T }> => {
    try {
      const module = await importer()
      if (typeof window !== 'undefined') {
        window.sessionStorage.removeItem(`konbu:lazy-retry:${key}`)
      }
      return module
    } catch (error) {
      if (typeof window !== 'undefined') {
        const retryKey = `konbu:lazy-retry:${key}`
        const hasRetried = window.sessionStorage.getItem(retryKey) === '1'
        if (!hasRetried && isRetryableImportError(error)) {
          window.sessionStorage.setItem(retryKey, '1')
          window.location.reload()
          return new Promise<{ default: T }>(() => {})
        }
        window.sessionStorage.removeItem(retryKey)
      }
      throw error
    }
  })
}
