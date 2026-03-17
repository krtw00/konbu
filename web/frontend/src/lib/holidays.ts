// Japanese public holidays
// Spring/Autumn equinox approximation valid for 2000-2099

function springEquinox(year: number): number {
  return Math.floor(20.8431 + 0.242194 * (year - 1980) - Math.floor((year - 1980) / 4))
}

function autumnEquinox(year: number): number {
  return Math.floor(23.2488 + 0.242194 * (year - 1980) - Math.floor((year - 1980) / 4))
}

function nthMonday(year: number, month: number, n: number): number {
  const first = new Date(year, month - 1, 1).getDay()
  return 1 + ((8 - first) % 7) + (n - 1) * 7
}

export function getHolidays(year: number): Map<string, string> {
  const holidays = new Map<string, string>()

  const add = (m: number, d: number, name: string) => {
    holidays.set(`${year}-${String(m).padStart(2, '0')}-${String(d).padStart(2, '0')}`, name)
  }

  add(1, 1, '元日')
  add(1, nthMonday(year, 1, 2), '成人の日')
  add(2, 11, '建国記念の日')
  add(2, 23, '天皇誕生日')
  add(3, springEquinox(year), '春分の日')
  add(4, 29, '昭和の日')
  add(5, 3, '憲法記念日')
  add(5, 4, 'みどりの日')
  add(5, 5, 'こどもの日')
  add(7, nthMonday(year, 7, 3), '海の日')
  add(8, 11, '山の日')
  add(9, nthMonday(year, 9, 3), '敬老の日')
  add(9, autumnEquinox(year), '秋分の日')
  add(10, nthMonday(year, 10, 2), 'スポーツの日')
  add(11, 3, '文化の日')
  add(11, 23, '勤労感謝の日')

  // Substitute holidays (振替休日): if holiday falls on Sunday, next Monday is a holiday
  const entries = [...holidays.entries()]
  for (const [dateStr] of entries) {
    const d = new Date(dateStr + 'T00:00:00')
    if (d.getDay() === 0) {
      const next = new Date(d)
      next.setDate(next.getDate() + 1)
      const nextStr = next.toISOString().slice(0, 10)
      if (!holidays.has(nextStr)) {
        holidays.set(nextStr, '振替休日')
      }
    }
  }

  return holidays
}

export function getHolidayName(year: number, month: number, day: number): string | undefined {
  const key = `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
  return getHolidays(year).get(key)
}
