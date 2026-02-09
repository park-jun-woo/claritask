import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  useSchedules, useAddSchedule, useDeleteSchedule, useToggleSchedule, useScheduleRuns, useProjects
} from '@/hooks/useClaribot'
import { Plus, Trash2, Clock, History, Power, PowerOff, Bot, Terminal } from 'lucide-react'

export default function Schedules() {
  const { projectId } = useParams<{ projectId?: string }>()

  const currentProject = projectId
  const isGlobal = !currentProject

  // When global: show all schedules; when project selected: filter by project
  const { data: schedulesData } = useSchedules(isGlobal, isGlobal ? undefined : currentProject)
  const { data: projectsData } = useProjects()
  const addSchedule = useAddSchedule()
  const deleteSchedule = useDeleteSchedule()
  const toggleSchedule = useToggleSchedule()

  const [showAdd, setShowAdd] = useState(false)
  const [addForm, setAddForm] = useState({ cronExpr: '', message: '', projectId: '', once: false, type: 'claude' as 'claude' | 'bash' })
  const [showRuns, setShowRuns] = useState<number | null>(null)

  const scheduleItems = parseItems(schedulesData?.data)
  const projectList = parseItems(projectsData?.data)

  const handleAdd = async () => {
    if (!addForm.cronExpr || !addForm.message) return
    await addSchedule.mutateAsync({
      cronExpr: addForm.cronExpr,
      message: addForm.message,
      projectId: addForm.projectId || undefined,
      once: addForm.once,
      type: addForm.type,
    })
    setAddForm({ cronExpr: '', message: '', projectId: '', once: false, type: 'claude' })
    setShowAdd(false)
  }

  const handleDelete = (id: number) => {
    if (confirm('Delete this schedule?')) {
      deleteSchedule.mutate(id)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h1 className="text-2xl md:text-3xl font-bold">Schedules</h1>
          {currentProject && (
            <span className="text-xs font-normal text-muted-foreground bg-muted px-2 py-0.5 rounded">
              {currentProject}
            </span>
          )}
        </div>
        <Button onClick={() => setShowAdd(!showAdd)} size="sm" className="min-h-[44px]">
          <Plus className="h-4 w-4 mr-1" /> Add Schedule
        </Button>
      </div>

      {/* Add Form */}
      {showAdd && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">New Schedule</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div>
              <label className="text-sm font-medium">Cron Expression</label>
              <Input
                placeholder="0 9 * * 1-5 (Weekdays 09:00)"
                value={addForm.cronExpr}
                onChange={e => setAddForm(f => ({ ...f, cronExpr: e.target.value }))}
              />
              <p className="text-xs text-muted-foreground mt-1">min hour day month weekday</p>
            </div>
            <div>
              <label className="text-sm font-medium">Type</label>
              <select
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={addForm.type}
                onChange={e => setAddForm(f => ({ ...f, type: e.target.value as 'claude' | 'bash' }))}
              >
                <option value="claude">Claude (AI)</option>
                <option value="bash">Bash (Command)</option>
              </select>
            </div>
            <div>
              <label className="text-sm font-medium">Message</label>
              <Textarea
                placeholder={addForm.type === 'bash' ? 'Shell command to execute' : 'Message to send to Claude'}
                value={addForm.message}
                onChange={e => setAddForm(f => ({ ...f, message: e.target.value }))}
                rows={2}
              />
            </div>
            <div>
              <label className="text-sm font-medium">Project</label>
              <select
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={addForm.projectId}
                onChange={e => setAddForm(f => ({ ...f, projectId: e.target.value }))}
              >
                <option value="">(Global)</option>
                {projectList.map((p: any) => (
                  <option key={p.id || p.ID} value={p.id || p.ID}>{p.id || p.ID}</option>
                ))}
              </select>
            </div>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={addForm.once}
                onChange={e => setAddForm(f => ({ ...f, once: e.target.checked }))}
              />
              Run once only
            </label>
          </CardContent>
          <CardFooter className="gap-2">
            <Button size="sm" className="min-h-[44px]" onClick={handleAdd} disabled={addSchedule.isPending}>Add</Button>
            <Button size="sm" variant="ghost" className="min-h-[44px]" onClick={() => setShowAdd(false)}>Cancel</Button>
          </CardFooter>
        </Card>
      )}

      {/* Schedule List */}
      <div className="space-y-3">
        {scheduleItems.map((s: any) => {
          const id = s.id || s.ID
          const cronExpr = s.cron_expr || s.CronExpr || ''
          const message = s.message || s.Message || ''
          const scheduleType = s.type || s.Type || 'claude'
          const enabled = s.enabled ?? s.Enabled ?? false
          const runOnce = s.run_once ?? s.RunOnce ?? false
          const projectId = s.project_id || s.ProjectID || null
          const lastRun = s.last_run || s.LastRun || null
          const nextRun = s.next_run || s.NextRun || null

          return (
            <Card key={id}>
              <CardContent className="p-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0 space-y-2">
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-medium">#{id}</span>
                      <Badge variant={enabled ? 'success' : 'secondary'}>
                        {enabled ? 'ON' : 'OFF'}
                      </Badge>
                      <Badge variant="outline" className="text-xs flex items-center gap-1">
                        {scheduleType === 'bash' ? <Terminal className="h-3 w-3" /> : <Bot className="h-3 w-3" />}
                        {scheduleType === 'bash' ? 'Bash' : 'Claude'}
                      </Badge>
                      {runOnce && <Badge variant="outline" className="text-xs">run_once</Badge>}
                      {projectId && <Badge variant="info" className="text-xs">{projectId}</Badge>}
                    </div>

                    <div className="flex items-center gap-2 text-sm overflow-x-auto">
                      <Clock className="h-4 w-4 text-muted-foreground shrink-0" />
                      <code className="bg-muted px-2 py-0.5 rounded text-xs whitespace-nowrap">{cronExpr}</code>
                      <span className="text-muted-foreground text-xs whitespace-nowrap">{describeCron(cronExpr)}</span>
                    </div>

                    <p className="text-sm">{message}</p>

                    <div className="flex gap-4 text-xs text-muted-foreground">
                      {lastRun && <span>Last: {formatTime(lastRun)}</span>}
                      {nextRun && <span>Next: {formatTime(nextRun)}</span>}
                    </div>
                  </div>

                  <div className="flex flex-col gap-1 shrink-0">
                    <Button
                      size="sm"
                      variant="ghost"
                      className="min-h-[44px] min-w-[44px]"
                      onClick={() => toggleSchedule.mutate({ id, enable: !enabled })}
                      title={enabled ? 'Disable' : 'Enable'}
                    >
                      {enabled ? <PowerOff className="h-4 w-4" /> : <Power className="h-4 w-4" />}
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      className="min-h-[44px] min-w-[44px]"
                      onClick={() => setShowRuns(showRuns === id ? null : id)}
                      title="Run history"
                    >
                      <History className="h-4 w-4" />
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      className="text-destructive hover:text-destructive min-h-[44px] min-w-[44px]"
                      onClick={() => handleDelete(id)}
                      title="Delete"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>

                {/* Run History */}
                {showRuns === id && <RunHistory scheduleId={id} scheduleType={scheduleType} />}
              </CardContent>
            </Card>
          )
        })}
      </div>

      {scheduleItems.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          No schedules configured. Click "Add Schedule" to create one.
        </div>
      )}
    </div>
  )
}

function RunHistory({ scheduleId, scheduleType }: { scheduleId: number; scheduleType: string }) {
  const { data: runsData } = useScheduleRuns(scheduleId)
  const runs = parseItems(runsData?.data)

  return (
    <div className="mt-4 border-t pt-3">
      <h4 className="text-sm font-medium mb-2">Run History</h4>
      {runs.length === 0 ? (
        <p className="text-xs text-muted-foreground">No runs yet</p>
      ) : (
        <div className="space-y-2">
          {runs.map((r: any) => {
            const id = r.id || r.ID
            const status = r.status || r.Status || 'running'
            const startedAt = r.started_at || r.StartedAt || ''
            const result = r.result || r.Result || ''
            const error = r.error || r.Error || ''
            return (
              <div key={id} className="text-sm border rounded p-2">
                <div className="flex items-center gap-2">
                  <span className="text-xs text-muted-foreground">#{id}</span>
                  {scheduleType === 'bash'
                    ? <Terminal className="h-3 w-3 text-muted-foreground" />
                    : <Bot className="h-3 w-3 text-muted-foreground" />
                  }
                  <Badge variant={status === 'done' ? 'success' : status === 'failed' ? 'destructive' : 'warning'}>
                    {status}
                  </Badge>
                  <span className="text-xs text-muted-foreground">{formatTime(startedAt)}</span>
                </div>
                {(result || error) && (
                  <pre className="mt-1 text-xs whitespace-pre-wrap bg-muted rounded p-2 max-h-[200px] overflow-auto">
                    {error ? `Error: ${error}` : result}
                  </pre>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

function describeCron(expr: string): string {
  const parts = expr.split(' ')
  if (parts.length < 5) return ''
  const [min, hour, day, month, weekday] = parts
  const descriptions: string[] = []
  if (weekday === '1-5') descriptions.push('Weekdays')
  else if (weekday === '*') descriptions.push('Daily')
  if (hour !== '*' && min !== '*') descriptions.push(`${hour.padStart(2, '0')}:${min.padStart(2, '0')}`)
  if (hour.startsWith('*/')) descriptions.push(`Every ${hour.slice(2)}h`)
  return descriptions.join(' ') || expr
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
