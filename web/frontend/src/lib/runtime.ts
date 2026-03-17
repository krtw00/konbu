function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, '')
}

function browserOrigin(): string {
  if (typeof window === 'undefined') return ''
  return window.location.origin
}

const configuredAPIBase = import.meta.env.VITE_API_BASE_URL?.trim()
const configuredPublicAppURL = import.meta.env.VITE_PUBLIC_APP_URL?.trim()

export const apiBaseURL = configuredAPIBase
  ? trimTrailingSlash(configuredAPIBase)
  : '/api/v1'

export const publicAppURL = trimTrailingSlash(configuredPublicAppURL || browserOrigin())

export function apiPath(path: string): string {
  return `${apiBaseURL}${path}`
}

export function appURL(path: string): string {
  if (!publicAppURL) return path
  return new URL(path, `${publicAppURL}/`).toString()
}

