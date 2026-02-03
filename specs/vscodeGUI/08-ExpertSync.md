# VSCode Extension Expert 파일 동기화

> **버전**: v0.0.4

## 자동 활성화

Extension은 `.claritask/db.clt` 파일이 있는 워크스페이스에서 자동 활성화:

```json
// package.json
{
  "activationEvents": [
    "workspaceContains:.claritask/db.clt"
  ]
}
```

---

## FileSystemWatcher 설정

Extension 활성화 시 Expert 파일 감시 시작 (db.clt를 열지 않아도 동작):

```typescript
// extension.ts
export function activate(context: vscode.ExtensionContext) {
    // Expert 파일 감시
    const expertWatcher = vscode.workspace.createFileSystemWatcher(
        '**/.claritask/experts/**/EXPERT.md'
    );

    // 파일 수정 시 → DB에 백업
    expertWatcher.onDidChange(uri => {
        syncExpertToDB(uri);
    });

    // 파일 생성 시 → DB에 등록
    expertWatcher.onDidCreate(uri => {
        syncExpertToDB(uri);
    });

    // 파일 삭제 시 → DB 백업에서 복구
    expertWatcher.onDidDelete(uri => {
        restoreExpertFromDB(uri);
    });

    context.subscriptions.push(expertWatcher);
}
```

---

## 동기화 흐름

```
┌─────────────────────────────────────────────────────────┐
│ 평소 동작                                               │
├─────────────────────────────────────────────────────────┤
│ EXPERT.md 수정 감지                                     │
│     ↓                                                   │
│ 파일 읽기 → 해시 계산                                   │
│     ↓                                                   │
│ 해시 변경됨? → DB content 컬럼에 백업                   │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ 파일 삭제 감지 시                                       │
├─────────────────────────────────────────────────────────┤
│ DB에서 백업된 content 조회                              │
│     ↓                                                   │
│ content 있음? → 파일 자동 복구 (조용히)                 │
│     ↓                                                   │
│ content 없음? → 템플릿으로 재생성                       │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ UI에서 삭제 버튼 클릭 시                                │
├─────────────────────────────────────────────────────────┤
│ DB 레코드 삭제 + .md 파일 삭제 (폴더 포함)              │
└─────────────────────────────────────────────────────────┘
```

---

## DB 스키마 (Expert 백업용)

```sql
CREATE TABLE IF NOT EXISTS experts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT DEFAULT '1.0.0',
    domain TEXT DEFAULT '',
    language TEXT DEFAULT '',
    framework TEXT DEFAULT '',
    path TEXT NOT NULL,
    content TEXT DEFAULT '',       -- 백업용: EXPERT.md 전체 내용
    content_hash TEXT DEFAULT '',  -- 변경 감지용 해시
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL       -- 마지막 동기화 시간
);
```

---

## 메시지 프로토콜 (Expert 관련)

### Webview → Extension

```typescript
// Expert 파일 열기 요청
{ type: 'openExpertFile', expertId: string }

// Expert 할당
{ type: 'assignExpert', expertId: string }

// Expert 할당 해제
{ type: 'unassignExpert', expertId: string }

// Expert 생성
{ type: 'createExpert', expertId: string }

// Expert 삭제
{ type: 'deleteExpert', expertId: string }
```

### Extension → Webview

```typescript
// Expert 목록 업데이트
{ type: 'expertsUpdated', experts: Expert[] }

// Expert 파일 열기 결과
{ type: 'expertFileOpened', expertId: string, success: boolean }
```

---

*Claritask VSCode Extension Spec v0.0.4 - 2026-02-03*
