// VSCode API types are from @types/vscode

export interface ProjectData {
  project: Project | null;
  features: Feature[];
  tasks: Task[];
  taskEdges: Edge[];
  featureEdges: Edge[];
  experts: Expert[];
  projectExperts: string[];
  context: Record<string, any> | null;
  tech: Record<string, any> | null;
  design: Record<string, any> | null;
  state: Record<string, string>;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  status: string;
  created_at: string;
}

export interface Feature {
  id: number;
  project_id: string;
  name: string;
  description: string;
  spec: string;
  fdl: string;
  fdl_hash: string;
  skeleton_generated: number;
  status: string;
  version: number;
  created_at: string;
}

export interface Task {
  id: number;
  feature_id: number;
  skeleton_id: number | null;
  status: 'pending' | 'doing' | 'done' | 'failed';
  title: string;
  content: string;
  target_file: string;
  target_line: number | null;
  target_function: string;
  result: string;
  error: string;
  version: number;
  created_at: string;
  started_at: string | null;
  completed_at: string | null;
  failed_at: string | null;
}

export interface Edge {
  from_id: number;
  to_id: number;
  created_at: string;
}

export interface Expert {
  id: string;
  name: string;
  version: string;
  domain: string;
  language: string;
  framework: string;
  path: string;
  description: string;
  content: string;
  content_hash: string;
  status: 'active' | 'archived';
  assigned: boolean;
  created_at: string;
  updated_at: string;
}

// Webview Messages: Extension → Webview
export interface SyncMessage {
  type: 'sync';
  data: ProjectData;
  timestamp: number;
}

export interface ConflictMessage {
  type: 'conflict';
  table: 'tasks' | 'features';
  id: number;
}

export interface ErrorMessage {
  type: 'error';
  message: string;
}

export interface SaveResultMessage {
  type: 'saveResult';
  success: boolean;
  table?: string;
  id?: number;
  error?: string;
}

export interface EdgeResultMessage {
  type: 'edgeResult';
  success: boolean;
  action?: 'add' | 'remove';
  error?: string;
}

export interface CreateResultMessage {
  type: 'createResult';
  success: boolean;
  table?: string;
  id?: number;
  error?: string;
}

export interface ExpertResultMessage {
  type: 'expertResult';
  success: boolean;
  action?: 'assign' | 'unassign' | 'create';
  expertId?: string;
  error?: string;
}

export type ExtensionMessage =
  | SyncMessage
  | ConflictMessage
  | ErrorMessage
  | SaveResultMessage
  | EdgeResultMessage
  | CreateResultMessage
  | ExpertResultMessage;

// Webview Messages: Webview → Extension
export interface SaveMessage {
  type: 'save';
  table: 'tasks' | 'features';
  id: number;
  data: Record<string, any>;
  version: number;
}

export interface RefreshMessage {
  type: 'refresh';
}

export interface AddEdgeMessage {
  type: 'addEdge';
  fromId: number;
  toId: number;
}

export interface RemoveEdgeMessage {
  type: 'removeEdge';
  fromId: number;
  toId: number;
}

export interface CreateTaskMessage {
  type: 'createTask';
  featureId: number;
  title: string;
  content: string;
}

export interface CreateFeatureMessage {
  type: 'createFeature';
  name: string;
  description: string;
}

export interface AssignExpertMessage {
  type: 'assignExpert';
  expertId: string;
}

export interface UnassignExpertMessage {
  type: 'unassignExpert';
  expertId: string;
}

export interface CreateExpertMessage {
  type: 'createExpert';
  expertId: string;
}

export interface OpenExpertFileMessage {
  type: 'openExpertFile';
  expertId: string;
}

export type WebviewMessage =
  | SaveMessage
  | RefreshMessage
  | AddEdgeMessage
  | RemoveEdgeMessage
  | CreateTaskMessage
  | CreateFeatureMessage
  | AssignExpertMessage
  | UnassignExpertMessage
  | CreateExpertMessage
  | OpenExpertFileMessage;
