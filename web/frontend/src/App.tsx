import { useState, useCallback, useEffect, Suspense } from 'react'
import { useAppStore } from '@/stores/app'
import { useTranslation } from 'react-i18next'
import { Sidebar } from '@/components/layout/Sidebar'
import { MobileHeader } from '@/components/layout/MobileHeader'
import { CommandPalette } from '@/components/CommandPalette'
import { AppErrorBoundary } from '@/components/AppErrorBoundary'
import { lazyWithRetry } from '@/lib/lazy'

const MemoEditPage = lazyWithRetry(() => import('@/pages/MemoEditPage').then(m => ({ default: m.MemoEditPage })), 'memo-edit')
const ChatPage = lazyWithRetry(() => import('@/pages/ChatPage').then(m => ({ default: m.ChatPage })), 'chat')
const HomePage = lazyWithRetry(() => import('@/pages/HomePage').then(m => ({ default: m.HomePage })), 'home')
const MemosPage = lazyWithRetry(() => import('@/pages/MemosPage').then(m => ({ default: m.MemosPage })), 'memos')
const TodosPage = lazyWithRetry(() => import('@/pages/TodosPage').then(m => ({ default: m.TodosPage })), 'todos')
const CalendarPage = lazyWithRetry(() => import('@/pages/CalendarPage').then(m => ({ default: m.CalendarPage })), 'calendar')
const TablesPage = lazyWithRetry(() => import('@/pages/TablesPage').then(m => ({ default: m.TablesPage })), 'tables')
const ToolsPage = lazyWithRetry(() => import('@/pages/ToolsPage').then(m => ({ default: m.ToolsPage })), 'tools')
const SettingsPage = lazyWithRetry(() => import('@/pages/SettingsPage').then(m => ({ default: m.SettingsPage })), 'settings')
const SearchPage = lazyWithRetry(() => import('@/pages/SearchPage').then(m => ({ default: m.SearchPage })), 'search')
const LoginPage = lazyWithRetry(() => import('@/pages/LoginPage').then(m => ({ default: m.LoginPage })), 'login')
const SetupPage = lazyWithRetry(() => import('@/pages/SetupPage').then(m => ({ default: m.SetupPage })), 'setup')
const PublicPage = lazyWithRetry(() => import('@/pages/PublicPage').then(m => ({ default: m.PublicPage })), 'public')
const FeedbackPage = lazyWithRetry(() => import('@/pages/FeedbackPage').then(m => ({ default: m.FeedbackPage })), 'feedback')

function LoadingScreen() {
  return (
    <div className="flex h-screen items-center justify-center bg-background text-sm text-muted-foreground">
      Loading...
    </div>
  )
}

function ContentLoading() {
  return (
    <main className="flex-1">
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        Loading...
      </div>
    </main>
  )
}

function App() {
  const { t } = useTranslation()
  const { currentPage, setPage, isAuthenticated, isLoading, needsSetup, checkAuth } = useAppStore()
  const [editingMemoId, setEditingMemoId] = useState<string | null>(null)
  const publicToken = typeof window !== 'undefined'
    ? window.location.pathname.match(/^\/public\/([a-zA-Z0-9]+)/)?.[1] ?? null
    : null
  const isFeedbackPage = typeof window !== 'undefined' && window.location.pathname === '/feedback'

  useEffect(() => {
    checkAuth()
  }, [checkAuth])

  const handleEditMemo = useCallback((id: string) => {
    setEditingMemoId(id)
    setPage('memo-edit')
  }, [setPage])

  useEffect(() => {
    function onOpenMemo(e: Event) {
      const id = (e as CustomEvent).detail
      if (id) handleEditMemo(id)
    }
    window.addEventListener('open-memo', onOpenMemo)
    return () => window.removeEventListener('open-memo', onOpenMemo)
  }, [handleEditMemo])

  const handleEditTable = useCallback((id: string) => {
    setEditingMemoId(id)
    setPage('table-edit')
  }, [setPage])

  const handleCloseMemoEdit = useCallback(() => {
    setEditingMemoId(null)
    const prev = useAppStore.getState().currentPage
    setPage(prev === 'table-edit' ? 'tables' : 'memos')
  }, [setPage])

  if (publicToken) {
    return (
      <Suspense fallback={<LoadingScreen />}>
        <PublicPage token={publicToken} />
      </Suspense>
    )
  }

  if (isFeedbackPage) {
    return (
      <Suspense fallback={<LoadingScreen />}>
        <FeedbackPage />
      </Suspense>
    )
  }

  if (isLoading) {
    return <LoadingScreen />
  }

  if (needsSetup) {
    return (
      <Suspense fallback={<LoadingScreen />}>
        <SetupPage />
      </Suspense>
    )
  }

  if (!isAuthenticated) {
    return (
      <Suspense fallback={<LoadingScreen />}>
        <LoginPage />
      </Suspense>
    )
  }

  return (
    <div className="flex flex-col md:flex-row h-screen bg-background">
      <Sidebar />
      <div className="flex-1 flex flex-col min-h-0">
        <MobileHeader />
        <AppErrorBoundary
          resetKey={`${currentPage}:${editingMemoId ?? ''}`}
          title={t('appError.title')}
          description={t('appError.description')}
          reloadLabel={t('appError.reload')}
          homeLabel={t('appError.home')}
          onHome={() => {
            setEditingMemoId(null)
            setPage('home')
          }}
        >
          <Suspense fallback={<ContentLoading />}>
            {(currentPage === 'memo-edit' || currentPage === 'table-edit') && editingMemoId ? (
              <main className="flex-1 overflow-hidden">
                <MemoEditPage memoId={editingMemoId} onClose={handleCloseMemoEdit} />
              </main>
            ) : currentPage === 'chat' ? (
              <main className="flex-1 overflow-hidden">
                <ChatPage />
              </main>
            ) : currentPage === 'calendar' ? (
              <main className="flex-1 overflow-auto p-2 md:p-4">
                <CalendarPage />
              </main>
            ) : (
              <main className="flex-1 overflow-auto">
                <div className="max-w-5xl mx-auto p-4 md:p-6">
                  {currentPage === 'home' && <HomePage onEditMemo={handleEditMemo} />}
                  {currentPage === 'memos' && <MemosPage onEditMemo={handleEditMemo} />}
                  {currentPage === 'tables' && <TablesPage onEditTable={handleEditTable} />}
                  {currentPage === 'todos' && <TodosPage />}
                  {currentPage === 'tools' && <ToolsPage />}
                  {currentPage === 'search' && <SearchPage onOpenMemo={handleEditMemo} />}
                  {currentPage === 'settings' && <SettingsPage />}
                </div>
              </main>
            )}
          </Suspense>
        </AppErrorBoundary>
      </div>
      <CommandPalette onOpenMemo={handleEditMemo} />
    </div>
  )
}

export default App
