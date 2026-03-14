import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { FileText, CheckSquare, Calendar, Search } from 'lucide-react'

export function LoginPage() {
  const { t } = useTranslation()
  const { checkAuth, openRegistration } = useAppStore()
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setError('')
    if (mode === 'register' && password !== confirmPassword) {
      setError(t('settings.passwordMismatch'))
      return
    }
    setLoading(true)
    try {
      if (mode === 'register') {
        await api.register({ email, password, name })
      }
      await api.login({ email, password })
      await checkAuth()
    } catch (err) {
      setError(err instanceof Error ? err.message : t('auth.loginError'))
    } finally {
      setLoading(false)
    }
  }

  const switchMode = () => {
    setMode(mode === 'login' ? 'register' : 'login')
    setError('')
  }

  const features = [
    { icon: FileText, label: t('login.featureMemos') },
    { icon: CheckSquare, label: t('login.featureTodos') },
    { icon: Calendar, label: t('login.featureCalendar') },
    { icon: Search, label: t('login.featureSearch') },
  ]

  return (
    <div className="flex min-h-screen bg-background">
      {/* Left: Hero */}
      <div className="hidden lg:flex lg:w-1/2 flex-col justify-center items-center bg-muted/30 p-12 relative overflow-hidden">
        <div className="max-w-md z-10">
          <div className="flex items-center gap-2 mb-6">
            <img className="w-8 h-8" src="/favicon.svg" alt="K" />
            <span className="font-semibold text-xl">konbu</span>
          </div>
          <h1 className="text-2xl font-bold mb-2">{t('login.heroTitle')}</h1>
          <p className="text-muted-foreground mb-8">{t('login.heroDescription')}</p>
          <div className="flex flex-col gap-3 mb-8">
            {features.map(({ icon: Icon, label }) => (
              <div key={label} className="flex items-center gap-3 text-sm">
                <Icon size={18} className="text-primary shrink-0" />
                <span>{label}</span>
              </div>
            ))}
          </div>
        </div>
        <div className="mt-4 w-full max-w-lg">
          <img
            src="/hero.png"
            alt="konbu dashboard"
            className="rounded-lg shadow-lg border"
          />
        </div>
      </div>

      {/* Right: Form */}
      <div className="flex-1 flex items-center justify-center p-4">
        <Card className="w-full max-w-sm">
          <CardHeader>
            <div className="flex items-center gap-2 mb-2 lg:hidden">
              <img className="w-6 h-6" src="/favicon.svg" alt="K" />
              <span className="font-semibold text-sm">konbu</span>
            </div>
            <CardTitle>{mode === 'login' ? t('auth.login') : t('auth.register')}</CardTitle>
            <CardDescription>
              {mode === 'login' ? t('auth.loginDescription') : t('auth.registerDescription')}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="flex flex-col gap-3">
              {mode === 'register' && (
                <Input
                  type="text"
                  placeholder={t('auth.name')}
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                  autoFocus
                />
              )}
              <Input
                type="email"
                placeholder={t('auth.email')}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoFocus={mode === 'login'}
              />
              <Input
                type="password"
                placeholder={t('auth.password')}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
              {mode === 'register' && (
                <Input
                  type="password"
                  placeholder={t('auth.confirmPassword')}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required
                />
              )}
              {error && <p className="text-sm text-destructive">{error}</p>}
              <Button type="submit" disabled={loading}>
                {mode === 'login' ? t('auth.loginButton') : t('auth.registerButton')}
              </Button>
            </form>
            {openRegistration && (
              <p className="mt-4 text-center text-sm text-muted-foreground">
                {mode === 'login' ? t('auth.noAccount') : t('auth.haveAccount')}{' '}
                <button
                  type="button"
                  onClick={switchMode}
                  className="text-primary underline-offset-4 hover:underline"
                >
                  {mode === 'login' ? t('auth.createAccount') : t('auth.backToLogin')}
                </button>
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
