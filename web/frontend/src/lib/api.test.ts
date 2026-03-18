import { beforeEach, describe, expect, it, vi } from 'vitest'

describe('apiFetch', () => {
  beforeEach(() => {
    vi.resetModules()
  })

  it('uses apiPath and includes credentials', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response('{}', { status: 200 }))
    vi.stubGlobal('fetch', fetchMock)

    const { apiFetch } = await import('@/lib/api')
    await apiFetch('/auth/me', { method: 'GET' })

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/auth/me', expect.objectContaining({
      method: 'GET',
      credentials: 'include',
    }))
  })
})
