import { NavLink, Link, useNavigate, useLocation } from 'react-router-dom'
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
  FileCode,
  RefreshCw,
  Layers,
  Wifi,
  WifiOff,
  LogOut,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useStatus, useHealth, useProjectStats, useProjects } from '@/hooks/useClaribot'
import { useLogout } from '@/hooks/useAuth'
import { ProjectSelector } from '@/components/ProjectSelector'
import type { ProjectStats, StatusResponse } from '@/types'

interface SidebarProps {
  collapsed: boolean
  onToggle: () => void
}

// Global navigation items (always available, project-independent)
export const globalNavItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/projects', icon: FolderOpen, label: 'Projects' },
  { to: '/messages', icon: MessageSquare, label: 'Messages' },
  { to: '/schedules', icon: Clock, label: 'Schedules' },
  { to: '/settings', icon: Settings, label: 'Settings' },
]

// Project-specific navigation items (context-aware)
export const projectNavItems = [
  { to: '/specs', icon: BookOpen, label: 'Specs' },
  { to: '/tasks', icon: CheckSquare, label: 'Tasks' },
  { to: '/files', icon: FileCode, label: 'Files' },
  { to: '/messages', icon: MessageSquare, label: 'Messages' },
  { to: '/schedules', icon: Clock, label: 'Schedules' },
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
  const { data: healthData } = useHealth()
  const { data: statsData } = useProjectStats()
  const { data: projectsData } = useProjects()
  const logout = useLogout()
  const navigate = useNavigate()
  const location = useLocation()

  // Detect project from URL (preferred) or status message (fallback)
  const projectFromUrl = location.pathname.match(/^\/projects\/([^/]+)/)?.[1]
  const currentProject = projectFromUrl || status?.project_id || 'GLOBAL'
  const hasProject = currentProject !== 'GLOBAL'

  // Claude status
  const claudeInfo = status?.message?.match(/\u{1F916} Claude: (\d+)\/(\d+)/u)
  const claudeUsed = claudeInfo?.[1] || '0'
  const claudeMax = claudeInfo?.[2] || '3'
  const isConnected = !!healthData

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
      {/* Logo + Toggle */}
      <div className="flex items-center justify-between p-2">
        {!collapsed && (
          <Link to="/" className="flex items-center gap-2 pl-1 hover:opacity-80 transition-opacity">
            <Layers className="h-5 w-5 text-primary shrink-0" />
            <span className="font-bold text-sm">Claribot</span>
          </Link>
        )}
        <Button variant="ghost" size="icon" onClick={onToggle} className="min-h-[44px] min-w-[44px]">
          {collapsed ? <PanelLeft className="h-4 w-4" /> : <PanelLeftClose className="h-4 w-4" />}
        </Button>
      </div>

      {/* Project Selector - always at top */}
      <div className="px-2 pb-2 border-b">
        <ProjectSelector collapsed={collapsed} onProjectSelect={(id) => {
          if (id !== 'none') {
            // Navigate to project-scoped tasks page
            navigate(`/projects/${id}/tasks`)
          } else {
            // Navigate to global dashboard
            navigate('/')
          }
        }} />
      </div>

      {/* Status Card */}
      {!collapsed && (
        <div className="mx-3 my-2 p-2 rounded-md border bg-card text-card-foreground">
          <div className="flex items-center justify-between">
            <Badge variant={Number(claudeUsed) > 0 ? "warning" : "secondary"} className="text-[10px] h-5 gap-1">
              Claude {claudeUsed}/{claudeMax}
            </Badge>
            {isConnected ? (
              <Badge variant="success" className="text-[10px] h-5 gap-1">
                <Wifi className="h-3 w-3" /> Connected
              </Badge>
            ) : (
              <Badge variant="destructive" className="text-[10px] h-5 gap-1">
                <WifiOff className="h-3 w-3" /> Offline
              </Badge>
            )}
          </div>
        </div>
      )}

      <nav className="flex-1 px-2 overflow-auto">
        {/* Project Stats Card - only when project selected and not collapsed */}
        {hasProject && !collapsed && (
          <div className="mx-1 my-2 p-2 rounded-md border bg-card text-card-foreground">
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

        {/* Project Navigation - only when project selected */}
        {hasProject && (
          <div className="space-y-1">
            {!collapsed && (
              <div className="px-3 py-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Project
              </div>
            )}
            <NavItem to={`/projects/${currentProject}/edit`} icon={Pencil} label="Edit" collapsed={collapsed} />
            {projectNavItems.map((item) => (
              <NavItem key={item.to} to={`/projects/${currentProject}${item.to}`} icon={item.icon} label={item.label} collapsed={collapsed} />
            ))}
          </div>
        )}

        {/* Global Navigation - only when GLOBAL (no project selected) */}
        {!hasProject && (
          <div className="space-y-1 mt-2">
            {globalNavItems.map((item) => (
              <NavItem key={item.to} {...item} collapsed={collapsed} />
            ))}
          </div>
        )}
      </nav>

      {/* Logout button - fixed at bottom */}
      <div className="border-t p-2">
        <Button
          variant="ghost"
          className={cn(
            "w-full justify-start gap-3 text-muted-foreground hover:text-foreground",
            collapsed && "justify-center px-0"
          )}
          onClick={() => logout.mutate()}
          disabled={logout.isPending}
          title="로그아웃"
        >
          <LogOut className="h-4 w-4 shrink-0" />
          {!collapsed && <span className="text-sm">로그아웃</span>}
        </Button>
      </div>
    </aside>
  )
}
