import type { ClaribotResponse, StatusResponse } from '@/types'

const API_BASE = '/api'

// --- Common fetch helpers ---

async function apiGet<T = ClaribotResponse>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, { credentials: 'include' })
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`)
  }
  return res.json()
}

async function apiPost<T = ClaribotResponse>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    credentials: 'include',
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`)
  }
  return res.json()
}

async function apiPatch<T = ClaribotResponse>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`)
  }
  return res.json()
}

async function apiPut<T = ClaribotResponse>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`)
  }
  return res.json()
}

async function apiDelete<T = ClaribotResponse>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'DELETE',
    credentials: 'include',
  })
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`)
  }
  return res.json()
}

// --- Health check ---

export async function health(): Promise<{ version: string; uptime: number }> {
  return apiGet<{ version: string; uptime: number }>('/health')
}

// --- Config YAML API ---

export const configYamlAPI = {
  get: () => apiGet('/config-yaml'),
  set: (content: string) => apiPut('/config-yaml', { content }),
}

// --- Usage API ---

export interface UsageData {
  success: boolean
  message: string      // stats-cache.json 기반 통계
  live?: string        // PTY /usage 결과 (rate limit)
  updated_at?: string  // live 데이터 마지막 업데이트 시간
}

export const usageAPI = {
  get: () => apiGet<UsageData>('/usage'),
  refresh: () => apiPost('/usage/refresh'),
}

// --- Project API ---

export const projectAPI = {
  list: (all = false) =>
    apiGet(`/projects${all ? '?all=true' : ''}`),
  get: (id?: string) =>
    apiGet(`/projects/${id}`),
  add: (path: string, desc?: string) =>
    apiPost('/projects', { path, description: desc }),
  create: (id: string, desc?: string) =>
    apiPost('/projects', { id, description: desc }),
  delete: (id: string) =>
    apiDelete(`/projects/${id}`),
  switch: (id: string) =>
    apiPost(`/projects/${id}/switch`),
  switchNone: () =>
    apiPost('/projects/none/switch'),
  set: (id: string, field: string, value: string) =>
    apiPatch(`/projects/${id}`, { field, value }),
  update: (id: string, data: { description?: string; parallel?: number }) =>
    apiPatch(`/projects/${id}`, data),
  stats: () =>
    apiGet('/projects/stats'),
}

// --- Task API ---

export const taskAPI = {
  list: (parentId?: number, _all = false, tree = false) => {
    const params = new URLSearchParams()
    if (tree) params.set('tree', 'true')
    if (parentId !== undefined) params.set('parent_id', String(parentId))
    if (_all) params.set('all', 'true')
    const qs = params.toString()
    return apiGet(`/tasks${qs ? '?' + qs : ''}`)
  },
  get: (id: number | string) =>
    apiGet(`/tasks/${id}`),
  add: (title: string, parentId?: number, spec?: string) =>
    apiPost('/tasks', { title, parent_id: parentId, spec }),
  set: (id: number | string, field: string, value: string) =>
    apiPatch(`/tasks/${id}`, { field, value }),
  delete: (id: number | string) =>
    apiDelete(`/tasks/${id}`),
  plan: (id: number | string) =>
    apiPost(`/tasks/${id}/plan`),
  planAll: () =>
    apiPost('/tasks/plan-all'),
  run: (id: number | string) =>
    apiPost(`/tasks/${id}/run`),
  runAll: () =>
    apiPost('/tasks/run-all'),
  cycle: () =>
    apiPost('/tasks/cycle'),
  stop: () =>
    apiPost('/tasks/stop'),
}

// --- Message API ---

export const messageAPI = {
  list: (all = false) =>
    apiGet(`/messages${all ? '?all=true' : ''}`),
  get: (id: number | string) =>
    apiGet(`/messages/${id}`),
  send: (content: string) =>
    apiPost('/messages', { content, source: 'gui' }),
  status: () =>
    apiGet('/messages/status'),
  processing: () =>
    apiGet('/messages/processing'),
}

// --- Schedule API ---

export const scheduleAPI = {
  list: (all = false) =>
    apiGet(`/schedules${all ? '?all=true' : ''}`),
  get: (id: number | string) =>
    apiGet(`/schedules/${id}`),
  add: (cronExpr: string, message: string, projectId?: string, once = false, type: 'claude' | 'bash' = 'claude') =>
    apiPost('/schedules', {
      cron_expr: cronExpr,
      message,
      project_id: projectId,
      run_once: once,
      type,
    }),
  delete: (id: number | string) =>
    apiDelete(`/schedules/${id}`),
  enable: (id: number | string) =>
    apiPost(`/schedules/${id}/enable`),
  disable: (id: number | string) =>
    apiPost(`/schedules/${id}/disable`),
  update: (id: number | string, field: string, value: string) =>
    apiPatch(`/schedules/${id}`, { field, value }),
  setProject: (id: number | string, projectId: string | null) =>
    apiPatch(`/schedules/${id}`, { field: 'project', value: projectId ?? 'none' }),
  runs: (scheduleId: number | string) =>
    apiGet(`/schedules/${scheduleId}/runs`),
  run: (runId: number | string) =>
    apiGet(`/schedule-runs/${runId}`),
}

// --- Spec API ---

export const specAPI = {
  list: (all = false) =>
    apiGet(`/specs${all ? '?all=true' : ''}`),
  get: (id: number | string) =>
    apiGet(`/specs/${id}`),
  add: (title: string, content?: string) =>
    apiPost('/specs', { title, content }),
  set: (id: number | string, field: string, value: string) =>
    apiPatch(`/specs/${id}`, { field, value }),
  delete: (id: number | string) =>
    apiDelete(`/specs/${id}`),
}

// --- Config API ---

export const configAPI = {
  list: () =>
    apiGet('/configs'),
  get: (key: string) =>
    apiGet(`/configs/${key}`),
  set: (key: string, value: string) =>
    apiPut(`/configs/${key}`, { value }),
  delete: (key: string) =>
    apiDelete(`/configs/${key}`),
}

// --- Status API ---

export const statusAPI = {
  get: () => apiGet<StatusResponse>('/status'),
}

// --- Auth API ---

export const authAPI = {
  setup: async (password: string): Promise<{ totp_uri?: string; success?: boolean }> => {
    const res = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ password }),
    })
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || 'Setup failed')
    }
    return res.json()
  },
  setupVerify: async (password: string, code: string): Promise<{ success: boolean }> => {
    const res = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ password, totp_code: code }),
    })
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || 'Verification failed')
    }
    return res.json()
  },
  totpSetup: async (): Promise<{ totp_uri: string }> => {
    const res = await fetch('/api/auth/totp-setup', { credentials: 'include' })
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || 'Failed to get TOTP setup')
    }
    return res.json()
  },
  login: async (password: string, code: string): Promise<{ success: boolean }> => {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ password, totp_code: code }),
    })
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || 'Login failed')
    }
    return res.json()
  },
  logout: async (): Promise<{ success: boolean }> => {
    const res = await fetch('/api/auth/logout', {
      method: 'POST',
      credentials: 'include',
    })
    if (!res.ok) {
      const err = await res.json()
      throw new Error(err.error || 'Logout failed')
    }
    return res.json()
  },
  status: async (): Promise<{ setup_completed: boolean; is_authenticated: boolean }> => {
    const res = await fetch('/api/auth/status', { credentials: 'include' })
    if (!res.ok) {
      throw new Error('Failed to check auth status')
    }
    return res.json()
  },
}
