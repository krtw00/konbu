import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Bug, Lightbulb, HelpCircle, Mail } from 'lucide-react'
import { api } from '@/lib/api'
import { useAppStore } from '@/stores/app'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

const categories = [
  { value: 'bug', icon: Bug },
  { value: 'feature', icon: Lightbulb },
  { value: 'question', icon: HelpCircle },
  { value: 'other', icon: Mail },
] as const

type FeedbackCategory = typeof categories[number]['value']

export function FeedbackPage() {
  const { t } = useTranslation()
  const user = useAppStore((s) => s.user)
  const [email, setEmail] = useState(user?.email ?? '')
  const [category, setCategory] = useState<FeedbackCategory>('bug')
  const [bugSummary, setBugSummary] = useState('')
  const [bugSteps, setBugSteps] = useState('')
  const [bugExpected, setBugExpected] = useState('')
  const [bugActual, setBugActual] = useState('')
  const [bugEnvironment, setBugEnvironment] = useState('')
  const [featureSummary, setFeatureSummary] = useState('')
  const [featureProblem, setFeatureProblem] = useState('')
  const [featureProposal, setFeatureProposal] = useState('')
  const [featureBenefit, setFeatureBenefit] = useState('')
  const [questionSubject, setQuestionSubject] = useState('')
  const [questionDetails, setQuestionDetails] = useState('')
  const [otherSubject, setOtherSubject] = useState('')
  const [otherMessage, setOtherMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitted, setSubmitted] = useState(false)
  const [error, setError] = useState('')

  const sourcePage = useMemo(() => {
    if (typeof window === 'undefined') return '/feedback'
    return window.location.pathname
  }, [])

  function buildMessage() {
    switch (category) {
      case 'bug':
        return [
          `Summary: ${bugSummary.trim()}`,
          '',
          'Steps to reproduce:',
          bugSteps.trim(),
          '',
          'Expected behavior:',
          bugExpected.trim(),
          '',
          'Actual behavior:',
          bugActual.trim(),
          bugEnvironment.trim() ? '' : null,
          bugEnvironment.trim() ? `Environment: ${bugEnvironment.trim()}` : null,
        ].filter(Boolean).join('\n')
      case 'feature':
        return [
          `Summary: ${featureSummary.trim()}`,
          '',
          'Current problem:',
          featureProblem.trim(),
          '',
          'Proposed change:',
          featureProposal.trim(),
          featureBenefit.trim() ? '' : null,
          featureBenefit.trim() ? 'Expected benefit:' : null,
          featureBenefit.trim() || null,
        ].filter(Boolean).join('\n')
      case 'question':
        return [
          `Subject: ${questionSubject.trim()}`,
          '',
          'Details:',
          questionDetails.trim(),
        ].filter(Boolean).join('\n')
      case 'other':
      default:
        return [
          `Subject: ${otherSubject.trim()}`,
          '',
          'Message:',
          otherMessage.trim(),
        ].filter(Boolean).join('\n')
    }
  }

  function validate() {
    switch (category) {
      case 'bug':
        return bugSummary.trim() && bugSteps.trim() && bugExpected.trim() && bugActual.trim()
      case 'feature':
        return featureSummary.trim() && featureProblem.trim() && featureProposal.trim()
      case 'question':
        return questionSubject.trim() && questionDetails.trim()
      case 'other':
      default:
        return otherSubject.trim() && otherMessage.trim()
    }
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    if (!validate()) {
      setError(t('feedback.validation'))
      return
    }

    setLoading(true)
    try {
      await api.submitFeedback({
        email,
        category,
        message: buildMessage(),
        source_page: sourcePage,
      })
      setSubmitted(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : t('feedback.submitError'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-background px-4 py-8 md:py-10">
      <div className="mx-auto flex max-w-4xl flex-col gap-6">
        <div className="space-y-3">
          <a href="/" className="text-sm text-primary hover:underline">
            {t('feedback.back')}
          </a>
          <div className="rounded-2xl border border-border bg-gradient-to-br from-primary/10 via-background to-sidebar-accent/70 p-6">
            <h1 className="text-3xl font-semibold">{t('feedback.title')}</h1>
            <p className="mt-2 max-w-2xl text-sm text-muted-foreground">{t('feedback.description')}</p>
          </div>
        </div>

        <div className="grid gap-6 lg:grid-cols-[1.6fr_0.9fr]">
          <Card>
            <CardHeader>
              <CardTitle>{t('feedback.formTitle')}</CardTitle>
              <CardDescription>{t('feedback.formDescription')}</CardDescription>
            </CardHeader>
            <CardContent>
              {submitted ? (
                <div className="rounded-md border border-primary/20 bg-primary/5 p-4 text-sm text-foreground">
                  {t('feedback.thanks')}
                </div>
              ) : (
                <form onSubmit={handleSubmit} className="flex flex-col gap-5">
                  <div className="flex flex-col gap-1.5">
                    <label className="text-sm font-medium" htmlFor="feedback-email">{t('feedback.email')}</label>
                    <Input
                      id="feedback-email"
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      required
                    />
                  </div>

                  <div className="flex flex-col gap-2">
                    <div className="text-sm font-medium">{t('feedback.category')}</div>
                    <div className="grid gap-2 sm:grid-cols-2">
                      {categories.map(({ value, icon: Icon }) => {
                        const active = category === value
                        return (
                          <button
                            key={value}
                            type="button"
                            onClick={() => setCategory(value)}
                            className={`flex items-start gap-3 rounded-xl border px-4 py-3 text-left transition-colors ${
                              active ? 'border-primary bg-primary/8 text-foreground' : 'border-border hover:bg-accent/40'
                            }`}
                          >
                            <Icon size={18} className={active ? 'text-primary' : 'text-muted-foreground'} />
                            <div>
                              <div className="text-sm font-medium">{t(`feedback.categories.${value}`)}</div>
                              <div className="text-xs text-muted-foreground">{t(`feedback.categoryHints.${value}`)}</div>
                            </div>
                          </button>
                        )
                      })}
                    </div>
                  </div>

                  {category === 'bug' && (
                    <div className="grid gap-4">
                      <Field label={t('feedback.fields.summary')} htmlFor="feedback-bug-summary">
                        <Input id="feedback-bug-summary" value={bugSummary} onChange={(e) => setBugSummary(e.target.value)} required />
                      </Field>
                      <Field label={t('feedback.fields.steps')} htmlFor="feedback-bug-steps">
                        <textarea id="feedback-bug-steps" className="min-h-28 rounded-md border border-input bg-background px-3 py-2 text-sm" value={bugSteps} onChange={(e) => setBugSteps(e.target.value)} required />
                      </Field>
                      <div className="grid gap-4 md:grid-cols-2">
                        <Field label={t('feedback.fields.expected')} htmlFor="feedback-bug-expected">
                          <textarea id="feedback-bug-expected" className="min-h-24 rounded-md border border-input bg-background px-3 py-2 text-sm" value={bugExpected} onChange={(e) => setBugExpected(e.target.value)} required />
                        </Field>
                        <Field label={t('feedback.fields.actual')} htmlFor="feedback-bug-actual">
                          <textarea id="feedback-bug-actual" className="min-h-24 rounded-md border border-input bg-background px-3 py-2 text-sm" value={bugActual} onChange={(e) => setBugActual(e.target.value)} required />
                        </Field>
                      </div>
                      <Field label={t('feedback.fields.environment')} htmlFor="feedback-bug-environment" optional optionalLabel={t('feedback.optional')}>
                        <Input id="feedback-bug-environment" value={bugEnvironment} onChange={(e) => setBugEnvironment(e.target.value)} />
                      </Field>
                    </div>
                  )}

                  {category === 'feature' && (
                    <div className="grid gap-4">
                      <Field label={t('feedback.fields.summary')} htmlFor="feedback-feature-summary">
                        <Input id="feedback-feature-summary" value={featureSummary} onChange={(e) => setFeatureSummary(e.target.value)} required />
                      </Field>
                      <Field label={t('feedback.fields.problem')} htmlFor="feedback-feature-problem">
                        <textarea id="feedback-feature-problem" className="min-h-28 rounded-md border border-input bg-background px-3 py-2 text-sm" value={featureProblem} onChange={(e) => setFeatureProblem(e.target.value)} required />
                      </Field>
                      <Field label={t('feedback.fields.proposal')} htmlFor="feedback-feature-proposal">
                        <textarea id="feedback-feature-proposal" className="min-h-28 rounded-md border border-input bg-background px-3 py-2 text-sm" value={featureProposal} onChange={(e) => setFeatureProposal(e.target.value)} required />
                      </Field>
                      <Field label={t('feedback.fields.benefit')} htmlFor="feedback-feature-benefit" optional optionalLabel={t('feedback.optional')}>
                        <textarea id="feedback-feature-benefit" className="min-h-24 rounded-md border border-input bg-background px-3 py-2 text-sm" value={featureBenefit} onChange={(e) => setFeatureBenefit(e.target.value)} />
                      </Field>
                    </div>
                  )}

                  {category === 'question' && (
                    <div className="grid gap-4">
                      <Field label={t('feedback.fields.subject')} htmlFor="feedback-question-subject">
                        <Input id="feedback-question-subject" value={questionSubject} onChange={(e) => setQuestionSubject(e.target.value)} required />
                      </Field>
                      <Field label={t('feedback.fields.details')} htmlFor="feedback-question-details">
                        <textarea id="feedback-question-details" className="min-h-32 rounded-md border border-input bg-background px-3 py-2 text-sm" value={questionDetails} onChange={(e) => setQuestionDetails(e.target.value)} required />
                      </Field>
                    </div>
                  )}

                  {category === 'other' && (
                    <div className="grid gap-4">
                      <Field label={t('feedback.fields.subject')} htmlFor="feedback-other-subject">
                        <Input id="feedback-other-subject" value={otherSubject} onChange={(e) => setOtherSubject(e.target.value)} required />
                      </Field>
                      <Field label={t('feedback.fields.details')} htmlFor="feedback-other-message">
                        <textarea id="feedback-other-message" className="min-h-32 rounded-md border border-input bg-background px-3 py-2 text-sm" value={otherMessage} onChange={(e) => setOtherMessage(e.target.value)} required />
                      </Field>
                    </div>
                  )}

                  <p className="text-xs text-muted-foreground">{t('feedback.privacy')}</p>
                  {error && <p className="text-sm text-destructive">{error}</p>}
                  <div className="flex flex-wrap gap-2">
                    <Button type="submit" disabled={loading}>
                      {loading ? t('feedback.sending') : t('feedback.submit')}
                    </Button>
                    <a
                      href="https://github.com/krtw00/konbu/issues/new/choose"
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex items-center rounded-md border border-input px-4 py-2 text-sm hover:bg-accent"
                    >
                      {t('feedback.github')}
                    </a>
                  </div>
                </form>
              )}
            </CardContent>
          </Card>

          <div className="flex flex-col gap-4">
            <Card className="border-primary/20 bg-primary/5">
              <CardHeader className="pb-3">
                <CardTitle className="text-base">{t('feedback.guideTitle')}</CardTitle>
                <CardDescription>{t('feedback.guideDescription')}</CardDescription>
              </CardHeader>
              <CardContent className="space-y-3 text-sm text-muted-foreground">
                <p>{t('feedback.guideBug')}</p>
                <p>{t('feedback.guideFeature')}</p>
                <p>{t('feedback.guideResponse')}</p>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  )
}

function Field({
  label,
  htmlFor,
  children,
  optional = false,
  optionalLabel = 'optional',
}: {
  label: string
  htmlFor: string
  children: React.ReactNode
  optional?: boolean
  optionalLabel?: string
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <label className="text-sm font-medium" htmlFor={htmlFor}>
        {label}
        {optional && <span className="ml-1 text-xs text-muted-foreground">({optionalLabel})</span>}
      </label>
      {children}
    </div>
  )
}
