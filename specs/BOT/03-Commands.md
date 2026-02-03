# Claribot Commands

> **현재 버전**: v0.0.1

---

## 명령어 개요

```
/start          # 봇 시작 및 도움말
/help           # 명령어 도움말
/project        # 프로젝트 관리
/task           # 태스크 관리
/msg            # 메시지 관리
/expert         # Expert 호출
/status         # 전체 상태 요약
/settings       # 설정 관리
```

---

## /start

봇 시작 및 초기 설정

### 사용법
```
/start
```

### 응답 예시
```
👋 Claribot에 오신 것을 환영합니다!

현재 프로젝트: blog-project
진행률: 45% (9/20 태스크 완료)

명령어 도움말: /help
```

---

## /help

명령어 도움말 표시

### 사용법
```
/help
/help <command>
```

### 응답 예시
```
📖 Claribot 명령어

프로젝트
  /project list    - 프로젝트 목록
  /project status  - 현재 프로젝트 상태
  /project switch  - 프로젝트 전환

태스크
  /task list       - 태스크 목록
  /task add        - 태스크 추가
  /task start <id> - 태스크 시작
  /task done <id>  - 태스크 완료

메시지
  /msg send        - 메시지 전송
  /msg list        - 메시지 목록

Expert
  /expert list     - Expert 목록
  /expert ask      - Expert에게 질문

전체 상태
  /status          - 요약 대시보드
```

---

## /project

프로젝트 관리

### 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `/project list` | 프로젝트 목록 |
| `/project status` | 현재 프로젝트 상태 |
| `/project switch <id>` | 프로젝트 전환 |
| `/project info` | 상세 정보 |

### /project list

```
/project list
```

**응답:**
```
📁 프로젝트 목록

1. blog-project ⭐ (현재)
   진행률: 45% | 태스크: 9/20

2. api-server
   진행률: 80% | 태스크: 16/20

3. mobile-app
   진행률: 10% | 태스크: 2/20
```

### /project status

```
/project status
```

**응답:**
```
📊 blog-project 상태

진행률: ████████░░░░░░░░ 45%

태스크 현황:
  ✅ 완료: 9
  🔄 진행중: 2
  ⏳ 대기: 9

최근 완료:
  • TASK-015: 로그인 API 구현
  • TASK-014: 회원가입 폼 검증

현재 진행:
  • TASK-016: 비밀번호 재설정
  • TASK-017: 이메일 인증
```

### /project switch

```
/project switch api-server
```

**응답:**
```
✅ 프로젝트가 전환되었습니다.
현재 프로젝트: api-server
```

---

## /task

태스크 관리

### 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `/task list` | 태스크 목록 |
| `/task add` | 태스크 추가 (대화형) |
| `/task get <id>` | 태스크 상세 |
| `/task start <id>` | 태스크 시작 |
| `/task done <id>` | 태스크 완료 |
| `/task fail <id>` | 태스크 실패 |

### /task list

```
/task list
/task list pending
/task list in_progress
```

**응답:**
```
📋 태스크 목록 (blog-project)

진행중:
  🔄 16. 비밀번호 재설정 [@backend]
  🔄 17. 이메일 인증 [@backend]

대기중:
  ⏳ 18. 소셜 로그인 [@frontend]
  ⏳ 19. 프로필 페이지 [@frontend]
  ⏳ 20. 설정 페이지 [@frontend]

[1/2] 더 보기 ▶
```

### /task add (대화형)

```
/task add
```

**대화 흐름:**
```
Bot: 태스크 제목을 입력하세요:
User: 검색 기능 구현
Bot: 설명을 입력하세요 (스킵: /skip):
User: 블로그 포스트 검색 기능
Bot: Expert를 지정하세요 (스킵: /skip):
     [backend] [frontend] [스킵]
User: (버튼 클릭: backend)
Bot: ✅ 태스크가 생성되었습니다.
     TASK-021: 검색 기능 구현
     Expert: backend
```

### /task get

```
/task get 16
```

**응답:**
```
📌 TASK-016: 비밀번호 재설정

상태: 🔄 진행중
Expert: backend
Feature: 회원기능

설명:
비밀번호 재설정 이메일 발송 및 토큰 검증 구현

생성일: 2026-02-01
시작일: 2026-02-03

[완료] [실패] [상세보기]
```

### /task start, done, fail

```
/task start 18
/task done 16
/task fail 17 "API 키 이슈"
```

**응답:**
```
✅ TASK-016이 완료되었습니다.

다음 태스크:
  ⏳ 18. 소셜 로그인 [@frontend]

[시작하기]
```

---

## /msg

메시지 관리 (CLI message 연동)

### 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `/msg send` | 메시지 전송 |
| `/msg list` | 메시지 목록 |
| `/msg get <id>` | 메시지 상세 |

### /msg send (대화형)

```
/msg send
```

**대화 흐름:**
```
Bot: 메시지 제목을 입력하세요:
User: API 설계 검토 요청
Bot: 내용을 입력하세요:
User: 로그인 API 엔드포인트 설계를 검토해주세요.
Bot: 수신자를 선택하세요:
     [모든 Expert] [backend] [frontend]
User: (버튼 클릭: backend)
Bot: ✅ 메시지가 전송되었습니다.
     MSG-007: API 설계 검토 요청
```

### /msg list

```
/msg list
```

**응답:**
```
💬 메시지 목록

최신:
  📩 007. API 설계 검토 요청
       → backend | 2분 전
  📩 006. 테스트 결과 공유
       → all | 1시간 전
  📨 005. 배포 일정 확인
       ← system | 어제

[더 보기]
```

---

## /expert

Expert 관리 및 호출

### 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `/expert list` | Expert 목록 |
| `/expert ask` | Expert에게 질문 |
| `/expert status` | Expert별 태스크 현황 |

### /expert list

```
/expert list
```

**응답:**
```
👥 Expert 목록

🟢 backend (활성)
   담당: API, 데이터베이스
   태스크: 3개 진행중

🟢 frontend (활성)
   담당: UI, 사용자 경험
   태스크: 2개 진행중

⚪ devops (비활성)
   담당: 배포, 인프라
   태스크: 0개
```

### /expert ask

```
/expert ask backend
```

**대화 흐름:**
```
Bot: backend에게 질문할 내용을 입력하세요:
User: JWT 토큰 만료 시간은 어떻게 설정하면 좋을까요?
Bot: ✅ 질문이 전송되었습니다.
     Expert backend에게 알림이 전달됩니다.
```

---

## /status

전체 상태 대시보드

```
/status
```

**응답:**
```
📊 Claritask 대시보드

프로젝트: blog-project
진행률: ████████░░░░░░░░ 45%

태스크 요약:
  ✅ 완료: 9
  🔄 진행: 2
  ⏳ 대기: 9

오늘의 활동:
  • 완료된 태스크: 2
  • 새 메시지: 3
  • 활성 Expert: 2

[프로젝트] [태스크] [메시지]
```

---

## /settings

봇 설정 관리

### 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `/settings notify` | 알림 설정 |
| `/settings project` | 기본 프로젝트 |

### /settings notify

```
/settings notify
```

**응답 (인라인 버튼):**
```
🔔 알림 설정

태스크 완료: [켜짐 ✓] [꺼짐]
태스크 실패: [켜짐 ✓] [꺼짐]
새 메시지:   [켜짐 ✓] [꺼짐]
일일 요약:   [켜짐] [꺼짐 ✓]
```

---

## 인라인 버튼 규칙

### 콜백 데이터 형식

```
<action>:<entity>:<id>:<extra>
```

### 예시

| 버튼 | 콜백 데이터 |
|------|-------------|
| 태스크 시작 | `task:start:16` |
| 태스크 완료 | `task:done:16` |
| 프로젝트 전환 | `project:switch:api-server` |
| 더 보기 | `task:list:page:2` |

---

## 에러 메시지

| 상황 | 메시지 |
|------|--------|
| 권한 없음 | ❌ 권한이 없습니다. |
| 프로젝트 없음 | ❌ 프로젝트를 찾을 수 없습니다. |
| 태스크 없음 | ❌ 태스크를 찾을 수 없습니다. |
| 잘못된 명령 | ❓ 알 수 없는 명령어입니다. /help를 확인하세요. |
| 서버 오류 | ⚠️ 일시적인 오류가 발생했습니다. 다시 시도해주세요. |

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [01-Overview.md](01-Overview.md) | 전체 개요 |
| [02-Architecture.md](02-Architecture.md) | 아키텍처 |
| [CLI/15-Message.md](../CLI/15-Message.md) | CLI Message |

---

*Claribot Commands v0.0.1*
