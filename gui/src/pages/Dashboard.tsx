import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useStatus, useTasks, useMessages, useSchedules } from '@/hooks/useClaribot'
import { Bot, CheckSquare, MessageSquare, Clock, Activity } from 'lucide-react'

export default function Dashboard() {
  const { data: status } = useStatus()
  const { data: tasksData } = useTasks()
  const { data: messagesData } = useMessages()
  const { data: schedulesData } = useSchedules()

  // Parse status
  const claudeMatch = status?.message?.match(/Claude: (\d+)\/(\d+)/)
  const claudeUsed = claudeMatch?.[1] || '0'
  const claudeMax = claudeMatch?.[2] || '3'

  // Parse task stats from data
  const stats = status?.data as any
  const taskItems = parseItems(tasksData?.data)
  const messageItems = parseItems(messagesData?.data)
  const scheduleItems = parseItems(schedulesData?.data)

  const taskTotal = taskItems.length
  const taskDone = taskItems.filter((t: any) => t.status === 'done' || t.Status === 'done').length
  const taskPending = taskItems.filter((t: any) => {
    const s = t.status || t.Status
    return s === 'spec_ready' || s === 'plan_ready'
  }).length

  const msgProcessing = messageItems.filter((m: any) => (m.status || m.Status) === 'processing').length
  const msgDone = messageItems.filter((m: any) => (m.status || m.Status) === 'done').length

  const schedActive = scheduleItems.filter((s: any) => s.enabled || s.Enabled).length

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Dashboard</h1>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Claude</CardTitle>
            <Bot className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{claudeUsed}/{claudeMax}</div>
            <p className="text-xs text-muted-foreground">
              {Number(claudeUsed) > 0 ? 'Running' : 'Idle'}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Tasks</CardTitle>
            <CheckSquare className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{taskTotal}</div>
            <p className="text-xs text-muted-foreground">
              {taskDone} done, {taskPending} pending
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Messages</CardTitle>
            <MessageSquare className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{msgProcessing} in progress</div>
            <p className="text-xs text-muted-foreground">
              {msgDone} completed
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Schedules</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{schedActive} active</div>
            <p className="text-xs text-muted-foreground">
              {scheduleItems.length} total
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Task Status Distribution */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Task Status</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <StatusBar label="spec_ready" count={countByStatus(taskItems, 'spec_ready')} total={taskTotal} color="bg-gray-400" />
              <StatusBar label="subdivided" count={countByStatus(taskItems, 'subdivided')} total={taskTotal} color="bg-blue-400" />
              <StatusBar label="plan_ready" count={countByStatus(taskItems, 'plan_ready')} total={taskTotal} color="bg-yellow-400" />
              <StatusBar label="done" count={countByStatus(taskItems, 'done')} total={taskTotal} color="bg-green-400" />
              <StatusBar label="failed" count={countByStatus(taskItems, 'failed')} total={taskTotal} color="bg-red-400" />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Recent Messages</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {messageItems.slice(0, 5).map((msg: any, i: number) => {
                const s = msg.status || msg.Status || 'pending'
                const content = msg.content || msg.Content || ''
                const source = msg.source || msg.Source || 'cli'
                return (
                  <div key={i} className="flex items-center gap-2 text-sm">
                    <Badge variant={s === 'done' ? 'success' : s === 'processing' ? 'warning' : 'secondary'} className="text-xs">
                      {s}
                    </Badge>
                    <span className="text-xs text-muted-foreground">[{source}]</span>
                    <span className="truncate flex-1">{content}</span>
                  </div>
                )
              })}
              {messageItems.length === 0 && (
                <p className="text-sm text-muted-foreground">No messages yet</p>
              )}
            </div>
          </CardContent>
        </Card>
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

function countByStatus(items: any[], status: string): number {
  return items.filter(t => (t.status || t.Status) === status).length
}

function StatusBar({ label, count, total, color }: { label: string; count: number; total: number; color: string }) {
  const pct = total > 0 ? (count / total) * 100 : 0
  return (
    <div className="space-y-1">
      <div className="flex justify-between text-sm">
        <span>{label}</span>
        <span className="text-muted-foreground">{count}</span>
      </div>
      <div className="h-2 rounded-full bg-secondary">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${pct}%` }} />
      </div>
    </div>
  )
}
