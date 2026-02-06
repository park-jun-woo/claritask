import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useStatus, useProjectStats, useSwitchProject, useMessages, useSchedules, useTaskCycle, useTaskStop, useProjects } from '@/hooks/useClaribot'
import { useNavigate } from 'react-router-dom'
import { Bot, MessageSquare, Clock, FolderOpen, RefreshCw, Pencil, ListTodo, Play, ArrowRight, Square } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { ProjectStats, StatusResponse } from '@/types'

export default function Dashboard() {
  const { data: status } = useStatus() as { data: StatusResponse | undefined }
  const { data: statsData } = useProjectStats()
  const { data: projectsData } = useProjects()
  const { data: messagesData } = useMessages()
  const { data: schedulesData } = useSchedules()
  const switchProject = useSwitchProject()
  const taskCycle = useTaskCycle()
  const taskStop = useTaskStop()
  const navigate = useNavigate()

  // Create category map from projects data
  const categoryMap = new Map<string, string>()
  const projectItems = parseItems(projectsData?.data)
  projectItems.forEach((p: any) => {
    const id = p.id || p.ID
    const cat = p.category || p.Category || ''
    if (cat) categoryMap.set(id, cat)
  })

  // Parse status
  const claudeMatch = status?.message?.match(/Claude: (\d+)\/(\d+)/)
  const claudeUsed = claudeMatch?.[1] || '0'
  const claudeMax = claudeMatch?.[2] || '3'

  // Cycle status
  const cycleStatus = status?.cycle_status

  const messageItems = parseItems(messagesData?.data)
  const scheduleItems = parseItems(schedulesData?.data)

  const msgProcessing = messageItems.filter((m: any) => (m.status || m.Status) === 'processing').length
  const msgDone = messageItems.filter((m: any) => (m.status || m.Status) === 'done').length
  const schedActive = scheduleItems.filter((s: any) => s.enabled || s.Enabled).length

  // Project stats from API
  const projects: ProjectStats[] = parseItems(statsData?.data)

  return (
    <div className="space-y-6">
      <h1 className="text-2xl md:text-3xl font-bold">Dashboard</h1>

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
            <CardTitle className="text-sm font-medium">Cycle</CardTitle>
            <RefreshCw className={`h-4 w-4 text-muted-foreground ${cycleStatus?.status === 'running' ? 'animate-spin' : ''}`} />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {cycleStatus?.status === 'running' && (
                <span className="text-green-600 dark:text-green-400">Running</span>
              )}
              {cycleStatus?.status === 'interrupted' && (
                <span className="text-yellow-600 dark:text-yellow-400">Interrupted</span>
              )}
              {(!cycleStatus || cycleStatus.status === 'idle') && (
                <span className="text-muted-foreground">Idle</span>
              )}
            </div>
            <p className="text-xs text-muted-foreground">
              {cycleStatus?.status === 'running' && (
                <>
                  {cycleStatus.type} {cycleStatus.phase && `(${cycleStatus.phase})`}
                  {cycleStatus.current_task_id ? ` Task #${cycleStatus.current_task_id}` : ''}
                  {cycleStatus.elapsed_sec != null && ` ${formatElapsed(cycleStatus.elapsed_sec)}`}
                </>
              )}
              {cycleStatus?.status === 'interrupted' && (
                <>
                  {cycleStatus.type}
                  {cycleStatus.current_task_id ? ` stopped at #${cycleStatus.current_task_id}` : ''}
                </>
              )}
              {(!cycleStatus || cycleStatus.status === 'idle') && 'No active cycle'}
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

      {/* Recent Messages */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-lg">Recent Messages</CardTitle>
          <Button variant="ghost" size="sm" onClick={() => navigate('/messages')}>
            <ArrowRight className="h-4 w-4" />
          </Button>
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

                    {/* Stacked status bar (leaf tasks only) */}
                    {leafTotal > 0 && (
                      <div className="h-2 rounded-full bg-secondary flex overflow-hidden">
                        {s.done > 0 && <div className="bg-green-400 h-full" style={{ width: `${(s.done / leafTotal) * 100}%` }} />}
                        {s.planned > 0 && <div className="bg-yellow-400 h-full" style={{ width: `${(s.planned / leafTotal) * 100}%` }} />}
                        {s.todo > 0 && <div className="bg-gray-400 h-full" style={{ width: `${(s.todo / leafTotal) * 100}%` }} />}
                        {s.failed > 0 && <div className="bg-red-400 h-full" style={{ width: `${(s.failed / leafTotal) * 100}%` }} />}
                      </div>
                    )}

                    {/* Progress bar (leaf-based) */}
                    <div className="space-y-1">
                      <div className="h-1.5 rounded-full bg-secondary overflow-hidden">
                        <div
                          className="h-full bg-primary transition-all"
                          style={{ width: `${progress}%` }}
                        />
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
                        onClick={() => navigate(`/projects/${p.project_id}/edit`)}
                      >
                        <Pencil className="h-3 w-3 mr-1" />
                        Edit
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        className="flex-1 h-8 text-xs"
                        onClick={() => {
                          switchProject.mutate(p.project_id, {
                            onSuccess: () => navigate('/tasks'),
                          })
                        }}
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
                          onClick={() => {
                            switchProject.mutate(p.project_id, {
                              onSuccess: () => taskCycle.mutate(),
                            })
                          }}
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

function formatElapsed(sec: number): string {
  if (sec < 60) return `${sec}s`
  const m = Math.floor(sec / 60)
  const s = sec % 60
  if (m < 60) return `${m}m ${s}s`
  const h = Math.floor(m / 60)
  return `${h}h ${m % 60}m`
}
