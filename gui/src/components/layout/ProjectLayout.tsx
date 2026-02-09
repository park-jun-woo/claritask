import { useEffect, useRef } from 'react'
import { Outlet, useParams } from 'react-router-dom'
import { useStatus, useSwitchProject } from '@/hooks/useClaribot'
import { Loader2 } from 'lucide-react'
import type { StatusResponse } from '@/types'

export function ProjectLayout() {
  const { projectId } = useParams<{ projectId: string }>()
  const { data: status } = useStatus() as { data: StatusResponse | undefined }
  const switchProject = useSwitchProject()
  const switchedRef = useRef<string | null>(null)

  const serverProject = status?.project_id || 'GLOBAL'

  // Auto-switch project if URL project differs from server project
  useEffect(() => {
    if (!projectId || !status) return
    if (serverProject === projectId) {
      switchedRef.current = null
      return
    }
    // Avoid duplicate switch calls
    if (switchedRef.current === projectId) return
    if (switchProject.isPending) return

    switchedRef.current = projectId
    switchProject.mutate(projectId)
  }, [projectId, serverProject, status, switchProject.isPending])

  // Show loading while switching
  if (switchProject.isPending) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="flex items-center gap-2 text-muted-foreground">
          <Loader2 className="h-5 w-5 animate-spin" />
          <span className="text-sm">프로젝트 전환 중...</span>
        </div>
      </div>
    )
  }

  return <Outlet context={{ projectId }} />
}
