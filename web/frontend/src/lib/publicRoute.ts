export type ExternalRoute =
  | { kind: 'public-share'; token: string }
  | { kind: 'published-memo'; slug: string }
  | null

export function parseExternalRoute(pathname: string): ExternalRoute {
  const shareMatch = pathname.match(/^\/public\/([a-zA-Z0-9]+)$/)
  if (shareMatch) {
    return { kind: 'public-share', token: shareMatch[1] }
  }

  const memoMatch = pathname.match(/^\/memo\/([^/]+)$/)
  if (memoMatch) {
    return { kind: 'published-memo', slug: decodeURIComponent(memoMatch[1]) }
  }

  return null
}
