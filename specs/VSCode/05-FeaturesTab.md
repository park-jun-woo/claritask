# VSCode Extension Features 탭

> **현재 버전**: v0.0.6 ([변경이력](../HISTORY.md))

---

## Features 탭 레이아웃

```
┌─────────────────────────────────────────────────────────┐
│  [Project]  [Features]  [Tasks]                         │
├────────────┬────────────────────────────────────────────┤
│            │                                            │
│  Features  │      Feature Detail                        │
│  ──────    │      ──────────────                        │
│  ▸ user_auth │    Name: user_auth                       │
│  ▸ blog_post │    Status: active                        │
│  + Add...    │    Description: ...                      │
│              │                                          │
│              │    ┌─ Spec (Rendered) ────────────────┐  │
│              │    │ # User Authentication            │  │
│              │    │ 사용자 인증 시스템 구현...        │  │
│              │    │          [Edit] [Open File]      │  │
│              │    └──────────────────────────────────┘  │
│              │                                          │
│              │    ┌─ FDL ────────────────────────────┐  │
│              │    │ feature: user_auth               │  │
│              │    │ ...                              │  │
│              │    └──────────────────────────────────┘  │
├────────────┴────────────────────────────────────────────┤
│  Status: Connected │ Last sync: 2s ago │ WAL mode: ON   │
└─────────────────────────────────────────────────────────┘
```

---

## Feature 관리 기능

- Feature 목록 트리 뷰
- Feature 추가/삭제/편집
- Feature 스펙 (Markdown) 렌더링 표시
- FDL 코드 편집 (코드 에디터 내장)

---

## Feature 생성 다이얼로그

`[+ Add...]` 버튼 클릭 시 Feature 생성 다이얼로그 표시:

```
┌─────────────────────────────────────────────────────────┐
│  Create New Feature                              [×]   │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Name:        [user_auth________________]               │
│  Description: [사용자 인증 시스템________]               │
│                                                         │
│  ┌─ FDL (선택) ──────────────────────────────────────┐  │
│  │ feature: user_auth                                │  │
│  │ version: 1.0.0                                    │  │
│  │ layers:                                           │  │
│  │   data:                                           │  │
│  │     ...                                           │  │
│  └───────────────────────────────────────────────────┘  │
│                                                         │
│  ☑ Task 자동 생성 (FDL 필수)                            │
│  ☐ 스켈레톤 코드 생성                                   │
│                                                         │
│  [FDL 검증]                    [Cancel] [Create]       │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### 동작 흐름

1. 사용자가 Name, Description 입력
2. (선택) FDL YAML 입력
3. [FDL 검증] 버튼 → `clari fdl validate` 호출
4. [Create] 버튼 → `clari feature create` 호출

### CLI 호출

```typescript
// CreateFeatureDialog.tsx
const handleCreate = async () => {
  const result = await vscode.postMessage({
    type: 'createFeature',
    data: {
      name: featureName,
      description: description,
      fdl: fdlContent || undefined,
      generateTasks: generateTasks,
      generateSkeleton: generateSkeleton
    }
  });
};
```

### 응답 처리

```typescript
// 성공 시
{ success: true, feature_id: 1, tasks_created: 5, ... }
→ Feature 목록 새로고침, 생성된 Feature 선택

// FDL 검증 실패 시
{ success: false, fdl_valid: false, fdl_errors: [...] }
→ 에러 메시지 표시, FDL 에디터에 에러 위치 하이라이트
```

---

## Feature 파일 동기화

### 파일 구조

```
project/
├── features/
│   ├── user_auth.md      ← Feature 스펙 파일
│   └── blog_post.md
└── .claritask/
    └── db.clt
```

### FileSystemWatcher 설정

Extension 활성화 시 Feature 파일 감시 시작:

```typescript
// extension.ts
export function activate(context: vscode.ExtensionContext) {
    // Feature 파일 감시
    const featureWatcher = vscode.workspace.createFileSystemWatcher(
        '**/features/*.md'
    );

    // 파일 수정 시 → DB에 동기화
    featureWatcher.onDidChange(uri => {
        syncFeatureToDB(uri);
    });

    // 파일 생성 시 → DB에 등록
    featureWatcher.onDidCreate(uri => {
        syncFeatureToDB(uri);
    });

    // 파일 삭제 시 → DB에서 file_path 클리어 (레코드 유지)
    featureWatcher.onDidDelete(uri => {
        clearFeatureFilePath(uri);
    });

    context.subscriptions.push(featureWatcher);
}
```

### 동기화 흐름

```
┌─────────────────────────────────────────────────────────┐
│ Feature 파일 수정 감지                                   │
├─────────────────────────────────────────────────────────┤
│ features/<name>.md 수정 감지                            │
│     ↓                                                   │
│ 파일 읽기 → 해시 계산                                   │
│     ↓                                                   │
│ 해시 변경됨? → DB spec/content 컬럼에 동기화            │
│     ↓                                                   │
│ Webview에 업데이트 알림                                 │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ Feature 파일 삭제 감지                                   │
├─────────────────────────────────────────────────────────┤
│ DB 레코드 유지 (file_path만 클리어)                     │
│     ↓                                                   │
│ UI에 "파일 없음" 상태 표시                              │
│     ↓                                                   │
│ [파일 재생성] 버튼 제공                                 │
└─────────────────────────────────────────────────────────┘
```

### DB 스키마 확장 (Feature 파일 동기화용)

```sql
-- features 테이블에 파일 관련 필드 추가
ALTER TABLE features ADD COLUMN file_path TEXT DEFAULT '';
ALTER TABLE features ADD COLUMN content TEXT DEFAULT '';
ALTER TABLE features ADD COLUMN content_hash TEXT DEFAULT '';
```

---

## Markdown 렌더링 뷰

### 렌더링 표시

Feature 선택 시 `features/<name>.md` 파일 내용을 렌더링하여 표시:

```typescript
// FeatureDetail.tsx
const FeatureDetail = ({ feature }) => {
  return (
    <div className="feature-detail">
      <div className="spec-section">
        <h3>Spec</h3>
        {/* Markdown을 HTML로 렌더링 */}
        <div
          className="markdown-body"
          dangerouslySetInnerHTML={{ __html: renderMarkdown(feature.content) }}
        />
        <div className="actions">
          <button onClick={() => openFile(feature.file_path)}>Open File</button>
          <button onClick={() => editInline()}>Edit</button>
        </div>
      </div>
    </div>
  );
};
```

### 버튼 동작

- **Open File**: VSCode 에디터에서 `features/<name>.md` 파일 열기
- **Edit**: 인라인 Markdown 에디터 토글

---

## 메시지 프로토콜 (Feature 파일 관련)

### Webview → Extension

```typescript
// Feature 파일 열기 요청
{ type: 'openFeatureFile', featureId: number }

// Feature 파일 재생성 요청
{ type: 'regenerateFeatureFile', featureId: number }

// Feature 스펙 인라인 편집 저장
{ type: 'updateFeatureSpec', featureId: number, content: string }
```

### Extension → Webview

```typescript
// Feature 목록 업데이트 (파일 상태 포함)
{
  type: 'featuresUpdated',
  features: Feature[],
  // Feature에 file_path, content, content_hash 포함
}

// Feature 파일 열기 결과
{ type: 'featureFileOpened', featureId: number, success: boolean }
```

---

*Claritask VSCode Extension Spec v0.0.6*
