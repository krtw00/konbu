import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { useChatStore } from '@/stores/chat'
import { useAppStore } from '@/stores/app'
import { Button } from '@/components/ui/button'
import { Send, Plus, Trash2, Loader2, MessageCircle, Menu } from 'lucide-react'

export function ChatPage() {
  const { t } = useTranslation()
  const user = useAppStore(s => s.user)
  const {
    sessions, loadSessions,
    currentSessionId, selectSession, newSession, deleteSession,
    messages, sendMessage,
    isStreaming, streamingContent, toolStatus,
  } = useChatStore()
  const [input, setInput] = useState('')
  const [drawerOpen, setDrawerOpen] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    if (user) loadSessions()
  }, [user])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, streamingContent])

  if (!user) return null

  function autoResize(el: HTMLTextAreaElement) {
    el.style.height = 'auto'
    el.style.height = Math.min(el.scrollHeight, 160) + 'px'
  }

  function handleSend() {
    if (!input.trim() || isStreaming) return
    sendMessage(input.trim())
    setInput('')
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault()
      handleSend()
    }
  }

  function handleSelectSession(id: string) {
    selectSession(id)
    setDrawerOpen(false)
  }


  const sessionList = (
    <>
      <div className="p-3 border-b border-border flex items-center justify-between">
        <h2 className="font-semibold text-sm">{t('chat.sessions')}</h2>
        <Button variant="ghost" size="sm" onClick={() => { newSession(); setDrawerOpen(false) }}>
          <Plus size={14} />
        </Button>
      </div>
      <div className="flex-1 overflow-y-auto">
        {sessions.length === 0 ? (
          <p className="text-xs text-muted-foreground p-3">{t('chat.noSessions')}</p>
        ) : (
          sessions.map(s => (
            <div
              key={s.id}
              className={`group flex items-center justify-between px-3 py-2.5 text-sm cursor-pointer hover:bg-accent/50 ${s.id === currentSessionId ? 'bg-accent' : ''}`}
              onClick={() => handleSelectSession(s.id)}
            >
              <span className="truncate flex-1">{s.title || t('chat.newSession')}</span>
              <button
                onClick={(e) => { e.stopPropagation(); deleteSession(s.id) }}
                className="text-muted-foreground hover:text-destructive ml-2 opacity-100 md:opacity-0 md:group-hover:opacity-100 shrink-0"
              >
                <Trash2 size={12} />
              </button>
            </div>
          ))
        )}
      </div>
    </>
  )

  return (
    <div className="flex h-full relative">
      {/* Session list: desktop */}
      <div className="hidden md:flex w-64 border-r border-border flex-col">
        {sessionList}
      </div>

      {/* Session list: mobile drawer */}
      {drawerOpen && (
        <>
          <div className="fixed inset-0 bg-black/40 z-40 md:hidden" onClick={() => setDrawerOpen(false)} />
          <div className="fixed left-0 top-0 bottom-0 w-64 bg-background border-r border-border z-50 flex flex-col md:hidden">
            {sessionList}
          </div>
        </>
      )}

      {/* Chat area */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Mobile header */}
        <div className="md:hidden flex items-center gap-2 px-3 py-2 border-b border-border">
          <button onClick={() => setDrawerOpen(true)} className="p-1 text-muted-foreground">
            <Menu size={18} />
          </button>
          <span className="text-sm font-medium truncate flex-1">
            {sessions.find(s => s.id === currentSessionId)?.title || t('chat.newSession')}
          </span>
          <Button variant="ghost" size="sm" onClick={() => { newSession(); setDrawerOpen(false) }}>
            <Plus size={14} />
          </Button>
        </div>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto p-3 md:p-4 space-y-3">
          {messages.length === 0 && !isStreaming ? (
            <div className="flex flex-col items-center justify-center h-full text-center">
              <MessageCircle size={40} className="mb-3 text-muted-foreground" />
              <p className="text-muted-foreground text-sm">{t('chat.placeholder')}</p>
            </div>
          ) : (
            <>
              {messages.filter(msg => msg.role === 'user' || (msg.role === 'assistant' && msg.content)).map(msg => (
                <div key={msg.id} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                  <div className={`max-w-[85%] md:max-w-[70%] rounded-lg px-3 py-2 text-sm whitespace-pre-wrap break-words ${
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
                  <div className="max-w-[85%] md:max-w-[70%] rounded-lg px-3 py-2 text-sm bg-muted">
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
        <div className="p-3 md:p-4 border-t border-border">
          <div className="flex gap-2 items-end">
            <textarea
              ref={textareaRef}
              value={input}
              onChange={e => { setInput(e.target.value); autoResize(e.target) }}
              onKeyDown={handleKeyDown}
              placeholder={t('chat.inputPlaceholder')}
              disabled={isStreaming}
              rows={1}
              className="flex-1 min-w-0 resize-none rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              style={{ maxHeight: '160px', minHeight: '38px' }}
            />
            <Button className="shrink-0" onClick={handleSend} disabled={!input.trim() || isStreaming}>
              <Send size={16} />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
