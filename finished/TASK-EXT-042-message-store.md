# TASK-EXT-042: Message Store 확장

## 목표
Zustand store에 Messages 상태 관리 추가

## 작업 내용

### 1. webview-ui/src/store.ts 수정

#### Import 추가
```typescript
import type { Project, Feature, Task, Edge, Expert, Message, ProjectData } from './types';
```

#### AppState 확장
```typescript
interface AppState {
  // ... existing fields

  // Messages
  messages: Message[];
  selectedMessageId: number | null;

  // Actions
  setSelectedMessage: (id: number | null) => void;

  // Getters
  getMessage: (id: number) => Message | undefined;
  getMessagesForFeature: (featureId: number) => Message[];
}
```

#### 초기 상태 추가
```typescript
// Initial Data
messages: [],

// Initial UI State
selectedMessageId: null,
```

#### setData 수정
```typescript
setData: (data) =>
  set({
    // ... existing fields
    messages: data.messages || [],
  }),
```

#### Actions 추가
```typescript
setSelectedMessage: (id) => set({ selectedMessageId: id }),
```

#### Getters 추가
```typescript
getMessage: (id) => get().messages.find((m) => m.id === id),
getMessagesForFeature: (featureId) =>
  get().messages.filter((m) => m.feature_id === featureId),
```

## 완료 조건
- [ ] messages 상태 추가
- [ ] selectedMessageId UI 상태 추가
- [ ] setSelectedMessage action 추가
- [ ] getMessage, getMessagesForFeature getter 추가
- [ ] setData에서 messages 처리
