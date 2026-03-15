import { useState, useCallback, useEffect } from 'react'
import { useAppStore } from '@/stores/app'
import { Sidebar } from '@/components/layout/Sidebar'
import { MobileHeader } from '@/components/layout/MobileHeader'
import { CommandPalette } from '@/components/CommandPalette'
import { ChatPanel } from '@/components/ChatPanel'
import { HomePage } from '@/pages/HomePage'
import { MemosPage } from '@/pages/MemosPage'
import { MemoEditPage } from '@/pages/MemoEditPage'
import { TodosPage } from '@/pages/TodosPage'
import { CalendarPage } from '@/pages/CalendarPage'
import { ToolsPage } from '@/pages/ToolsPage'
import { SettingsPage } from '@/pages/SettingsPage'
import { LoginPage } from '@/pages/LoginPage'
import { SetupPage } from '@/pages/SetupPage'

function App() {
  const { currentPage, setPage, isAuthenticated, isLoading, needsSetup, checkAuth } = useAppStore()
  const [editingMemoId, setEditingMemoId] = useState<string | null>(null)

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

  const handleCloseMemoEdit = useCallback(() => {
    setEditingMemoId(null)
    setPage('memos')
  }, [setPage])

  if (isLoading) {
    return <div className="flex h-screen items-center justify-center bg-background" />
  }

  if (needsSetup) {
    return <SetupPage />
  }

  if (!isAuthenticated) {
    return <LoginPage />
  }

  return (
    <div className="flex flex-col md:flex-row h-screen bg-background">
      <Sidebar />
      <div className="flex-1 flex flex-col min-h-0">
        <MobileHeader />
        {currentPage === 'memo-edit' && editingMemoId ? (
          <main className="flex-1 overflow-hidden">
            <MemoEditPage memoId={editingMemoId} onClose={handleCloseMemoEdit} />
          </main>
        ) : (
          <main className="flex-1 overflow-auto">
            <div className="max-w-5xl mx-auto p-4 md:p-6">
              {currentPage === 'home' && <HomePage onEditMemo={handleEditMemo} />}
              {currentPage === 'memos' && <MemosPage onEditMemo={handleEditMemo} />}
              {currentPage === 'todos' && <TodosPage />}
              {currentPage === 'calendar' && <CalendarPage />}
              {currentPage === 'tools' && <ToolsPage />}
              {currentPage === 'settings' && <SettingsPage />}
            </div>
          </main>
        )}
      </div>
      <CommandPalette onOpenMemo={handleEditMemo} />
      <ChatPanel />
    </div>
  )
}

export default App
