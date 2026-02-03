# VSCode Extension 메시지 프로토콜

> **현재 버전**: v0.0.6 ([변경이력](../HISTORY.md))

---

## Extension → Webview

```typescript
// 전체 데이터 동기화
{ type: 'sync', data: ProjectData }

// 부분 업데이트
{ type: 'update', table: 'tasks', id: 3, data: TaskData }

// 충돌 알림
{ type: 'conflict', table: 'tasks', id: 3 }

// Context/Tech/Design 저장 결과
{ type: 'settingSaveResult', section: 'context' | 'tech' | 'design', success: boolean, error?: string }

// CLI 실행 결과
{ type: 'cliResult', command: string, success: boolean, data?: any, error?: string }

// CLI 진행 상태
{ type: 'cliProgress', command: string, step: string, message: string }
```

---

## Webview → Extension

```typescript
// 데이터 저장 요청
{ type: 'save', table: 'tasks', id: 3, data: TaskData, version: 5 }

// Edge 생성
{ type: 'addEdge', fromId: 2, toId: 1 }

// 새로고침 요청
{ type: 'refresh' }

// Context 저장
{ type: 'saveContext', data: ContextData }

// Tech 저장
{ type: 'saveTech', data: TechData }

// Design 저장
{ type: 'saveDesign', data: DesignData }

// Feature 통합 생성 (CLI 호출)
{ type: 'createFeature', data: CreateFeatureData }

// FDL 검증 (CLI 호출)
{ type: 'validateFDL', featureId: number }

// Task 생성 (CLI 호출)
{ type: 'generateTasks', featureId: number }

// 스켈레톤 생성 (CLI 호출)
{ type: 'generateSkeleton', featureId: number, dryRun?: boolean }
```

---

## CreateFeatureData 타입

```typescript
interface CreateFeatureData {
  name: string;           // Feature 이름 (snake_case)
  description: string;    // 설명
  fdl?: string;           // FDL YAML (선택)
  generateTasks?: boolean;    // Task 자동 생성
  generateSkeleton?: boolean; // 스켈레톤 생성
}
```

---

*Claritask VSCode Extension Spec v0.0.6*
