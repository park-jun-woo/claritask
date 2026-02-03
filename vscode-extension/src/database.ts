import initSqlJs, { Database as SqlJsDatabase } from 'sql.js';
import * as fs from 'fs';
import * as path from 'path';

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

let SQL: Awaited<ReturnType<typeof initSqlJs>> | null = null;
let extensionPath: string = '';

export function setExtensionPath(extPath: string): void {
  extensionPath = extPath;
}

async function getSqlJs(): Promise<typeof SQL> {
  if (!SQL) {
    const wasmPath = path.join(extensionPath, 'node_modules', 'sql.js', 'dist', 'sql-wasm.wasm');
    SQL = await initSqlJs({
      locateFile: () => wasmPath,
    });
  }
  return SQL;
}

export class Database {
  private db: SqlJsDatabase | null = null;
  private dbPath: string;

  constructor(dbPath: string) {
    this.dbPath = dbPath;
  }

  async init(): Promise<void> {
    const SqlJs = await getSqlJs();
    const buffer = fs.readFileSync(this.dbPath);
    this.db = new SqlJs!.Database(buffer);
  }

  private getDb(): SqlJsDatabase {
    if (!this.db) {
      throw new Error('Database not initialized. Call init() first.');
    }
    return this.db;
  }

  close(): void {
    if (this.db) {
      this.db.close();
      this.db = null;
    }
  }

  save(): void {
    if (this.db) {
      const data = this.db.export();
      const buffer = Buffer.from(data);
      fs.writeFileSync(this.dbPath, buffer);
    }
  }

  reload(): void {
    if (this.db) {
      const buffer = fs.readFileSync(this.dbPath);
      const SqlJs = SQL!;
      this.db.close();
      this.db = new SqlJs.Database(buffer);
    }
  }

  private queryOne<T>(sql: string, params: any[] = []): T | null {
    const db = this.getDb();
    const stmt = db.prepare(sql);
    stmt.bind(params);
    if (stmt.step()) {
      const row = stmt.getAsObject() as T;
      stmt.free();
      return row;
    }
    stmt.free();
    return null;
  }

  private queryAll<T>(sql: string, params: any[] = []): T[] {
    const db = this.getDb();
    const results: T[] = [];
    const stmt = db.prepare(sql);
    stmt.bind(params);
    while (stmt.step()) {
      results.push(stmt.getAsObject() as T);
    }
    stmt.free();
    return results;
  }

  private run(sql: string, params: any[] = []): number {
    const db = this.getDb();
    db.run(sql, params);
    return db.getRowsModified();
  }

  readAll(): ProjectData {
    const project = this.getProject();
    const projectExperts = project ? this.getProjectExperts(project.id) : [];
    const experts = this.getExperts(projectExperts);

    return {
      project,
      features: this.getFeatures(),
      tasks: this.getTasks(),
      taskEdges: this.getTaskEdges(),
      featureEdges: this.getFeatureEdges(),
      experts,
      projectExperts,
      context: this.getContext(),
      tech: this.getTech(),
      design: this.getDesign(),
      state: this.getState(),
    };
  }

  getProject(): Project | null {
    return this.queryOne<Project>('SELECT * FROM projects LIMIT 1');
  }

  getFeatures(): Feature[] {
    return this.queryAll<Feature>('SELECT * FROM features ORDER BY id');
  }

  getTasks(): Task[] {
    return this.queryAll<Task>('SELECT * FROM tasks ORDER BY id');
  }

  getTaskEdges(): Edge[] {
    return this.queryAll<Edge>(
      'SELECT from_task_id as from_id, to_task_id as to_id, created_at FROM task_edges'
    );
  }

  getFeatureEdges(): Edge[] {
    return this.queryAll<Edge>(
      'SELECT from_feature_id as from_id, to_feature_id as to_id, created_at FROM feature_edges'
    );
  }

  getContext(): Record<string, any> | null {
    const row = this.queryOne<{ data: string }>('SELECT data FROM context WHERE id = 1');
    return row ? JSON.parse(row.data) : null;
  }

  getTech(): Record<string, any> | null {
    const row = this.queryOne<{ data: string }>('SELECT data FROM tech WHERE id = 1');
    return row ? JSON.parse(row.data) : null;
  }

  getDesign(): Record<string, any> | null {
    const row = this.queryOne<{ data: string }>('SELECT data FROM design WHERE id = 1');
    return row ? JSON.parse(row.data) : null;
  }

  getState(): Record<string, string> {
    const rows = this.queryAll<{ key: string; value: string }>('SELECT key, value FROM state');
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

    const changes = this.run(
      `UPDATE tasks SET ${setClause}, version = version + 1 WHERE id = ? AND version = ?`,
      [...values, id, expectedVersion]
    );

    if (changes > 0) {
      this.save();
      return true;
    }
    return false;
  }

  updateFeature(id: number, data: Partial<Feature>, expectedVersion: number): boolean {
    const fields = Object.keys(data).filter((k) => k !== 'id' && k !== 'version');
    if (fields.length === 0) return false;

    const setClause = fields.map((f) => `${f} = ?`).join(', ');
    const values = fields.map((f) => (data as any)[f]);

    const changes = this.run(
      `UPDATE features SET ${setClause}, version = version + 1 WHERE id = ? AND version = ?`,
      [...values, id, expectedVersion]
    );

    if (changes > 0) {
      this.save();
      return true;
    }
    return false;
  }

  addTaskEdge(fromId: number, toId: number): void {
    const now = new Date().toISOString();
    this.run(
      'INSERT OR IGNORE INTO task_edges (from_task_id, to_task_id, created_at) VALUES (?, ?, ?)',
      [fromId, toId, now]
    );
    this.save();
  }

  removeTaskEdge(fromId: number, toId: number): void {
    this.run('DELETE FROM task_edges WHERE from_task_id = ? AND to_task_id = ?', [fromId, toId]);
    this.save();
  }

  createTask(featureId: number, title: string, content: string): number {
    const now = new Date().toISOString();
    this.run(
      "INSERT INTO tasks (feature_id, title, content, status, created_at) VALUES (?, ?, ?, 'pending', ?)",
      [featureId, title, content, now]
    );
    this.save();
    const row = this.queryOne<{ id: number }>('SELECT last_insert_rowid() as id');
    return row?.id ?? 0;
  }

  createFeature(name: string, description: string): number {
    const project = this.getProject();
    const projectId = project?.id ?? '';
    const now = new Date().toISOString();
    this.run(
      "INSERT INTO features (project_id, name, description, status, created_at) VALUES (?, ?, ?, 'pending', ?)",
      [projectId, name, description, now]
    );
    this.save();
    const row = this.queryOne<{ id: number }>('SELECT last_insert_rowid() as id');
    return row?.id ?? 0;
  }

  saveContext(data: Record<string, any>): void {
    const jsonData = JSON.stringify(data);
    const existing = this.queryOne<{ id: number }>('SELECT id FROM context WHERE id = 1');
    if (existing) {
      this.run('UPDATE context SET data = ? WHERE id = 1', [jsonData]);
    } else {
      this.run('INSERT INTO context (id, data) VALUES (1, ?)', [jsonData]);
    }
    this.save();
  }

  saveTech(data: Record<string, any>): void {
    const jsonData = JSON.stringify(data);
    const existing = this.queryOne<{ id: number }>('SELECT id FROM tech WHERE id = 1');
    if (existing) {
      this.run('UPDATE tech SET data = ? WHERE id = 1', [jsonData]);
    } else {
      this.run('INSERT INTO tech (id, data) VALUES (1, ?)', [jsonData]);
    }
    this.save();
  }

  saveDesign(data: Record<string, any>): void {
    const jsonData = JSON.stringify(data);
    const existing = this.queryOne<{ id: number }>('SELECT id FROM design WHERE id = 1');
    if (existing) {
      this.run('UPDATE design SET data = ? WHERE id = 1', [jsonData]);
    } else {
      this.run('INSERT INTO design (id, data) VALUES (1, ?)', [jsonData]);
    }
    this.save();
  }

  getExperts(projectExperts: string[]): Expert[] {
    const rows = this.queryAll<Omit<Expert, 'assigned'>>(`
      SELECT id, name, version, domain, language, framework,
             path, description, content, content_hash, status,
             created_at, updated_at
      FROM experts
      WHERE status = 'active'
      ORDER BY name
    `);
    return rows.map((row) => ({
      ...row,
      assigned: projectExperts.includes(row.id),
    }));
  }

  getProjectExperts(projectId: string): string[] {
    const rows = this.queryAll<{ expert_id: string }>(
      'SELECT expert_id FROM project_experts WHERE project_id = ?',
      [projectId]
    );
    return rows.map((row) => row.expert_id);
  }

  assignExpert(projectId: string, expertId: string): void {
    const now = new Date().toISOString();
    this.run(
      'INSERT OR IGNORE INTO project_experts (project_id, expert_id, assigned_at) VALUES (?, ?, ?)',
      [projectId, expertId, now]
    );
    this.save();
  }

  unassignExpert(projectId: string, expertId: string): void {
    this.run('DELETE FROM project_experts WHERE project_id = ? AND expert_id = ?', [
      projectId,
      expertId,
    ]);
    this.save();
  }

  updateExpertContent(expertId: string, content: string, contentHash: string): void {
    const now = new Date().toISOString();
    this.run('UPDATE experts SET content = ?, content_hash = ?, updated_at = ? WHERE id = ?', [
      content,
      contentHash,
      now,
      expertId,
    ]);
    this.save();
  }

  createExpert(
    expertId: string,
    name: string,
    path: string,
    content: string,
    contentHash: string
  ): void {
    const now = new Date().toISOString();
    this.run(
      `INSERT INTO experts (id, name, version, domain, language, framework, path, description, content, content_hash, status, created_at, updated_at)
       VALUES (?, ?, '1.0.0', '', '', '', ?, '', ?, ?, 'active', ?, ?)`,
      [expertId, name, path, content, contentHash, now, now]
    );
    this.save();
  }

  getExpertContentHash(expertId: string): string | null {
    const row = this.queryOne<{ content_hash: string }>(
      'SELECT content_hash FROM experts WHERE id = ?',
      [expertId]
    );
    return row?.content_hash ?? null;
  }

  getExpertContent(expertId: string): string | null {
    const row = this.queryOne<{ content: string }>('SELECT content FROM experts WHERE id = ?', [
      expertId,
    ]);
    return row?.content ?? null;
  }

  updateExpertMetadata(
    expertId: string,
    name: string,
    domain: string,
    language: string,
    framework: string
  ): void {
    this.run(
      'UPDATE experts SET name = ?, domain = ?, language = ?, framework = ? WHERE id = ?',
      [name, domain, language, framework, expertId]
    );
    this.save();
  }
}
