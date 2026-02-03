# TASK-EXT-040: Message Types 정의

## 목표
VSCode Extension webview에서 사용할 Message 관련 타입 정의

## 작업 내용

### 1. webview-ui/src/types.ts 수정

```typescript
// Message 인터페이스 추가
export interface Message {
  id: number;
  project_id: string;
  feature_id: number | null;
  content: string;
  response: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  error: string;
  created_at: string;
  completed_at: string | null;
}

export interface MessageListItem {
  id: number;
  content: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  feature_id: number | null;
  tasks_count: number;
  created_at: string;
}
```

### 2. ProjectData 확장
```typescript
export interface ProjectData {
  // ... existing fields
  messages: Message[];
}
```

### 3. MessageToWebview 확장
```typescript
| { type: 'messageResult'; success: boolean; action?: 'send' | 'delete'; messageId?: number; error?: string }
```

### 4. MessageFromWebview 확장
```typescript
| { type: 'sendMessage'; content: string; featureId?: number }
| { type: 'deleteMessage'; messageId: number }
| { type: 'getMessageDetail'; messageId: number }
```

## 완료 조건
- [ ] Message, MessageListItem 타입 추가
- [ ] ProjectData에 messages 필드 추가
- [ ] 메시지 프로토콜 타입 추가
