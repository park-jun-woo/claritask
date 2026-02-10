import { useState, useEffect } from 'react'
import { NavLink, Link, useLocation, useNavigate } from 'react-router-dom'
import { Layers, Menu, Wifi, WifiOff, FolderOpen, RefreshCw, Pencil, LogOut } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Sheet, SheetTrigger, SheetContent } from '@/components/ui/sheet'
import { globalNavItems, projectNavItems, getProjectSwitchTarget } from '@/components/layout/Sidebar'
import { useStatus, useHealth, useProjectStats, useProjects } from '@/hooks/useClaribot'
import { useLogout } from '@/hooks/useAuth'
import { ProjectSelector } from '@/components/ProjectSelector'
import type { ProjectStats, StatusResponse } from '@/types'

export function Header() {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const location = useLocation()

  // Close drawer on navigation
  useEffect(() => {
    setDrawerOpen(false)
  }, [location.pathname])

  const { data: status } = useStatus() as { data: StatusResponse | undefined }
  const { data: healthData } = useHealth()
  const { data: statsData } = useProjectStats()
  const { data: projectsData } = useProjects()
  const navigate = useNavigate()
  const logout = useLogout()

  // Detect project from URL only (global state when not in /projects/:id)
  const projectFromUrl = location.pathname.match(/^\/projects\/([^/]+)/)?.[1]
  const currentProject = projectFromUrl || 'GLOBAL'
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

  // Cycle status
  const isRunning = status?.cycle_statuses?.some(
    c => c.status === 'running' && c.project_id === currentProject
  ) || (status?.cycle_status?.status === 'running' && status?.cycle_status?.project_id === currentProject)

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 md:hidden">
      <div className="flex h-14 items-center px-2 gap-2">
        {/* Mobile Hamburger Menu */}
        <Sheet open={drawerOpen} onOpenChange={setDrawerOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="min-h-[44px] min-w-[44px]">
              <Menu className="h-5 w-5" />
              <span className="sr-only">메뉴</span>
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-[260px] p-0 pt-12 flex flex-col">
            {/* Project Selector */}
            <div className="px-3 pb-2 border-b">
              <ProjectSelector collapsed={false} onProjectSelect={(id) => {
                navigate(getProjectSwitchTarget(location.pathname, id))
              }} />
            </div>

            {/* Status Card */}
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

            {/* Project Stats Card - only when project selected */}
            {hasProject && (
              <div className="mx-3 my-1 p-2 rounded-md border bg-card text-card-foreground">
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
                <div className="flex flex-wrap gap-1 mb-2 text-[10px]">
                  {s.todo > 0 && <Badge variant="secondary" className="h-4 px-1">{s.todo} todo</Badge>}
                  {s.planned > 0 && <Badge variant="secondary" className="h-4 px-1 bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300">{s.planned} plan</Badge>}
                  {s.done > 0 && <Badge variant="secondary" className="h-4 px-1 bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300">{s.done} done</Badge>}
                  {s.failed > 0 && <Badge variant="destructive" className="h-4 px-1">{s.failed} fail</Badge>}
                </div>
                {s.total > 0 && (
                  <div className="h-1.5 rounded-full bg-secondary flex overflow-hidden mb-1">
                    {s.done > 0 && <div className="bg-green-400 h-full" style={{ width: `${(s.done / leafTotal) * 100}%` }} />}
                    {s.planned > 0 && <div className="bg-yellow-400 h-full" style={{ width: `${(s.planned / leafTotal) * 100}%` }} />}
                    {s.todo > 0 && <div className="bg-gray-400 h-full" style={{ width: `${(s.todo / leafTotal) * 100}%` }} />}
                    {s.failed > 0 && <div className="bg-red-400 h-full" style={{ width: `${(s.failed / leafTotal) * 100}%` }} />}
                  </div>
                )}
                <div className="flex justify-between text-[10px] text-muted-foreground">
                  <span>{leafDone}/{leafTotal}</span>
                  <span>{progress}%</span>
                </div>
              </div>
            )}

            {/* Navigation */}
            <nav className="flex-1 flex flex-col px-3 overflow-auto">
              {/* Project Navigation - only when project selected */}
              {hasProject && (
                <div className="space-y-1">
                  <div className="px-3 py-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                    Project
                  </div>
                  <NavLink
                    to={`/projects/${currentProject}/edit`}
                    className={({ isActive }) =>
                      cn(
                        "flex items-center gap-3 rounded-md px-3 py-3 text-sm font-medium transition-colors",
                        "hover:bg-accent hover:text-accent-foreground",
                        isActive ? "bg-accent text-accent-foreground" : "text-muted-foreground"
                      )
                    }
                  >
                    <Pencil className="h-5 w-5 shrink-0" />
                    <span>Edit</span>
                  </NavLink>
                  {projectNavItems.map(({ to, icon: Icon, label }) => (
                    <NavLink
                      key={to}
                      to={`/projects/${currentProject}${to}`}
                      end={to === '/'}
                      className={({ isActive }) =>
                        cn(
                          "flex items-center gap-3 rounded-md px-3 py-3 text-sm font-medium transition-colors",
                          "hover:bg-accent hover:text-accent-foreground",
                          isActive ? "bg-accent text-accent-foreground" : "text-muted-foreground"
                        )
                      }
                    >
                      <Icon className="h-5 w-5 shrink-0" />
                      <span>{label}</span>
                    </NavLink>
                  ))}
                </div>
              )}

              {/* Global Navigation - only when GLOBAL */}
              {!hasProject && (
                <div className="space-y-1 mt-2">
                  {globalNavItems.map(({ to, icon: Icon, label }) => (
                    <NavLink
                      key={to}
                      to={to}
                      end={to === '/'}
                      className={({ isActive }) =>
                        cn(
                          "flex items-center gap-3 rounded-md px-3 py-3 text-sm font-medium transition-colors",
                          "hover:bg-accent hover:text-accent-foreground",
                          isActive ? "bg-accent text-accent-foreground" : "text-muted-foreground"
                        )
                      }
                    >
                      <Icon className="h-5 w-5 shrink-0" />
                      <span>{label}</span>
                    </NavLink>
                  ))}
                </div>
              )}
            </nav>

            {/* Logout button - fixed at bottom */}
            <div className="border-t p-3">
              <Button
                variant="ghost"
                className="w-full justify-start gap-3 text-muted-foreground hover:text-foreground"
                onClick={() => logout.mutate()}
                disabled={logout.isPending}
                title="로그아웃"
              >
                <LogOut className="h-4 w-4 shrink-0" />
                <span className="text-sm">로그아웃</span>
              </Button>
            </div>
          </SheetContent>
        </Sheet>

        {/* Logo */}
        <Link to="/" className="flex items-center gap-2 font-bold text-lg hover:opacity-80 transition-opacity">
          <Layers className="h-6 w-6 text-primary" />
          <span>Claribot</span>
        </Link>
      </div>
    </header>
  )
}
