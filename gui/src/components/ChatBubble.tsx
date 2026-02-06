import { Badge } from '@/components/ui/badge'
import { Loader2, Check, AlertCircle, Eye } from 'lucide-react'

interface ChatBubbleProps {
  type: 'user' | 'bot'
  content: string
  status?: string
  source?: string
  result?: string
  time: string
  onDetailClick?: () => void
  isSelected?: boolean
}

export function ChatBubble({ type, content, status, source, result, time, onDetailClick, isSelected }: ChatBubbleProps) {
  const isUser = type === 'user'

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-3`}>
      <div className={`max-w-[80%] ${isUser ? 'order-1' : 'order-1'}`}>
        {/* Source label for user messages */}
        {isUser && source && (
          <div className="flex justify-end mb-0.5">
            <span className="text-[10px] text-muted-foreground">{sourceLabel(source)}</span>
          </div>
        )}

        {/* Bubble */}
        <div
          className={`rounded-2xl px-3.5 py-2.5 text-sm break-words transition-shadow ${
            isUser
              ? 'bg-primary text-primary-foreground rounded-br-md'
              : `bg-muted rounded-bl-md ${isSelected ? 'ring-2 ring-primary/50' : ''}`
          }`}
        >
          {/* Content text */}
          <p className="whitespace-pre-wrap">{content}</p>

          {/* Bot bubble: status + result summary + detail link */}
          {!isUser && (
            <div className="mt-1.5 space-y-1">
              <div className="flex items-center gap-2 flex-wrap">
                <StatusIndicator status={status} />
              </div>

              {/* Result summary: first 1-2 lines */}
              {result && (
                <p className="text-xs text-muted-foreground line-clamp-2 mt-1">
                  {getFirstLines(result)}
                </p>
              )}

              {/* Detail button */}
              {onDetailClick && (result || status === 'done' || status === 'failed') && (
                <button
                  onClick={onDetailClick}
                  className="inline-flex items-center gap-1 text-xs text-primary hover:underline cursor-pointer mt-0.5"
                >
                  <Eye className="h-3 w-3" />
                  자세히보기
                </button>
              )}
            </div>
          )}
        </div>

        {/* Timestamp */}
        <div className={`mt-0.5 ${isUser ? 'text-right' : 'text-left'}`}>
          <span className="text-[10px] text-muted-foreground">{time}</span>
        </div>
      </div>
    </div>
  )
}

function StatusIndicator({ status }: { status?: string }) {
  switch (status) {
    case 'processing':
      return (
        <Badge variant="warning" className="text-[10px] px-1.5 py-0 gap-1 min-h-0">
          <Loader2 className="h-3 w-3 animate-spin" />
          처리 중
        </Badge>
      )
    case 'done':
      return (
        <Badge variant="success" className="text-[10px] px-1.5 py-0 gap-1 min-h-0">
          <Check className="h-3 w-3" />
          완료
        </Badge>
      )
    case 'failed':
      return (
        <Badge variant="destructive" className="text-[10px] px-1.5 py-0 gap-1 min-h-0">
          <AlertCircle className="h-3 w-3" />
          실패
        </Badge>
      )
    case 'pending':
      return (
        <Badge variant="secondary" className="text-[10px] px-1.5 py-0 gap-1 min-h-0">
          대기 중
        </Badge>
      )
    default:
      return null
  }
}

function sourceLabel(source: string): string {
  switch (source) {
    case 'telegram': return 'Telegram'
    case 'cli': return 'CLI'
    case 'schedule': return 'Schedule'
    default: return source
  }
}

function getFirstLines(text: string): string {
  const lines = text
    .replace(/^#+\s/gm, '')
    .replace(/\*\*/g, '')
    .split('\n')
    .filter(l => l.trim())
  return lines.slice(0, 2).join(' ').slice(0, 150)
}
