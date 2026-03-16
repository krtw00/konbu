import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'
import { api } from '@/lib/api'
import type { ApiKey } from '@/types/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Copy, Trash2, Plus, Key, ExternalLink } from 'lucide-react'

function ProfileTab() {
  const { t } = useTranslation()
  const { user, setUser } = useAppStore()
  const [name, setName] = useState(user?.name ?? '')

  const handleBlur = async () => {
    if (name === user?.name || !name.trim()) return
    try {
      const res = await api.updateMe({ name: name.trim() })
      setUser(res.data)
    } catch {
      setName(user?.name ?? '')
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-1.5">
        <label className="text-sm font-medium">{t('settings.displayName')}</label>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          onBlur={handleBlur}
        />
      </div>
      <div className="flex flex-col gap-1.5">
        <label className="text-sm font-medium">{t('settings.email')}</label>
        <Input value={user?.email ?? ''} disabled />
      </div>
      <Separator />
      <div className="flex flex-col gap-1.5">
        <label className="text-sm font-medium">{t('settings.plan')}</label>
        <div className="flex items-center gap-3">
          <span className={`text-sm font-medium px-2 py-0.5 rounded ${user?.plan === 'sponsor' || user?.is_admin ? 'bg-primary/10 text-primary' : 'bg-muted'}`}>
            {user?.plan === 'sponsor' || user?.is_admin ? t('settings.planSponsor') : t('settings.planFree')}
          </span>
        </div>
        {(user?.plan === 'sponsor' || user?.is_admin) ? (
          <div className="mt-2 p-4 rounded-lg border border-primary/20 bg-primary/5">
            <h3 className="font-semibold text-sm mb-2">{t('settings.sponsorActiveTitle')}</h3>
            <ul className="text-sm text-muted-foreground space-y-1.5">
              <li className="flex items-center gap-2">
                <span className="text-primary">✓</span> {t('settings.upgradeFeatureChat')}
              </li>
              <li className="flex items-center gap-2">
                <span className="text-primary">✓</span> {t('settings.upgradeFeatureImage')}
              </li>
              <li className="flex items-center gap-2">
                <span className="text-primary">✓</span> {t('settings.upgradeFeatureApi')}
              </li>
            </ul>
          </div>
        ) : (
          <div className="mt-2 p-4 rounded-lg border border-border bg-card">
            <h3 className="font-semibold text-sm mb-2">{t('settings.upgradeTitle')}</h3>
            <ul className="text-sm text-muted-foreground space-y-1 mb-3">
              <li>• {t('settings.upgradeFeatureChat')}</li>
              <li>• {t('settings.upgradeFeatureImage')}</li>
              <li>• {t('settings.upgradeFeatureApi')}</li>
            </ul>
            <p className="text-xs text-muted-foreground mb-3">{t('settings.upgradeEmailNote')}</p>
            <iframe
              id="kofiframe"
              src="https://ko-fi.com/codenica000/?hidefeed=true&widget=true&embed=true"
              className="w-full border-0 rounded-lg"
              style={{ height: '712px' }}
              title="Ko-fi"
            />
          </div>
        )}
      </div>
    </div>
  )
}

function AppearanceTab() {
  const { t, i18n } = useTranslation()
  const [firstDayOfWeek, setFirstDayOfWeek] = useState(0)

  useEffect(() => {
    api.getSettings().then((r) => {
      if (r.data?.first_day_of_week !== undefined) {
        setFirstDayOfWeek(r.data.first_day_of_week)
      }
    }).catch(() => {})
  }, [])

  async function handleFirstDayChange(val: number) {
    setFirstDayOfWeek(val)
    try {
      const current = await api.getSettings()
      await api.updateSettings({ ...current.data, first_day_of_week: val })
    } catch {
      // ignore
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-1.5">
        <label className="text-sm font-medium">{t('settings.language')}</label>
        <select
          className="flex h-8 w-full rounded-md border border-input bg-background px-3 text-sm"
          value={i18n.language.startsWith('ja') ? 'ja' : 'en'}
          onChange={(e) => i18n.changeLanguage(e.target.value)}
        >
          <option value="en">English</option>
          <option value="ja">日本語</option>
        </select>
      </div>
      <div className="flex flex-col gap-1.5">
        <label className="text-sm font-medium">{t('settings.firstDayOfWeek')}</label>
        <select
          className="flex h-8 w-full rounded-md border border-input bg-background px-3 text-sm"
          value={firstDayOfWeek}
          onChange={(e) => handleFirstDayChange(Number(e.target.value))}
        >
          <option value={0}>{t('settings.sunday')}</option>
          <option value={1}>{t('settings.monday')}</option>
        </select>
      </div>
      <div className="flex flex-col gap-1.5">
        <label className="text-sm font-medium">{t('settings.theme')}</label>
        <Input value={t('settings.themeDefault')} disabled />
      </div>
    </div>
  )
}

function SecurityTab() {
  const { t } = useTranslation()
  const { logout } = useAppStore()
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [pwMsg, setPwMsg] = useState('')
  const [pwError, setPwError] = useState(false)
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([])
  const [newKeyName, setNewKeyName] = useState('')
  const [createdKey, setCreatedKey] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)
  const [deletePassword, setDeletePassword] = useState('')
  const [deleteError, setDeleteError] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState(false)

  useEffect(() => {
    api.listApiKeys().then((res) => setApiKeys(res.data)).catch(() => {})
  }, [])

  const handleChangePassword = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setPwMsg('')
    setPwError(false)
    if (newPassword !== confirmPassword) {
      setPwMsg(t('settings.passwordMismatch'))
      setPwError(true)
      return
    }
    try {
      await api.changePassword({ old_password: oldPassword, new_password: newPassword })
      setPwMsg(t('settings.passwordChanged'))
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (err) {
      setPwMsg(err instanceof Error ? err.message : 'Error')
      setPwError(true)
    }
  }

  const handleCreateKey = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (!newKeyName.trim()) return
    try {
      const res = await api.createApiKey({ name: newKeyName.trim() })
      setCreatedKey(res.data.key ?? null)
      setNewKeyName('')
      const list = await api.listApiKeys()
      setApiKeys(list.data)
    } catch {
      // ignore
    }
  }

  const handleCopyKey = async (key: string) => {
    await navigator.clipboard.writeText(key)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleDeleteKey = async (id: string) => {
    if (deletingId !== id) {
      setDeletingId(id)
      return
    }
    try {
      await api.deleteApiKey(id)
      setApiKeys((prev) => prev.filter((k) => k.id !== id))
    } catch {
      // ignore
    }
    setDeletingId(null)
  }

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <CardTitle>{t('settings.changePassword')}</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleChangePassword} className="flex flex-col gap-3">
            <Input
              type="password"
              placeholder={t('settings.oldPassword')}
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              required
            />
            <Input
              type="password"
              placeholder={t('settings.newPassword')}
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
            />
            <Input
              type="password"
              placeholder={t('settings.confirmPassword')}
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />
            {pwMsg && (
              <p className={`text-sm ${pwError ? 'text-destructive' : 'text-green-600'}`}>{pwMsg}</p>
            )}
            <Button type="submit">{t('settings.changePassword')}</Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('settings.apiKeys')}</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {useAppStore.getState().user?.plan !== 'sponsor' && !useAppStore.getState().user?.is_admin && (
            <div className="rounded-md bg-muted p-3 text-sm text-muted-foreground">
              <p>{t('settings.apiKeysSponsorOnly')}</p>
              <a
                href="https://ko-fi.com/codenica000"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 mt-2 text-primary hover:underline"
              >
                {t('settings.sponsorCTA')}
                <ExternalLink size={14} />
              </a>
            </div>
          )}
          {createdKey && (
            <div className="flex items-center gap-2 rounded-md bg-muted p-2 text-sm">
              <Key size={14} />
              <code className="flex-1 break-all">{createdKey}</code>
              <Button
                variant="ghost"
                size="icon-xs"
                onClick={() => handleCopyKey(createdKey)}
              >
                <Copy size={14} />
              </Button>
              <span className="text-xs text-muted-foreground">
                {copied ? t('settings.keyCopied') : t('settings.copyKey')}
              </span>
            </div>
          )}

          {apiKeys.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t('settings.noKeys')}</p>
          ) : (
            <div className="flex flex-col gap-2">
              {apiKeys.map((key) => (
                <div key={key.id} className="flex items-center justify-between rounded-md border p-2 text-sm">
                  <div>
                    <span className="font-medium">{key.name}</span>
                    <span className="ml-2 text-xs text-muted-foreground">
                      {new Date(key.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  <Button
                    variant={deletingId === key.id ? 'destructive' : 'ghost'}
                    size="icon-xs"
                    onClick={() => handleDeleteKey(key.id)}
                  >
                    <Trash2 size={14} />
                  </Button>
                </div>
              ))}
            </div>
          )}

          <Separator />
          <form onSubmit={handleCreateKey} className="flex gap-2">
            <Input
              placeholder={t('settings.keyName')}
              value={newKeyName}
              onChange={(e) => setNewKeyName(e.target.value)}
              className="flex-1"
            />
            <Button type="submit" variant="outline" size="sm">
              <Plus size={14} />
              {t('settings.createKey')}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card className="border-destructive/50">
        <CardHeader>
          <CardTitle className="text-destructive">{t('settings.deleteAccount')}</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <p className="text-sm text-muted-foreground">{t('settings.deleteAccountDescription')}</p>
          {!deleteConfirm ? (
            <Button variant="destructive" size="sm" onClick={() => setDeleteConfirm(true)}>
              {t('settings.deleteAccount')}
            </Button>
          ) : (
            <form onSubmit={async (e) => {
              e.preventDefault()
              setDeleteError('')
              try {
                await api.deleteAccount({ password: deletePassword })
                await logout()
              } catch (err) {
                setDeleteError(err instanceof Error ? err.message : 'Error')
              }
            }} className="flex flex-col gap-3">
              <Input
                type="password"
                placeholder={t('settings.deleteAccountConfirm')}
                value={deletePassword}
                onChange={(e) => setDeletePassword(e.target.value)}
                required
                autoFocus
              />
              {deleteError && <p className="text-sm text-destructive">{deleteError}</p>}
              <div className="flex gap-2">
                <Button type="submit" variant="destructive" size="sm">
                  {t('settings.deleteAccountButton')}
                </Button>
                <Button type="button" variant="outline" size="sm" onClick={() => {
                  setDeleteConfirm(false)
                  setDeletePassword('')
                  setDeleteError('')
                }}>
                  {t('common.cancel')}
                </Button>
              </div>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function DataTab() {
  const { t } = useTranslation()
  const [importing, setImporting] = useState(false)
  const [importMsg, setImportMsg] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)

  function downloadExport(format: 'json' | 'markdown') {
    const url = `/api/v1/export/${format}`
    const a = document.createElement('a')
    a.href = url
    a.download = ''
    a.click()
  }

  async function handleICalImport(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setImporting(true)
    setImportMsg('')
    try {
      const form = new FormData()
      form.append('file', file)
      const res = await fetch('/api/v1/import/ical', { method: 'POST', body: form })
      if (!res.ok) {
        const data = await res.json()
        throw new Error(data.error?.message || 'Import failed')
      }
      setImportMsg(t('settings.importSuccess'))
    } catch (err) {
      setImportMsg(err instanceof Error ? err.message : 'Error')
    }
    setImporting(false)
    if (fileRef.current) fileRef.current.value = ''
  }

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <CardTitle>{t('settings.export')}</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <p className="text-sm text-muted-foreground">{t('settings.exportDescription')}</p>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => downloadExport('json')}>
              {t('settings.exportJSON')}
            </Button>
            <Button variant="outline" size="sm" onClick={() => downloadExport('markdown')}>
              {t('settings.exportMarkdown')}
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t('settings.import')}</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <p className="text-sm text-muted-foreground">{t('settings.importDescription')}</p>
          <div>
            <label className="text-sm font-medium">{t('settings.importICal')}</label>
            <Input
              ref={fileRef}
              type="file"
              accept=".ics,.ical"
              onChange={handleICalImport}
              disabled={importing}
              className="mt-1"
            />
          </div>
          {importMsg && <p className="text-sm text-green-600">{importMsg}</p>}
        </CardContent>
      </Card>
    </div>
  )
}

function AITab() {
  const { t } = useTranslation()
  const [provider, setProvider] = useState('openai')
  const [openaiKey, setOpenaiKey] = useState('')
  const [anthropicKey, setAnthropicKey] = useState('')
  const [openaiMasked, setOpenaiMasked] = useState('')
  const [anthropicMasked, setAnthropicMasked] = useState('')
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    api.getChatConfig().then(r => {
      const cfg = r.data
      if (cfg.provider) setProvider(cfg.provider)
      if (cfg.openai_key_masked) setOpenaiMasked(cfg.openai_key_masked)
      if (cfg.anthropic_key_masked) setAnthropicMasked(cfg.anthropic_key_masked)
    }).catch(() => {})
  }, [])

  async function handleSave() {
    await api.updateChatConfig({
      provider,
      openai_key: openaiKey || undefined,
      anthropic_key: anthropicKey || undefined,
    })
    setOpenaiKey('')
    setAnthropicKey('')
    const r = await api.getChatConfig()
    if (r.data.openai_key_masked) setOpenaiMasked(r.data.openai_key_masked)
    if (r.data.anthropic_key_masked) setAnthropicMasked(r.data.anthropic_key_masked)
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{t('settings.ai')}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-col gap-1.5">
          <label className="text-sm font-medium">{t('settings.aiProvider')}</label>
          <select
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            value={provider}
            onChange={e => setProvider(e.target.value)}
          >
            <option value="openai">OpenAI (GPT-4o)</option>
            <option value="anthropic">Anthropic (Claude)</option>
          </select>
        </div>
        <Separator />
        <div className="flex flex-col gap-1.5">
          <label className="text-sm font-medium">OpenAI API Key</label>
          {openaiMasked && <p className="text-xs text-muted-foreground">{openaiMasked}</p>}
          <Input
            type="password"
            value={openaiKey}
            onChange={e => setOpenaiKey(e.target.value)}
            placeholder={t('settings.aiKeyPlaceholder')}
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <label className="text-sm font-medium">Anthropic API Key</label>
          {anthropicMasked && <p className="text-xs text-muted-foreground">{anthropicMasked}</p>}
          <Input
            type="password"
            value={anthropicKey}
            onChange={e => setAnthropicKey(e.target.value)}
            placeholder={t('settings.aiKeyPlaceholder')}
          />
        </div>
        <div className="flex items-center gap-2">
          <Button onClick={handleSave}>{t('common.save')}</Button>
          {saved && <span className="text-sm text-green-600">{t('settings.aiKeySaved')}</span>}
        </div>
      </CardContent>
    </Card>
  )
}

export function SettingsPage() {
  const { t } = useTranslation()

  return (
    <div>
      <h1 className="text-lg font-semibold mb-4">{t('settings.title')}</h1>
      <Tabs defaultValue="profile">
        <TabsList className="w-full overflow-x-auto flex-nowrap justify-start">
          <TabsTrigger value="profile" className="shrink-0 text-xs sm:text-sm">{t('settings.profile')}</TabsTrigger>
          <TabsTrigger value="appearance" className="shrink-0 text-xs sm:text-sm">{t('settings.appearance')}</TabsTrigger>
          <TabsTrigger value="security" className="shrink-0 text-xs sm:text-sm">{t('settings.security')}</TabsTrigger>
          <TabsTrigger value="data" className="shrink-0 text-xs sm:text-sm">{t('settings.data')}</TabsTrigger>
          <TabsTrigger value="ai" className="shrink-0 text-xs sm:text-sm">{t('settings.ai')}</TabsTrigger>
        </TabsList>
        <TabsContent value="profile" className="mt-4">
          <ProfileTab />
        </TabsContent>
        <TabsContent value="appearance" className="mt-4">
          <AppearanceTab />
        </TabsContent>
        <TabsContent value="security" className="mt-4">
          <SecurityTab />
        </TabsContent>
        <TabsContent value="data" className="mt-4">
          <DataTab />
        </TabsContent>
        <TabsContent value="ai" className="mt-4">
          <AITab />
        </TabsContent>
      </Tabs>
    </div>
  )
}
