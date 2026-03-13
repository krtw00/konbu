import { useState, useCallback } from 'react'
import { useAppStore } from '@/stores/app'
import { Sidebar } from '@/components/layout/Sidebar'
import { BottomNav } from '@/components/layout/BottomNav'
import { ThemeSwitcher } from '@/components/layout/ThemeSwitcher'
import { CommandPalette } from '@/components/CommandPalette'
import { HomePage } from '@/pages/HomePage'
import { MemosPage } from '@/pages/MemosPage'
import { MemoEditPage } from '@/pages/MemoEditPage'
import { TodosPage } from '@/pages/TodosPage'
import { CalendarPage } from '@/pages/CalendarPage'
import { ToolsPage } from '@/pages/ToolsPage'

function App() {
  const { currentPage, setPage, theme } = useAppStore()
  const [editingMemoId, setEditingMemoId] = useState<string | null>(null)

  const handleEditMemo = useCallback((id: string) => {
    setEditingMemoId(id)
    setPage('memo-edit')
  }, [setPage])

  const handleCloseMemoEdit = useCallback(() => {
    setEditingMemoId(null)
    setPage('memos')
  }, [setPage])

  return (
    <div className="flex h-screen bg-background" data-theme={theme}>
      <Sidebar />
      <main className="flex-1 overflow-auto pb-16 md:pb-0">
        <div className="max-w-5xl mx-auto p-4 md:p-6">
          {currentPage === 'home' && <HomePage />}
          {currentPage === 'memos' && <MemosPage onEditMemo={handleEditMemo} />}
          {currentPage === 'memo-edit' && editingMemoId && (
            <MemoEditPage memoId={editingMemoId} onClose={handleCloseMemoEdit} />
          )}
          {currentPage === 'todos' && <TodosPage />}
          {currentPage === 'calendar' && <CalendarPage />}
          {currentPage === 'tools' && <ToolsPage />}
        </div>
      </main>
      <BottomNav />
      <ThemeSwitcher />
      <CommandPalette onOpenMemo={handleEditMemo} />
    </div>
  )
}

export default App
