import { useState, useEffect } from 'react'
import { Layers, ChevronDown, Wifi, WifiOff } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useProjects, useSwitchProject, useStatus, useHealth } from '@/hooks/useClaribot'

export function Header() {
  const [showProjects, setShowProjects] = useState(false)
  const { data: projects } = useProjects()
  const { data: status } = useStatus()
  const { data: healthData } = useHealth()
  const switchProject = useSwitchProject()

  // Parse current project from status message
  const currentProject = status?.message?.match(/\u{1F4C1} \uD504\uB85C\uC81D\uD2B8: (.+)/u)?.[1] || '(none)'
  const claudeInfo = status?.message?.match(/\u{1F916} Claude: (\d+)\/(\d+)/u)
  const claudeUsed = claudeInfo?.[1] || '0'
  const claudeMax = claudeInfo?.[2] || '3'
  const isConnected = !!healthData

  const handleSwitch = (id: string) => {
    switchProject.mutate(id)
    setShowProjects(false)
  }

  // Close dropdown on outside click
  useEffect(() => {
    if (!showProjects) return
    const handler = () => setShowProjects(false)
    document.addEventListener('click', handler)
    return () => document.removeEventListener('click', handler)
  }, [showProjects])

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

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-14 items-center px-4 gap-4">
        {/* Logo */}
        <div className="flex items-center gap-2 font-bold text-lg">
          <Layers className="h-6 w-6 text-primary" />
          <span>Claribot</span>
        </div>

        {/* Project Selector */}
        <div className="relative ml-4" onClick={e => e.stopPropagation()}>
          <Button
            variant="outline"
            size="sm"
            className="gap-1 min-w-[160px] justify-between"
            onClick={() => setShowProjects(!showProjects)}
          >
            <span className="truncate">{currentProject}</span>
            <ChevronDown className="h-4 w-4 shrink-0 opacity-50" />
          </Button>
          {showProjects && (
            <div className="absolute top-full mt-1 left-0 z-50 min-w-[200px] rounded-md border bg-popover p-1 shadow-md">
              <button
                className="w-full text-left px-2 py-1.5 text-sm rounded-sm hover:bg-accent"
                onClick={() => handleSwitch('none')}
              >
                (global)
              </button>
              {projectList.map(p => (
                <button
                  key={p.id}
                  className="w-full text-left px-2 py-1.5 text-sm rounded-sm hover:bg-accent"
                  onClick={() => handleSwitch(p.id)}
                >
                  {p.id}
                  {p.description && <span className="ml-2 text-muted-foreground text-xs">{p.description}</span>}
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="flex-1" />

        {/* Claude Status */}
        <Badge variant={Number(claudeUsed) > 0 ? "warning" : "secondary"} className="gap-1">
          Claude {claudeUsed}/{claudeMax}
        </Badge>

        {/* Connection Status */}
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
    </header>
  )
}
