# TASK-EXT-023: Expert 타입 정의 및 Store 확장

## 개요
Expert 관련 TypeScript 타입 정의 및 Zustand store 확장

## 배경
- **스펙**: specs/VSCode/12-DataTypes.md
- **현재 상태**: Expert 타입 및 store 데이터 없음

## 작업 내용

### 1. types.ts 확장
**파일**: `vscode-extension/src/types.ts`, `vscode-extension/webview-ui/src/types.ts`

```typescript
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
  contentHash: string;
  status: 'active' | 'archived';
  assigned: boolean;  // 현재 프로젝트에 할당 여부
  createdAt: string;
  updatedAt: string;
}

export interface ProjectExpert {
  projectId: string;
  expertId: string;
  assignedAt: string;
}

export interface ExpertAssignment {
  expertId: string;
  featureId: number;
  createdAt: string;
}
```

### 2. ProjectData 타입 확장
```typescript
export interface ProjectData {
  project: Project | null;
  features: Feature[];
  tasks: Task[];
  taskEdges: TaskEdge[];
  featureEdges: FeatureEdge[];
  context: Record<string, any>;
  tech: Record<string, any>;
  design: Record<string, any>;
  state: Record<string, string>;
  experts: Expert[];           // 추가
  projectExperts: string[];    // 추가: 할당된 expert ID 목록
}
```

### 3. store.ts 확장
**파일**: `vscode-extension/webview-ui/src/store.ts`

```typescript
interface AppState {
  // 기존 상태...
  experts: Expert[];
  selectedExpertId: string | null;

  // 액션 추가
  setExperts: (experts: Expert[]) => void;
  setSelectedExpertId: (id: string | null) => void;
  assignExpert: (expertId: string) => void;
  unassignExpert: (expertId: string) => void;
}
```

## 완료 기준
- [ ] Expert 인터페이스 정의
- [ ] ProjectData에 experts 필드 추가
- [ ] store에 experts 상태 및 액션 추가
