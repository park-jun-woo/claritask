import { useState, useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import {
  useTasks, useAddTask, useDeleteTask, useTaskPlan, useTaskRun, useTaskCycle, useSetTask
} from '@/hooks/useClaribot'
import {
  Plus, Play, RefreshCw, ChevronRight, ChevronDown, X, TreePine, List, FileText
} from 'lucide-react'

type ViewMode = 'tree' | 'list'

export default function Tasks() {
  const { data: tasksData } = useTasks()
  const addTask = useAddTask()
  const deleteTask = useDeleteTask()
  const taskPlan = useTaskPlan()
  const taskRun = useTaskRun()
  const taskCycle = useTaskCycle()
  const setTask = useSetTask()

  const [viewMode, setViewMode] = useState<ViewMode>('tree')
  const [showAdd, setShowAdd] = useState(false)
  const [addForm, setAddForm] = useState({ title: '', parentId: '', spec: '' })
  const [selectedTask, setSelectedTask] = useState<any>(null)
  const [expandedNodes, setExpandedNodes] = useState<Set<number>>(new Set())
  const [editField, setEditField] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')

  const taskItems = useMemo(() => parseItems(tasksData?.data), [tasksData])

  // Build tree structure
  const treeData = useMemo(() => buildTree(taskItems), [taskItems])

  const handleAdd = async () => {
    if (!addForm.title) return
    await addTask.mutateAsync({
      title: addForm.title,
      parentId: addForm.parentId ? Number(addForm.parentId) : undefined,
      spec: addForm.spec || undefined,
    })
    setAddForm({ title: '', parentId: '', spec: '' })
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
    // Update selectedTask
    if (selectedTask && (selectedTask.id || selectedTask.ID) === taskId) {
      setSelectedTask((prev: any) => ({ ...prev, [field]: value }))
    }
  }

  return (
    <div className="flex gap-4 h-[calc(100vh-8rem)]">
      {/* Main Content */}
      <div className="flex-1 flex flex-col space-y-4 min-w-0">
        {/* Toolbar */}
        <div className="flex items-center justify-between flex-wrap gap-2">
          <h1 className="text-3xl font-bold">Tasks</h1>
          <div className="flex items-center gap-2">
            {/* View Mode Toggle */}
            <div className="flex border rounded-md">
              <Button
                variant={viewMode === 'tree' ? 'secondary' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('tree')}
                className="rounded-r-none"
              >
                <TreePine className="h-4 w-4" />
              </Button>
              <Button
                variant={viewMode === 'list' ? 'secondary' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('list')}
                className="rounded-l-none"
              >
                <List className="h-4 w-4" />
              </Button>
            </div>

            <Separator orientation="vertical" className="h-6" />

            <Button size="sm" onClick={() => setShowAdd(!showAdd)}>
              <Plus className="h-4 w-4 mr-1" /> Task
            </Button>
            <Button size="sm" variant="outline" onClick={() => taskPlan.mutate(undefined)} disabled={taskPlan.isPending}>
              <Play className="h-4 w-4 mr-1" /> Plan All
            </Button>
            <Button size="sm" variant="outline" onClick={() => taskRun.mutate(undefined)} disabled={taskRun.isPending}>
              <Play className="h-4 w-4 mr-1" /> Run All
            </Button>
            <Button size="sm" variant="outline" onClick={() => taskCycle.mutate()} disabled={taskCycle.isPending}>
              <RefreshCw className="h-4 w-4 mr-1" /> Cycle
            </Button>
          </div>
        </div>

        {/* Action Status */}
        {(taskPlan.isPending || taskRun.isPending || taskCycle.isPending) && (
          <div className="flex items-center gap-2 px-3 py-2 bg-yellow-50 dark:bg-yellow-950 rounded-md border border-yellow-200 dark:border-yellow-800">
            <RefreshCw className="h-4 w-4 animate-spin" />
            <span className="text-sm">
              {taskPlan.isPending && 'Planning...'}
              {taskRun.isPending && 'Running...'}
              {taskCycle.isPending && 'Cycling...'}
            </span>
          </div>
        )}

        {/* Add Form */}
        {showAdd && (
          <Card>
            <CardContent className="p-4 space-y-2">
              <Input
                placeholder="Task title"
                value={addForm.title}
                onChange={e => setAddForm(f => ({ ...f, title: e.target.value }))}
              />
              <div className="flex gap-2">
                <Input
                  placeholder="Parent ID (optional)"
                  value={addForm.parentId}
                  onChange={e => setAddForm(f => ({ ...f, parentId: e.target.value }))}
                  className="w-[150px]"
                />
                <Input
                  placeholder="Spec (optional)"
                  value={addForm.spec}
                  onChange={e => setAddForm(f => ({ ...f, spec: e.target.value }))}
                  className="flex-1"
                />
              </div>
              <div className="flex gap-2">
                <Button size="sm" onClick={handleAdd} disabled={addTask.isPending}>Add</Button>
                <Button size="sm" variant="ghost" onClick={() => setShowAdd(false)}>Cancel</Button>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Task View */}
        <ScrollArea className="flex-1 border rounded-md">
          <div className="p-4">
            {viewMode === 'tree' ? (
              <TreeView
                nodes={treeData}
                expandedNodes={expandedNodes}
                onToggle={toggleNode}
                onSelect={setSelectedTask}
                selectedId={selectedTask?.id || selectedTask?.ID}
              />
            ) : (
              <ListView
                items={taskItems}
                onSelect={setSelectedTask}
                selectedId={selectedTask?.id || selectedTask?.ID}
              />
            )}
            {taskItems.length === 0 && (
              <p className="text-center text-muted-foreground py-8">
                No tasks yet. Add a task to get started.
              </p>
            )}
          </div>
        </ScrollArea>
      </div>

      {/* Detail Panel */}
      {selectedTask && (
        <div className="w-[380px] border rounded-md flex flex-col shrink-0">
          <div className="flex items-center justify-between p-4 border-b">
            <h3 className="font-semibold">Task #{selectedTask.id || selectedTask.ID}</h3>
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setSelectedTask(null)}>
              <X className="h-4 w-4" />
            </Button>
          </div>
          <ScrollArea className="flex-1 p-4 space-y-4">
            <div className="space-y-4">
              <div>
                <h4 className="text-lg font-medium">{selectedTask.title || selectedTask.Title}</h4>
                <div className="flex items-center gap-2 mt-1">
                  <StatusBadge status={selectedTask.status || selectedTask.Status} />
                  <span className="text-xs text-muted-foreground">
                    depth: {selectedTask.depth ?? selectedTask.Depth ?? 0}
                  </span>
                  {(selectedTask.is_leaf || selectedTask.IsLeaf) && (
                    <Badge variant="outline" className="text-xs">leaf</Badge>
                  )}
                </div>
              </div>

              {/* Spec */}
              <DetailSection
                title="Spec"
                content={selectedTask.spec || selectedTask.Spec || ''}
                isEditing={editField === 'spec'}
                onEdit={() => { setEditField('spec'); setEditValue(selectedTask.spec || selectedTask.Spec || '') }}
                onSave={() => handleEdit(selectedTask.id || selectedTask.ID, 'spec', editValue)}
                onCancel={() => setEditField(null)}
                editValue={editValue}
                onEditChange={setEditValue}
              />

              {/* Plan */}
              <DetailSection
                title="Plan"
                content={selectedTask.plan || selectedTask.Plan || ''}
                isEditing={editField === 'plan'}
                onEdit={() => { setEditField('plan'); setEditValue(selectedTask.plan || selectedTask.Plan || '') }}
                onSave={() => handleEdit(selectedTask.id || selectedTask.ID, 'plan', editValue)}
                onCancel={() => setEditField(null)}
                editValue={editValue}
                onEditChange={setEditValue}
              />

              {/* Report */}
              <DetailSection
                title="Report"
                content={selectedTask.report || selectedTask.Report || ''}
                isEditing={editField === 'report'}
                onEdit={() => { setEditField('report'); setEditValue(selectedTask.report || selectedTask.Report || '') }}
                onSave={() => handleEdit(selectedTask.id || selectedTask.ID, 'report', editValue)}
                onCancel={() => setEditField(null)}
                editValue={editValue}
                onEditChange={setEditValue}
              />

              <Separator />

              {/* Actions */}
              <div className="flex gap-2 flex-wrap">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => taskPlan.mutate(selectedTask.id || selectedTask.ID)}
                  disabled={taskPlan.isPending}
                >
                  <Play className="h-3 w-3 mr-1" /> Plan
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => taskRun.mutate(selectedTask.id || selectedTask.ID)}
                  disabled={taskRun.isPending}
                >
                  <Play className="h-3 w-3 mr-1" /> Run
                </Button>
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => {
                    if (confirm('Delete this task?')) {
                      deleteTask.mutate(selectedTask.id || selectedTask.ID)
                      setSelectedTask(null)
                    }
                  }}
                >
                  Delete
                </Button>
              </div>
            </div>
          </ScrollArea>
        </div>
      )}
    </div>
  )
}

// --- Sub-components ---

function StatusBadge({ status }: { status: string }) {
  const variants: Record<string, { variant: any; icon: string }> = {
    spec_ready: { variant: 'secondary', icon: '\u25CB' },
    subdivided: { variant: 'info', icon: '\u25D0' },
    plan_ready: { variant: 'warning', icon: '\u25CF' },
    done: { variant: 'success', icon: '\u2705' },
    failed: { variant: 'destructive', icon: '\u274C' },
  }
  const v = variants[status] || { variant: 'secondary', icon: '?' }
  return <Badge variant={v.variant}>{v.icon} {status}</Badge>
}

function DetailSection({
  title, content, isEditing, onEdit, onSave, onCancel, editValue, onEditChange
}: {
  title: string
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
      <div className="flex items-center justify-between">
        <h5 className="text-sm font-medium text-muted-foreground">{title}</h5>
        {!isEditing && (
          <Button variant="ghost" size="sm" className="h-6 text-xs" onClick={onEdit}>
            <FileText className="h-3 w-3 mr-1" /> Edit
          </Button>
        )}
      </div>
      {isEditing ? (
        <div className="mt-1 space-y-1">
          <Textarea value={editValue} onChange={e => onEditChange(e.target.value)} rows={4} />
          <div className="flex gap-1">
            <Button size="sm" className="h-7 text-xs" onClick={onSave}>Save</Button>
            <Button size="sm" variant="ghost" className="h-7 text-xs" onClick={onCancel}>Cancel</Button>
          </div>
        </div>
      ) : (
        <pre className="mt-1 text-sm whitespace-pre-wrap bg-muted rounded p-2 max-h-[200px] overflow-auto">
          {content || '(empty)'}
        </pre>
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
        const title = node.task.title || node.task.Title || '(untitled)'
        const status = node.task.status || node.task.Status || 'spec_ready'
        const hasChildren = node.children.length > 0
        const isExpanded = expandedNodes.has(id)
        const isSelected = id === selectedId

        return (
          <div key={id}>
            <div
              className={`flex items-center gap-1 py-1 px-2 rounded cursor-pointer text-sm hover:bg-accent ${isSelected ? 'bg-accent' : ''}`}
              style={{ paddingLeft: `${depth * 20 + 8}px` }}
              onClick={() => onSelect(node.task)}
            >
              {hasChildren ? (
                <button onClick={e => { e.stopPropagation(); onToggle(id) }} className="p-0.5">
                  {isExpanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
                </button>
              ) : (
                <span className="w-4" />
              )}
              <StatusDot status={status} />
              <span className="text-muted-foreground text-xs mr-1">#{id}</span>
              <span className="truncate">{title}</span>
              <span className="ml-auto text-xs text-muted-foreground">{status}</span>
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

function StatusDot({ status }: { status: string }) {
  const colors: Record<string, string> = {
    spec_ready: 'bg-gray-400',
    subdivided: 'bg-blue-400',
    plan_ready: 'bg-yellow-400',
    done: 'bg-green-400',
    failed: 'bg-red-400',
  }
  return <span className={`w-2 h-2 rounded-full shrink-0 ${colors[status] || 'bg-gray-300'}`} />
}

function ListView({ items, onSelect, selectedId }: { items: any[]; onSelect: (t: any) => void; selectedId?: number }) {
  return (
    <table className="w-full text-sm">
      <thead>
        <tr className="border-b text-left">
          <th className="py-2 px-2 w-[50px]">ID</th>
          <th className="py-2 px-2">Title</th>
          <th className="py-2 px-2 w-[100px]">Status</th>
          <th className="py-2 px-2 w-[60px]">Depth</th>
          <th className="py-2 px-2 w-[60px]">Parent</th>
        </tr>
      </thead>
      <tbody>
        {items.map((t: any) => {
          const id = t.id || t.ID
          const title = t.title || t.Title || '(untitled)'
          const status = t.status || t.Status || 'spec_ready'
          const depth = t.depth ?? t.Depth ?? 0
          const parentId = t.parent_id ?? t.ParentID
          const isSelected = id === selectedId

          return (
            <tr
              key={id}
              className={`border-b cursor-pointer hover:bg-accent ${isSelected ? 'bg-accent' : ''}`}
              onClick={() => onSelect(t)}
            >
              <td className="py-2 px-2 text-muted-foreground">#{id}</td>
              <td className="py-2 px-2">{title}</td>
              <td className="py-2 px-2"><StatusBadge status={status} /></td>
              <td className="py-2 px-2 text-center">{depth}</td>
              <td className="py-2 px-2 text-muted-foreground">{parentId ? `#${parentId}` : '-'}</td>
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
