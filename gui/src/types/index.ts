// API Response type matching Go types.Result

export interface ClaribotResponse {
  success: boolean
  message: string
  data?: any
  needs_input?: boolean
  prompt?: string
  context?: string
}

// Project
export interface Project {
  id: string
  name: string
  path: string
  type: string
  description: string
  status: string
  created_at: string
  updated_at: string
}

// Task
export interface Task {
  id: number
  parent_id: number | null
  title: string
  spec: string
  plan: string
  report: string
  status: 'todo' | 'split' | 'planned' | 'done' | 'failed'
  error: string
  is_leaf: boolean
  depth: number
  created_at: string
  updated_at: string
}

// Message
export interface Message {
  id: number
  project_id: string | null
  content: string
  source: 'telegram' | 'cli' | 'schedule'
  status: 'pending' | 'processing' | 'done' | 'failed'
  result: string
  error: string
  created_at: string
  completed_at: string | null
}

// Schedule
export interface Schedule {
  id: number
  project_id: string | null
  cron_expr: string
  message: string
  type: 'claude' | 'bash'
  enabled: boolean
  run_once: boolean
  last_run: string | null
  next_run: string | null
  created_at: string
  updated_at: string
}

// Schedule Run
export interface ScheduleRun {
  id: number
  schedule_id: number
  status: 'running' | 'done' | 'failed'
  result: string
  error: string
  started_at: string
  completed_at: string | null
}

// Claude Status
export interface ClaudeStatus {
  used: number
  max: number
  available: number
}

// Task Stats
export interface TaskStats {
  total: number
  leaf: number
  todo: number
  planned: number
  split: number
  done: number
  failed: number
}

// Project Stats (from /api/projects/stats)
export interface ProjectStats {
  project_id: string
  project_name: string
  stats: TaskStats & { in_progress: number }
}

// Pagination
export interface PaginatedList<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}
