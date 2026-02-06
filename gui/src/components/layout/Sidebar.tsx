import { NavLink } from 'react-router-dom'
import {
  LayoutDashboard,
  FolderOpen,
  CheckSquare,
  MessageSquare,
  Clock,
  Settings,
  PanelLeft,
  PanelLeftClose,
  Pencil,
  BookOpen,
  RefreshCw,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useStatus, useProjectStats, useProjects } from '@/hooks/useClaribot'
import type { ProjectStats, StatusResponse } from '@/types'

interface SidebarProps {
  collapsed: boolean
  onToggle: () => void
}

// Global navigation items (always available, project-independent)
export const globalNavItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/messages', icon: MessageSquare, label: 'Messages' },
  { to: '/projects', icon: FolderOpen, label: 'Projects' },
  { to: '/schedules', icon: Clock, label: 'Schedules' },
  { to: '/settings', icon: Settings, label: 'Settings' },
]

// Project-specific navigation items (context-aware)
export const projectNavItems = [
  { to: '/specs', icon: BookOpen, label: 'Specs' },
  { to: '/tasks', icon: CheckSquare, label: 'Tasks' },
]

// Combined for Header mobile menu
export const navItems = [...globalNavItems, ...projectNavItems]

function NavItem({ to, icon: Icon, label, collapsed }: { to: string; icon: React.ComponentType<{ className?: string }>; label: string; collapsed: boolean }) {
  return (
    <NavLink
      to={to}
      end={to === '/'}
      className={({ isActive }) =>
        cn(
          "flex items-center gap-3 rounded-md px-3 py-3 text-sm font-medium transition-colors",
          "hover:bg-accent hover:text-accent-foreground",
          isActive
            ? "bg-accent text-accent-foreground"
            : "text-muted-foreground"
        )
      }
    >
      <Icon className="h-5 w-5 shrink-0" />
      {!collapsed && <span>{label}</span>}
    </NavLink>
  )
}

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const { data: status } = useStatus() as { data: StatusResponse | undefined }
  const { data: statsData } = useProjectStats()
  const { data: projectsData } = useProjects()

  // Parse current project from status message (📌 project-id — ...)
  const currentProject = status?.message?.match(/📌 (.+?) —/u)?.[1] || 'GLOBAL'

  // Get current project stats
  const projectStats: ProjectStats[] = statsData?.data
    ? (Array.isArray(statsData.data) ? statsData.data : statsData.data.items || [])
    : []
  const currentStats = projectStats.find(p => p.project_id === currentProject)
  const s = currentStats?.stats || { total: 0, leaf: 0, todo: 0, planned: 0, split: 0, done: 0, failed: 0 }
  const leafTotal = s.leaf || 1
  const leafDone = s.done
  const progress = leafTotal > 0 ? Math.round((leafDone / leafTotal) * 100) : 0

  // Get category
  const projectItems = projectsData?.data
    ? (Array.isArray(projectsData.data) ? projectsData.data : projectsData.data.items || [])
    : []
  const currentProjectData = projectItems.find((p: any) => (p.id || p.ID) === currentProject)
  const category = currentProjectData?.category || currentProjectData?.Category || ''

  // Cycle status - check cycle_statuses array for multiple running cycles
  const isRunning = status?.cycle_statuses?.some(
    c => c.status === 'running' && c.project_id === currentProject
  ) || (status?.cycle_status?.status === 'running' && status?.cycle_status?.project_id === currentProject)

  return (
    <aside
      className={cn(
        "flex flex-col border-r bg-background transition-all duration-200",
        collapsed ? "w-[60px]" : "w-[220px]"
      )}
    >
      <div className="flex items-center justify-end p-2">
        <Button variant="ghost" size="icon" onClick={onToggle} className="min-h-[44px] min-w-[44px]">
          {collapsed ? <PanelLeft className="h-4 w-4" /> : <PanelLeftClose className="h-4 w-4" />}
        </Button>
      </div>
      <nav className="flex-1 px-2">
        <div className="space-y-1">
          {!collapsed && (
            <div className="px-3 py-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              Project
            </div>
          )}

          {/* Current Project Card (collapsed: icon only) */}
          {!collapsed && currentProject !== 'GLOBAL' && (
            <div className="mx-1 mb-2 p-2 rounded-md border bg-card text-card-foreground">
              {/* Project name & category */}
              <div className="flex items-center gap-1.5 mb-2">
                {isRunning ? (
                  <RefreshCw className="h-3.5 w-3.5 text-green-500 animate-spin shrink-0" />
                ) : (
                  <FolderOpen className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                )}
                <span className="text-xs font-medium truncate">{currentProject}</span>
                {category && (
                  <Badge variant="outline" className="text-[9px] h-4 px-1 shrink-0">
                    {category}
                  </Badge>
                )}
              </div>

              {/* Status counts */}
              <div className="flex flex-wrap gap-1 mb-2 text-[10px]">
                {s.todo > 0 && <Badge variant="secondary" className="h-4 px-1">{s.todo} todo</Badge>}
                {s.planned > 0 && <Badge variant="secondary" className="h-4 px-1 bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300">{s.planned} plan</Badge>}
                {s.done > 0 && <Badge variant="secondary" className="h-4 px-1 bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300">{s.done} done</Badge>}
                {s.failed > 0 && <Badge variant="destructive" className="h-4 px-1">{s.failed} fail</Badge>}
              </div>

              {/* Stacked status bar */}
              {s.total > 0 && (
                <div className="h-1.5 rounded-full bg-secondary flex overflow-hidden mb-1">
                  {s.done > 0 && <div className="bg-green-400 h-full" style={{ width: `${(s.done / leafTotal) * 100}%` }} />}
                  {s.planned > 0 && <div className="bg-yellow-400 h-full" style={{ width: `${(s.planned / leafTotal) * 100}%` }} />}
                  {s.todo > 0 && <div className="bg-gray-400 h-full" style={{ width: `${(s.todo / leafTotal) * 100}%` }} />}
                  {s.failed > 0 && <div className="bg-red-400 h-full" style={{ width: `${(s.failed / leafTotal) * 100}%` }} />}
                </div>
              )}

              {/* Progress text */}
              <div className="flex justify-between text-[10px] text-muted-foreground">
                <span>{leafDone}/{leafTotal}</span>
                <span>{progress}%</span>
              </div>
            </div>
          )}

          {/* Edit Project - dynamic link */}
          <NavItem to={`/projects/${currentProject}/edit`} icon={Pencil} label="Edit" collapsed={collapsed} />
          {projectNavItems.map((item) => (
            <NavItem key={item.to} {...item} collapsed={collapsed} />
          ))}
        </div>
      </nav>
    </aside>
  )
}
