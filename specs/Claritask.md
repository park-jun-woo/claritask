# Claritask - Task And LLM Operating System

> **버전**: v0.0.1

## 변경이력

| 버전 | 날짜 | 내용 |
|------|------|------|
| v0.0.1 | 2026-02-03 | 최초 작성 |

---

## 개요

LLM 기반 프로젝트 자동 실행 시스템

**목표**:
- 프로젝트 수동 세팅 자동화 (30-50분 절약)
- 무제한 무인 작업 가능 (Task 수 제한 없음)
- 컨텍스트 한계 완전 극복 (매 Task마다 초기화)

**철학**:
- **Claritask가 오케스트레이터**, Claude는 실행기
- Task 단위 독립 실행으로 컨텍스트 격리
- **FDL(Feature Definition Language)로 계약 정의**, 스켈레톤 자동 생성
- **LLM은 TODO만 채움** - 함수명/타입/API 경로는 확정적
- 한 줄 명령으로 프로젝트 완성

---

## 아키텍처: 제어 역전

### 기존 구조의 한계

기존에는 Claude Code가 Claritask를 도구로 사용했다. 이 구조는 단일 작업에는 적합하지만, 대규모 자동화에는 치명적인 한계가 있다.

- **컨텍스트 누적**: Task를 처리할수록 대화 컨텍스트가 쌓인다
- **세션 의존성**: Claude Code 세션이 끊기면 작업도 중단된다
- **확장 불가**: Task가 100개, 1000개로 늘어나면 단일 세션으로 처리 불가능

### 새로운 구조: Claritask가 오케스트레이터

**제어권을 역전한다.** Claritask가 드라이버가 되고, Claude는 순수 실행기가 된다.

```
┌─────────────────────────────────────────────────────────────┐
│                        Claritask                            │
│                     (Orchestrator)                          │
│                                                             │
│   ┌─────────┐    ┌─────────┐    ┌─────────┐                │
│   │ Task 1  │───▶│ Task 2  │───▶│ Task N  │───▶ 완료       │
│   └────┬────┘    └────┬────┘    └────┬────┘                │
│        │              │              │                      │
│        ▼              ▼              ▼                      │
│   claude --print claude --print claude --print              │
│   (독립 컨텍스트) (독립 컨텍스트) (독립 컨텍스트)              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

| 측면 | 기존 (Claude 드라이버) | 신규 (Claritask 드라이버) |
|------|------------------------|---------------------------|
| 컨텍스트 | 누적되어 폭발 | 매 Task마다 초기화 |
| 세션 | 끊기면 중단 | 프로세스 기반, 복구 가능 |
| 확장성 | 수십 개 한계 | 수천 개도 가능 |
| 상태 관리 | Claude 메모리 의존 | DB에 영속화 |
| 재시작 | 처음부터 다시 | 마지막 Task부터 재개 |

### 두 가지 모드 공존

**1. 자동화 모드 (Claritask 드라이버)**
```bash
clari project start
# → Task 전체 순회, claude --print 반복 호출
```

**2. 대화형 모드 (Claude/사용자 드라이버)**
```bash
clari task list
clari task get 3
clari memo add --scope task --id 3 "JWT 만료 시간 수정"
```

---

## 기술 스택

- **Go + SQLite**: 단일 바이너리, 고성능
- **Python**: FDL 파서 및 스켈레톤 생성기
- **파일**: `.claritask/db.clt` 하나로 모든 것 관리
- **동시성**: WAL 모드로 CLI/GUI 동시 접근 지원

---

## 데이터 구조: 그래프 기반

### project → feature → task (with edges)

```
project: Blog Platform
├─ feature: 로그인
│  ├─ task: user_table_sql
│  ├─ task: user_model ─────────depends_on────▶ user_table_sql
│  └─ task: login_api ──────────depends_on────▶ user_model
│
├─ feature: 결제 ───────────────depends_on────▶ 로그인 (Feature Edge)
│  └─ task: payment_api ────────depends_on────▶ user_model
│
└─ feature: 블로그
   └─ task: post_api ───────────depends_on────▶ auth_service
```

**특징**:
- **project**: 프로젝트 전체
- **feature**: 기능 단위 (로그인, 결제 등)
- **task**: 실제 실행 단위
- **edge**: Task/Feature 간 의존성 (그래프 구조)

그래프 구조의 장점:
- **컨텍스트 정밀 주입**: 해당 Task + 의존 Task 결과만 주입
- **실행 순서 자동 결정**: Topological Sort로 의존성 해결된 Task부터 실행
- **토큰 최소화**: 전체 manifest 대신 필요한 것만

---

## FDL 통합

FDL(Feature Definition Language)로 계약을 정의하면 스켈레톤이 자동 생성된다.

```
FDL (YAML)  →  Python Parser  →  Skeleton Code  →  Task (TODO 채우기)
     ↓              ↓                  ↓                    ↓
  계약 정의      AST 변환         코드 틀 생성        LLM이 내용만 작성
```

**LLM의 역할이 "코드 전체 작성"에서 "TODO 채우기"로 축소됨**

| 기존 (result만 공유) | FDL + 스켈레톤 |
|---------------------|---------------|
| LLM이 함수명 결정 → 오타 가능 | FDL에서 확정 |
| LLM이 타입 결정 → 불일치 가능 | FDL에서 확정 |
| Task 간 import 경로 불일치 | 스켈레톤이 Single Source |
| 전체 코드 작성 | TODO만 채우기 |

**"LLM의 창의성은 로직 구현에만, 구조는 확정적으로"**

> FDL 상세 스펙: [specs/FDL/](FDL/01-Overview.md)

---

## 워크플로우 요약

```
1. clari init <project>           # 프로젝트 초기화
2. clari plan features            # Feature 목록 산출 (LLM)
3. clari fdl create <feature>     # FDL 작성
4. clari fdl skeleton <feature>   # 스켈레톤 생성 (Python)
5. clari fdl tasks <feature>      # Task 자동 생성
6. clari project start            # 자동 실행
7. clari fdl verify <feature>     # 검증
```

> 명령어 상세: [specs/CLI/](CLI/01-Overview.md)

---

## 핵심 가치

1. **제어 역전**: Claritask가 오케스트레이터, Claude는 실행기
2. **FDL 기반 계약**: 함수명, 타입, API 경로를 먼저 확정
3. **스켈레톤 자동 생성**: Python이 FDL → 코드 틀 생성 (오타 원천 차단)
4. **TODO만 채우기**: LLM은 로직만 작성, 구조 변경 불가
5. **그래프 기반**: Task 간 의존성을 Edge로 명시, 정밀한 컨텍스트 주입
6. **검증 가능**: 구현이 FDL과 일치하는지 자동 검사
7. **무제한 확장**: Task 수천 개도 자동 처리
8. **복구 가능**: 실패 시 해당 Task부터 재개

```
사람: FDL로 "무엇을" 정의
      ↓
Python: "어떤 구조로" 스켈레톤 생성
      ↓
LLM: "어떻게" TODO 채우기
```

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [FDL Spec](FDL/01-Overview.md) | Feature Definition Language 명세 |
| [CLI Commands](CLI/01-Overview.md) | 명령어 레퍼런스 |
| [VSCode Extension](VSCode/01-Overview.md) | VSCode 확장 명세 |
| [TTY Handover](TTY-Handover.md) | TTY 핸드오버 명세 |

---

*Claritask Specification v0.0.1 - 2026-02-03*
