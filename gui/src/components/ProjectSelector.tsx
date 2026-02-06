import { useState, useMemo, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useProjects, useSwitchProject, useStatus } from '@/hooks/useClaribot'
import { projectAPI } from '@/api/client'
import {
  FolderOpen, ChevronDown, Search, Pin, PinOff, ArrowUpDown,
  Clock, Calendar, ListTodo
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

export function ProjectSelector() {
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [sortField, setSortField] = useState<SortField>('last_accessed')
  const [sortDir, setSortDir] = useState<SortDir>('desc')
  const [categoryFilter, setCategoryFilter] = useState<string | null>(null)

  const { data: status } = useStatus()
  const { data: projects, refetch } = useProjects()
  const switchProject = useSwitchProject()

  const currentProject = status?.message?.match(/📌 (.+?) —/u)?.[1] || 'GLOBAL'

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
    switchProject.mutate(id)
    setOpen(false)
  }

  const handleTogglePin = async (e: React.MouseEvent, id: string, currentPinned: boolean) => {
    e.stopPropagation()
    await projectAPI.set(id, 'pinned', currentPinned ? '0' : '1')
    refetch()
  }

  const handleSetCategory = async (id: string, category: string) => {
    await projectAPI.set(id, 'category', category)
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

  return (
    <div className="relative ml-2 project-selector" onClick={e => e.stopPropagation()}>
      <Button
        variant="outline"
        size="sm"
        className="gap-1 h-8"
        onClick={() => setOpen(!open)}
      >
        <FolderOpen className="h-4 w-4" />
        <span className="hidden sm:inline max-w-[100px] truncate">{currentProject}</span>
        <ChevronDown className="h-3 w-3 shrink-0 opacity-50" />
      </Button>

      {open && (
        <div className="absolute top-full mt-1 left-0 z-50 w-[320px] rounded-md border bg-popover shadow-lg">
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

          {/* Sort & Filter */}
          <div className="px-2 py-1.5 border-b flex items-center gap-1 flex-wrap">
            {/* Sort */}
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

            <div className="w-px h-4 bg-border mx-1" />

            {/* Category Filter */}
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

          {/* Project List */}
          <ScrollArea className="max-h-[300px]">
            <div className="p-1">
              {/* GLOBAL option */}
              <button
                className={cn(
                  "w-full text-left px-2 py-2 text-xs rounded-sm hover:bg-accent flex items-center gap-2",
                  currentProject === 'GLOBAL' && "bg-accent"
                )}
                onClick={() => handleSwitch('none')}
              >
                <span className="font-medium">GLOBAL</span>
              </button>

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

                  {/* Category selector (on hover) */}
                  <select
                    className="opacity-0 group-hover:opacity-100 h-5 text-[10px] bg-transparent border rounded px-1 cursor-pointer"
                    value={p.category}
                    onClick={e => e.stopPropagation()}
                    onChange={e => handleSetCategory(p.id, e.target.value)}
                  >
                    <option value="">-</option>
                    {categories.map(cat => (
                      <option key={cat} value={cat}>{cat}</option>
                    ))}
                  </select>
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
        </div>
      )}
    </div>
  )
}
