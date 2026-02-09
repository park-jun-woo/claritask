import { useState, useMemo, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import {
  useTasks, useTask, useAddTask, useDeleteTask, useTaskCycle, useTaskStop, useSetTask, useStatus
} from '@/hooks/useClaribot'
import type { StatusResponse } from '@/types'
import {
  Plus, Play, RefreshCw, ChevronRight, ChevronDown, X, TreePine, List, FileText, Square
} from 'lucide-react'
import { MarkdownRenderer } from '@/components/MarkdownRenderer'

function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() =>
    typeof window !== 'undefined' ? window.matchMedia(query).matches : false
  )
  useEffect(() => {
    const mql = window.matchMedia(query)
    const handler = (e: MediaQueryListEvent) => setMatches(e.matches)
    mql.addEventListener('change', handler)
    return () => mql.removeEventListener('change', handler)
  }, [query])
  return matches
}

type ViewMode = 'tree' | 'list'

export default function Tasks() {
  const { projectId, taskId: taskIdParam } = useParams<{ projectId?: string; taskId?: string }>()
  const navigate = useNavigate()

  const currentProject = projectId || 'GLOBAL'
  const basePath = projectId ? `/projects/${projectId}/tasks` : '/tasks'

  const { data: tasksData } = useTasks()
  const addTask = useAddTask()
  const deleteTask = useDeleteTask()
  const taskCycle = useTaskCycle()
  const taskStop = useTaskStop()
  const setTask = useSetTask()

  const [viewMode, setViewMode] = useState<ViewMode>('list')
  const [showAdd, setShowAdd] = useState(false)
  const [addForm, setAddForm] = useState({ spec: '', parentId: '' })

  // URL-based task selection
  const selectedTaskId = taskIdParam ? Number(taskIdParam) : null
  const setSelectedTaskId = (id: number | null) => {
    if (id !== null) {
      navigate(`${basePath}/${id}`)
    } else {
      navigate(basePath)
    }
  }

  const { data: taskDetail } = useTask(selectedTaskId ?? undefined)
  const selectedTask = taskDetail?.data ?? null
  const [expandedNodes, setExpandedNodes] = useState<Set<number>>(new Set())
  const [editField, setEditField] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const [statusFilter, setStatusFilter] = useState<string | null>(null)

  // Check if current project is running
  const { data: statusData } = useStatus() as { data: StatusResponse | undefined }
  const isProjectRunning = statusData?.cycle_statuses?.some(
    c => c.status === 'running' && c.project_id === currentProject
  ) || (statusData?.cycle_status?.status === 'running' && statusData?.cycle_status?.project_id === currentProject)

  const taskItems = useMemo(() => parseItems(tasksData?.data), [tasksData])

  // Filtered items based on status
  const filteredItems = useMemo(() => {
    if (!statusFilter) return taskItems
    return taskItems.filter((t: any) => (t.status || t.Status || 'todo') === statusFilter)
  }, [taskItems, statusFilter])

  // Build tree structure (use all items for tree, filter shows flat list)
  const treeData = useMemo(() => buildTree(taskItems), [taskItems])
  const filteredTreeData = useMemo(() => statusFilter ? buildTree(filteredItems) : treeData, [statusFilter, filteredItems, treeData])

  const handleAdd = async () => {
    if (!addForm.spec) return
    await addTask.mutateAsync({
      spec: addForm.spec,
      parentId: addForm.parentId ? Number(addForm.parentId) : undefined,
    })
    setAddForm({ spec: '', parentId: '' })
    setShowAdd(false)
  }

  const toggleNode = (id: number) => {
    setExpandedNodes(prev => {
      const next = new Set(prev)
      next.has(id) ? next.delete(id) : next.add(id)
      return next
    })
  }

  const handleEdit = async (taskId: number, field: string, value: string) => {
    await setTask.mutateAsync({ id: taskId, field, value })
    setEditField(null)
  }

  const handleSelect = (task: any) => {
    const id = task.id || task.ID
    setSelectedTaskId(id)
  }

  const isDesktop = useMediaQuery('(min-width: 768px)')

  const [activeTab, setActiveTab] = useState('spec')

  const detailContent = selectedTask ? (
    <div className="flex flex-col flex-1 min-h-0">
      <ScrollArea className="flex-1 min-h-0 p-4">
        <div className="space-y-4">
          {/* Title & Status */}
          <div>
            <h4 className="text-lg font-medium">{selectedTask.title || selectedTask.Title}</h4>
            <div className="flex items-center gap-2 mt-1">
              <StatusDot status={selectedTask.status || selectedTask.Status} />
              <span className="text-xs text-muted-foreground">
                depth: {selectedTask.depth ?? selectedTask.Depth ?? 0}
              </span>
              {(selectedTask.is_leaf || selectedTask.IsLeaf) && (
                <Badge variant="outline" className="text-xs">leaf</Badge>
              )}
            </div>
          </div>

          {/* Actions */}
          <div className="flex gap-2 flex-wrap">
            <Button
              size="sm"
              variant="destructive"
              onClick={() => {
                if (confirm('Delete this task?')) {
                  deleteTask.mutate(selectedTask.id || selectedTask.ID)
                  setSelectedTaskId(null)
                }
              }}
            >
              Delete
            </Button>
          </div>

          {/* Tabs: Spec / Plan / Report */}
          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="spec">Spec</TabsTrigger>
              <TabsTrigger value="plan">Plan</TabsTrigger>
              <TabsTrigger value="report">Report</TabsTrigger>
            </TabsList>

            <TabsContent value="spec">
              <TabSection
                content={selectedTask.spec || selectedTask.Spec || ''}
                isEditing={editField === 'spec'}
                onEdit={() => { setEditField('spec'); setEditValue(selectedTask.spec || selectedTask.Spec || '') }}
                onSave={() => handleEdit(selectedTask.id || selectedTask.ID, 'spec', editValue)}
                onCancel={() => setEditField(null)}
                editValue={editValue}
                onEditChange={setEditValue}
              />
            </TabsContent>

            <TabsContent value="plan">
              <TabSection
                content={selectedTask.plan || selectedTask.Plan || ''}
                isEditing={editField === 'plan'}
                onEdit={() => { setEditField('plan'); setEditValue(selectedTask.plan || selectedTask.Plan || '') }}
                onSave={() => handleEdit(selectedTask.id || selectedTask.ID, 'plan', editValue)}
                onCancel={() => setEditField(null)}
                editValue={editValue}
                onEditChange={setEditValue}
              />
            </TabsContent>

            <TabsContent value="report">
              <TabSection
                content={selectedTask.report || selectedTask.Report || ''}
                isEditing={editField === 'report'}
                onEdit={() => { setEditField('report'); setEditValue(selectedTask.report || selectedTask.Report || '') }}
                onSave={() => handleEdit(selectedTask.id || selectedTask.ID, 'report', editValue)}
                onCancel={() => setEditField(null)}
                editValue={editValue}
                onEditChange={setEditValue}
              />
            </TabsContent>
          </Tabs>
        </div>
      </ScrollArea>
    </div>
  ) : null

  return (
    <div className="flex flex-col md:flex-row gap-4 flex-1 min-h-0 h-full overflow-hidden">
      {/* List Panel */}
      <div className={cn(
        "flex flex-col space-y-2 min-w-0",
        isDesktop ? "w-1/2" : "w-full flex-1"
      )}>
        {/* Toolbar */}
        <div className="flex flex-wrap items-center justify-between gap-1">
          <h1 className="text-xl font-bold shrink-0">Tasks</h1>
          <div className="flex flex-wrap items-center gap-1">
            <div className="flex border rounded-md">
              <Button
                variant={viewMode === 'tree' ? 'secondary' : 'ghost'}
                size="icon"
                onClick={() => setViewMode('tree')}
                className="rounded-r-none h-8 w-8"
              >
                <TreePine className="h-4 w-4" />
              </Button>
              <Button
                variant={viewMode === 'list' ? 'secondary' : 'ghost'}
                size="icon"
                onClick={() => setViewMode('list')}
                className="rounded-l-none h-8 w-8"
              >
                <List className="h-4 w-4" />
              </Button>
            </div>
            <Button size="icon" className="h-8 w-8" onClick={() => setShowAdd(!showAdd)}>
              <Plus className="h-4 w-4" />
            </Button>
            {/* Cycle/Stop Toggle */}
            {isProjectRunning ? (
              <Button size="icon" variant="outline" className="h-8 w-8" onClick={() => taskStop.mutate()} disabled={taskStop.isPending}>
                <RefreshCw className="h-4 w-4 animate-spin" />
              </Button>
            ) : (
              <Button size="icon" variant="outline" className="h-8 w-8" onClick={() => taskCycle.mutate(currentProject !== 'GLOBAL' ? currentProject : undefined)} disabled={taskCycle.isPending || currentProject === 'GLOBAL'}>
                <Play className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>

        {/* Add Form */}
        {showAdd && (
          <Card>
            <CardContent className="p-3 space-y-2">
              <Textarea
                placeholder="Task spec (first line becomes title)"
                value={addForm.spec}
                onChange={e => setAddForm(f => ({ ...f, spec: e.target.value }))}
                rows={3}
                className="text-sm"
              />
              <Input
                placeholder="Parent ID (optional)"
                value={addForm.parentId}
                onChange={e => setAddForm(f => ({ ...f, parentId: e.target.value }))}
                className="h-8 text-sm"
              />
              <div className="flex gap-2">
                <Button size="sm" className="h-7 text-xs" onClick={handleAdd} disabled={addTask.isPending || !addForm.spec.trim()}>Add</Button>
                <Button size="sm" variant="ghost" className="h-7 text-xs" onClick={() => setShowAdd(false)}>Cancel</Button>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Action Status */}
        {(taskCycle.isPending || taskStop.isPending) && (
          <div className="flex items-center gap-2 px-2 py-1 bg-yellow-50 dark:bg-yellow-950 rounded-md border border-yellow-200 dark:border-yellow-800">
            <RefreshCw className="h-3 w-3 animate-spin" />
            <span className="text-xs">
              {taskCycle.isPending && 'Starting cycle...'}
              {taskStop.isPending && 'Stopping...'}
            </span>
          </div>
        )}

        {/* Task Status Bar */}
        {taskItems.length > 0 && (
          <TaskStatusBar
            items={taskItems}
            statusFilter={statusFilter}
            onFilterChange={setStatusFilter}
            currentProject={currentProject}
          />
        )}

        {/* Task View */}
        <ScrollArea className="flex-1 border rounded-md">
          <div className="p-2">
            {viewMode === 'tree' ? (
              <TreeView
                nodes={filteredTreeData}
                expandedNodes={expandedNodes}
                onToggle={toggleNode}
                onSelect={handleSelect}
                selectedId={selectedTaskId ?? undefined}
              />
            ) : (
              <ListView
                items={filteredItems}
                onSelect={handleSelect}
                selectedId={selectedTaskId ?? undefined}
                isMobile={!isDesktop}
              />
            )}
            {filteredItems.length === 0 && taskItems.length > 0 && (
              <p className="text-center text-muted-foreground py-8 text-sm">
                No tasks match the filter.
              </p>
            )}
            {taskItems.length === 0 && (
              <p className="text-center text-muted-foreground py-8 text-sm">
                No tasks yet.
              </p>
            )}
          </div>
        </ScrollArea>
      </div>

      {/* Desktop: Detail Panel (3/4) */}
      {isDesktop && (
        <div className="w-1/2 border rounded-md flex flex-col min-w-0 min-h-0 overflow-hidden">
          {selectedTask ? (
            <>
              <div className="flex items-center justify-between p-4 border-b">
                <h3 className="font-semibold">Task #{selectedTask.id || selectedTask.ID}</h3>
                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setSelectedTaskId(null)}>
                  <X className="h-4 w-4" />
                </Button>
              </div>
              {detailContent}
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center text-muted-foreground">
              <div className="text-center">
                <FileText className="h-12 w-12 mx-auto mb-3 opacity-30" />
                <p className="text-sm">Select a task to view details</p>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Mobile: Full-screen overlay */}
      {!isDesktop && selectedTask && (
        <div className="fixed inset-0 z-50 bg-background flex flex-col">
          <div className="flex items-center justify-between p-4 border-b">
            <h3 className="font-semibold">Task #{selectedTask.id || selectedTask.ID}</h3>
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setSelectedTaskId(null)}>
              <X className="h-4 w-4" />
            </Button>
          </div>
          {detailContent}
        </div>
      )}
    </div>
  )
}

// --- Sub-components ---

function TabSection({
  content, isEditing, onEdit, onSave, onCancel, editValue, onEditChange
}: {
  content: string
  isEditing: boolean
  onEdit: () => void
  onSave: () => void
  onCancel: () => void
  editValue: string
  onEditChange: (v: string) => void
}) {
  return (
    <div>
      {isEditing ? (
        <div className="space-y-2">
          <Textarea value={editValue} onChange={e => onEditChange(e.target.value)} rows={8} />
          <div className="flex gap-2">
            <Button size="sm" onClick={onSave}>Save</Button>
            <Button size="sm" variant="ghost" onClick={onCancel}>Cancel</Button>
          </div>
        </div>
      ) : (
        <div>
          <div className="flex justify-end mb-2">
            <Button variant="ghost" size="sm" className="text-xs h-8" onClick={onEdit}>
              <FileText className="h-3 w-3 mr-1" /> Edit
            </Button>
          </div>
          <div className="bg-muted rounded p-3">
            {content ? (
              <MarkdownRenderer content={content} />
            ) : (
              <p className="text-sm text-muted-foreground">No content</p>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

interface TreeNode {
  task: any
  children: TreeNode[]
}

function buildTree(items: any[]): TreeNode[] {
  const map = new Map<number, TreeNode>()
  const roots: TreeNode[] = []

  items.forEach((t: any) => {
    const id = t.id || t.ID
    map.set(id, { task: t, children: [] })
  })

  items.forEach((t: any) => {
    const id = t.id || t.ID
    const parentId = t.parent_id ?? t.ParentID
    const node = map.get(id)!
    if (parentId && map.has(parentId)) {
      map.get(parentId)!.children.push(node)
    } else {
      roots.push(node)
    }
  })

  // Sort roots and children by id DESC (newest first)
  const sortDesc = (a: TreeNode, b: TreeNode) =>
    (b.task.id || b.task.ID) - (a.task.id || a.task.ID)
  roots.sort(sortDesc)
  map.forEach(node => node.children.sort(sortDesc))

  return roots
}

function TreeView({
  nodes, expandedNodes, onToggle, onSelect, selectedId, depth = 0
}: {
  nodes: TreeNode[]
  expandedNodes: Set<number>
  onToggle: (id: number) => void
  onSelect: (task: any) => void
  selectedId?: number
  depth?: number
}) {
  return (
    <div className="space-y-0.5">
      {nodes.map(node => {
        const id = node.task.id || node.task.ID
        const title = node.task.title || node.task.Title || (node.task.spec || node.task.Spec || '').split('\n')[0] || '(untitled)'
        const status = node.task.status || node.task.Status || 'todo'
        const hasChildren = node.children.length > 0
        const isExpanded = expandedNodes.has(id)
        const isSelected = id === selectedId

        return (
          <div key={id}>
            <div
              className={`flex items-center gap-1 py-2.5 px-2 rounded cursor-pointer text-sm hover:bg-accent ${isSelected ? 'bg-accent' : ''}`}
              style={{ paddingLeft: `${depth * 12 + 8}px` }}
              onClick={() => onSelect(node.task)}
            >
              {hasChildren ? (
                <button onClick={e => { e.stopPropagation(); onToggle(id) }} className="p-2">
                  {isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                </button>
              ) : (
                <span className="w-4" />
              )}
              <StatusDot status={status} />
              <span className="text-muted-foreground text-xs mr-1">#{id}</span>
              <span className="truncate">{title}</span>
            </div>
            {hasChildren && isExpanded && (
              <TreeView
                nodes={node.children}
                expandedNodes={expandedNodes}
                onToggle={onToggle}
                onSelect={onSelect}
                selectedId={selectedId}
                depth={depth + 1}
              />
            )}
          </div>
        )
      })}
    </div>
  )
}

function TaskStatusBar({ items, statusFilter, onFilterChange, currentProject }: {
  items: any[]
  statusFilter: string | null
  onFilterChange: (status: string | null) => void
  currentProject: string
}) {
  const { data: status } = useStatus() as { data: StatusResponse | undefined }
  // Only show cycle status if it matches current project
  const cycleStatus = status?.cycle_status?.project_id === currentProject ? status.cycle_status : undefined

  const statuses = ['todo', 'split', 'planned', 'done', 'failed'] as const
  const colors: Record<string, string> = {
    todo: 'bg-gray-400',
    split: 'bg-blue-400',
    planned: 'bg-yellow-400',
    done: 'bg-green-400',
    failed: 'bg-red-400',
  }
  const counts = useMemo(() => {
    const map: Record<string, number> = {}
    for (const s of statuses) map[s] = 0
    for (const t of items) {
      const s = t.status || t.Status || 'todo'
      map[s] = (map[s] || 0) + 1
    }
    return map
  }, [items])

  // leaf count & done ratio
  const leafItems = items.filter((t: any) => t.is_leaf || t.IsLeaf)
  const leafDone = leafItems.filter((t: any) => (t.status || t.Status) === 'done').length
  const leafTotal = leafItems.length
  const progress = leafTotal > 0 ? Math.round((leafDone / leafTotal) * 100) : 0

  return (
    <div className="space-y-1">
      {/* Cycle status row */}
      {cycleStatus && cycleStatus.status !== 'idle' && (
        <div className={cn(
          "flex flex-wrap items-center gap-2 px-2 py-1 rounded-md border text-xs",
          cycleStatus.status === 'running'
            ? 'bg-green-50 dark:bg-green-950 border-green-200 dark:border-green-800'
            : 'bg-yellow-50 dark:bg-yellow-950 border-yellow-200 dark:border-yellow-800'
        )}>
          <RefreshCw className={`h-3 w-3 ${cycleStatus.status === 'running' ? 'animate-spin text-green-600' : 'text-yellow-600'}`} />
          <span className="font-medium">
            {cycleStatus.status === 'running' ? 'Running' : 'Interrupted'}
          </span>
          <span className="text-muted-foreground">{cycleStatus.type}</span>
          {cycleStatus.phase && (
            <Badge variant="outline" className="text-[10px] h-4 px-1">{cycleStatus.phase}</Badge>
          )}
          {cycleStatus.current_task_id ? (
            <span>Task #{cycleStatus.current_task_id}</span>
          ) : null}
          {cycleStatus.target_total != null && cycleStatus.target_total > 0 && (
            <span className="text-muted-foreground">
              {cycleStatus.completed ?? 0}/{cycleStatus.target_total}
            </span>
          )}
          {cycleStatus.elapsed_sec != null && (
            <span className="text-muted-foreground ml-auto">
              {formatElapsed(cycleStatus.elapsed_sec)}
            </span>
          )}
        </div>
      )}

      {/* Status counts row */}
      <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
        {statuses.map(s => (
          <button
            key={s}
            className={cn(
              "flex items-center gap-1 px-1.5 py-0.5 rounded transition-colors",
              statusFilter === s
                ? "bg-accent text-accent-foreground"
                : "hover:bg-muted"
            )}
            onClick={() => onFilterChange(statusFilter === s ? null : s)}
          >
            <span className={`w-2 h-2 rounded-full ${colors[s]}`} />
            <span>{s}</span>
            <span className="font-medium text-foreground">{counts[s]}</span>
          </button>
        ))}
        <div className="flex items-center gap-1 ml-auto">
          <span>done/leaf</span>
          <span className="font-medium text-foreground">{leafDone}/{leafTotal}</span>
          <span className="text-muted-foreground">({progress}%)</span>
        </div>
      </div>
    </div>
  )
}

function formatElapsed(sec: number): string {
  if (sec < 60) return `${sec}s`
  const m = Math.floor(sec / 60)
  const s = sec % 60
  if (m < 60) return `${m}m ${s}s`
  const h = Math.floor(m / 60)
  return `${h}h ${m % 60}m`
}

function StatusDot({ status }: { status: string }) {
  const colors: Record<string, string> = {
    todo: 'bg-gray-400',
    split: 'bg-blue-400',
    planned: 'bg-yellow-400',
    done: 'bg-green-400',
    failed: 'bg-red-400',
  }
  return <span className={`inline-block w-3 h-3 rounded-full shrink-0 ${colors[status] || 'bg-gray-300'}`} />
}

function ListView({ items, onSelect, selectedId, isMobile }: { items: any[]; onSelect: (t: any) => void; selectedId?: number; isMobile?: boolean }) {
  const sortedItems = useMemo(() =>
    [...items].sort((a, b) => (b.id || b.ID) - (a.id || a.ID)),
    [items]
  )

  if (isMobile) {
    return (
      <div className="space-y-2">
        {sortedItems.map((t: any) => {
          const id = t.id || t.ID
          const title = t.title || t.Title || (t.spec || t.Spec || '').split('\n')[0] || '(untitled)'
          const status = t.status || t.Status || 'todo'
          const depth = t.depth ?? t.Depth ?? 0
          const parentId = t.parent_id ?? t.ParentID
          const isSelected = id === selectedId

          return (
            <div
              key={id}
              className={cn(
                "border rounded-md p-3 cursor-pointer hover:bg-accent",
                isSelected && "bg-accent border-primary"
              )}
              onClick={() => onSelect(t)}
            >
              <div className="flex items-center justify-between mb-1">
                <span className="text-xs text-muted-foreground">#{id}</span>
                <StatusDot status={status} />
              </div>
              <div className="font-medium text-sm truncate">{title}</div>
              <div className="flex items-center gap-3 mt-1.5 text-xs text-muted-foreground">
                <span>depth: {depth}</span>
                <span>parent: {parentId ? `#${parentId}` : '-'}</span>
              </div>
            </div>
          )
        })}
      </div>
    )
  }

  return (
    <table className="w-full text-sm">
      <thead>
        <tr className="border-b text-left">
          <th className="py-2.5 px-2 w-[50px]">ID</th>
          <th className="py-2.5 px-2">Title</th>
          <th className="py-2.5 px-2 w-[100px]">Status</th>
          <th className="py-2.5 px-2 w-[60px]">Depth</th>
          <th className="py-2.5 px-2 w-[60px]">Parent</th>
        </tr>
      </thead>
      <tbody>
        {sortedItems.map((t: any) => {
          const id = t.id || t.ID
          const title = t.title || t.Title || (t.spec || t.Spec || '').split('\n')[0] || '(untitled)'
          const status = t.status || t.Status || 'todo'
          const depth = t.depth ?? t.Depth ?? 0
          const parentId = t.parent_id ?? t.ParentID
          const isSelected = id === selectedId

          return (
            <tr
              key={id}
              className={`border-b cursor-pointer hover:bg-accent ${isSelected ? 'bg-accent' : ''}`}
              onClick={() => onSelect(t)}
            >
              <td className="py-2.5 px-2 text-muted-foreground">#{id}</td>
              <td className="py-2.5 px-2">{title}</td>
              <td className="py-2.5 px-2"><StatusDot status={status} /></td>
              <td className="py-2.5 px-2 text-center">{depth}</td>
              <td className="py-2.5 px-2 text-muted-foreground">{parentId ? `#${parentId}` : '-'}</td>
            </tr>
          )
        })}
      </tbody>
    </table>
  )
}

function parseItems(data: any): any[] {
  if (!data) return []
  if (Array.isArray(data)) return data
  if (data.items && Array.isArray(data.items)) return data.items
  return []
}
