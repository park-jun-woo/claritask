import { useState, useRef, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { useMessages, useMessage, useSendMessage, useStatus } from '@/hooks/useClaribot'
import { Send, MessageSquare, ArrowLeft } from 'lucide-react'
import { MarkdownRenderer } from '@/components/MarkdownRenderer'
import { ChatBubble } from '@/components/ChatBubble'

type MobileView = 'chat' | 'detail'

export default function Messages() {
  const { data: statusData } = useStatus()

  // Get current project from status (📌 project-id — ...)
  const currentProject = statusData?.message?.match(/📌 (.+?) —/u)?.[1]
  const isGlobal = !currentProject

  // When global: show all messages; when project selected: filter by project
  const { data: messagesData } = useMessages(isGlobal, isGlobal ? undefined : currentProject)
  const sendMessage = useSendMessage()
  const [input, setInput] = useState('')
  const [selectedMessageId, setSelectedMessageId] = useState<number | null>(null)
  const [mobileView, setMobileView] = useState<MobileView>('chat')
  const [pendingMessages, setPendingMessages] = useState<Array<{ id: string; content: string; created_at: string }>>([])

  const { data: messageDetail } = useMessage(selectedMessageId ?? undefined)
  const selectedMessage = messageDetail?.data ?? null

  const messageItems = parseItems(messagesData?.data)

  // Merge pending messages with actual messages
  const allMessages = [
    ...messageItems,
    ...pendingMessages.map(pm => ({
      id: pm.id,
      content: pm.content,
      status: 'pending',
      source: 'gui',
      created_at: pm.created_at,
      result: '',
      error: '',
    }))
  ]

  const sortedMessages = [...allMessages].sort((a, b) => {
    const ta = new Date(a.created_at || a.CreatedAt || 0).getTime()
    const tb = new Date(b.created_at || b.CreatedAt || 0).getTime()
    return ta - tb
  })

  const chatEndRef = useRef<HTMLDivElement>(null)
  const isInitialLoad = useRef(true)

  // Auto-scroll to bottom when messages change
  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: isInitialLoad.current ? 'instant' : 'smooth' })
    isInitialLoad.current = false
  }, [sortedMessages.length])

  const handleSend = () => {
    if (!input.trim()) return
    const content = input.trim()
    const tempId = `pending-${Date.now()}`
    const now = new Date().toISOString()

    // Add to pending messages immediately
    setPendingMessages(prev => [...prev, { id: tempId, content, created_at: now }])
    setInput('')

    // Send to server
    sendMessage.mutate(
      { content, projectId: currentProject },
      {
        onSuccess: () => {
          // Remove from pending after successful send (actual message will appear from query)
          setPendingMessages(prev => prev.filter(m => m.id !== tempId))
        },
        onError: () => {
          // Remove from pending on error
          setPendingMessages(prev => prev.filter(m => m.id !== tempId))
        },
      }
    )
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      handleSend()
    }
  }

  const handleDetailClick = (id: number) => {
    setSelectedMessageId(id)
    setMobileView('detail')
  }

  const handleBackToChat = () => {
    setMobileView('chat')
  }

  // Group messages by date for date separators
  const groupedMessages = groupByDate(sortedMessages)

  return (
    <div className="flex gap-0 md:gap-4 flex-1 min-h-0 h-full overflow-hidden">
      {/* Left Panel: Chat Area */}
      <div className={`w-full md:w-1/2 flex flex-col border rounded-lg ${mobileView === 'detail' ? 'hidden md:flex' : 'flex'}`}>
        {/* Chat Header */}
        <div className="px-4 py-3 border-b shrink-0">
          <h1 className="text-lg font-semibold flex items-center gap-2">
            <MessageSquare className="h-5 w-5" />
            Messages
            {currentProject && (
              <span className="text-xs font-normal text-muted-foreground bg-muted px-2 py-0.5 rounded">
                {currentProject}
              </span>
            )}
          </h1>
        </div>

        {/* Chat Messages */}
        <ScrollArea className="flex-1 px-4">
          <div className="py-4">
            {groupedMessages.map(({ date, messages: msgs }) => (
              <div key={date}>
                {/* Date Separator */}
                <DateSeparator date={date} />

                {/* Message pairs: user content + bot result */}
                {msgs.map((msg: any) => {
                  const id = msg.id || msg.ID
                  const content = msg.content || msg.Content || ''
                  const status = msg.status || msg.Status || 'pending'
                  const source = msg.source || msg.Source || 'cli'
                  const result = msg.result || msg.Result || ''
                  const error = msg.error || msg.Error || ''
                  const createdAt = msg.created_at || msg.CreatedAt || ''
                  const completedAt = msg.completed_at || msg.CompletedAt || ''
                  const isSelected = id === selectedMessageId

                  return (
                    <div key={id}>
                      {/* User message bubble */}
                      <ChatBubble
                        type="user"
                        content={content}
                        source={source}
                        time={formatTime(createdAt)}
                      />

                      {/* Bot response bubble */}
                      <ChatBubble
                        type="bot"
                        content={
                          error
                            ? error.slice(0, 120) + (error.length > 120 ? '...' : '')
                            : result
                              ? getResponseSummary(result)
                              : statusMessage(status)
                        }
                        result={result || undefined}
                        status={status}
                        time={formatTime(completedAt || createdAt)}
                        isSelected={isSelected}
                        onDetailClick={() => handleDetailClick(id)}
                      />
                    </div>
                  )
                })}
              </div>
            ))}
            {sortedMessages.length === 0 && (
              <div className="flex flex-col items-center justify-center py-20 text-muted-foreground">
                <MessageSquare className="h-10 w-10 mb-2 opacity-30" />
                <p className="text-sm">메시지가 없습니다</p>
              </div>
            )}
            <div ref={chatEndRef} />
          </div>
        </ScrollArea>

        {/* Input Area - Fixed at bottom */}
        <div className="p-3 border-t shrink-0">
          <div className="flex gap-2 items-end">
            <Textarea
              placeholder="메시지를 입력하세요... (Ctrl+Enter)"
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              rows={2}
              className="text-sm resize-none flex-1"
            />
            <Button
              size="sm"
              onClick={handleSend}
              disabled={!input.trim()}
              className="h-[52px] px-4 shrink-0"
            >
              <Send className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Right Panel: Message Detail */}
      <div className={`w-full md:w-1/2 border rounded-lg flex flex-col min-w-0 ${mobileView === 'chat' ? 'hidden md:flex' : 'flex'}`}>
        {selectedMessage ? (
          <MessageDetail
            message={selectedMessage}
            onBack={handleBackToChat}
          />
        ) : (
          <div className="flex-1 flex items-center justify-center text-muted-foreground">
            <div className="text-center space-y-2">
              <MessageSquare className="h-12 w-12 mx-auto opacity-30" />
              <p className="text-sm">메시지를 선택하면 상세 내용을 볼 수 있습니다</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

// --- Sub-components ---

function DateSeparator({ date }: { date: string }) {
  return (
    <div className="flex items-center gap-3 my-4">
      <div className="flex-1 h-px bg-border" />
      <span className="text-[11px] text-muted-foreground font-medium">{date}</span>
      <div className="flex-1 h-px bg-border" />
    </div>
  )
}

function MessageDetail({ message, onBack }: { message: any; onBack: () => void }) {
  const id = message.id || message.ID
  const content = message.content || message.Content || ''
  const source = message.source || message.Source || 'cli'
  const status = message.status || message.Status || 'pending'
  const result = message.result || message.Result || ''
  const error = message.error || message.Error || ''
  const createdAt = message.created_at || message.CreatedAt || ''

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="p-4 border-b shrink-0">
        <div className="flex items-center gap-2 flex-wrap">
          <Button variant="ghost" size="sm" onClick={onBack} className="md:hidden mr-1 -ml-2">
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <h3 className="font-semibold">Message #{id}</h3>
          <Badge variant={
            status === 'done' ? 'success'
              : status === 'processing' ? 'warning'
                : status === 'failed' ? 'destructive'
                  : 'secondary'
          }>
            {status}
          </Badge>
          <SourceBadge source={source} />
          <span className="text-xs text-muted-foreground ml-auto">{formatTime(createdAt)}</span>
        </div>
      </div>

      {/* Body */}
      <ScrollArea className="flex-1 p-4">
        <div className="space-y-4">
          {/* Content */}
          <div>
            <h5 className="text-sm font-medium text-muted-foreground mb-1">Content</h5>
            <p className="text-sm whitespace-pre-wrap">{content}</p>
          </div>

          {/* Result */}
          {result && (
            <>
              <Separator />
              <div>
                <h5 className="text-sm font-medium text-muted-foreground mb-1">Result</h5>
                <div className="bg-muted rounded p-3">
                  <MarkdownRenderer content={result} />
                </div>
              </div>
            </>
          )}

          {/* Error */}
          {error && (
            <>
              <Separator />
              <div>
                <h5 className="text-sm font-medium text-destructive mb-1">Error</h5>
                <pre className="text-sm whitespace-pre-wrap bg-destructive/10 rounded p-3 text-destructive">
                  {error}
                </pre>
              </div>
            </>
          )}
        </div>
      </ScrollArea>
    </div>
  )
}

function SourceBadge({ source }: { source: string }) {
  const icons: Record<string, string> = {
    telegram: '\uD83D\uDCE8',
    cli: '\uD83D\uDCBB',
    gui: '\uD83D\uDDA5\uFE0F',
    schedule: '\u23F0',
  }
  return (
    <Badge variant="outline" className="text-xs gap-1 min-h-0">
      {icons[source] || ''} {source}
    </Badge>
  )
}

function formatTime(ts: string): string {
  if (!ts) return ''
  try {
    const d = new Date(ts)
    return d.toLocaleString('ko-KR', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
  } catch {
    return ts
  }
}

function formatDate(ts: string): string {
  if (!ts) return ''
  try {
    const d = new Date(ts)
    return d.toLocaleDateString('ko-KR', { year: 'numeric', month: 'long', day: 'numeric' })
  } catch {
    return ts
  }
}

function parseItems(data: any): any[] {
  if (!data) return []
  if (Array.isArray(data)) return data
  if (data.items && Array.isArray(data.items)) return data.items
  return []
}

function getResponseSummary(result: string): string {
  if (!result) return ''
  const plain = result.replace(/[#*`>\-\[\]()!]/g, '').trim()
  const first = plain.split('\n').find(l => l.trim())?.trim() || ''
  return first.length > 100 ? first.slice(0, 100) + '...' : first
}

function statusMessage(status: string): string {
  switch (status) {
    case 'pending': return '대기 중...'
    case 'processing': return '처리 중...'
    case 'done': return '완료'
    case 'failed': return '실패'
    default: return status
  }
}

function groupByDate(messages: any[]): { date: string; messages: any[] }[] {
  const groups: Map<string, any[]> = new Map()
  for (const msg of messages) {
    const ts = msg.created_at || msg.CreatedAt || ''
    const dateKey = formatDate(ts) || 'Unknown'
    if (!groups.has(dateKey)) {
      groups.set(dateKey, [])
    }
    groups.get(dateKey)!.push(msg)
  }
  return Array.from(groups.entries()).map(([date, messages]) => ({ date, messages }))
}
