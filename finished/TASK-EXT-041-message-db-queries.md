# TASK-EXT-041: Message DB Queries

## 목표
VSCode Extension database.ts에 Message 관련 쿼리 메서드 추가

## 작업 내용

### 1. src/database.ts 수정

#### Message 인터페이스 추가
```typescript
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
```

#### 쿼리 메서드 추가
```typescript
getMessages(): Message[] {
  return this.queryAll<Message>(
    'SELECT * FROM messages ORDER BY created_at DESC'
  );
}

getMessage(id: number): Message | null {
  return this.queryOne<Message>(
    'SELECT * FROM messages WHERE id = ?',
    [id]
  );
}

createMessage(projectId: string, content: string, featureId?: number): number {
  const now = new Date().toISOString();
  this.run(
    `INSERT INTO messages (project_id, feature_id, content, status, created_at)
     VALUES (?, ?, ?, 'pending', ?)`,
    [projectId, featureId ?? null, content, now]
  );
  this.save();
  const row = this.queryOne<{ id: number }>('SELECT last_insert_rowid() as id');
  return row?.id ?? 0;
}

deleteMessage(id: number): void {
  // Delete message_tasks first
  this.run('DELETE FROM message_tasks WHERE message_id = ?', [id]);
  this.run('DELETE FROM messages WHERE id = ?', [id]);
  this.save();
}

getMessageTasksCount(messageId: number): number {
  const row = this.queryOne<{ count: number }>(
    'SELECT COUNT(*) as count FROM message_tasks WHERE message_id = ?',
    [messageId]
  );
  return row?.count ?? 0;
}
```

### 2. readAll() 수정
```typescript
readAll(): ProjectData {
  // ... existing code
  let messages: Message[] = [];
  try {
    messages = this.getMessages();
  } catch {
    // messages table may not exist
  }

  return {
    // ... existing fields
    messages,
  };
}
```

### 3. ProjectData 인터페이스 수정
```typescript
export interface ProjectData {
  // ... existing fields
  messages: Message[];
}
```

## 완료 조건
- [ ] Message 인터페이스 추가
- [ ] getMessages, getMessage, createMessage, deleteMessage 메서드 추가
- [ ] readAll()에 messages 포함
- [ ] ProjectData에 messages 필드 추가
