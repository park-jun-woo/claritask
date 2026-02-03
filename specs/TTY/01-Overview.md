# TTY Handover: Overview

> **버전**: v0.0.1

## 변경이력

| 버전 | 날짜 | 내용 |
|------|------|------|
| v0.0.1 | 2026-02-03 | 최초 작성 |

---

## 개요

TTY Handover는 Claritask가 Claude Code에게 **터미널 제어권을 완전히 넘기고**, Claude Code가 종료되면 다시 제어권을 가져오는 방식입니다.

```
Claritask ──TTY Handover──▶ Claude Code ──종료──▶ Claritask
```

---

## 2-Phase 구조

Claritask는 두 개의 Phase로 동작하며, 두 Phase 모두 TTY Handover를 사용합니다:

| Phase | 모드 | 주체 | 목적 | TTY Handover 용도 |
|-------|------|------|------|------------------|
| Phase 1 | 대화형 | 사용자 + Claude Code | 요구사항 수립 | 사용자와 대화 |
| Phase 2 | 자동화 | Claritask + Claude Code | 실행 | Task별 코딩/테스트/디버깅 |

---

## 왜 TTY Handover인가?

| 방식 | 가능한 작업 | 한계 |
|------|------------|------|
| `claude --print` | 코드 생성 | 테스트 실행 불가, 에러 확인 불가, 디버깅 불가 |
| TTY Handover | 코딩 + 테스트 + 디버깅 + 사용자 대화 | 없음 (완전한 Claude Code 세션) |

**핵심**: Claude Code가 **실제 터미널에서 동작**하므로:
- 테스트 실행 가능
- 에러 로그 확인 가능
- 코드 수정 후 재실행 가능
- 필요 시 사용자 개입 가능

---

## 문서 목차

| 문서 | 내용 |
|------|------|
| [01-Overview.md](01-Overview.md) | 개요 및 2-Phase 구조 |
| [02-Architecture.md](02-Architecture.md) | 프로세스 아키텍처 및 설계 원칙 |
| [03-Phase1.md](03-Phase1.md) | Phase 1: 요구사항 수립 |
| [04-Phase2.md](04-Phase2.md) | Phase 2: 자동 실행 |
| [05-Implementation.md](05-Implementation.md) | Go 구현 및 CLI 명령어 |
| [06-ClaudeCLI.md](06-ClaudeCLI.md) | Claude CLI 옵션 및 프롬프트 전략 |
| [07-Scenarios.md](07-Scenarios.md) | 사용 시나리오 |

---

*TTY Handover Specification v0.0.1 - 2026-02-03*
