import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { useMessages, useSendMessage } from '@/hooks/useClaribot'
import { Send, ChevronDown, ChevronRight, MessageSquare } from 'lucide-react'

export default function Messages() {
  const { data: messagesData } = useMessages()
  const sendMessage = useSendMessage()
  const [input, setInput] = useState('')
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())

  const messageItems = parseItems(messagesData?.data)

  const handleSend = async () => {
    if (!input.trim()) return
    await sendMessage.mutateAsync(input.trim())
    setInput('')
  }

  const toggleExpand = (id: number) => {
    setExpandedIds(prev => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      handleSend()
    }
  }

  return (
    <div className="space-y-6 max-w-4xl">
      <h1 className="text-3xl font-bold">Messages</h1>

      {/* Send Form */}
      <Card>
        <CardContent className="p-4">
          <div className="space-y-2">
            <Textarea
              placeholder="Send a message to Claude... (Ctrl+Enter to send)"
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              rows={3}
            />
            <div className="flex justify-end">
              <Button
                onClick={handleSend}
                disabled={sendMessage.isPending || !input.trim()}
              >
                <Send className="h-4 w-4 mr-1" />
                {sendMessage.isPending ? 'Sending...' : 'Send'}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Message List */}
      <div className="space-y-3">
        {messageItems.map((msg: any) => {
          const id = msg.id || msg.ID
          const content = msg.content || msg.Content || ''
          const source = msg.source || msg.Source || 'cli'
          const status = msg.status || msg.Status || 'pending'
          const result = msg.result || msg.Result || ''
          const error = msg.error || msg.Error || ''
          const createdAt = msg.created_at || msg.CreatedAt || ''
          const isExpanded = expandedIds.has(id)
          const hasResult = result || error

          return (
            <Card key={id}>
              <CardContent className="p-4">
                <div className="flex items-start gap-3">
                  <MessageSquare className="h-5 w-5 text-muted-foreground mt-0.5 shrink-0" />
                  <div className="flex-1 min-w-0 space-y-2">
                    {/* Header */}
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="text-xs text-muted-foreground">#{id}</span>
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

                    {/* Content */}
                    <p className="text-sm">{content}</p>

                    {/* Result Toggle */}
                    {hasResult && (
                      <div>
                        <button
                          className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
                          onClick={() => toggleExpand(id)}
                        >
                          {isExpanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                          {status === 'processing' ? 'Live output' : 'Show result'}
                        </button>
                        {isExpanded && (
                          <pre className="mt-2 text-sm whitespace-pre-wrap bg-muted rounded p-3 max-h-[400px] overflow-auto">
                            {error ? `Error: ${error}` : result}
                          </pre>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {messageItems.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          No messages yet. Send a message above to get started.
        </div>
      )}
    </div>
  )
}

function SourceBadge({ source }: { source: string }) {
  const icons: Record<string, string> = {
    telegram: '\uD83D\uDCE8',
    cli: '\uD83D\uDCBB',
    schedule: '\u23F0',
  }
  return (
    <Badge variant="outline" className="text-xs gap-1">
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

function parseItems(data: any): any[] {
  if (!data) return []
  if (Array.isArray(data)) return data
  if (data.items && Array.isArray(data.items)) return data.items
  return []
}
