import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'
import { Card, CardContent } from '@/components/ui/card'
import { FileText, CheckSquare, Calendar, Monitor, Search, MessageCircle, Lightbulb, ExternalLink, ArrowRight } from 'lucide-react'

const sections = [
  { key: 'memos', icon: FileText, page: 'memos' as const, color: 'text-blue-500' },
  { key: 'todos', icon: CheckSquare, page: 'todos' as const, color: 'text-emerald-500' },
  { key: 'calendar', icon: Calendar, page: 'calendar' as const, color: 'text-rose-500' },
  { key: 'tools', icon: Monitor, page: 'tools' as const, color: 'text-violet-500' },
  { key: 'search', icon: Search, page: 'search' as const, color: 'text-amber-500' },
  { key: 'chat', icon: MessageCircle, page: 'chat' as const, color: 'text-cyan-500' },
  { key: 'general', icon: Lightbulb, page: null, color: 'text-teal-500' },
] as const

export function HelpPage() {
  const { t } = useTranslation()
  const setPage = useAppStore((s) => s.setPage)

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl font-bold">{t('help.title')}</h1>
        <p className="mt-1 text-muted-foreground">{t('help.description')}</p>
      </div>

      <Card>
        <CardContent className="pt-4">
          <h2 className="mb-2 text-sm font-medium text-muted-foreground">{t('help.toc')}</h2>
          <nav className="flex flex-wrap gap-2">
            {sections.map(({ key, icon: Icon, color }) => (
              <a
                key={key}
                href={`#help-${key}`}
                className="inline-flex items-center gap-1.5 rounded-md bg-muted/50 px-3 py-1.5 text-sm transition-colors hover:bg-muted"
              >
                <Icon size={14} className={color} />
                {t(`help.${key}.title`)}
              </a>
            ))}
          </nav>
        </CardContent>
      </Card>

      {sections.map(({ key, icon: Icon, page, color }) => (
        <Card key={key} id={`help-${key}`}>
          <CardContent className="pt-5">
            <div className="flex items-start justify-between gap-4">
              <div className="mb-2 flex items-center gap-2">
                <Icon size={20} className={color} />
                <h2 className="text-lg font-semibold">{t(`help.${key}.title`)}</h2>
              </div>
              {page && (
                <button
                  onClick={() => setPage(page)}
                  className="inline-flex shrink-0 items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
                >
                  {t('help.tryIt', { page: t(`help.${key}.title`) })}
                  <ArrowRight size={14} />
                </button>
              )}
            </div>
            <p className="mb-3 text-sm text-muted-foreground">
              {t(`help.${key}.description`)}
            </p>
            <ul className="space-y-1.5">
              {(t(`help.${key}.tips`, { returnObjects: true }) as string[]).map((tip, i) => (
                <li key={i} className="flex gap-2 text-sm">
                  <span className="shrink-0 text-muted-foreground">•</span>
                  <span>{tip}</span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      ))}

      <div className="pb-4 text-center text-sm text-muted-foreground">
        {t('help.githubFooter')}{' '}
        <a
          href="https://github.com/krtw00/konbu"
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1 text-foreground hover:underline"
        >
          {t('help.githubLink')}
          <ExternalLink size={12} />
        </a>
      </div>
    </div>
  )
}
