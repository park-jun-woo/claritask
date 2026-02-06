import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useHealth } from '@/hooks/useClaribot'
import { configYamlAPI } from '@/api/client'
import { Server, Bot, FileText, Save, Settings2, MessageSquare, FolderOpen, List, ScrollText } from 'lucide-react'
import YAML from 'yaml'

interface ConfigData {
  service: {
    host: string
    port: number
  }
  telegram: {
    token: string
    allowed_users: number[]
    admin_chat_id: number
  }
  claude: {
    timeout: number
    max_timeout: number
    max: number
  }
  project: {
    path: string
  }
  pagination: {
    page_size: number
  }
  log: {
    level: string
    file: string
  }
}

const defaultConfig: ConfigData = {
  service: { host: '127.0.0.1', port: 9847 },
  telegram: { token: '', allowed_users: [], admin_chat_id: 0 },
  claude: { timeout: 1200, max_timeout: 1800, max: 10 },
  project: { path: '' },
  pagination: { page_size: 10 },
  log: { level: 'info', file: '' },
}

export default function Settings() {
  const { data: healthData, isError: isHealthError } = useHealth()

  const [config, setConfig] = useState<ConfigData>(defaultConfig)
  const [configLoading, setConfigLoading] = useState(true)
  const [configSaving, setConfigSaving] = useState(false)
  const [configError, setConfigError] = useState('')
  const [configSuccess, setConfigSuccess] = useState('')

  const version = healthData?.version || 'unknown'
  const uptime = healthData?.uptime ? formatUptime(healthData.uptime) : 'N/A'

  // Load config.yaml on mount
  useEffect(() => {
    loadConfig()
  }, [])

  const loadConfig = async () => {
    setConfigLoading(true)
    setConfigError('')
    try {
      const res = await configYamlAPI.get()
      const yamlContent = (res.data as string) || ''
      if (yamlContent) {
        const parsed = YAML.parse(yamlContent) as Partial<ConfigData>
        setConfig({
          service: { ...defaultConfig.service, ...parsed.service },
          telegram: { ...defaultConfig.telegram, ...parsed.telegram },
          claude: { ...defaultConfig.claude, ...parsed.claude },
          project: { ...defaultConfig.project, ...parsed.project },
          pagination: { ...defaultConfig.pagination, ...parsed.pagination },
          log: { ...defaultConfig.log, ...parsed.log },
        })
      }
    } catch (e) {
      setConfigError('Failed to load config')
    } finally {
      setConfigLoading(false)
    }
  }

  const saveConfig = async () => {
    setConfigSaving(true)
    setConfigError('')
    setConfigSuccess('')
    try {
      // Build clean config object (omit empty/default values for cleaner YAML)
      const cleanConfig: Record<string, unknown> = {}

      // Service
      if (config.service.host !== defaultConfig.service.host || config.service.port !== defaultConfig.service.port) {
        cleanConfig.service = {}
        if (config.service.host !== defaultConfig.service.host) (cleanConfig.service as Record<string, unknown>).host = config.service.host
        if (config.service.port !== defaultConfig.service.port) (cleanConfig.service as Record<string, unknown>).port = config.service.port
      }

      // Telegram
      if (config.telegram.token || config.telegram.allowed_users.length > 0 || config.telegram.admin_chat_id) {
        cleanConfig.telegram = {}
        if (config.telegram.token) (cleanConfig.telegram as Record<string, unknown>).token = config.telegram.token
        if (config.telegram.allowed_users.length > 0) (cleanConfig.telegram as Record<string, unknown>).allowed_users = config.telegram.allowed_users
        if (config.telegram.admin_chat_id) (cleanConfig.telegram as Record<string, unknown>).admin_chat_id = config.telegram.admin_chat_id
      }

      // Claude
      cleanConfig.claude = {}
      if (config.claude.timeout !== defaultConfig.claude.timeout) (cleanConfig.claude as Record<string, unknown>).timeout = config.claude.timeout
      if (config.claude.max_timeout !== defaultConfig.claude.max_timeout) (cleanConfig.claude as Record<string, unknown>).max_timeout = config.claude.max_timeout
      if (config.claude.max !== defaultConfig.claude.max) (cleanConfig.claude as Record<string, unknown>).max = config.claude.max
      if (Object.keys(cleanConfig.claude as object).length === 0) delete cleanConfig.claude

      // Project
      if (config.project.path) {
        cleanConfig.project = { path: config.project.path }
      }

      // Pagination
      if (config.pagination.page_size !== defaultConfig.pagination.page_size) {
        cleanConfig.pagination = { page_size: config.pagination.page_size }
      }

      // Log
      if (config.log.level !== defaultConfig.log.level || config.log.file) {
        cleanConfig.log = {}
        if (config.log.level !== defaultConfig.log.level) (cleanConfig.log as Record<string, unknown>).level = config.log.level
        if (config.log.file) (cleanConfig.log as Record<string, unknown>).file = config.log.file
      }

      const yamlContent = YAML.stringify(cleanConfig)
      const res = await configYamlAPI.set(yamlContent)
      setConfigSuccess(res.message || 'Saved')
      setTimeout(() => setConfigSuccess(''), 5000)
    } catch (e: unknown) {
      setConfigError((e as Error).message || 'Failed to save')
    } finally {
      setConfigSaving(false)
    }
  }

  const updateConfig = <K extends keyof ConfigData>(
    section: K,
    field: keyof ConfigData[K],
    value: unknown
  ) => {
    setConfig(prev => ({
      ...prev,
      [section]: {
        ...prev[section],
        [field]: value,
      },
    }))
  }

  return (
    <div className="space-y-6 max-w-3xl">
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
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Service</span>
            {isHealthError ? (
              <Badge variant="destructive">Offline</Badge>
            ) : (
              <Badge variant="success">Connected</Badge>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Config Editor - Structured */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <FileText className="h-5 w-5" /> Config (config.yaml)
          </CardTitle>
          <CardDescription>
            Edit ~/.claribot/config.yaml. Restart claribot to apply changes.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {configLoading ? (
            <div className="text-sm text-muted-foreground">Loading...</div>
          ) : (
            <>
              {/* Service Section */}
              <ConfigSection icon={Settings2} title="Service">
                <ConfigField label="Host" hint="default: 127.0.0.1">
                  <Input
                    value={config.service.host}
                    onChange={e => updateConfig('service', 'host', e.target.value)}
                    placeholder="127.0.0.1"
                  />
                </ConfigField>
                <ConfigField label="Port" hint="default: 9847">
                  <Input
                    type="number"
                    value={config.service.port}
                    onChange={e => updateConfig('service', 'port', parseInt(e.target.value) || 0)}
                    placeholder="9847"
                  />
                </ConfigField>
              </ConfigSection>

              {/* Telegram Section */}
              <ConfigSection icon={MessageSquare} title="Telegram">
                <ConfigField label="Bot Token" hint="required for telegram bot">
                  <Input
                    value={config.telegram.token}
                    onChange={e => updateConfig('telegram', 'token', e.target.value)}
                    placeholder="123456789:ABC..."
                    type="password"
                  />
                </ConfigField>
                <ConfigField label="Admin Chat ID" hint="for schedule notifications">
                  <Input
                    type="number"
                    value={config.telegram.admin_chat_id || ''}
                    onChange={e => updateConfig('telegram', 'admin_chat_id', parseInt(e.target.value) || 0)}
                    placeholder="123456789"
                  />
                </ConfigField>
                <ConfigField label="Allowed Users" hint="comma-separated IDs (empty = allow all)">
                  <Input
                    value={config.telegram.allowed_users.join(', ')}
                    onChange={e => {
                      const ids = e.target.value
                        .split(',')
                        .map(s => parseInt(s.trim()))
                        .filter(n => !isNaN(n) && n > 0)
                      updateConfig('telegram', 'allowed_users', ids)
                    }}
                    placeholder="123, 456, 789"
                  />
                </ConfigField>
              </ConfigSection>

              {/* Claude Section */}
              <ConfigSection icon={Bot} title="Claude">
                <ConfigField label="Max Concurrent" hint="1-10, default: 10">
                  <Input
                    type="number"
                    min={1}
                    max={10}
                    value={config.claude.max}
                    onChange={e => updateConfig('claude', 'max', parseInt(e.target.value) || 1)}
                  />
                </ConfigField>
                <ConfigField label="Idle Timeout (sec)" hint="default: 1200 (20min)">
                  <Input
                    type="number"
                    min={60}
                    value={config.claude.timeout}
                    onChange={e => updateConfig('claude', 'timeout', parseInt(e.target.value) || 60)}
                  />
                </ConfigField>
                <ConfigField label="Max Timeout (sec)" hint="60-7200, default: 1800 (30min)">
                  <Input
                    type="number"
                    min={60}
                    max={7200}
                    value={config.claude.max_timeout}
                    onChange={e => updateConfig('claude', 'max_timeout', parseInt(e.target.value) || 60)}
                  />
                </ConfigField>
              </ConfigSection>

              {/* Project Section */}
              <ConfigSection icon={FolderOpen} title="Project">
                <ConfigField label="Default Path" hint="default project creation path">
                  <Input
                    value={config.project.path}
                    onChange={e => updateConfig('project', 'path', e.target.value)}
                    placeholder="/home/user/projects"
                  />
                </ConfigField>
              </ConfigSection>

              {/* Pagination Section */}
              <ConfigSection icon={List} title="Pagination">
                <ConfigField label="Page Size" hint="1-100, default: 10">
                  <Input
                    type="number"
                    min={1}
                    max={100}
                    value={config.pagination.page_size}
                    onChange={e => updateConfig('pagination', 'page_size', parseInt(e.target.value) || 10)}
                  />
                </ConfigField>
              </ConfigSection>

              {/* Log Section */}
              <ConfigSection icon={ScrollText} title="Log">
                <ConfigField label="Level" hint="debug, info, warn, error">
                  <select
                    value={config.log.level}
                    onChange={e => updateConfig('log', 'level', e.target.value)}
                    className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  >
                    <option value="debug">debug</option>
                    <option value="info">info</option>
                    <option value="warn">warn</option>
                    <option value="error">error</option>
                  </select>
                </ConfigField>
                <ConfigField label="File" hint="log file path (empty = stdout only)">
                  <Input
                    value={config.log.file}
                    onChange={e => updateConfig('log', 'file', e.target.value)}
                    placeholder="~/.claribot/claribot.log"
                  />
                </ConfigField>
              </ConfigSection>
            </>
          )}
          {configError && (
            <p className="text-sm text-destructive">{configError}</p>
          )}
          {configSuccess && (
            <p className="text-sm text-green-600">{configSuccess}</p>
          )}
        </CardContent>
        <CardFooter>
          <Button onClick={saveConfig} disabled={configSaving || configLoading}>
            <Save className="h-4 w-4 mr-2" />
            {configSaving ? 'Saving...' : 'Save Config'}
          </Button>
        </CardFooter>
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

interface ConfigSectionProps {
  icon: React.ComponentType<{ className?: string }>
  title: string
  children: React.ReactNode
}

function ConfigSection({ icon: Icon, title, children }: ConfigSectionProps) {
  return (
    <div className="space-y-3">
      <h3 className="text-sm font-medium flex items-center gap-2 text-muted-foreground">
        <Icon className="h-4 w-4" />
        {title}
      </h3>
      <div className="grid gap-3 pl-6">
        {children}
      </div>
    </div>
  )
}

interface ConfigFieldProps {
  label: string
  hint?: string
  children: React.ReactNode
}

function ConfigField({ label, hint, children }: ConfigFieldProps) {
  return (
    <div className="grid gap-1.5">
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium">{label}</label>
        {hint && <span className="text-xs text-muted-foreground">{hint}</span>}
      </div>
      {children}
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
