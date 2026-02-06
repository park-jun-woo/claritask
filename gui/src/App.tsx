import { Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from '@/components/layout/Layout'
import { useAuthStatus } from '@/hooks/useAuth'
import Dashboard from '@/pages/Dashboard'
import Projects from '@/pages/Projects'
import Tasks from '@/pages/Tasks'
import Messages from '@/pages/Messages'
import Schedules from '@/pages/Schedules'
import Settings from '@/pages/Settings'
import Setup from '@/pages/Setup'
import Login from '@/pages/Login'
import { Loader2 } from 'lucide-react'

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
        <Route path="tasks" element={<Tasks />} />
        <Route path="messages" element={<Messages />} />
        <Route path="schedules" element={<Schedules />} />
        <Route path="settings" element={<Settings />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
