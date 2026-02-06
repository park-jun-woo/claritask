import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useProjects, useSwitchProject, useProjectStats, useTaskCycle } from '@/hooks/useClaribot'
import { projectAPI } from '@/api/client'
import { Plus, Pencil, ListTodo, Play } from 'lucide-react'
import type { ProjectStats } from '@/types'

export default function Projects() {
  const { data: projects, refetch } = useProjects()
  const { data: statsData } = useProjectStats()
  const switchProject = useSwitchProject()
  const taskCycle = useTaskCycle()
  const navigate = useNavigate()
  const [showAdd, setShowAdd] = useState(false)
  const [addForm, setAddForm] = useState({ path: '', description: '' })

  const projectList = parseItems(projects?.data)
  const projectStats: ProjectStats[] = parseItems(statsData?.data)

  // Create a map of project stats by project_id
  const statsMap = new Map(projectStats.map(p => [p.project_id, p]))

  const handleAdd = async () => {
    if (!addForm.path) return
    await projectAPI.add(addForm.path, addForm.description || undefined)
    setAddForm({ path: '', description: '' })
    setShowAdd(false)
    refetch()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl md:text-3xl font-bold">Projects</h1>
        <Button onClick={() => setShowAdd(!showAdd)} size="sm" className="min-h-[44px]">
          <Plus className="h-4 w-4 mr-1" /> Add Project
        </Button>
      </div>

      {/* Add Form */}
      {showAdd && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Add Project</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Input
              placeholder="Project path (e.g., /home/user/my-project)"
              value={addForm.path}
              onChange={e => setAddForm(f => ({ ...f, path: e.target.value }))}
            />
            <Textarea
              placeholder="Description"
              value={addForm.description}
              onChange={e => setAddForm(f => ({ ...f, description: e.target.value }))}
              rows={2}
            />
          </CardContent>
          <CardFooter className="gap-2">
            <Button size="sm" className="min-h-[44px]" onClick={handleAdd}>Add</Button>
            <Button size="sm" variant="ghost" className="min-h-[44px]" onClick={() => setShowAdd(false)}>Cancel</Button>
          </CardFooter>
        </Card>
      )}

      {/* Project List */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {projectList.map((p: any) => {
          const id = p.id || p.ID
          const desc = p.description || p.Description || ''
          const stats = statsMap.get(id)
          const s = stats?.stats || { total: 0, leaf: 0, todo: 0, planned: 0, done: 0, failed: 0 }
          const leafTotal = s.leaf || 1
          const leafDone = s.done
          const progress = leafTotal > 0 ? Math.round((leafDone / leafTotal) * 100) : 0

          return (
            <Card
              key={id}
              className="cursor-pointer hover:border-primary/50 transition-colors"
              onClick={() => {
                switchProject.mutate(id, {
                  onSuccess: () => navigate('/tasks'),
                })
              }}
            >
              <CardHeader className="pb-2">
                <CardTitle className="text-base flex items-center justify-between">
                  <span className="truncate">{id}</span>
                  <Badge variant="outline" className="ml-2 shrink-0 text-xs">
                    {s.total} tasks
                  </Badge>
                </CardTitle>
                {desc && (
                  <p className="text-xs text-muted-foreground truncate">{desc}</p>
                )}
              </CardHeader>
              <CardContent className="space-y-3">
                {/* Status counts */}
                <div className="flex flex-wrap gap-2 text-xs">
                  {s.todo > 0 && <Badge variant="secondary">{s.todo} todo</Badge>}
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
                    onClick={(e) => {
                      e.stopPropagation()
                      switchProject.mutate(id, {
                        onSuccess: () => navigate('/tasks'),
                      })
                    }}
                  >
                    <ListTodo className="h-3 w-3 mr-1" />
                    Tasks
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="flex-1 h-8 text-xs"
                    onClick={(e) => {
                      e.stopPropagation()
                      switchProject.mutate(id, {
                        onSuccess: () => taskCycle.mutate(),
                      })
                    }}
                  >
                    <Play className="h-3 w-3 mr-1" />
                    Cycle
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="flex-1 h-8 text-xs"
                    onClick={(e) => {
                      e.stopPropagation()
                      navigate(`/projects/${id}/edit`)
                    }}
                  >
                    <Pencil className="h-3 w-3 mr-1" />
                    Edit
                  </Button>
                </div>
              </CardContent>
            </Card>
          )
        })}
      </div>

      {projectList.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          No projects registered yet. Click "Add Project" to get started.
        </div>
      )}
    </div>
  )
}

function parseItems(data: any): any[] {
  if (!data) return []
  if (Array.isArray(data)) return data
  if (data.items && Array.isArray(data.items)) return data.items
  return []
}
