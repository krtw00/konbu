import { describe, expect, it } from 'vitest'
import { parseExternalRoute } from '@/lib/publicRoute'

describe('parseExternalRoute', () => {
  it('parses token-based public share routes', () => {
    expect(parseExternalRoute('/public/abc123')).toEqual({
      kind: 'public-share',
      token: 'abc123',
    })
  })

  it('parses published memo routes', () => {
    expect(parseExternalRoute('/memo/hello-konbu')).toEqual({
      kind: 'published-memo',
      slug: 'hello-konbu',
    })
  })

  it('returns null for app-internal routes', () => {
    expect(parseExternalRoute('/settings')).toBeNull()
  })
})
