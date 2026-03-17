import { useState, useCallback, useEffect, lazy, Suspense } from 'react'
import { useAppStore } from '@/stores/app'
import { Sidebar } from '@/components/layout/Sidebar'
import { MobileHeader } from '@/components/layout/MobileHeader'
import { CommandPalette } from '@/components/CommandPalette'

const MemoEditPage = lazy(() => import('@/pages/MemoEditPage').then(m => ({ default: m.MemoEditPage })))
const ChatPage = lazy(() => import('@/pages/ChatPage').then(m => ({ default: m.ChatPage })))
const HomePage = lazy(() => import('@/pages/HomePage').then(m => ({ default: m.HomePage })))
const MemosPage = lazy(() => import('@/pages/MemosPage').then(m => ({ default: m.MemosPage })))
const TodosPage = lazy(() => import('@/pages/TodosPage').then(m => ({ default: m.TodosPage })))
const CalendarPage = lazy(() => import('@/pages/CalendarPage').then(m => ({ default: m.CalendarPage })))
const TablesPage = lazy(() => import('@/pages/TablesPage').then(m => ({ default: m.TablesPage })))
const ToolsPage = lazy(() => import('@/pages/ToolsPage').then(m => ({ default: m.ToolsPage })))
const SettingsPage = lazy(() => import('@/pages/SettingsPage').then(m => ({ default: m.SettingsPage })))
const SearchPage = lazy(() => import('@/pages/SearchPage').then(m => ({ default: m.SearchPage })))
const LoginPage = lazy(() => import('@/pages/LoginPage').then(m => ({ default: m.LoginPage })))
const SetupPage = lazy(() => import('@/pages/SetupPage').then(m => ({ default: m.SetupPage })))
const PublicPage = lazy(() => import('@/pages/PublicPage').then(m => ({ default: m.PublicPage })))

function LoadingScreen() {
  return <div className="flex h-screen items-center justify-center bg-background" />
}

function App() {
  const { currentPage, setPage, isAuthenticated, isLoading, needsSetup, checkAuth } = useAppStore()
  const [editingMemoId, setEditingMemoId] = useState<string | null>(null)
  const publicToken = typeof window !== 'undefined'
    ? window.location.pathname.match(/^\/public\/([a-zA-Z0-9]+)/)?.[1] ?? null
    : null

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
        <Suspense fallback={<div className="flex-1 bg-background" />}>
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
      </div>
      <CommandPalette onOpenMemo={handleEditMemo} />
    </div>
  )
}

export default App
