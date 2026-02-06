import { useState, useEffect } from 'react'
import { NavLink, Link, useLocation } from 'react-router-dom'
import { Layers, Wifi, WifiOff, Menu, LogOut, ChevronDown, FolderOpen } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Sheet, SheetTrigger, SheetContent } from '@/components/ui/sheet'
import { globalNavItems, projectNavItems } from '@/components/layout/Sidebar'
import { useStatus, useHealth, useProjects, useSwitchProject } from '@/hooks/useClaribot'
import { useLogout } from '@/hooks/useAuth'

export function Header() {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [showProjects, setShowProjects] = useState(false)
  const location = useLocation()

  // Close drawer on navigation
  useEffect(() => {
    setDrawerOpen(false)
  }, [location.pathname])

  // Close project dropdown on outside click
  useEffect(() => {
    if (!showProjects) return
    const handler = () => setShowProjects(false)
    document.addEventListener('click', handler)
    return () => document.removeEventListener('click', handler)
  }, [showProjects])

  const { data: status } = useStatus()
  const { data: healthData } = useHealth()
  const { data: projects } = useProjects()
  const switchProject = useSwitchProject()
  const logout = useLogout()

  const claudeInfo = status?.message?.match(/\u{1F916} Claude: (\d+)\/(\d+)/u)
  const claudeUsed = claudeInfo?.[1] || '0'
  const claudeMax = claudeInfo?.[2] || '3'
  const isConnected = !!healthData

  // Parse current project from status message (📌 project-id — ...)
  const currentProject = status?.message?.match(/📌 (.+?) —/u)?.[1] || 'GLOBAL'

  // Parse project list from data
  const projectList: { id: string; description: string }[] = []
  if (projects?.data) {
    const data = projects.data as any
    if (Array.isArray(data)) {
      data.forEach((p: any) => projectList.push({ id: p.id || p.ID, description: p.description || p.Description || '' }))
    } else if (data.items && Array.isArray(data.items)) {
      data.items.forEach((p: any) => projectList.push({ id: p.id || p.ID, description: p.description || p.Description || '' }))
    }
  }

  const handleSwitch = (id: string) => {
    switchProject.mutate(id)
    setShowProjects(false)
  }

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-14 items-center px-2 md:px-4 gap-2 md:gap-4">
        {/* Mobile Hamburger Menu */}
        <Sheet open={drawerOpen} onOpenChange={setDrawerOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="md:hidden min-h-[44px] min-w-[44px]">
              <Menu className="h-5 w-5" />
              <span className="sr-only">메뉴</span>
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-[260px] p-0 pt-12">
            <nav className="flex flex-col px-3">
              {/* Global Section */}
              <div className="px-3 py-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Global
              </div>
              {globalNavItems.map(({ to, icon: Icon, label }) => (
                <NavLink
                  key={to}
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
                  <span>{label}</span>
                </NavLink>
              ))}

              {/* Project Section - only show when a project is selected */}
              {currentProject !== 'GLOBAL' && (
                <>
                  {/* Separator */}
                  <div className="my-3 mx-3 h-px bg-border" />

                  <div className="px-3 py-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                    Project
                  </div>
                  {projectNavItems.map(({ to, icon: Icon, label }) => (
                    <NavLink
                      key={to}
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
                      <span>{label}</span>
                    </NavLink>
                  ))}
                </>
              )}
            </nav>
          </SheetContent>
        </Sheet>

        {/* Logo */}
        <Link to="/" className="flex items-center gap-2 font-bold text-lg hover:opacity-80 transition-opacity">
          <Layers className="h-6 w-6 text-primary" />
          <span className="hidden sm:inline">Claribot</span>
        </Link>

        {/* Project Selector */}
        <div className="relative ml-2" onClick={e => e.stopPropagation()}>
          <Button
            variant="outline"
            size="sm"
            className="gap-1 h-8"
            onClick={() => setShowProjects(!showProjects)}
          >
            <FolderOpen className="h-4 w-4" />
            <span className="hidden sm:inline max-w-[100px] truncate">{currentProject}</span>
            <ChevronDown className="h-3 w-3 shrink-0 opacity-50" />
          </Button>
          {showProjects && (
            <div className="absolute top-full mt-1 left-0 z-50 min-w-[150px] rounded-md border bg-popover p-1 shadow-md">
              <button
                className={cn(
                  "w-full text-left px-2 py-2 text-xs rounded-sm hover:bg-accent truncate",
                  currentProject === 'GLOBAL' && "bg-accent"
                )}
                onClick={() => handleSwitch('none')}
              >
                GLOBAL
              </button>
              {projectList.map(p => (
                <button
                  key={p.id}
                  className={cn(
                    "w-full text-left px-2 py-2 text-xs rounded-sm hover:bg-accent truncate",
                    currentProject === p.id && "bg-accent"
                  )}
                  onClick={() => handleSwitch(p.id)}
                  title={p.description || p.id}
                >
                  {p.id}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Global Navigation - desktop only */}
        <nav className="hidden md:flex items-center gap-1 ml-4">
          {globalNavItems.map(({ to, icon: Icon, label }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors",
                  "hover:bg-accent hover:text-accent-foreground",
                  isActive
                    ? "bg-accent text-accent-foreground"
                    : "text-muted-foreground"
                )
              }
            >
              <Icon className="h-4 w-4" />
              <span className="hidden lg:inline">{label}</span>
            </NavLink>
          ))}
        </nav>

        <div className="flex-1" />

        {/* Claude Status & Connection Status - hidden on mobile */}
        <div className="hidden md:flex items-center gap-2">
          <Badge variant={Number(claudeUsed) > 0 ? "warning" : "secondary"} className="gap-1">
            Claude {claudeUsed}/{claudeMax}
          </Badge>

          {isConnected ? (
            <Badge variant="success" className="gap-1">
              <Wifi className="h-3 w-3" /> Connected
            </Badge>
          ) : (
            <Badge variant="destructive" className="gap-1">
              <WifiOff className="h-3 w-3" /> Offline
            </Badge>
          )}
        </div>

        {/* Logout Button */}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => logout.mutate()}
          disabled={logout.isPending}
          title="로그아웃"
        >
          <LogOut className="h-4 w-4" />
          <span className="sr-only">로그아웃</span>
        </Button>
      </div>
    </header>
  )
}
