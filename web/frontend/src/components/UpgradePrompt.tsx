import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'

interface Props {
  feature: string
}

export function UpgradePrompt({ feature }: Props) {
  const { t } = useTranslation()
  const setPage = useAppStore(s => s.setPage)

  return (
    <div className="text-center py-6 space-y-3">
      <p className="text-sm text-muted-foreground">
        {t(`upgrade.${feature}`)}
      </p>
      <button
        onClick={() => setPage('settings')}
        className="inline-flex items-center gap-2 text-sm font-medium text-white px-5 py-2 rounded-md bg-primary hover:opacity-90 transition-opacity"
      >
        {t('upgrade.cta')}
      </button>
    </div>
  )
}
