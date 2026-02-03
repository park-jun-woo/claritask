# TASK-EXT-024: Expert DB 쿼리 함수

## 개요
database.ts에 Expert 테이블 읽기/쓰기 함수 추가

## 배경
- **스펙**: specs/VSCode/08-ExpertSync.md
- **현재 상태**: experts, project_experts 테이블 쿼리 없음

## 작업 내용

### 1. database.ts에 함수 추가
**파일**: `vscode-extension/src/database.ts`

```typescript
// Expert 목록 조회
export function readExperts(db: Database): Expert[] {
  const stmt = db.prepare(`
    SELECT id, name, version, domain, language, framework,
           path, description, content, content_hash, status,
           created_at, updated_at
    FROM experts
    WHERE status = 'active'
    ORDER BY name
  `);
  // ...
}

// 프로젝트에 할당된 Expert 목록
export function readProjectExperts(db: Database, projectId: string): string[] {
  const stmt = db.prepare(`
    SELECT expert_id FROM project_experts WHERE project_id = ?
  `);
  // ...
}

// Expert 할당
export function assignExpertToProject(
  db: Database,
  projectId: string,
  expertId: string
): void {
  const now = new Date().toISOString();
  db.run(`
    INSERT OR IGNORE INTO project_experts (project_id, expert_id, assigned_at)
    VALUES (?, ?, ?)
  `, [projectId, expertId, now]);
}

// Expert 할당 해제
export function unassignExpertFromProject(
  db: Database,
  projectId: string,
  expertId: string
): void {
  db.run(`
    DELETE FROM project_experts WHERE project_id = ? AND expert_id = ?
  `, [projectId, expertId]);
}

// Expert 내용 업데이트 (파일 동기화용)
export function updateExpertContent(
  db: Database,
  expertId: string,
  content: string,
  contentHash: string
): void {
  const now = new Date().toISOString();
  db.run(`
    UPDATE experts SET content = ?, content_hash = ?, updated_at = ?
    WHERE id = ?
  `, [content, contentHash, now, expertId]);
}
```

### 2. readAll 함수 확장
```typescript
export function readAll(db: Database): ProjectData {
  // 기존 코드...

  const experts = readExperts(db);
  const project = readProject(db);
  const projectExperts = project ? readProjectExperts(db, project.id) : [];

  // experts에 assigned 플래그 추가
  const expertsWithAssigned = experts.map(e => ({
    ...e,
    assigned: projectExperts.includes(e.id)
  }));

  return {
    // 기존 필드...
    experts: expertsWithAssigned,
    projectExperts
  };
}
```

## 완료 기준
- [ ] readExperts 함수 구현
- [ ] readProjectExperts 함수 구현
- [ ] assignExpertToProject 함수 구현
- [ ] unassignExpertFromProject 함수 구현
- [ ] updateExpertContent 함수 구현
- [ ] readAll에 experts 데이터 포함
