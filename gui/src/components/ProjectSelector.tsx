import { useState, useMemo, useEffect } from 'react'
import { useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useProjects } from '@/hooks/useClaribot'
import { projectAPI } from '@/api/client'
import {
  FolderOpen, ChevronDown, Search, Pin, PinOff, ArrowUpDown,
  Clock, Calendar, ListTodo, ExternalLink,
} from 'lucide-react'

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

interface ProjectSelectorProps {
  collapsed?: boolean
  onProjectSelect?: (projectId: string) => void
}

export function ProjectSelector({ collapsed = false, onProjectSelect }: ProjectSelectorProps) {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [sortField, setSortField] = useState<SortField>('last_accessed')
  const [sortDir, setSortDir] = useState<SortDir>('desc')
  const [categoryFilter, setCategoryFilter] = useState<string | null>(null)
  const location = useLocation()
  const { data: projects, refetch } = useProjects()

  // Detect project from URL (consistent with Sidebar/Header)
  const projectFromUrl = location.pathname.match(/^\/projects\/([^/]+)/)?.[1]
  const currentProject = projectFromUrl || 'GLOBAL'

  // Parse project list
  const projectList: ProjectItem[] = useMemo(() => {
    if (!projects?.data) return []
    const data = projects.data as any
    const items = Array.isArray(data) ? data : data.items || []
    return items.map((p: any) => ({
      id: p.id || p.ID,
      description: p.description || p.Description || '',
      category: p.category || p.Category || '',
      pinned: p.pinned || p.Pinned || false,
      last_accessed: p.last_accessed || p.LastAccessed || '',
      created_at: p.created_at || p.CreatedAt || '',
      task_count: p.task_count,
    }))
  }, [projects])

  // Current project data
  const currentProjectData = projectList.find(p => p.id === currentProject)

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
      // Pinned first
      if (a.pinned !== b.pinned) return a.pinned ? -1 : 1

      // Then by selected field
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

  // Close on outside click
  useEffect(() => {
    if (!open) return
    const handler = (e: MouseEvent) => {
      const target = e.target as HTMLElement
      if (!target.closest('.project-selector')) {
        setOpen(false)
      }
    }
    document.addEventListener('click', handler)
    return () => document.removeEventListener('click', handler)
  }, [open])

  const handleSwitch = (id: string) => {
    // Navigate to project URL; ProjectLayout handles backend switch
    onProjectSelect?.(id)
    setOpen(false)
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

  const toggleSortDir = () => {
    setSortDir(d => d === 'asc' ? 'desc' : 'asc')
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

  // Collapsed mode: icon only
  if (collapsed) {
    return (
      <div className="relative project-selector" onClick={e => e.stopPropagation()}>
        <button
          className="flex items-center justify-center w-full p-2 rounded-md hover:bg-accent transition-colors"
          onClick={() => setOpen(!open)}
          title={currentProject === 'GLOBAL' ? '프로젝트 선택' : currentProject}
        >
          <div className="flex items-center justify-center h-8 w-8 rounded-md bg-primary/10">
            <FolderOpen className="h-4 w-4 text-primary" />
          </div>
        </button>

        {open && (
          <div className="absolute top-full mt-1 left-0 z-50 w-[320px] rounded-md border bg-popover shadow-lg">
            {renderDropdown()}
          </div>
        )}
      </div>
    )
  }

  function renderDropdown() {
    return (
      <>
        {/* Search */}
        <div className="p-2 border-b">
          <div className="relative">
            <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-muted-foreground" />
            <Input
              placeholder="검색..."
              value={search}
              onChange={e => setSearch(e.target.value)}
              className="h-7 pl-7 text-xs"
              autoFocus
            />
          </div>
        </div>

        {/* Sort */}
        <div className="px-2 py-1 border-b flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            className="h-6 text-[10px] gap-1 px-1.5"
            onClick={cycleSortField}
            title="정렬 기준 변경"
          >
            <SortIcon className="h-3 w-3" />
            {sortLabel[sortField]}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="h-6 w-6 p-0"
            onClick={toggleSortDir}
            title={sortDir === 'desc' ? '내림차순' : '오름차순'}
          >
            <ArrowUpDown className={cn("h-3 w-3", sortDir === 'asc' && "rotate-180")} />
          </Button>
        </div>

        {/* Category Filter */}
        {categories.length > 0 && (
          <div className="px-2 py-1 border-b flex items-center gap-1 flex-wrap">
            <Button
              variant={categoryFilter === null ? "secondary" : "ghost"}
              size="sm"
              className="h-6 text-[10px] px-1.5"
              onClick={() => setCategoryFilter(null)}
            >
              전체
            </Button>
            {categories.map(cat => (
              <Button
                key={cat}
                variant={categoryFilter === cat ? "secondary" : "ghost"}
                size="sm"
                className="h-6 text-[10px] px-1.5"
                onClick={() => setCategoryFilter(cat)}
              >
                {cat}
              </Button>
            ))}
          </div>
        )}

        {/* Project List */}
        <ScrollArea className="max-h-[300px]">
          <div className="p-1">
            {/* Projects */}
            {filteredProjects.map(p => (
              <div
                key={p.id}
                className={cn(
                  "w-full text-left px-2 py-2 text-xs rounded-sm hover:bg-accent flex items-center gap-2 cursor-pointer group",
                  currentProject === p.id && "bg-accent"
                )}
                onClick={() => handleSwitch(p.id)}
              >
                {/* Pin button */}
                <button
                  className="opacity-50 hover:opacity-100 shrink-0"
                  onClick={e => handleTogglePin(e, p.id, p.pinned)}
                  title={p.pinned ? '고정 해제' : '고정'}
                >
                  {p.pinned ? (
                    <Pin className="h-3 w-3 text-primary" />
                  ) : (
                    <PinOff className="h-3 w-3 opacity-0 group-hover:opacity-50" />
                  )}
                </button>

                {/* Project info */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-1">
                    <span className="font-medium truncate">{p.id}</span>
                    {p.category && (
                      <Badge variant="outline" className="text-[9px] h-4 px-1">
                        {p.category}
                      </Badge>
                    )}
                  </div>
                  {p.description && (
                    <div className="text-[10px] text-muted-foreground truncate">
                      {p.description}
                    </div>
                  )}
                </div>

              </div>
            ))}

            {filteredProjects.length === 0 && projectList.length > 0 && (
              <div className="text-center text-muted-foreground py-4 text-xs">
                검색 결과 없음
              </div>
            )}

            {projectList.length === 0 && (
              <div className="text-center text-muted-foreground py-4 text-xs">
                프로젝트 없음
              </div>
            )}
          </div>
        </ScrollArea>

        {/* GLOBAL link */}
        <div className="border-t p-1">
          <button
            className="w-full text-left px-2 py-2 text-xs rounded-sm hover:bg-accent flex items-center gap-2 text-muted-foreground"
            onClick={() => handleSwitch('none')}
          >
            <ExternalLink className="h-3 w-3" />
            <span>GLOBAL</span>
          </button>
        </div>
      </>
    )
  }

  return (
    <div className="relative project-selector" onClick={e => e.stopPropagation()}>
      {/* Full-width sidebar button - gozip BuildingSelector style */}
      <button
        className="flex items-center gap-3 w-full p-2 rounded-md hover:bg-accent transition-colors text-left"
        onClick={() => setOpen(!open)}
      >
        <div className="flex items-center justify-center h-8 w-8 rounded-md bg-primary/10 shrink-0">
          <FolderOpen className="h-4 w-4 text-primary" />
        </div>
        <div className="flex-1 min-w-0">
          {currentProject === 'GLOBAL' ? (
            <div className="text-xs text-muted-foreground">프로젝트 선택</div>
          ) : (
            <>
              <div className="text-sm font-medium truncate">{currentProject}</div>
              {currentProjectData?.category && (
                <div className="text-[10px] text-muted-foreground truncate">
                  {currentProjectData.category}
                </div>
              )}
            </>
          )}
        </div>
        <ChevronDown className={cn("h-4 w-4 shrink-0 text-muted-foreground transition-transform", open && "rotate-180")} />
      </button>

      {/* Dropdown */}
      {open && (
        <div className="absolute left-0 right-0 top-full z-50 rounded-md border bg-popover shadow-lg">
          {renderDropdown()}
        </div>
      )}
    </div>
  )
}
