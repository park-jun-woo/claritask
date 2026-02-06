import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useStatus, useHealth } from '@/hooks/useClaribot'
import { Server, Bot, Wifi, Database } from 'lucide-react'

export default function Settings() {
  const { data: status } = useStatus()
  const { data: healthData, isError: isHealthError } = useHealth()

  const claudeMatch = status?.message?.match(/Claude: (\d+)\/(\d+)/)
  const claudeUsed = claudeMatch?.[1] || '0'
  const claudeMax = claudeMatch?.[2] || '3'

  const version = healthData?.version || 'unknown'
  const uptime = healthData?.uptime ? formatUptime(healthData.uptime) : 'N/A'

  return (
    <div className="space-y-6 max-w-2xl">
      <h1 className="text-2xl md:text-3xl font-bold">Settings</h1>

      {/* System Info */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Server className="h-5 w-5" /> System Info
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <InfoRow label="Claribot Version" value={`v${version}`} />
          <InfoRow label="Uptime" value={uptime} />
          <InfoRow label="DB Path" value="~/.claribot/db.clt" />
        </CardContent>
      </Card>

      {/* Claude Code */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Bot className="h-5 w-5" /> Claude Code
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <InfoRow label="Max Concurrent" value={claudeMax} />
          <InfoRow label="Currently Used" value={claudeUsed} />
          <InfoRow label="Available" value={String(Number(claudeMax) - Number(claudeUsed))} />
        </CardContent>
      </Card>

      {/* Connection Status */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Wifi className="h-5 w-5" /> Connection Status
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm">Service</span>
            {isHealthError ? (
              <Badge variant="destructive">Offline</Badge>
            ) : (
              <Badge variant="success">Connected (127.0.0.1:9847)</Badge>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="text-sm font-mono">{value}</span>
    </div>
  )
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  const parts: string[] = []
  if (days > 0) parts.push(`${days}d`)
  if (hours > 0) parts.push(`${hours}h`)
  parts.push(`${mins}m`)
  return parts.join(' ')
}
