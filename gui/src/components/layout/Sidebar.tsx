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
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { useStatus } from '@/hooks/useClaribot'

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
  const { data: status } = useStatus()

  // Parse current project from status message (📌 project-id — ...)
  const currentProject = status?.message?.match(/📌 (.+?) —/u)?.[1] || 'GLOBAL'

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
