# VSCode Extension 데이터 타입

> **버전**: v0.0.4

## ProjectData

```typescript
interface ProjectData {
  project: Project;
  features: Feature[];
  tasks: Task[];
  taskEdges: Edge[];
  featureEdges: Edge[];
  experts: Expert[];
  context: Record<string, any>;
  tech: Record<string, any>;
  design: Record<string, any>;
  state: Record<string, string>;
  memos: Memo[];
}
```

---

## Expert

```typescript
interface Expert {
  id: string;
  name: string;
  version: string;
  domain: string;
  language: string;
  framework: string;
  content: string;
  content_hash: string;
  assigned: boolean;
  fileExists: boolean;
}
```

---

## Feature

```typescript
interface Feature {
  id: number;
  name: string;
  spec: string;
  fdl: string;
  fdl_hash: string;
  status: string;
  version: number;
}
```

---

## Task

```typescript
interface Task {
  id: number;
  feature_id: number;
  parent_id: number | null;
  title: string;
  content: string;
  level: string;
  skill: string;
  status: 'pending' | 'doing' | 'done' | 'failed';
  version: number;
}
```

---

## Edge

```typescript
interface Edge {
  from_id: number;
  to_id: number;
}
```

---

*Claritask VSCode Extension Spec v0.0.4 - 2026-02-03*
