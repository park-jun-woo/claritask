export interface ClaribotResponse {
  success: boolean;
  message: string;
  data?: any;
  needs_input?: boolean;
  prompt?: string;
  context?: string;
}

export interface Project {
  id: string;
  name: string;
  path: string;
  type: string;
  description: string;
  status: string;
  category: string;
  pinned: boolean;
  last_accessed: string;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: number;
  parent_id: number | null;
  title: string;
  spec: string;
  plan: string;
  report: string;
  status: 'todo' | 'split' | 'planned' | 'done' | 'failed';
  error: string;
  is_leaf: boolean;
  depth: number;
  priority: number;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: number;
  project_id: string | null;
  content: string;
  source: 'telegram' | 'cli' | 'gui' | 'schedule';
  status: 'pending' | 'processing' | 'done' | 'failed';
  result: string;
  error: string;
  created_at: string;
  completed_at: string | null;
}

export interface Schedule {
  id: number;
  project_id: string | null;
  cron_expr: string;
  message: string;
  type: 'claude' | 'bash';
  enabled: boolean;
  run_once: boolean;
  last_run: string | null;
  next_run: string | null;
  created_at: string;
  updated_at: string;
}

export interface ScheduleRun {
  id: number;
  schedule_id: number;
  status: 'running' | 'done' | 'failed';
  result: string;
  error: string;
  started_at: string;
  completed_at: string | null;
}

export interface Spec {
  id: number;
  title: string;
  content: string;
  status: 'draft' | 'review' | 'approved' | 'deprecated';
  priority: number;
  created_at: string;
  updated_at: string;
}

export interface ClaudeStatus {
  used: number;
  max: number;
  available: number;
}

export interface TaskStats {
  total: number;
  leaf: number;
  todo: number;
  planned: number;
  split: number;
  done: number;
  failed: number;
}

export interface ProjectStats {
  project_id: string;
  project_name: string;
  project_description: string;
  stats: TaskStats & {in_progress: number};
}

export interface CycleStatus {
  status: 'idle' | 'running' | 'interrupted';
  type?: string;
  project_id?: string;
  started_at?: string;
  current_task_id?: number;
  active_workers?: number;
  phase?: string;
  target_total?: number;
  completed?: number;
  elapsed_sec?: number;
}

export interface StatusResponse {
  success: boolean;
  message: string;
  data?: ClaudeStatus;
  project_id?: string;
  cycle_status: CycleStatus;
  cycle_statuses?: CycleStatus[];
  task_stats?: TaskStats;
}

export interface PaginatedList<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}
