import { useTranslation } from 'react-i18next'

interface Props {
  feature: string
}

export function UpgradePrompt({ feature }: Props) {
  const { t } = useTranslation()

  return (
    <div className="text-center py-6 space-y-3">
      <p className="text-sm text-muted-foreground">
        {t(`upgrade.${feature}`)}
      </p>
      <a
        href="https://ko-fi.com/codenica000"
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-2 text-sm font-medium text-white px-5 py-2 rounded-md bg-primary hover:opacity-90 transition-opacity"
      >
        {t('upgrade.cta')}
      </a>
    </div>
  )
}
