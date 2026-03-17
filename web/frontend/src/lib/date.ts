import i18n from '@/i18n'

function currentLocale(): string {
  const lng = i18n.language || 'en'
  return lng === 'ja' ? 'ja-JP' : 'en-US'
}

export function relativeTime(iso: string): string {
  if (!iso) return ''
  const t = i18n.t.bind(i18n)
  const d = new Date(iso)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 60_000) return t('date.now')
  if (diff < 3_600_000) return t('date.minutesAgo', { count: Math.floor(diff / 60_000) })
  if (diff < 86_400_000) return t('date.hoursAgo', { count: Math.floor(diff / 3_600_000) })
  if (diff < 604_800_000) return t('date.daysAgo', { count: Math.floor(diff / 86_400_000) })
  return d.toLocaleDateString(currentLocale(), { month: 'short', day: 'numeric' })
}

export function formatTime(iso: string): string {
  if (!iso) return ''
  return new Date(iso).toLocaleTimeString(currentLocale(), { hour: '2-digit', minute: '2-digit' })
}

export function formatDate(iso: string): string {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString(currentLocale(), { month: 'short', day: 'numeric' })
}

export function dateKey(y: number, m: number, d: number): string {
  return `${y}-${String(m + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`
}

export function dueFmt(dueDate: string | null): { text: string; className: string } | null {
  if (!dueDate) return null
  const t = i18n.t.bind(i18n)
  const today = new Date()
  const due = new Date(dueDate + 'T00:00:00')
  const diff = Math.floor(
    (due.getTime() - new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime()) / 86_400_000
  )
  if (diff < 0) return { text: t('todos.overdue', { days: Math.abs(diff) }), className: 'text-destructive' }
  if (diff === 0) return { text: t('todos.today'), className: 'text-orange-500' }
  if (diff === 1) return { text: t('todos.tomorrow'), className: '' }
  if (diff <= 7) return { text: `${diff}d`, className: '' }
  return { text: due.toLocaleDateString(currentLocale(), { month: 'short', day: 'numeric' }), className: '' }
}

const pad2 = (n: number) => String(n).padStart(2, '0')

/** Convert a datetime-local value ("YYYY-MM-DDTHH:mm") to ISO 8601 with timezone offset for API */
export function localToISO(local: string): string {
  if (!local) return ''
  const d = new Date(local)
  const off = -d.getTimezoneOffset()
  const sign = off >= 0 ? '+' : '-'
  const absOff = Math.abs(off)
  const oh = pad2(Math.floor(absOff / 60))
  const om = pad2(absOff % 60)
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}T${pad2(d.getHours())}:${pad2(d.getMinutes())}:00${sign}${oh}:${om}`
}

export function localDateToISO(localDate: string): string {
  if (!localDate) return ''
  return localToISO(`${localDate}T00:00`)
}

/** Convert an ISO 8601 string to datetime-local value ("YYYY-MM-DDTHH:mm") */
export function isoToLocal(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())}T${pad2(d.getHours())}:${pad2(d.getMinutes())}`
}

export function isoToDateInput(iso: string): string {
  if (!iso) return ''
  return isoToLocal(iso).slice(0, 10)
}

export function dateDelta(n: number): string {
  const d = new Date()
  d.setDate(d.getDate() + n)
  return d.toISOString().slice(0, 10)
}

export function nextMonday(): string {
  const d = new Date()
  const day = d.getDay()
  const add = day === 0 ? 1 : 8 - day
  d.setDate(d.getDate() + add)
  return d.toISOString().slice(0, 10)
}
