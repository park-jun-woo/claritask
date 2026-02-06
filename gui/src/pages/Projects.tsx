import { useState } from 'react'
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useProjects, useSwitchProject, useDeleteProject } from '@/hooks/useClaribot'
import { projectAPI } from '@/api/client'
import { FolderOpen, Plus, Trash2, ArrowRight } from 'lucide-react'

export default function Projects() {
  const { data: projects, refetch } = useProjects()
  const switchProject = useSwitchProject()
  const deleteProject = useDeleteProject()
  const [showAdd, setShowAdd] = useState(false)
  const [addForm, setAddForm] = useState({ path: '', type: '', description: '' })

  const projectList = parseItems(projects?.data)

  const handleAdd = async () => {
    if (!addForm.path) return
    await projectAPI.add(addForm.path, addForm.description || undefined)
    setAddForm({ path: '', type: '', description: '' })
    setShowAdd(false)
    refetch()
  }

  const handleDelete = (id: string) => {
    if (confirm(`Delete project "${id}"?`)) {
      deleteProject.mutate(id)
    }
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
            <Input
              placeholder="Type (e.g., dev.platform)"
              value={addForm.type}
              onChange={e => setAddForm(f => ({ ...f, type: e.target.value }))}
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
          const path = p.path || p.Path || ''
          const type = p.type || p.Type || ''
          const desc = p.description || p.Description || ''
          const status = p.status || p.Status || 'active'

          return (
            <Card key={id} className="flex flex-col">
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-2">
                    <FolderOpen className="h-5 w-5 text-muted-foreground" />
                    <CardTitle className="text-lg">{id}</CardTitle>
                  </div>
                  {type && <Badge variant="secondary" className="text-xs">{type}</Badge>}
                </div>
              </CardHeader>
              <CardContent className="flex-1 space-y-2">
                <p className="text-xs text-muted-foreground font-mono truncate">{path}</p>
                {desc && <p className="text-sm">{desc}</p>}
              </CardContent>
              <CardFooter className="gap-2 pt-3">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => switchProject.mutate(id)}
                  className="flex-1 min-h-[44px]"
                >
                  <ArrowRight className="h-4 w-4 mr-1" /> Select
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => handleDelete(id)}
                  className="text-destructive hover:text-destructive min-h-[44px] min-w-[44px]"
                >
                  <Trash2 className="h-4 w-4" />
                </Button>
              </CardFooter>
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
