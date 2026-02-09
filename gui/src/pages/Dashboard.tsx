import { useState, useRef, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useStatus, useProjectStats, useMessages, useSendMessage, useTaskCycle, useTaskStop, useProjects } from '@/hooks/useClaribot'
import { useNavigate } from 'react-router-dom'
import { MessageSquare, FolderOpen, RefreshCw, ListTodo, Play, Square, Send } from 'lucide-react'
import { ChatBubble } from '@/components/ChatBubble'
import type { ProjectStats, StatusResponse } from '@/types'

export default function Dashboard() {
  const { data: status } = useStatus() as { data: StatusResponse | undefined }
  const { data: statsData } = useProjectStats()
  const { data: projectsData } = useProjects()
  const { data: messagesData } = useMessages(true)
  const sendMessage = useSendMessage()
  const taskCycle = useTaskCycle()
  const taskStop = useTaskStop()
  const navigate = useNavigate()

  // Chat state
  const [input, setInput] = useState('')
  const [pendingMessages, setPendingMessages] = useState<Array<{ id: string; content: string; created_at: string }>>([])
  const chatEndRef = useRef<HTMLDivElement>(null)
  const isInitialLoad = useRef(true)

  // Create category map from projects data
  const categoryMap = new Map<string, string>()
  const projectItems = parseItems(projectsData?.data)
  projectItems.forEach((p: any) => {
    const id = p.id || p.ID
    const cat = p.category || p.Category || ''
    if (cat) categoryMap.set(id, cat)
  })

  // Cycle status
  const cycleStatus = status?.cycle_status

  // Messages
  const messageItems = parseItems(messagesData?.data)
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

  // Auto-scroll
  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: isInitialLoad.current ? 'instant' : 'smooth' })
    isInitialLoad.current = false
  }, [sortedMessages.length])

  const handleSend = () => {
    if (!input.trim()) return
    const content = input.trim()
    const tempId = `pending-${Date.now()}`
    const now = new Date().toISOString()
    setPendingMessages(prev => [...prev, { id: tempId, content, created_at: now }])
    setInput('')
    sendMessage.mutate(
      { content },
      {
        onSuccess: () => setPendingMessages(prev => prev.filter(m => m.id !== tempId)),
        onError: () => setPendingMessages(prev => prev.filter(m => m.id !== tempId)),
      }
    )
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      handleSend()
    }
  }

  // Project stats from API
  const projects: ProjectStats[] = parseItems(statsData?.data)

  return (
    <div className="flex flex-col h-full gap-4 min-h-0">
      {/* Top: Global Messages Chat */}
      <div className="flex-1 flex flex-col border rounded-lg min-h-0">
        <div className="px-4 py-3 border-b shrink-0">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <MessageSquare className="h-5 w-5" />
            Messages
            <span className="text-xs font-normal text-muted-foreground bg-muted px-2 py-0.5 rounded">Global</span>
          </h2>
        </div>

        <ScrollArea className="flex-1 px-4">
          <div className="py-4">
            {sortedMessages.map((msg: any) => {
              const content = msg.content || msg.Content || ''
              const msgStatus = msg.status || msg.Status || 'pending'
              const source = msg.source || msg.Source || 'cli'
              const result = msg.result || msg.Result || ''
              const error = msg.error || msg.Error || ''
              const createdAt = msg.created_at || msg.CreatedAt || ''
              const completedAt = msg.completed_at || msg.CompletedAt || ''

              return (
                <div key={msg.id || msg.ID}>
                  <ChatBubble
                    type="user"
                    content={content}
                    source={source}
                    time={formatTime(createdAt)}
                  />
                  <ChatBubble
                    type="bot"
                    content={
                      error
                        ? error.slice(0, 120) + (error.length > 120 ? '...' : '')
                        : result
                          ? getResponseSummary(result)
                          : statusMessage(msgStatus)
                    }
                    result={result || undefined}
                    status={msgStatus}
                    time={formatTime(completedAt || createdAt)}
                    onDetailClick={() => navigate(`/messages?id=${msg.id || msg.ID}`)}
                  />
                </div>
              )
            })}
            {sortedMessages.length === 0 && (
              <div className="flex flex-col items-center justify-center py-10 text-muted-foreground">
                <MessageSquare className="h-8 w-8 mb-2 opacity-30" />
                <p className="text-sm">메시지가 없습니다</p>
              </div>
            )}
            <div ref={chatEndRef} />
          </div>
        </ScrollArea>

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

      {/* Bottom: Projects Grid */}
      <div className="flex-1 flex flex-col min-h-0">
        <h2 className="text-lg font-semibold mb-3 flex items-center gap-2 shrink-0">
          <FolderOpen className="h-5 w-5" />
          Projects
        </h2>
        {projects.length === 0 ? (
          <p className="text-sm text-muted-foreground">No projects found</p>
        ) : (
          <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3 overflow-auto flex-1 min-h-0">
            {projects.map((p) => {
              const s = p.stats
              const leafDone = s.done
              const leafTotal = s.leaf || 1
              const progress = leafTotal > 0 ? Math.round((leafDone / leafTotal) * 100) : 0
              const isRunning = status?.cycle_statuses?.some(
                c => c.status === 'running' && c.project_id === p.project_id
              ) || (cycleStatus?.status === 'running' && cycleStatus?.project_id === p.project_id)
              const category = categoryMap.get(p.project_id)
              return (
                <Card
                  key={p.project_id}
                  className="hover:border-primary/50 transition-colors"
                >
                  <CardHeader className="pb-2">
                    <CardTitle className="text-base flex items-center gap-2">
                      {isRunning ? (
                        <RefreshCw className="h-4 w-4 text-green-500 animate-spin shrink-0" />
                      ) : (
                        <FolderOpen className="h-4 w-4 text-muted-foreground shrink-0" />
                      )}
                      <span className="truncate">{p.project_name || p.project_id}</span>
                      {category && (
                        <Badge variant="outline" className="text-[10px] shrink-0">
                          {category}
                        </Badge>
                      )}
                    </CardTitle>
                    {p.project_description && (
                      <p className="text-xs text-muted-foreground truncate">{p.project_description}</p>
                    )}
                  </CardHeader>
                  <CardContent className="space-y-3">
                    {/* Status counts */}
                    <div className="flex flex-wrap gap-2 text-xs">
                      {s.todo > 0 && <Badge variant="secondary">{s.todo} todo</Badge>}
                      {s.split > 0 && <Badge variant="secondary" className="bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">{s.split} split</Badge>}
                      {s.planned > 0 && <Badge variant="secondary" className="bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300">{s.planned} planned</Badge>}
                      {s.done > 0 && <Badge variant="secondary" className="bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300">{s.done} done</Badge>}
                      {s.failed > 0 && <Badge variant="destructive">{s.failed} failed</Badge>}
                    </div>

                    {/* Stacked status bar */}
                    {leafTotal > 0 && (
                      <div className="h-2 rounded-full bg-secondary flex overflow-hidden">
                        {s.done > 0 && <div className="bg-green-400 h-full" style={{ width: `${(s.done / leafTotal) * 100}%` }} />}
                        {s.planned > 0 && <div className="bg-yellow-400 h-full" style={{ width: `${(s.planned / leafTotal) * 100}%` }} />}
                        {s.todo > 0 && <div className="bg-gray-400 h-full" style={{ width: `${(s.todo / leafTotal) * 100}%` }} />}
                        {s.failed > 0 && <div className="bg-red-400 h-full" style={{ width: `${(s.failed / leafTotal) * 100}%` }} />}
                      </div>
                    )}

                    {/* Progress */}
                    <div className="space-y-1">
                      <div className="h-1.5 rounded-full bg-secondary overflow-hidden">
                        <div className="h-full bg-primary transition-all" style={{ width: `${progress}%` }} />
                      </div>
                      <div className="flex justify-between text-xs text-muted-foreground">
                        <span>Done/Task: {leafDone}/{leafTotal}</span>
                        <span>{progress}%</span>
                      </div>
                    </div>

                    {/* Action buttons */}
                    <div className="flex gap-2 pt-1">
                      <Button
                        variant="outline"
                        size="sm"
                        className="flex-1 h-8 text-xs"
                        onClick={() => navigate(`/projects/${p.project_id}/messages`)}
                      >
                        <MessageSquare className="h-3 w-3 mr-1" />
                        Msgs
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        className="flex-1 h-8 text-xs"
                        onClick={() => navigate(`/projects/${p.project_id}/tasks`)}
                      >
                        <ListTodo className="h-3 w-3 mr-1" />
                        Tasks
                      </Button>
                      {isRunning ? (
                        <Button
                          variant="outline"
                          size="sm"
                          className="flex-1 h-8 text-xs"
                          onClick={() => taskStop.mutate()}
                        >
                          <Square className="h-3 w-3 mr-1" />
                          Stop
                        </Button>
                      ) : (
                        <Button
                          variant="outline"
                          size="sm"
                          className="flex-1 h-8 text-xs"
                          onClick={() => taskCycle.mutate(p.project_id)}
                        >
                          <Play className="h-3 w-3 mr-1" />
                          Cycle
                        </Button>
                      )}
                    </div>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}

function parseItems(data: any): any[] {
  if (!data) return []
  if (Array.isArray(data)) return data
  if (data.items && Array.isArray(data.items)) return data.items
  return []
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
