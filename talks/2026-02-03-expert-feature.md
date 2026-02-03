# Expert 기능 구현 대화록

> **날짜**: 2026-02-03
> **주제**: Expert(전문가) 기능 설계, 구현, VSCode Extension 연동

---

## 1. 명칭 결정: Skills vs Expert

### 토론 내용
- Claude Code에서 이미 "Skills" 용어 사용 중
- Claritask에서도 Skills 사용 시 혼동 가능성

### 결론
**Expert/Experts 채택**
- Claritask의 기존 컨셉(Context, Tech, Design)과 어울림
- Claude Code Skills와 개념적 분리 확실
- `clari expert add "Go Backend Developer"` - 자연스러운 CLI 문법

---

## 2. Expert 샘플 작성

`backend-expert.md` 파일 생성 - Go GIN 프레임워크 전문가 정의

### Expert 포맷 구조
```
1. Metadata (테이블) - ID, Name, Version, Domain, Language, Framework
2. Role Definition - 한 문장 역할 정의
3. Tech Stack - Core / Supporting 구분
4. Architecture Pattern - 디렉토리 구조
5. Coding Rules - 패턴별 코드 템플릿
6. Error Handling - 에러 타입/처리 패턴
7. Testing Rules - 테스트 코드 템플릿
8. Performance Guidelines - 체크리스트
9. Security Checklist - 체크박스
10. References - 외부 문서 링크
```

---

## 3. 저장 방식 결정

### 토론 옵션
- A. 파일 시스템 (MD 파일)
- B. SQLite TEXT 저장
- C. 하이브리드

### 결론: 하이브리드 채택
```
.claritask/
├── db.clt                    # 메타데이터, 프로젝트 연결
└── experts/
    └── <expert-id>/
        └── EXPERT.md         # 실제 내용
```

- **편집**: VSCode 기본 MD 에디터 + 프리뷰 사용
- **버전관리**: Git으로 Expert 변경 추적
- **DB 역할**: 프로젝트-Expert 연결, 메타데이터만 관리

---

## 4. CLI 명령어 설계

```bash
clari expert add <id>           # Expert 생성
clari expert list [--assigned|--available]
clari expert get <id>
clari expert edit <id>          # 에디터로 열기
clari expert remove <id> [--force]
clari expert assign <id>        # 프로젝트에 할당
clari expert unassign <id>
```

specs/Commands.md 업데이트 완료

---

## 5. 개발 실행

### 완료된 TASK (DEV-068 ~ DEV-073)

| TASK | 파일 | 내용 |
|------|------|------|
| DEV-068 | models.go | Expert, ProjectExpert, ExpertInfo 모델 |
| DEV-069 | db.go | experts, project_experts 테이블 |
| DEV-070 | expert_service.go | 신규 - 9개 함수 |
| DEV-071 | expert.go (cmd) | 신규 - 7개 서브커맨드 |
| DEV-072 | task_service.go | PopTaskFull에 Expert manifest 연동 |
| DEV-073 | expert_service_test.go | 13개 테스트 케이스 |

모든 테스트 통과 확인

---

## 6. VSCode Extension UI 토론

### .clt 에디터에서 외부 파일 열기
- **질문**: .clt에서 DB 데이터도 관리하고, 외부 파일(EXPERT.md)도 열어주는 게 이상한가?
- **결론**: 전혀 이상하지 않음. 표준 패턴.
  - VS Code Settings UI → "Edit in settings.json" 버튼
  - Database 클라이언트 → Export to SQL
  - Docker Extension → View Logs, Inspect

### Experts 탭 설계
```
┌────────────────────────────────────────┐
│ Assigned Experts (2)                   │
│ ┌────────────────────────────────────┐ │
│ │ 🔧 backend-go-gin          [Edit]  │ │
│ │ (렌더링된 마크다운 내용)            │ │
│ │                        [Unassign]  │ │
│ └────────────────────────────────────┘ │
│                                        │
│ Available Experts (1)                  │
│ ┌────────────────────────────────────┐ │
│ │ ☁️ devops-k8s              [Assign] │ │
│ └────────────────────────────────────┘ │
│                                        │
│ [+ Create New Expert]                  │
└────────────────────────────────────────┘
```

- **마크다운 렌더링**: react-markdown 사용
- **[Edit] 버튼**: `vscode.commands.executeCommand('vscode.open', uri)`

---

## 7. Expert 파일 동기화 방식

### 토론: md 파일 삭제 시 처리
- 옵션 A: DB도 삭제
- 옵션 B: DB 유지 + 재생성
- 옵션 C: DB 유지 + 경고

### 결론: B + 백업 방식
```
평소: md 읽을 때마다 DB에 content 백업
파일 삭제: DB 백업에서 자동 복구 (조용히)
UI 삭제: DB + 파일 모두 삭제
```

### FileSystemWatcher 채택 (옵션 2)

```typescript
// extension.ts - db.clt 안 열어도 동작
export function activate(context) {
    const watcher = vscode.workspace.createFileSystemWatcher(
        '**/.claritask/experts/**/EXPERT.md'
    );

    watcher.onDidChange(uri => syncExpertToDB(uri));
    watcher.onDidDelete(uri => restoreExpertFromDB(uri));
}
```

**activationEvents**:
```json
{
  "activationEvents": [
    "workspaceContains:.claritask/db.clt"
  ]
}
```

---

## 8. specs 문서 업데이트

### Commands.md (v0.0.3)
- Expert DB 스키마 백업 필드 추가
  - `content` - EXPERT.md 전체 내용 백업
  - `content_hash` - 변경 감지용 해시
  - `updated_at` - 마지막 동기화 시간
- 동기화 정책 명시

### VscodeGUI.md (v0.0.4)
- Experts 탭 레이아웃 및 기능
- FileSystemWatcher 섹션
- 메시지 프로토콜 Expert 관련
- 로드맵 Phase 4: Experts 탭

---

## 9. TODO (미완료)

실제 Go 코드에 백업 필드 추가 필요:
- [ ] db.go - content, content_hash, updated_at 컬럼
- [ ] expert_service.go - 백업/복구 로직
- [ ] VSCode Extension - FileSystemWatcher 구현
- [ ] VSCode Extension - Experts 탭 UI

---

*2026-02-03 Expert 기능 구현 대화록*
