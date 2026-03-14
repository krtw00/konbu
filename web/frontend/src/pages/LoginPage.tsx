import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useAppStore } from '@/stores/app'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { FileText, CheckSquare, Calendar, Search } from 'lucide-react'
import heroImg from '@/assets/hero.png'

export function LoginPage() {
  const { t } = useTranslation()
  const { checkAuth, openRegistration, googleAuth } = useAppStore()
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
            src={heroImg}
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
            {googleAuth && mode === 'login' && (
              <>
                <div className="relative my-4">
                  <div className="absolute inset-0 flex items-center">
                    <span className="w-full border-t" />
                  </div>
                  <div className="relative flex justify-center text-xs uppercase">
                    <span className="bg-card px-2 text-muted-foreground">or</span>
                  </div>
                </div>
                <a
                  href="/auth/google/login"
                  className="inline-flex w-full items-center justify-center gap-2 rounded-md border border-input bg-background px-4 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground"
                >
                  <svg className="h-4 w-4" viewBox="0 0 24 24">
                    <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4" />
                    <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853" />
                    <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05" />
                    <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335" />
                  </svg>
                  {t('auth.googleLogin')}
                </a>
              </>
            )}
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
