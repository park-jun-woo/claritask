import { useState, useEffect, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, CardContent, CardFooter, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useProject, useUpdateProject, useDeleteProject, useProjects } from '@/hooks/useClaribot'
import { projectAPI } from '@/api/client'
import { ArrowLeft, Save, Trash2, Plus } from 'lucide-react'

export default function ProjectEdit() {
  const { projectId: id } = useParams<{ projectId: string }>()
  const navigate = useNavigate()
  const { data: projectData, isLoading } = useProject(id)
  const { data: projectsData } = useProjects()
  const updateProject = useUpdateProject()
  const deleteProject = useDeleteProject()

  const [description, setDescription] = useState('')
  const [parallel, setParallel] = useState(3)
  const [category, setCategory] = useState('')
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [deleteConfirm, setDeleteConfirm] = useState('')
  const [showAddCategory, setShowAddCategory] = useState(false)
  const [newCategory, setNewCategory] = useState('')

  // Get unique categories from all projects
  const categories = useMemo(() => {
    const items = projectsData?.data
    if (!items) return []
    const list = Array.isArray(items) ? items : items.items || []
    const cats = new Set<string>()
    list.forEach((p: any) => {
      const cat = p.category || p.Category
      if (cat) cats.add(cat)
    })
    return Array.from(cats).sort()
  }, [projectsData])

  // Initialize form when project data loads
  useEffect(() => {
    if (projectData?.data) {
      const p = projectData.data as any
      setDescription(p.description || p.Description || '')
      setParallel(p.parallel || p.Parallel || 3)
      setCategory(p.category || p.Category || '')
    }
  }, [projectData])

  const handleSave = async () => {
    if (!id) return
    // Save category separately
    await projectAPI.set(id, 'category', category)
    updateProject.mutate(
      { id, data: { description, parallel } },
      {
        onSuccess: () => navigate('/projects'),
      }
    )
  }

  const handleAddCategory = () => {
    if (newCategory.trim() && !categories.includes(newCategory.trim())) {
      setCategory(newCategory.trim())
      setShowAddCategory(false)
      setNewCategory('')
    }
  }

  const handleDelete = () => {
    if (!id || deleteConfirm !== id) return
    deleteProject.mutate(id, {
      onSuccess: () => navigate('/projects'),
    })
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">Loading...</p>
      </div>
    )
  }

  const project = projectData?.data as any

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => navigate('/projects')}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <h1 className="text-2xl md:text-3xl font-bold">Edit Project</h1>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{id}</CardTitle>
          {project?.path && (
            <p className="text-xs text-muted-foreground font-mono">{project.path}</p>
          )}
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Project ID</label>
            <Input value={id || ''} disabled className="bg-muted" />
            <p className="text-xs text-muted-foreground">Project ID cannot be changed</p>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Description</label>
            <Textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter project description"
              rows={3}
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Category</label>
            <div className="flex flex-wrap gap-2">
              <Button
                type="button"
                variant={category === '' ? 'default' : 'outline'}
                size="sm"
                className="h-9"
                onClick={() => setCategory('')}
              >
                없음
              </Button>
              {/* Show current category if not in list */}
              {category && !categories.includes(category) && (
                <Button
                  type="button"
                  variant="default"
                  size="sm"
                  className="h-9"
                >
                  {category}
                </Button>
              )}
              {categories.map(cat => (
                <Button
                  key={cat}
                  type="button"
                  variant={category === cat ? 'default' : 'outline'}
                  size="sm"
                  className="h-9"
                  onClick={() => setCategory(cat)}
                >
                  {cat}
                </Button>
              ))}
              {!showAddCategory ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="h-9 w-9 p-0"
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
                    className="h-9 w-32 text-sm"
                    onKeyDown={e => {
                      if (e.key === 'Enter') {
                        e.preventDefault()
                        handleAddCategory()
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
                    className="h-9"
                    onClick={handleAddCategory}
                  >
                    추가
                  </Button>
                </div>
              )}
            </div>
            <p className="text-xs text-muted-foreground">
              프로젝트 분류를 위한 카테고리 (Save 클릭 시 저장)
            </p>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Parallel Claude Count</label>
            <Input
              type="number"
              min={1}
              max={10}
              value={parallel}
              onChange={(e) => setParallel(Number(e.target.value))}
            />
            <p className="text-xs text-muted-foreground">
              Number of Claude instances to run in parallel for this project (1-10)
            </p>
          </div>
        </CardContent>
        <CardFooter className="gap-2">
          <Button onClick={handleSave} disabled={updateProject.isPending}>
            <Save className="h-4 w-4 mr-2" />
            {updateProject.isPending ? 'Saving...' : 'Save'}
          </Button>
          <Button variant="outline" onClick={() => navigate('/projects')}>
            Cancel
          </Button>
        </CardFooter>
      </Card>

      {/* Danger Zone */}
      <Card className="border-destructive">
        <CardHeader>
          <CardTitle className="text-lg text-destructive">Danger Zone</CardTitle>
          <CardDescription>
            Deleting a project will remove it from Claribot. The actual files on disk will not be deleted.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {!showDeleteConfirm ? (
            <Button
              variant="destructive"
              onClick={() => setShowDeleteConfirm(true)}
            >
              <Trash2 className="h-4 w-4 mr-2" />
              Delete Project
            </Button>
          ) : (
            <div className="space-y-4">
              <div className="space-y-2">
                <label className="text-sm font-medium">
                  To confirm, type <span className="font-mono font-bold">{id}</span> below:
                </label>
                <Input
                  value={deleteConfirm}
                  onChange={(e) => setDeleteConfirm(e.target.value)}
                  placeholder={`Type ${id} to confirm`}
                  className="font-mono"
                  autoFocus
                />
              </div>
              <div className="flex gap-2">
                <Button
                  variant="destructive"
                  onClick={handleDelete}
                  disabled={deleteConfirm !== id || deleteProject.isPending}
                >
                  <Trash2 className="h-4 w-4 mr-2" />
                  {deleteProject.isPending ? 'Deleting...' : 'Confirm Delete'}
                </Button>
                <Button
                  variant="outline"
                  onClick={() => {
                    setShowDeleteConfirm(false)
                    setDeleteConfirm('')
                  }}
                >
                  Cancel
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
