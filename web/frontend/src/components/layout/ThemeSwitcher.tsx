import { useAppStore } from '@/stores/app'

const themes = [
  { id: 'konbu', label: 'Konbu' },
  { id: 'notion', label: 'Notion' },
  { id: 'solarized', label: 'Solarized' },
  { id: 'latte', label: 'Latte' },
  { id: 'nord', label: 'Nord' },
  { id: 'linear', label: 'Linear' },
  { id: 'mocha', label: 'Mocha' },
]

export function ThemeSwitcher() {
  const { theme, setTheme } = useAppStore()

  return (
    <div className="fixed top-2 right-2 flex gap-1 z-50">
      {themes.map((t) => (
        <button
          key={t.id}
          onClick={() => setTheme(t.id)}
          title={t.label}
          className={`w-3 h-3 rounded-full border transition-transform ${
            theme === t.id ? 'scale-125 border-primary' : 'border-border hover:scale-110'
          }`}
          data-theme-dot={t.id}
        />
      ))}
    </div>
  )
}
