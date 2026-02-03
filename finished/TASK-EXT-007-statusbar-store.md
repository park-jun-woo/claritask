# TASK-EXT-007: StatusBar 및 상태 관리

## 목표
상태 바 컴포넌트 및 Zustand 상태 관리 구현.

## 파일

### 1. webview-ui/src/components/StatusBar.tsx

```typescript
import React from 'react';
import { useProjectStore } from '../stores/projectStore';

export function StatusBar() {
  const { lastSync, isConnected, tasks } = useProjectStore();

  const totalTasks = tasks.length;
  const doneTasks = tasks.filter((t) => t.status === 'done').length;
  const doingTasks = tasks.filter((t) => t.status === 'doing').length;
  const pendingTasks = tasks.filter((t) => t.status === 'pending').length;
  const failedTasks = tasks.filter((t) => t.status === 'failed').length;

  const formatTime = (timestamp: number | null): string => {
    if (!timestamp) return 'Never';
    const seconds = Math.floor((Date.now() - timestamp) / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.floor(seconds / 60);
    return `${minutes}m ago`;
  };

  return (
    <footer className="h-6 px-4 flex items-center text-xs border-t border-gray-700 bg-gray-800">
      {/* Connection Status */}
      <div className="flex items-center mr-4">
        <span
          className={`w-2 h-2 rounded-full mr-1 ${
            isConnected ? 'bg-green-500' : 'bg-red-500'
          }`}
        />
        <span>{isConnected ? 'Connected' : 'Disconnected'}</span>
      </div>

      {/* Last Sync */}
      <div className="mr-4 text-gray-400">
        Last sync: {formatTime(lastSync)}
      </div>

      {/* Task Summary */}
      <div className="flex items-center space-x-3 ml-auto">
        <span className="text-gray-400">
          Total: <span className="text-white">{totalTasks}</span>
        </span>
        <span className="text-green-400">
          Done: {doneTasks}
        </span>
        <span className="text-blue-400">
          Doing: {doingTasks}
        </span>
        <span className="text-gray-400">
          Pending: {pendingTasks}
        </span>
        {failedTasks > 0 && (
          <span className="text-red-400">
            Failed: {failedTasks}
          </span>
        )}
      </div>

      {/* WAL Mode Indicator */}
      <div className="ml-4 text-gray-500">
        WAL: ON
      </div>
    </footer>
  );
}
```

### 2. webview-ui/src/stores/projectStore.ts

```typescript
import { create } from 'zustand';

export interface Project {
  id: string;
  name: string;
  description: string;
  status: string;
}

export interface Feature {
  id: number;
  project_id: string;
  name: string;
  description: string;
  spec: string;
  fdl: string;
  status: string;
  version: number;
}

export interface Task {
  id: number;
  feature_id: number;
  status: 'pending' | 'doing' | 'done' | 'failed';
  title: string;
  content: string;
  target_file: string;
  version: number;
}

export interface Edge {
  from_id: number;
  to_id: number;
}

interface ProjectState {
  // Data
  project: Project | null;
  features: Feature[];
  tasks: Task[];
  taskEdges: Edge[];
  featureEdges: Edge[];
  context: Record<string, any> | null;
  tech: Record<string, any> | null;
  design: Record<string, any> | null;
  state: Record<string, string>;

  // UI State
  isConnected: boolean;
  lastSync: number | null;
  selectedFeatureId: number | null;
  selectedTaskId: number | null;

  // Actions
  setData: (data: Partial<ProjectState>) => void;
  setSync: (timestamp: number) => void;
  setConnected: (connected: boolean) => void;
  selectFeature: (id: number | null) => void;
  selectTask: (id: number | null) => void;
}

export const useProjectStore = create<ProjectState>((set) => ({
  // Initial Data
  project: null,
  features: [],
  tasks: [],
  taskEdges: [],
  featureEdges: [],
  context: null,
  tech: null,
  design: null,
  state: {},

  // Initial UI State
  isConnected: false,
  lastSync: null,
  selectedFeatureId: null,
  selectedTaskId: null,

  // Actions
  setData: (data) => set((state) => ({ ...state, ...data })),
  setSync: (timestamp) => set({ lastSync: timestamp, isConnected: true }),
  setConnected: (connected) => set({ isConnected: connected }),
  selectFeature: (id) => set({ selectedFeatureId: id, selectedTaskId: null }),
  selectTask: (id) => set({ selectedTaskId: id }),
}));
```

### 3. webview-ui/src/hooks/useSync.ts

```typescript
import { useEffect } from 'react';
import { useProjectStore } from '../stores/projectStore';

// VSCode API 타입
declare function acquireVsCodeApi(): {
  postMessage: (message: any) => void;
  getState: () => any;
  setState: (state: any) => void;
};

const vscode = acquireVsCodeApi();

export function useSync() {
  const { setData, setSync, setConnected } = useProjectStore();

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      const message = event.data;

      switch (message.type) {
        case 'sync':
          setData(message.data);
          setSync(message.timestamp);
          break;

        case 'error':
          console.error('Sync error:', message.message);
          setConnected(false);
          break;

        case 'conflict':
          // TODO: 충돌 처리 UI (Phase 4)
          console.warn('Conflict detected:', message.table, message.id);
          break;
      }
    };

    window.addEventListener('message', handleMessage);

    return () => {
      window.removeEventListener('message', handleMessage);
    };
  }, [setData, setSync, setConnected]);

  // 메시지 전송 헬퍼
  const postMessage = (message: any) => {
    vscode.postMessage(message);
  };

  return { postMessage };
}

// 전역 메시지 전송 함수 (컴포넌트 외부에서 사용)
export function sendMessage(message: any) {
  const vscode = acquireVsCodeApi();
  vscode.postMessage(message);
}
```

## 완료 조건
- [ ] StatusBar 컴포넌트 구현
- [ ] Zustand store 구현
- [ ] useSync 훅 구현
- [ ] 메시지 수신 처리
- [ ] 연결 상태 표시
- [ ] Task 통계 표시
