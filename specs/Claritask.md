# Claritask - Task And LLM Operating System

> **현재 버전**: v0.0.4 ([변경이력](HISTORY.md))

---

## 개요

LLM 기반 프로젝트 자동 실행 시스템

**목표**:
- 프로젝트 수동 세팅 자동화 (30-50분 절약)
- 무제한 무인 작업 가능 (Task 수 제한 없음)
- 컨텍스트 한계 완전 극복 (매 Task마다 초기화)

**철학**:
- **Claritask가 오케스트레이터**, Claude Code는 실행기
- Task 단위 독립 실행으로 컨텍스트 격리
- **FDL(Feature Definition Language)로 계약 정의**, 스켈레톤 자동 생성
- **LLM은 TODO만 채움** - 함수명/타입/API 경로는 확정적
- 한 줄 명령으로 프로젝트 완성

---

## 아키텍처: 2-Phase + TTY Handover

### 전체 구조

```
clari init
  └─▶ Claude Code [Phase 1: 요구사항 수립]
        │
        │  clari feature add '...'
        │  사용자: "개발해"
        │
        └─▶ clari project start
              │
              ├─▶ Claude Code [Task 1] ─▶ 완료
              ├─▶ Claude Code [Task 2] ─▶ 완료
              ├─▶ Claude Code [Task N] ─▶ 완료
              │
              └─▶ 최종 보고
```

### TTY Handover

Claritask가 Claude Code에게 **터미널 제어권을 완전히 넘기고**, 종료 시 복귀:

```
Claritask ──TTY Handover──▶ Claude Code ──종료──▶ Claritask
```

| 방식 | 가능한 작업 | 한계 |
|------|------------|------|
| `claude --print` | 코드 생성 | 테스트/디버깅 불가 |
| TTY Handover | 코딩 + 테스트 + 디버깅 | 없음 |

### 설계 원칙

| 원칙 | 설명 |
|------|------|
| 순차 실행 | 프로세스 중첩 아님, 하나 끝나면 다음 |
| Stateless | 각 clari 프로세스는 stateless, 상태는 DB에 |
| 복구 용이 | 실패 시 마지막 Task부터 재개 |

> 상세: [TTY/01-Overview.md](TTY/01-Overview.md)

---

## 기술 스택

- **Go + SQLite**: 단일 바이너리, 고성능
- **Python**: FDL 파서 및 스켈레톤 생성기
- **파일**: `.claritask/db.clt` 하나로 모든 것 관리
- **동시성**: WAL 모드로 CLI/GUI 동시 접근 지원

> 상세: [DB/01-Overview.md](DB/01-Overview.md)

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

> FDL 상세 스펙: [FDL/01-Overview.md](FDL/01-Overview.md)

---

## 워크플로우 요약

```
1. clari init <project>           # 프로젝트 초기화 + TTY Handover
2. (대화) Feature 확정             # Claude Code가 feature add 실행
3. clari project start            # 자동 실행 시작
4. (자동) Task별 TTY Handover      # 코딩 + 테스트 + 디버깅
5. 완료 또는 실패 보고
```

> 명령어 상세: [CLI/01-Overview.md](CLI/01-Overview.md)

---

## 핵심 가치

1. **제어 역전**: Claritask가 오케스트레이터, Claude Code는 실행기
2. **TTY Handover**: 테스트/디버깅까지 완전한 Claude Code 세션
3. **FDL 기반 계약**: 함수명, 타입, API 경로를 먼저 확정
4. **스켈레톤 자동 생성**: Python이 FDL → 코드 틀 생성 (오타 원천 차단)
5. **TODO만 채우기**: LLM은 로직만 작성, 구조 변경 불가
6. **그래프 기반**: Task 간 의존성을 Edge로 명시, 정밀한 컨텍스트 주입
7. **검증 가능**: 구현이 FDL과 일치하는지 자동 검사
8. **무제한 확장**: Task 수천 개도 자동 처리
9. **복구 가능**: 실패 시 해당 Task부터 재개

```
사람: FDL로 "무엇을" 정의
      ↓
Python: "어떤 구조로" 스켈레톤 생성
      ↓
Claude Code: "어떻게" TODO 채우기 (TTY Handover)
```

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [TTY/](TTY/01-Overview.md) | 2-Phase 흐름 및 TTY Handover 명세 |
| [DB/](DB/01-Overview.md) | 데이터베이스 스키마 |
| [FDL/](FDL/01-Overview.md) | Feature Definition Language 명세 |
| [CLI/](CLI/01-Overview.md) | 명령어 레퍼런스 |
| [VSCode/](VSCode/01-Overview.md) | VSCode 확장 명세 |
| [HISTORY.md](HISTORY.md) | 전체 변경이력 |

---

*Claritask Specification v0.0.4*
