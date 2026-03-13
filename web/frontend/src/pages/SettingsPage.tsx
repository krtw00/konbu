import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'
import { api } from '@/lib/api'
import type { ApiKey } from '@/types/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Copy, Trash2, Plus, Key } from 'lucide-react'

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
    </div>
  )
}

function AppearanceTab() {
  const { t, i18n } = useTranslation()

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
        <label className="text-sm font-medium">{t('settings.theme')}</label>
        <Input value={t('settings.themeDefault')} disabled />
      </div>
    </div>
  )
}

function SecurityTab() {
  const { t } = useTranslation()
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
    </div>
  )
}

export function SettingsPage() {
  const { t } = useTranslation()

  return (
    <div>
      <h1 className="text-lg font-semibold mb-4">{t('settings.title')}</h1>
      <Tabs defaultValue="profile">
        <TabsList>
          <TabsTrigger value="profile">{t('settings.profile')}</TabsTrigger>
          <TabsTrigger value="appearance">{t('settings.appearance')}</TabsTrigger>
          <TabsTrigger value="security">{t('settings.security')}</TabsTrigger>
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
      </Tabs>
    </div>
  )
}
