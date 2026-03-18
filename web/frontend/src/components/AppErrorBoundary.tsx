import { Component, type ReactNode } from 'react'
import { Button } from '@/components/ui/button'

interface AppErrorBoundaryProps {
  children: ReactNode
  resetKey: string
  title: string
  description: string
  reloadLabel: string
  homeLabel: string
  onHome: () => void
}

interface AppErrorBoundaryState {
  hasError: boolean
}

export class AppErrorBoundary extends Component<AppErrorBoundaryProps, AppErrorBoundaryState> {
  state: AppErrorBoundaryState = { hasError: false }

  static getDerivedStateFromError() {
    return { hasError: true }
  }

  componentDidCatch(error: Error) {
    console.error('App render error', error)
  }

  componentDidUpdate(prevProps: AppErrorBoundaryProps) {
    if (this.state.hasError && prevProps.resetKey !== this.props.resetKey) {
      this.setState({ hasError: false })
    }
  }

  render() {
    if (!this.state.hasError) return this.props.children

    return (
      <main className="flex-1 overflow-auto">
        <div className="mx-auto flex min-h-full max-w-xl items-center justify-center p-6">
          <div className="w-full rounded-2xl border border-border bg-background p-6 text-center shadow-sm">
            <h2 className="text-lg font-semibold">{this.props.title}</h2>
            <p className="mt-2 text-sm text-muted-foreground">{this.props.description}</p>
            <div className="mt-4 flex flex-wrap items-center justify-center gap-2">
              <Button type="button" onClick={() => window.location.reload()}>
                {this.props.reloadLabel}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  this.setState({ hasError: false })
                  this.props.onHome()
                }}
              >
                {this.props.homeLabel}
              </Button>
            </div>
          </div>
        </div>
      </main>
    )
  }
}
