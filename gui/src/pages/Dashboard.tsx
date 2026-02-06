import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useStatus, useProjectStats, useSwitchProject, useMessages, useSchedules } from '@/hooks/useClaribot'
import { useNavigate } from 'react-router-dom'
import { Bot, MessageSquare, Clock, FolderOpen } from 'lucide-react'
import type { ProjectStats } from '@/types'

export default function Dashboard() {
  const { data: status } = useStatus()
  const { data: statsData } = useProjectStats()
  const { data: messagesData } = useMessages()
  const { data: schedulesData } = useSchedules()
  const switchProject = useSwitchProject()
  const navigate = useNavigate()

  // Parse status
  const claudeMatch = status?.message?.match(/Claude: (\d+)\/(\d+)/)
  const claudeUsed = claudeMatch?.[1] || '0'
  const claudeMax = claudeMatch?.[2] || '3'

  const messageItems = parseItems(messagesData?.data)
  const scheduleItems = parseItems(schedulesData?.data)

  const msgProcessing = messageItems.filter((m: any) => (m.status || m.Status) === 'processing').length
  const msgDone = messageItems.filter((m: any) => (m.status || m.Status) === 'done').length
  const schedActive = scheduleItems.filter((s: any) => s.enabled || s.Enabled).length

  // Project stats from API
  const projects: ProjectStats[] = parseItems(statsData?.data)

  const handleProjectClick = (projectId: string) => {
    switchProject.mutate(projectId, {
      onSuccess: () => navigate('/tasks'),
    })
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl md:text-3xl font-bold">Dashboard</h1>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
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

      {/* Project Stats Board */}
      <div>
        <h2 className="text-lg font-semibold mb-3 flex items-center gap-2">
          <FolderOpen className="h-5 w-5" />
          Projects
        </h2>
        {projects.length === 0 ? (
          <p className="text-sm text-muted-foreground">No projects found</p>
        ) : (
          <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
            {projects.map((p) => {
              const s = p.stats
              const leafDone = s.done
              const leafTotal = s.leaf || 1
              const progress = leafTotal > 0 ? Math.round((leafDone / leafTotal) * 100) : 0
              return (
                <Card
                  key={p.project_id}
                  className="cursor-pointer hover:border-primary/50 transition-colors"
                  onClick={() => handleProjectClick(p.project_id)}
                >
                  <CardHeader className="pb-2">
                    <CardTitle className="text-base flex items-center justify-between">
                      <span className="truncate">{p.project_name || p.project_id}</span>
                      <Badge variant="outline" className="ml-2 shrink-0 text-xs">
                        {s.total} tasks
                      </Badge>
                    </CardTitle>
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
                    {s.total > 0 && (
                      <div className="h-2 rounded-full bg-secondary flex overflow-hidden">
                        {s.done > 0 && <div className="bg-green-400 h-full" style={{ width: `${(s.done / s.total) * 100}%` }} />}
                        {s.planned > 0 && <div className="bg-yellow-400 h-full" style={{ width: `${(s.planned / s.total) * 100}%` }} />}
                        {s.todo > 0 && <div className="bg-gray-400 h-full" style={{ width: `${(s.todo / s.total) * 100}%` }} />}
                        {s.split > 0 && <div className="bg-blue-400 h-full" style={{ width: `${(s.split / s.total) * 100}%` }} />}
                        {s.failed > 0 && <div className="bg-red-400 h-full" style={{ width: `${(s.failed / s.total) * 100}%` }} />}
                      </div>
                    )}

                    {/* Progress */}
                    <div className="flex justify-between text-xs text-muted-foreground">
                      <span>done/leaf: {leafDone}/{leafTotal}</span>
                      <span>{progress}%</span>
                    </div>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        )}
      </div>

      {/* Recent Messages */}
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
  )
}

function parseItems(data: any): any[] {
  if (!data) return []
  if (Array.isArray(data)) return data
  if (data.items && Array.isArray(data.items)) return data.items
  return []
}
