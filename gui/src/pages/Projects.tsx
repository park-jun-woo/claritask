import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useProjects, useSwitchProject, useProjectStats, useTaskCycle, useTaskStop, useStatus } from '@/hooks/useClaribot'
import { projectAPI } from '@/api/client'
import { cn } from '@/lib/utils'
import {
  Plus, Pencil, ListTodo, Play, Search, Pin, PinOff,
  Clock, Calendar, ArrowUpDown, Square, RefreshCw
} from 'lucide-react'
import type { StatusResponse } from '@/types'
import type { ProjectStats } from '@/types'

type SortField = 'last_accessed' | 'created_at' | 'task_count'
type SortDir = 'asc' | 'desc'

interface ProjectItem {
  id: string
  description: string
  category: string
  pinned: boolean
  last_accessed: string
  created_at: string
  task_count?: number
}

export default function Projects() {
  const { data: projects, refetch } = useProjects()
  const { data: statsData } = useProjectStats()
  const { data: status } = useStatus() as { data: StatusResponse | undefined }
  const switchProject = useSwitchProject()
  const taskCycle = useTaskCycle()
  const taskStop = useTaskStop()
  const navigate = useNavigate()

  // Check if a project is running
  const isProjectRunning = (projectId: string) => {
    return status?.cycle_statuses?.some(
      c => c.status === 'running' && c.project_id === projectId
    ) || (status?.cycle_status?.status === 'running' && status?.cycle_status?.project_id === projectId)
  }

  // Add form state
  const [showAdd, setShowAdd] = useState(false)
  const [addForm, setAddForm] = useState({ path: '', description: '', category: '' })
  const [showAddCategory, setShowAddCategory] = useState(false)
  const [newCategory, setNewCategory] = useState('')

  // Filter & Sort state
  const [search, setSearch] = useState('')
  const [sortField, setSortField] = useState<SortField>('last_accessed')
  const [sortDir, setSortDir] = useState<SortDir>('desc')
  const [categoryFilter, setCategoryFilter] = useState<string | null>(null)

  const projectStats: ProjectStats[] = parseItems(statsData?.data)
  const statsMap = new Map(projectStats.map(p => [p.project_id, p]))

  // Parse project list with new fields
  const projectList: ProjectItem[] = useMemo(() => {
    const items = parseItems(projects?.data)
    return items.map((p: any) => {
      const id = p.id || p.ID
      const stats = statsMap.get(id)
      return {
        id,
        description: p.description || p.Description || '',
        category: p.category || p.Category || '',
        pinned: p.pinned || p.Pinned || false,
        last_accessed: p.last_accessed || p.LastAccessed || '',
        created_at: p.created_at || p.CreatedAt || '',
        task_count: stats?.stats?.total || 0,
      }
    })
  }, [projects, statsMap])

  // Get unique categories
  const categories = useMemo(() => {
    const cats = new Set<string>()
    projectList.forEach(p => {
      if (p.category) cats.add(p.category)
    })
    return Array.from(cats).sort()
  }, [projectList])

  // Filter and sort
  const filteredProjects = useMemo(() => {
    let items = [...projectList]

    // Search filter
    if (search.trim()) {
      const q = search.toLowerCase()
      items = items.filter(p =>
        p.id.toLowerCase().includes(q) ||
        p.description.toLowerCase().includes(q) ||
        p.category.toLowerCase().includes(q)
      )
    }

    // Category filter
    if (categoryFilter) {
      items = items.filter(p => p.category === categoryFilter)
    }

    // Sort - pinned always first
    items.sort((a, b) => {
      if (a.pinned !== b.pinned) return a.pinned ? -1 : 1

      let cmp = 0
      if (sortField === 'last_accessed') {
        cmp = (a.last_accessed || '').localeCompare(b.last_accessed || '')
      } else if (sortField === 'created_at') {
        cmp = (a.created_at || '').localeCompare(b.created_at || '')
      } else if (sortField === 'task_count') {
        cmp = (a.task_count || 0) - (b.task_count || 0)
      }
      return sortDir === 'desc' ? -cmp : cmp
    })

    return items
  }, [projectList, search, categoryFilter, sortField, sortDir])

  const handleAdd = async () => {
    if (!addForm.path) {
      alert('프로젝트 ID 또는 경로를 입력하세요')
      return
    }
    try {
      const input = addForm.path.trim()
      const isPath = input.includes('/') || input.includes('\\')

      let result
      let projectId: string

      if (isPath) {
        // Full path provided - use add (existing folder)
        result = await projectAPI.add(input, addForm.description || undefined)
        projectId = input.split('/').pop() || input.split('\\').pop() || ''
      } else {
        // Just ID provided - use create (new folder in DefaultPath)
        result = await projectAPI.create(input, addForm.description || undefined)
        projectId = input
      }

      if (!result?.success) {
        alert(result?.message || '프로젝트 추가 실패')
        return
      }

      // Set category if specified
      if (addForm.category && projectId) {
        await projectAPI.set(projectId, 'category', addForm.category)
      }

      setAddForm({ path: '', description: '', category: '' })
      setShowAdd(false)
      setShowAddCategory(false)
      setNewCategory('')
      refetch()
    } catch (err: any) {
      alert(err?.message || '프로젝트 추가 중 오류 발생')
    }
  }

  const handleAddNewCategory = () => {
    if (newCategory.trim() && !categories.includes(newCategory.trim())) {
      setAddForm(f => ({ ...f, category: newCategory.trim() }))
      setShowAddCategory(false)
      setNewCategory('')
    }
  }

  const handleTogglePin = async (e: React.MouseEvent, id: string, currentPinned: boolean) => {
    e.stopPropagation()
    await projectAPI.set(id, 'pinned', currentPinned ? '0' : '1')
    refetch()
  }

  const cycleSortField = () => {
    const fields: SortField[] = ['last_accessed', 'created_at', 'task_count']
    const idx = fields.indexOf(sortField)
    setSortField(fields[(idx + 1) % fields.length])
  }

  const sortLabel = {
    last_accessed: '최근사용',
    created_at: '생성일',
    task_count: 'Task수',
  }

  const SortIcon = {
    last_accessed: Clock,
    created_at: Calendar,
    task_count: ListTodo,
  }[sortField]

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-2xl md:text-3xl font-bold">Projects</h1>
        <Button onClick={() => setShowAdd(!showAdd)} size="sm" className="min-h-[44px]">
          <Plus className="h-4 w-4 mr-1" /> Add Project
        </Button>
      </div>

      {/* Search & Filter Bar */}
      <div className="flex flex-wrap items-center gap-2">
        {/* Search */}
        <div className="relative flex-1 min-w-[200px] max-w-[300px]">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="검색..."
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="pl-8 h-9"
          />
        </div>

        {/* Sort */}
        <Button
          variant="outline"
          size="sm"
          className="h-9 gap-1.5"
          onClick={cycleSortField}
          title="정렬 기준 변경"
        >
          <SortIcon className="h-4 w-4" />
          <span className="hidden sm:inline">{sortLabel[sortField]}</span>
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="h-9 w-9 p-0"
          onClick={() => setSortDir(d => d === 'asc' ? 'desc' : 'asc')}
          title={sortDir === 'desc' ? '내림차순' : '오름차순'}
        >
          <ArrowUpDown className={cn("h-4 w-4", sortDir === 'asc' && "rotate-180")} />
        </Button>

        <div className="w-px h-6 bg-border mx-1 hidden sm:block" />

        {/* Category Filter */}
        <div className="flex flex-wrap items-center gap-1">
          <Button
            variant={categoryFilter === null ? "secondary" : "ghost"}
            size="sm"
            className="h-8 text-xs"
            onClick={() => setCategoryFilter(null)}
          >
            전체
          </Button>
          {categories.map(cat => (
            <Button
              key={cat}
              variant={categoryFilter === cat ? "secondary" : "ghost"}
              size="sm"
              className="h-8 text-xs"
              onClick={() => setCategoryFilter(cat)}
            >
              {cat}
            </Button>
          ))}
        </div>
      </div>

      {/* Add Form */}
      {showAdd && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Add Project</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <Input
              placeholder="Project ID or path (e.g., my-project or /home/user/my-project)"
              value={addForm.path}
              onChange={e => setAddForm(f => ({ ...f, path: e.target.value }))}
            />
            <Textarea
              placeholder="Description"
              value={addForm.description}
              onChange={e => setAddForm(f => ({ ...f, description: e.target.value }))}
              rows={2}
            />
            {/* Category Selection */}
            <div className="space-y-2">
              <label className="text-sm font-medium">Category</label>
              <div className="flex flex-wrap gap-2">
                <Button
                  type="button"
                  variant={addForm.category === '' ? 'default' : 'outline'}
                  size="sm"
                  className="h-8"
                  onClick={() => setAddForm(f => ({ ...f, category: '' }))}
                >
                  없음
                </Button>
                {/* Show selected category if not in list */}
                {addForm.category && !categories.includes(addForm.category) && (
                  <Button
                    type="button"
                    variant="default"
                    size="sm"
                    className="h-8"
                  >
                    {addForm.category}
                  </Button>
                )}
                {categories.map(cat => (
                  <Button
                    key={cat}
                    type="button"
                    variant={addForm.category === cat ? 'default' : 'outline'}
                    size="sm"
                    className="h-8"
                    onClick={() => setAddForm(f => ({ ...f, category: cat }))}
                  >
                    {cat}
                  </Button>
                ))}
                {!showAddCategory ? (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="h-8 w-8 p-0"
                    onClick={() => setShowAddCategory(true)}
                    title="카테고리 추가"
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                ) : (
                  <div className="flex gap-1">
                    <Input
                      placeholder="새 카테고리..."
                      value={newCategory}
                      onChange={e => setNewCategory(e.target.value)}
                      className="h-8 w-28 text-sm"
                      onKeyDown={e => {
                        if (e.key === 'Enter') {
                          e.preventDefault()
                          handleAddNewCategory()
                        }
                        if (e.key === 'Escape') {
                          setShowAddCategory(false)
                          setNewCategory('')
                        }
                      }}
                      autoFocus
                    />
                    <Button
                      type="button"
                      size="sm"
                      className="h-8"
                      onClick={handleAddNewCategory}
                    >
                      추가
                    </Button>
                  </div>
                )}
              </div>
            </div>
          </CardContent>
          <CardFooter className="gap-2">
            <Button size="sm" className="min-h-[44px]" onClick={handleAdd}>Add</Button>
            <Button size="sm" variant="ghost" className="min-h-[44px]" onClick={() => setShowAdd(false)}>Cancel</Button>
          </CardFooter>
        </Card>
      )}

      {/* Project List */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {filteredProjects.map((p) => {
          const stats = statsMap.get(p.id)
          const s = stats?.stats || { total: 0, leaf: 0, todo: 0, planned: 0, done: 0, failed: 0 }
          const leafTotal = s.leaf || 1
          const leafDone = s.done
          const progress = leafTotal > 0 ? Math.round((leafDone / leafTotal) * 100) : 0

          return (
            <Card
              key={p.id}
              className={cn(
                "hover:border-primary/50 transition-colors relative group",
                p.pinned && "border-primary/30"
              )}
            >
              {/* Pin Button */}
              <button
                className="absolute top-2 right-2 p-1.5 rounded-md hover:bg-accent z-10"
                onClick={e => handleTogglePin(e, p.id, p.pinned)}
                title={p.pinned ? '고정 해제' : '고정'}
              >
                {p.pinned ? (
                  <Pin className="h-4 w-4 text-primary" />
                ) : (
                  <PinOff className="h-4 w-4 opacity-0 group-hover:opacity-50" />
                )}
              </button>

              <CardHeader className="pb-2 pr-10">
                <CardTitle className="text-base flex items-center gap-2">
                  {isProjectRunning(p.id) && (
                    <RefreshCw className="h-4 w-4 text-green-500 animate-spin shrink-0" />
                  )}
                  <span className="truncate">{p.id}</span>
                  {p.category && (
                    <Badge variant="outline" className="text-[10px] shrink-0">
                      {p.category}
                    </Badge>
                  )}
                </CardTitle>
                {p.description && (
                  <p className="text-xs text-muted-foreground truncate">{p.description}</p>
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
                    onClick={() => navigate(`/projects/${p.id}/edit`)}
                  >
                    <Pencil className="h-3 w-3 mr-1" />
                    Edit
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="flex-1 h-8 text-xs"
                    onClick={() => {
                      switchProject.mutate(p.id, {
                        onSuccess: () => navigate('/tasks'),
                      })
                    }}
                  >
                    <ListTodo className="h-3 w-3 mr-1" />
                    Tasks
                  </Button>
                  {isProjectRunning(p.id) ? (
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
                        switchProject.mutate(p.id, {
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

      {filteredProjects.length === 0 && projectList.length > 0 && (
        <div className="text-center py-12 text-muted-foreground">
          검색 결과가 없습니다.
        </div>
      )}

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
