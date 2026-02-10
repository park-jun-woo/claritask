import { Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { ProjectLayout } from '@/components/layout/ProjectLayout'
import { useAuthStatus } from '@/hooks/useAuth'
import { useStatus } from '@/hooks/useClaribot'
import Dashboard from '@/pages/Dashboard'
import Projects from '@/pages/Projects'
import ProjectEdit from '@/pages/ProjectEdit'
import Tasks from '@/pages/Tasks'
import Messages from '@/pages/Messages'
import Schedules from '@/pages/Schedules'
import Settings from '@/pages/Settings'
import Specs from '@/pages/Specs'
import Files from '@/pages/Files'
import Terminal from '@/pages/Terminal'
import Setup from '@/pages/Setup'
import Login from '@/pages/Login'
import { Loader2 } from 'lucide-react'
import { useEffect } from 'react'
import type { StatusResponse } from '@/types'

function AuthGuard({ children }: { children: React.ReactNode }) {
  const { data: auth, isLoading, isError } = useAuthStatus()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (isError) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-center space-y-2">
          <p className="text-muted-foreground">서버에 연결할 수 없습니다</p>
          <p className="text-xs text-muted-foreground">Claribot 서비스가 실행 중인지 확인하세요</p>
        </div>
      </div>
    )
  }

  if (!auth?.setup_completed) {
    return <Navigate to="/setup" replace />
  }

  if (!auth?.is_authenticated) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function LegacyRedirect({ to }: { to: string }) {
  const navigate = useNavigate()
  const { data: status } = useStatus() as { data: StatusResponse | undefined }

  useEffect(() => {
    const projectId = status?.project_id
    if (projectId && projectId !== 'GLOBAL') {
      navigate(`/projects/${projectId}/${to}`, { replace: true })
    } else {
      navigate('/projects', { replace: true })
    }
  }, [status, to, navigate])

  return (
    <div className="flex-1 flex items-center justify-center">
      <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
    </div>
  )
}

export default function App() {
  return (
    <Routes>
      <Route path="/setup" element={<Setup />} />
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <AuthGuard>
            <Layout />
          </AuthGuard>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="projects" element={<Projects />} />
        <Route path="projects/:projectId" element={<ProjectLayout />}>
          <Route index element={<Navigate to="tasks" replace />} />
          <Route path="edit" element={<ProjectEdit />} />
          <Route path="tasks" element={<Tasks />} />
          <Route path="tasks/:taskId" element={<Tasks />} />
          <Route path="specs" element={<Specs />} />
          <Route path="specs/:specId" element={<Specs />} />
          <Route path="messages" element={<Messages />} />
          <Route path="schedules" element={<Schedules />} />
          <Route path="files" element={<Files />} />
          <Route path="files/*" element={<Files />} />
          <Route path="terminal" element={<Terminal />} />
        </Route>
        <Route path="messages" element={<Messages />} />
        <Route path="schedules" element={<Schedules />} />
        <Route path="terminal" element={<Terminal />} />
        <Route path="settings" element={<Settings />} />
        {/* Legacy redirects */}
        <Route path="tasks" element={<LegacyRedirect to="tasks" />} />
        <Route path="specs" element={<LegacyRedirect to="specs" />} />
        <Route path="files" element={<LegacyRedirect to="files" />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
