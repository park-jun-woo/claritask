# TTY Handover: Claude CLI Options

> **버전**: v0.0.1

---

## 프롬프트 전달

```bash
# 대화형 모드 + 첫 프롬프트 (positional argument)
claude "테스트 실행하고 버그 고쳐줘"

# 시스템 프롬프트 + 첫 프롬프트
claude --system-prompt "너는 디버깅 전문가야" "pytest 실행해"
```

---

## 권한 모드

```bash
--permission-mode <mode>

# 옵션:
# - default: 기본 (도구 실행 전 확인)
# - acceptEdits: 편집 자동 승인
# - bypassPermissions: 모든 권한 확인 건너뛰기
# - plan: 계획 모드
```

**Claritask에서 사용**: `--permission-mode acceptEdits`
- 코드 편집을 자동 승인
- 테스트/빌드 명령은 확인 필요

---

## 세션 관리

```bash
--continue              # 가장 최근 대화 이어가기
--resume <session_id>   # 특정 세션 복원
--session-id <uuid>     # 특정 세션 ID 사용
```

### 세션 저장/복원

```bash
# 세션 ID 지정하여 실행
claude --session-id task-42 "implement createComment"

# 나중에 이어서
claude --resume task-42
```

---

## 프롬프트 전략

### Auto-Pilot Trigger

Claude가 대화형 모드에서 즉시 작업을 시작하게 하려면:

```text
[시스템 프롬프트 끝에]

IMPORTANT: Start working immediately without waiting for user input.
Your first action should be running the test command.
```

### 컨텍스트 압축

대화형 모드에서도 컨텍스트가 너무 크면 문제. 핵심만 전달:

```text
=== FDL (핵심만) ===
service:
  - name: createComment
    input: { userId, postId, content }
    steps: [validate, db insert, return]

=== 에러 로그 (최근 50줄) ===
...

=== 관련 코드 (TODO 부분만) ===
async def createComment(...):
    # TODO: implement
    raise NotImplementedError
```

### 종료 조건 명시

```text
COMPLETION:
When the test passes, summarize what you fixed and exit with /exit.
If you cannot fix it after 3 attempts, explain the blocker and exit.
```

---

## Claritask에서 사용하는 옵션 조합

### Phase 1 (요구사항 수립)

```bash
claude \
  --system-prompt "You are helping define project features..." \
  "프로젝트: ${PROJECT_NAME}, 설명: ${PROJECT_DESC}"
```

### Phase 2 (Task 실행)

```bash
claude \
  --system-prompt "${TASK_SYSTEM_PROMPT}" \
  --permission-mode acceptEdits \
  "${TASK_INITIAL_PROMPT}"
```

---

## 환경 변수

Claude CLI가 참조하는 환경 변수:

```bash
ANTHROPIC_API_KEY=sk-...      # API 키 (필수)
CLAUDE_CODE_CONFIG_DIR=...    # 설정 디렉토리
```

---

*TTY Handover Specification v0.0.1 - 2026-02-03*
