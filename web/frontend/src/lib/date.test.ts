import { describe, expect, it } from 'vitest'
import { isoToDateInput, localDateToISO, localToISO } from '@/lib/date'

describe('date helpers', () => {
  it('creates midnight ISO strings for all-day dates', () => {
    const iso = localDateToISO('2026-03-18')

    expect(iso).toMatch(/^2026-03-18T00:00:00[+-]\d{2}:\d{2}$/)
  })

  it('round-trips all-day dates through ISO conversion', () => {
    const iso = localDateToISO('2026-03-18')

    expect(isoToDateInput(iso)).toBe('2026-03-18')
  })

  it('preserves local datetime values when converting to ISO', () => {
    const iso = localToISO('2026-03-18T09:30')

    expect(iso).toMatch(/^2026-03-18T09:30:00[+-]\d{2}:\d{2}$/)
  })
})

