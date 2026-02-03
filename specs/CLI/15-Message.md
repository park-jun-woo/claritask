# Message 명령어

> **현재 버전**: v0.0.1 ([변경이력](../HISTORY.md))

---

## 개요

Message는 사용자의 수정 요청을 입력받아 Claude Code가 분석하고 Task로 변환하는 기능입니다.

**워크플로우**:
```
사용자 요청 → DB 저장 → Claude 분석 → Task 등록 → 결과 보고
```

---

## 명령어

### message send

수정 요청 메시지를 전송하고 Claude Code가 Task로 변환

```bash
clari message send "<요청 내용>"
clari message send --feature <feature_id> "<요청 내용>"
```

**플래그**:
| 플래그 | 타입 | 필수 | 설명 |
|--------|------|------|------|
| `--feature`, `-f` | int | 선택 | 관련 Feature ID |

**프로세스**:
1. Messages 테이블에 메시지 INSERT (status: `pending`)
2. TTY Handover로 Claude Code 호출
3. Claude가 요청 분석 후 `clari task add` 로 Task 등록
4. `.claritask/complete` 파일 생성 (내용: 결과 보고서 MD)
5. clari가 complete 감지 후:
   - Message의 `response` 필드에 결과 저장
   - `status`를 `completed`로 업데이트
   - `reports/` 폴더에 MD 파일 저장
   - Claude 프로세스 종료
   - complete 파일 삭제

**출력**:
```json
{
  "success": true,
  "message_id": 1,
  "tasks_created": 3,
  "report_path": "reports/2026-02-04-message-001.md"
}
```

**에러**:
```json
{
  "success": false,
  "error": "message send failed: <reason>"
}
```

---

### message list

메시지 목록 조회

```bash
clari message list
clari message list --status pending
clari message list --limit 10
```

**플래그**:
| 플래그 | 타입 | 기본값 | 설명 |
|--------|------|--------|------|
| `--status`, `-s` | string | all | 상태 필터 (pending, processing, completed, failed) |
| `--feature`, `-f` | int | - | Feature ID 필터 |
| `--limit`, `-l` | int | 20 | 최대 개수 |

**출력**:
```json
{
  "success": true,
  "messages": [
    {
      "id": 1,
      "content": "로그인 페이지에 소셜 로그인 추가해줘",
      "status": "completed",
      "feature_id": 2,
      "tasks_count": 3,
      "created_at": "2026-02-04T10:00:00Z"
    }
  ],
  "total": 1
}
```

---

### message get

특정 메시지 상세 조회

```bash
clari message get <message_id>
```

**출력**:
```json
{
  "success": true,
  "message": {
    "id": 1,
    "content": "로그인 페이지에 소셜 로그인 추가해줘",
    "status": "completed",
    "feature_id": 2,
    "response": "## 분석 결과\n\n3개의 Task가 생성되었습니다...",
    "created_at": "2026-02-04T10:00:00Z",
    "completed_at": "2026-02-04T10:05:00Z",
    "tasks": [
      { "id": 10, "title": "OAuth 모듈 추가", "status": "pending" },
      { "id": 11, "title": "소셜 로그인 버튼 UI", "status": "pending" },
      { "id": 12, "title": "소셜 로그인 API 연동", "status": "pending" }
    ]
  }
}
```

---

### message delete

메시지 삭제 (관련 Task는 유지)

```bash
clari message delete <message_id>
```

**출력**:
```json
{
  "success": true,
  "deleted_id": 1
}
```

---

## Claude Code 시스템 프롬프트

```
You are in Claritask Message Analysis Mode.

ROLE: Analyze user's modification request and create Tasks.

WORKFLOW:
1. Analyze the user's request
2. Break down into actionable Tasks
3. Register each Task using: clari task add '{"feature_id": N, "title": "...", "content": "..."}'
4. Write a summary report

COMPLETION:
When all Tasks are registered, create '.claritask/complete' file with the report content.
Report format (Markdown):
- Summary of request analysis
- List of created Tasks with IDs
- Any assumptions or clarifications needed

CONSTRAINTS:
- Always register Tasks, never implement directly
- Each Task should be specific and actionable
- Reference existing FDL if available
- Link Tasks to appropriate Feature
```

---

## 결과 보고서 형식

```markdown
# Message Analysis Report

## 요청 내용
> 로그인 페이지에 소셜 로그인 추가해줘

## 분석 결과

### 생성된 Task 목록

| ID | Title | Feature | Status |
|----|-------|---------|--------|
| 10 | OAuth 모듈 추가 | 회원기능 | pending |
| 11 | 소셜 로그인 버튼 UI | 회원기능 | pending |
| 12 | 소셜 로그인 API 연동 | 회원기능 | pending |

### 상세 설명

1. **OAuth 모듈 추가**: Google, GitHub OAuth 2.0 연동 모듈
2. **소셜 로그인 버튼 UI**: LoginPage에 소셜 로그인 버튼 추가
3. **소셜 로그인 API 연동**: /auth/oauth/{provider} 엔드포인트 구현

### 참고사항
- 기존 회원기능 FDL 구조와 호환되도록 설계
- 추가 OAuth provider 확장 고려
```

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [DB/02-C-Content.md](../DB/02-C-Content.md) | Messages 테이블 스키마 |
| [TTY/01-Overview.md](../TTY/01-Overview.md) | TTY Handover 아키텍처 |
| [04-Task.md](04-Task.md) | Task 명령어 |

---

*Claritask Message Command Specification v0.0.1*
