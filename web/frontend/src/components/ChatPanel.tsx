import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useChatStore } from '@/stores/chat'
import { useAppStore } from '@/stores/app'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { MessageCircle, X, Send, Plus, Trash2, Loader2 } from 'lucide-react'
import { UpgradePrompt } from '@/components/UpgradePrompt'

export function ChatPanel() {
  const { t } = useTranslation()
  const user = useAppStore(s => s.user)
  const {
    isOpen, toggle, close,
    sessions, loadSessions,
    currentSessionId, selectSession, newSession, deleteSession,
    messages, sendMessage,
    isStreaming, streamingContent, toolStatus,
  } = useChatStore()
  const [input, setInput] = useState('')
  const [showSessions, setShowSessions] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (isOpen && user) loadSessions()
  }, [isOpen, user])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, streamingContent])

  if (!user) return null

  const isSponsor = user.plan === 'sponsor'

  function handleSend() {
    if (!input.trim() || isStreaming) return
    sendMessage(input.trim())
    setInput('')
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <>
      {/* Toggle button */}
      <button
        onClick={toggle}
        className={`fixed bottom-4 right-4 z-50 w-12 h-12 rounded-full bg-primary text-primary-foreground shadow-lg flex items-center justify-center hover:opacity-90 transition-opacity ${isOpen ? 'hidden' : ''}`}
      >
        <MessageCircle size={20} />
      </button>

      {/* Panel */}
      {isOpen && (
        <div className="fixed right-0 top-0 h-full w-[400px] max-w-full bg-background border-l border-border shadow-xl z-50 flex flex-col">
          {/* Header */}
          <div className="flex items-center justify-between p-3 border-b border-border">
            <div className="flex items-center gap-2">
              <h2 className="font-semibold text-sm">{t('chat.title')}</h2>
              <Button variant="ghost" size="sm" onClick={() => setShowSessions(!showSessions)} className="text-xs">
                {t('chat.sessions')}
              </Button>
            </div>
            <div className="flex items-center gap-1">
              <Button variant="ghost" size="sm" onClick={newSession}>
                <Plus size={14} />
              </Button>
              <Button variant="ghost" size="sm" onClick={close}>
                <X size={14} />
              </Button>
            </div>
          </div>

          {/* Session list */}
          {showSessions && (
            <div className="border-b border-border max-h-48 overflow-y-auto">
              {sessions.length === 0 ? (
                <p className="text-xs text-muted-foreground p-3">{t('chat.noSessions')}</p>
              ) : (
                sessions.map(s => (
                  <div
                    key={s.id}
                    className={`flex items-center justify-between px-3 py-2 text-sm cursor-pointer hover:bg-accent/50 ${s.id === currentSessionId ? 'bg-accent' : ''}`}
                    onClick={() => { selectSession(s.id); setShowSessions(false) }}
                  >
                    <span className="truncate flex-1">{s.title || t('chat.newSession')}</span>
                    <button
                      onClick={(e) => { e.stopPropagation(); deleteSession(s.id) }}
                      className="text-muted-foreground hover:text-destructive ml-2"
                    >
                      <Trash2 size={12} />
                    </button>
                  </div>
                ))
              )}
            </div>
          )}

          {/* Messages area */}
          <div className="flex-1 overflow-y-auto p-3 space-y-3">
            {!isSponsor ? (
              <UpgradePrompt feature="chat" />
            ) : messages.length === 0 && !isStreaming ? (
              <div className="text-center py-8">
                <MessageCircle size={32} className="mx-auto mb-2 text-muted-foreground" />
                <p className="text-sm text-muted-foreground">{t('chat.placeholder')}</p>
              </div>
            ) : (
              <>
                {messages.map(msg => (
                  <div key={msg.id} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                    <div className={`max-w-[85%] rounded-lg px-3 py-2 text-sm whitespace-pre-wrap ${
                      msg.role === 'user'
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-muted'
                    }`}>
                      {msg.content}
                    </div>
                  </div>
                ))}
                {isStreaming && (
                  <div className="flex justify-start">
                    <div className="max-w-[85%] rounded-lg px-3 py-2 text-sm bg-muted">
                      {toolStatus && (
                        <div className="flex items-center gap-1 text-xs text-muted-foreground mb-1">
                          <Loader2 size={10} className="animate-spin" />
                          {toolStatus}
                        </div>
                      )}
                      {streamingContent || (
                        <Loader2 size={14} className="animate-spin text-muted-foreground" />
                      )}
                    </div>
                  </div>
                )}
              </>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Input */}
          {isSponsor && (
            <div className="p-3 border-t border-border">
              <div className="flex gap-2">
                <Input
                  value={input}
                  onChange={e => setInput(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder={t('chat.inputPlaceholder')}
                  disabled={isStreaming}
                  className="flex-1"
                />
                <Button size="sm" onClick={handleSend} disabled={!input.trim() || isStreaming}>
                  <Send size={14} />
                </Button>
              </div>
            </div>
          )}
        </div>
      )}
    </>
  )
}
