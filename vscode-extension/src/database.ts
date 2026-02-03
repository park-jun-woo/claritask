import BetterSqlite3 from 'better-sqlite3';

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

export interface ProjectData {
  project: Project | null;
  features: Feature[];
  tasks: Task[];
  taskEdges: Edge[];
  featureEdges: Edge[];
  context: Record<string, any> | null;
  tech: Record<string, any> | null;
  design: Record<string, any> | null;
  state: Record<string, string>;
}

export class Database {
  private db: BetterSqlite3.Database;

  constructor(dbPath: string) {
    this.db = new BetterSqlite3(dbPath, { readonly: false });
    this.db.pragma('journal_mode = WAL');
    this.db.pragma('busy_timeout = 5000');
  }

  close(): void {
    this.db.close();
  }

  readAll(): ProjectData {
    return {
      project: this.getProject(),
      features: this.getFeatures(),
      tasks: this.getTasks(),
      taskEdges: this.getTaskEdges(),
      featureEdges: this.getFeatureEdges(),
      context: this.getContext(),
      tech: this.getTech(),
      design: this.getDesign(),
      state: this.getState(),
    };
  }

  getProject(): Project | null {
    const row = this.db.prepare('SELECT * FROM projects LIMIT 1').get() as Project | undefined;
    return row ?? null;
  }

  getFeatures(): Feature[] {
    return this.db.prepare('SELECT * FROM features ORDER BY id').all() as Feature[];
  }

  getTasks(): Task[] {
    return this.db.prepare('SELECT * FROM tasks ORDER BY id').all() as Task[];
  }

  getTaskEdges(): Edge[] {
    return this.db
      .prepare('SELECT from_task_id as from_id, to_task_id as to_id, created_at FROM task_edges')
      .all() as Edge[];
  }

  getFeatureEdges(): Edge[] {
    return this.db
      .prepare('SELECT from_feature_id as from_id, to_feature_id as to_id, created_at FROM feature_edges')
      .all() as Edge[];
  }

  getContext(): Record<string, any> | null {
    const row = this.db.prepare('SELECT data FROM context WHERE id = 1').get() as
      | { data: string }
      | undefined;
    return row ? JSON.parse(row.data) : null;
  }

  getTech(): Record<string, any> | null {
    const row = this.db.prepare('SELECT data FROM tech WHERE id = 1').get() as
      | { data: string }
      | undefined;
    return row ? JSON.parse(row.data) : null;
  }

  getDesign(): Record<string, any> | null {
    const row = this.db.prepare('SELECT data FROM design WHERE id = 1').get() as
      | { data: string }
      | undefined;
    return row ? JSON.parse(row.data) : null;
  }

  getState(): Record<string, string> {
    const rows = this.db.prepare('SELECT key, value FROM state').all() as {
      key: string;
      value: string;
    }[];
    const state: Record<string, string> = {};
    for (const row of rows) {
      state[row.key] = row.value;
    }
    return state;
  }

  updateTask(id: number, data: Partial<Task>, expectedVersion: number): boolean {
    const fields = Object.keys(data).filter((k) => k !== 'id' && k !== 'version');
    if (fields.length === 0) return false;

    const setClause = fields.map((f) => `${f} = ?`).join(', ');
    const values = fields.map((f) => (data as any)[f]);

    const stmt = this.db.prepare(`
      UPDATE tasks
      SET ${setClause}, version = version + 1
      WHERE id = ? AND version = ?
    `);

    const result = stmt.run(...values, id, expectedVersion);
    return result.changes > 0;
  }

  updateFeature(id: number, data: Partial<Feature>, expectedVersion: number): boolean {
    const fields = Object.keys(data).filter((k) => k !== 'id' && k !== 'version');
    if (fields.length === 0) return false;

    const setClause = fields.map((f) => `${f} = ?`).join(', ');
    const values = fields.map((f) => (data as any)[f]);

    const stmt = this.db.prepare(`
      UPDATE features
      SET ${setClause}, version = version + 1
      WHERE id = ? AND version = ?
    `);

    const result = stmt.run(...values, id, expectedVersion);
    return result.changes > 0;
  }

  addTaskEdge(fromId: number, toId: number): void {
    const now = new Date().toISOString();
    this.db
      .prepare(
        `INSERT OR IGNORE INTO task_edges (from_task_id, to_task_id, created_at) VALUES (?, ?, ?)`
      )
      .run(fromId, toId, now);
  }

  removeTaskEdge(fromId: number, toId: number): void {
    this.db
      .prepare('DELETE FROM task_edges WHERE from_task_id = ? AND to_task_id = ?')
      .run(fromId, toId);
  }

  createTask(featureId: number, title: string, content: string): number {
    const now = new Date().toISOString();
    const result = this.db
      .prepare(
        `INSERT INTO tasks (feature_id, title, content, status, created_at) VALUES (?, ?, ?, 'pending', ?)`
      )
      .run(featureId, title, content, now);
    return Number(result.lastInsertRowid);
  }

  createFeature(name: string, description: string): number {
    const project = this.getProject();
    const projectId = project?.id ?? '';
    const now = new Date().toISOString();
    const result = this.db
      .prepare(
        `INSERT INTO features (project_id, name, description, status, created_at) VALUES (?, ?, ?, 'pending', ?)`
      )
      .run(projectId, name, description, now);
    return Number(result.lastInsertRowid);
  }
}
