# VSCode Extension 메시지 프로토콜

> **버전**: v0.0.4

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
```

---

*Claritask VSCode Extension Spec v0.0.4 - 2026-02-03*
