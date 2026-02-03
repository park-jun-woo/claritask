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

export type MessageToWebview =
  | { type: 'sync'; data: ProjectData; timestamp: number }
  | { type: 'error'; message: string }
  | { type: 'saveResult'; success: boolean; table?: string; id?: number; error?: string }
  | { type: 'conflict'; table: string; id: number }
  | { type: 'edgeResult'; success: boolean; action?: string; error?: string }
  | { type: 'createResult'; success: boolean; table?: string; id?: number; error?: string }
  | { type: 'settingSaveResult'; section: 'context' | 'tech' | 'design'; success: boolean; error?: string }
  | { type: 'expertResult'; success: boolean; action?: 'assign' | 'unassign' | 'create'; expertId?: string; error?: string };

export type MessageFromWebview =
  | { type: 'save'; table: string; id: number; data: any; version: number }
  | { type: 'refresh' }
  | { type: 'addEdge'; fromId: number; toId: number }
  | { type: 'removeEdge'; fromId: number; toId: number }
  | { type: 'createTask'; featureId: number; title: string; content: string }
  | { type: 'createFeature'; name: string; description: string }
  | { type: 'saveContext'; data: Record<string, any> }
  | { type: 'saveTech'; data: Record<string, any> }
  | { type: 'saveDesign'; data: Record<string, any> }
  | { type: 'assignExpert'; expertId: string }
  | { type: 'unassignExpert'; expertId: string }
  | { type: 'createExpert'; expertId: string }
  | { type: 'openExpertFile'; expertId: string };
