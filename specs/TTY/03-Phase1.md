# TTY Handover: Phase 1 - 요구사항 수립

> **버전**: v0.0.1

---

## 목적

사용자의 모호한 아이디어를 구체적인 Features 명세로 발전시킵니다.

---

## 실행 방법

```bash
$ clari init my-project "중고거래 플랫폼"
# → Claude Code 세션 시작 (TTY Handover)
# → 사용자와 대화하며 Features 확정
```

---

## 특징

- **구현**: TTY Handover로 Claude Code에게 위임 (직접 대화 UI 구현 X)
- **상태 저장**: Claude Code가 `clari feature add` 명령으로 DB에 저장
- **종료**: 사용자가 "개발해"라고 하면 `clari project start` 실행

---

## 플로우

```
┌─────────────────────────────────────────────────────────────┐
│  Phase 1: 요구사항 수립                                       │
│                                                             │
│  $ clari init my-project                                    │
│       ↓                                                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  TTY Handover → Claude Code                          │   │
│  │                                                      │   │
│  │  사용자: "중고거래 플랫폼, 당근마켓처럼"               │   │
│  │       ↓                                              │   │
│  │  Claude: Features 초안 제시                          │   │
│  │       ↓                                              │   │
│  │  사용자: "이거 추가해, 저거 빼"                       │   │
│  │       ↓                                              │   │
│  │  Claude: clari feature add '...'  ← DB 저장          │   │
│  │       ↓                                              │   │
│  │  사용자: "OK, 개발해"                                │   │
│  │       ↓                                              │   │
│  │  Claude: clari project start 실행                    │   │
│  │       ↓                                              │   │
│  │  Claude Code 종료                                    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 예시 대화

```
$ clari init carrot-clone "중고거래 플랫폼"

[Claude Code 세션 시작]

사용자: "당근마켓이랑 비슷하게 만들어줘"

Claude: "다음 기능들을 제안합니다:
  1. 회원가입/로그인
  2. 상품 등록
  3. 상품 검색/목록
  4. 채팅
  5. 거래 완료
  6. 사용자 평가

  추가하거나 수정할 기능이 있나요?"

사용자: "위치 기반 검색 추가해줘. 채팅은 나중에"

Claude: "수정했습니다. DB에 저장합니다."

$ clari feature add '{"name": "auth", "description": "회원가입/로그인"}'
$ clari feature add '{"name": "product", "description": "상품 등록"}'
$ clari feature add '{"name": "search", "description": "상품 검색 + 위치 기반"}'
$ clari feature add '{"name": "transaction", "description": "거래 완료"}'
$ clari feature add '{"name": "review", "description": "사용자 평가"}'

사용자: "좋아. 이제 개발해"

Claude: "Phase 2를 시작합니다."

$ clari project start

[Claude Code 종료 → clari project start가 새 오케스트레이터]
```

---

## 종료 조건

- 사용자가 Features 명세에 동의
- 명시적으로 실행 요청 ("개발해", "시작해", "만들어줘")
- Claude Code가 `clari project start` 실행 후 종료

---

## Phase 전환

Phase 1에서 Phase 2로의 전환:

1. Claude Code가 `clari project start` 실행
2. **새 프로세스**로 `clari project start`가 실행됨
3. Claude Code 종료
4. `clari project start`가 새 오케스트레이터가 됨

**Note**: 부모 `clari init`이 `clari project start`를 "감지"하는 게 아님. 단순히 새 프로세스로 교체됨.

---

*TTY Handover Specification v0.0.1 - 2026-02-03*
