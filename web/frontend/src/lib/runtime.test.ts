import { beforeEach, describe, expect, it, vi } from 'vitest'

describe('runtime helpers', () => {
  beforeEach(() => {
    vi.unstubAllEnvs()
    vi.resetModules()
  })

  it('uses configured API and public app URLs when provided', async () => {
    vi.stubEnv('VITE_API_BASE_URL', 'https://api.example.com/api/v1/')
    vi.stubEnv('VITE_PUBLIC_APP_URL', 'https://konbu.example.com/')

    const runtime = await import('@/lib/runtime')

    expect(runtime.apiBaseURL).toBe('https://api.example.com/api/v1')
    expect(runtime.publicAppURL).toBe('https://konbu.example.com')
    expect(runtime.apiPath('/memos/abc')).toBe('https://api.example.com/api/v1/memos/abc')
    expect(runtime.appURL('/memos/abc')).toBe('https://konbu.example.com/memos/abc')
  })

  it('falls back to local defaults when no runtime env is configured', async () => {
    const runtime = await import('@/lib/runtime')

    expect(runtime.apiBaseURL).toBe('/api/v1')
    expect(runtime.publicAppURL).toBe(window.location.origin)
    expect(runtime.apiPath('/health')).toBe('/api/v1/health')
    expect(runtime.appURL('/memos/abc')).toBe(`${window.location.origin}/memos/abc`)
  })
})
