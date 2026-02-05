import type { ClaribotRequest, ClaribotResponse } from '@/types'

const API_BASE = '/api'

export async function execute(req: ClaribotRequest): Promise<ClaribotResponse> {
  const res = await fetch(API_BASE, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  return res.json()
}

// Convenience wrappers

export async function cmd(command: string, ...args: string[]): Promise<ClaribotResponse> {
  return execute({ command, args: args.length ? args : undefined })
}

// Health check
export async function health(): Promise<{ version: string; uptime: number }> {
  const res = await fetch('/api/health')
  return res.json()
}

// --- Project API ---
export const projectAPI = {
  list: (all = false) => cmd('project', 'list', ...(all ? ['--all'] : [])),
  get: (id?: string) => cmd('project', 'get', ...(id ? [id] : [])),
  add: (path: string, type?: string, desc?: string) =>
    cmd('project', 'add', path, ...(type ? [type] : []), ...(desc ? [desc] : [])),
  create: (id: string, type?: string, desc?: string) =>
    cmd('project', 'create', id, ...(type ? [type] : []), ...(desc ? [desc] : [])),
  delete: (id: string, confirmed = false) =>
    cmd('project', 'delete', id, ...(confirmed ? ['yes'] : [])),
  switch: (id: string) => cmd('project', 'switch', id),
  switchNone: () => cmd('project', 'switch', 'none'),
}

// --- Task API ---
export const taskAPI = {
  list: (parentId?: number, all = false) =>
    cmd('task', 'list', ...(parentId !== undefined ? [String(parentId)] : []), ...(all ? ['--all'] : [])),
  get: (id: number | string) => cmd('task', 'get', String(id)),
  add: (title: string, parentId?: number, spec?: string) =>
    cmd('task', 'add', title, ...(parentId !== undefined ? ['--parent', String(parentId)] : []), ...(spec ? ['--spec', spec] : [])),
  set: (id: number | string, field: string, value: string) =>
    cmd('task', 'set', String(id), field, value),
  delete: (id: number | string, confirmed = false) =>
    cmd('task', 'delete', String(id), ...(confirmed ? ['yes'] : [])),
  plan: (id?: number | string) =>
    id !== undefined ? cmd('task', 'plan', String(id)) : cmd('task', 'plan', '--all'),
  planAll: () => cmd('task', 'plan', '--all'),
  run: (id?: number | string) =>
    id !== undefined ? cmd('task', 'run', String(id)) : cmd('task', 'run', '--all'),
  runAll: () => cmd('task', 'run', '--all'),
  cycle: () => cmd('task', 'cycle'),
}

// --- Edge API ---
export const edgeAPI = {
  list: (taskId?: number | string, all = false) =>
    cmd('edge', 'list', ...(taskId !== undefined ? [String(taskId)] : []), ...(all ? ['--all'] : [])),
  add: (fromId: number | string, toId: number | string) =>
    cmd('edge', 'add', String(fromId), String(toId)),
  get: (fromId: number | string, toId: number | string) =>
    cmd('edge', 'get', String(fromId), String(toId)),
  delete: (fromId: number | string, toId: number | string, confirmed = false) =>
    cmd('edge', 'delete', String(fromId), String(toId), ...(confirmed ? ['yes'] : [])),
}

// --- Message API ---
export const messageAPI = {
  list: (all = false) => cmd('message', 'list', ...(all ? ['--all'] : [])),
  get: (id: number | string) => cmd('message', 'get', String(id)),
  send: (content: string) => cmd('message', 'send', content),
  status: () => cmd('message', 'status'),
  processing: () => cmd('message', 'processing'),
}

// --- Schedule API ---
export const scheduleAPI = {
  list: (all = false) => cmd('schedule', 'list', ...(all ? ['--all'] : [])),
  get: (id: number | string) => cmd('schedule', 'get', String(id)),
  add: (cronExpr: string, message: string, projectId?: string, once = false) =>
    cmd('schedule', 'add', cronExpr, message, ...(projectId ? ['--project', projectId] : []), ...(once ? ['--once'] : [])),
  delete: (id: number | string, confirmed = false) =>
    cmd('schedule', 'delete', String(id), ...(confirmed ? ['yes'] : [])),
  enable: (id: number | string) => cmd('schedule', 'enable', String(id)),
  disable: (id: number | string) => cmd('schedule', 'disable', String(id)),
  setProject: (id: number | string, projectId: string | null) =>
    cmd('schedule', 'set', String(id), 'project', projectId ?? 'none'),
  runs: (scheduleId: number | string) => cmd('schedule', 'runs', String(scheduleId)),
  run: (runId: number | string) => cmd('schedule', 'run', String(runId)),
}

// --- Status API ---
export const statusAPI = {
  get: () => cmd('status'),
}
