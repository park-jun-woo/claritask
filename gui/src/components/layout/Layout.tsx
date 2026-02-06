import { useState } from 'react'
import { Outlet } from 'react-router-dom'
import { Header } from './Header'
import { Sidebar } from './Sidebar'
import { useStatus } from '@/hooks/useClaribot'

export function Layout() {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const { data: status } = useStatus()

  // Parse current project from status message (📌 project-id — ...)
  const currentProject = status?.message?.match(/📌 (.+?) —/u)?.[1] || 'GLOBAL'
  const hasProject = currentProject !== 'GLOBAL'

  return (
    <div className="flex h-screen flex-col">
      <Header />
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar - only show when a project is selected */}
        {hasProject && (
          <div className="hidden md:flex">
            <Sidebar
              collapsed={sidebarCollapsed}
              onToggle={() => setSidebarCollapsed(!sidebarCollapsed)}
            />
          </div>
        )}
        <main className="flex-1 overflow-auto p-3 sm:p-4 md:p-6 flex flex-col">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
