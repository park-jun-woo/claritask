import type {ClaribotResponse, StatusResponse} from '../types';

let baseURL = '';

export function setBaseURL(url: string) {
  baseURL = url.replace(/\/$/, '');
}

export function getBaseURL(): string {
  return baseURL;
}

let authToken = '';

export function setAuthToken(token: string) {
  authToken = token;
}

function headers(extra?: Record<string, string>): Record<string, string> {
  const h: Record<string, string> = {'Content-Type': 'application/json'};
  if (authToken) {
    h.Authorization = `Bearer ${authToken}`;
  }
  return {...h, ...extra};
}

function handleResponse<T>(res: Response): Promise<T> {
  if (res.status === 401) {
    authToken = '';
    onAuthError?.();
    throw new Error('Unauthorized');
  }
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

async function apiGet<T = ClaribotResponse>(path: string): Promise<T> {
  const res = await fetch(`${baseURL}/api${path}`, {headers: headers()});
  return handleResponse<T>(res);
}

async function apiPost<T = ClaribotResponse>(
  path: string,
  body?: unknown,
): Promise<T> {
  const res = await fetch(`${baseURL}/api${path}`, {
    method: 'POST',
    headers: headers(),
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  return handleResponse<T>(res);
}

async function apiPatch<T = ClaribotResponse>(
  path: string,
  body: unknown,
): Promise<T> {
  const res = await fetch(`${baseURL}/api${path}`, {
    method: 'PATCH',
    headers: headers(),
    body: JSON.stringify(body),
  });
  return handleResponse<T>(res);
}

async function apiPut<T = ClaribotResponse>(
  path: string,
  body: unknown,
): Promise<T> {
  const res = await fetch(`${baseURL}/api${path}`, {
    method: 'PUT',
    headers: headers(),
    body: JSON.stringify(body),
  });
  return handleResponse<T>(res);
}

async function apiDelete<T = ClaribotResponse>(path: string): Promise<T> {
  const res = await fetch(`${baseURL}/api${path}`, {
    method: 'DELETE',
    headers: headers(),
  });
  return handleResponse<T>(res);
}

function buildQuery(params: Record<string, string | boolean | undefined>): string {
  const parts: string[] = [];
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== false) {
      parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`);
    }
  }
  return parts.length > 0 ? '?' + parts.join('&') : '';
}

// --- Health check ---
export async function health(): Promise<{version: string; uptime: number}> {
  return apiGet<{version: string; uptime: number}>('/health');
}

// --- Project API ---
export const projectAPI = {
  list: (all = false) => apiGet(`/projects${all ? '?all=true' : ''}`),
  get: (id?: string) => apiGet(`/projects/${id}`),
  switch: (id: string) => apiPost(`/projects/${id}/switch`),
  switchNone: () => apiPost('/projects/none/switch'),
  stats: () => apiGet('/projects/stats'),
  delete: (id: string) => apiDelete(`/projects/${id}`),
  update: (id: string, data: {description?: string; parallel?: number}) =>
    apiPatch(`/projects/${id}`, data),
};

// --- Task API ---
export const taskAPI = {
  list: (parentId?: number, all = false, tree = false) => {
    const qs = buildQuery({
      tree: tree || undefined,
      parent_id: parentId !== undefined ? String(parentId) : undefined,
      all: all || undefined,
    });
    return apiGet(`/tasks${qs}`);
  },
  get: (id: number | string) => apiGet(`/tasks/${id}`),
  add: (spec: string, parentId?: number) =>
    apiPost('/tasks', {spec, parent_id: parentId}),
  set: (id: number | string, field: string, value: string) =>
    apiPatch(`/tasks/${id}`, {field, value}),
  delete: (id: number | string) => apiDelete(`/tasks/${id}`),
  plan: (id: number | string) => apiPost(`/tasks/${id}/plan`),
  planAll: () => apiPost('/tasks/plan-all'),
  run: (id: number | string) => apiPost(`/tasks/${id}/run`),
  runAll: () => apiPost('/tasks/run-all'),
  cycle: (projectId?: string) =>
    apiPost('/tasks/cycle', projectId ? {project_id: projectId} : undefined),
  stop: () => apiPost('/tasks/stop'),
};

// --- Message API ---
export const messageAPI = {
  list: (all = false, projectId?: string) => {
    const qs = buildQuery({
      all: all || undefined,
      project_id: projectId,
    });
    return apiGet(`/messages${qs}`);
  },
  get: (id: number | string) => apiGet(`/messages/${id}`),
  send: (content: string, projectId?: string) =>
    apiPost('/messages', {
      content,
      source: 'mobile',
      project_id: projectId || null,
    }),
  status: () => apiGet('/messages/status'),
};

// --- Schedule API ---
export const scheduleAPI = {
  list: (all = false, projectId?: string) => {
    const qs = buildQuery({
      all: all || undefined,
      project_id: projectId,
    });
    return apiGet(`/schedules${qs}`);
  },
  get: (id: number | string) => apiGet(`/schedules/${id}`),
  add: (
    cronExpr: string,
    message: string,
    projectId?: string,
    once = false,
    type: 'claude' | 'bash' = 'claude',
  ) =>
    apiPost('/schedules', {
      cron_expr: cronExpr,
      message,
      project_id: projectId,
      run_once: once,
      type,
    }),
  delete: (id: number | string) => apiDelete(`/schedules/${id}`),
  enable: (id: number | string) => apiPost(`/schedules/${id}/enable`),
  disable: (id: number | string) => apiPost(`/schedules/${id}/disable`),
  runs: (scheduleId: number | string) =>
    apiGet(`/schedules/${scheduleId}/runs`),
};

// --- Spec API ---
export const specAPI = {
  list: (all = false) => apiGet(`/specs${all ? '?all=true' : ''}`),
  get: (id: number | string) => apiGet(`/specs/${id}`),
  add: (title: string, content?: string) =>
    apiPost('/specs', {title, content}),
  set: (id: number | string, field: string, value: string) =>
    apiPatch(`/specs/${id}`, {field, value}),
  delete: (id: number | string) => apiDelete(`/specs/${id}`),
};

// --- Status API ---
export const statusAPI = {
  get: () => apiGet<StatusResponse>('/status'),
};

// --- Auth API ---
export const authAPI = {
  login: async (
    password: string,
    code: string,
  ): Promise<{success: boolean; token?: string}> => {
    const res = await fetch(`${baseURL}/api/auth/login`, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({password, totp_code: code}),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Login failed');
    }
    return res.json();
  },
  logout: async (): Promise<{success: boolean}> => {
    const res = await fetch(`${baseURL}/api/auth/logout`, {
      method: 'POST',
      headers: headers(),
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || 'Logout failed');
    }
    return res.json();
  },
  status: async (): Promise<{
    setup_completed: boolean;
    is_authenticated: boolean;
  }> => {
    const res = await fetch(`${baseURL}/api/auth/status`, {
      headers: headers(),
    });
    if (!res.ok) {
      throw new Error('Failed to check auth status');
    }
    return res.json();
  },
};

// Callback for auth errors (401)
let onAuthError: (() => void) | null = null;
export function setOnAuthError(cb: () => void) {
  onAuthError = cb;
}

// Helper to set server URL (delegates to setBaseURL)
export function setServerUrl(url: string) {
  setBaseURL(url);
}
