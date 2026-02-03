# VSCode Extension SQLite 동기화 전략

> **버전**: v0.0.4

## WAL 모드 활성화

```sql
PRAGMA journal_mode=WAL;
PRAGMA busy_timeout=5000;
```

---

## Polling 메커니즘

```typescript
// Extension Host
setInterval(async () => {
  const mtime = fs.statSync(dbPath).mtimeMs;
  if (mtime !== lastMtime) {
    lastMtime = mtime;
    const data = readDatabase();
    webview.postMessage({ type: 'sync', data });
  }
}, 1000);
```

---

## 충돌 방지: 낙관적 잠금

### 스키마 변경

```sql
ALTER TABLE tasks ADD COLUMN version INTEGER DEFAULT 1;
ALTER TABLE features ADD COLUMN version INTEGER DEFAULT 1;
```

### 업데이트 로직

```sql
UPDATE tasks
SET title = ?, status = ?, version = version + 1
WHERE id = ? AND version = ?;
-- affected_rows = 0 이면 충돌
```

### GUI 처리

```
저장 실패 → "외부에서 수정됨" 다이얼로그
         → [새로고침] [강제 저장] [취소]
```

---

## 실시간 동기화 기능

- CLI 변경 사항 자동 반영
- 충돌 감지 및 알림
- 수동 새로고침 버튼

---

*Claritask VSCode Extension Spec v0.0.4 - 2026-02-03*
