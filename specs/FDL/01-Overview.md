# ClariSpec: Feature Definition DSL - Overview

> **현재 버전**: v0.0.4 ([변경이력](../HISTORY.md))

---

## 개요

ClariSpec은 Claritask 시스템에서 하나의 기능(Feature)을 "Vertical Slice(수직적 격리)" 형태로 정의하기 위한 YAML 기반의 DSL(Domain Specific Language)입니다. 이 명세는 SQL, 백엔드 로직, API 인터페이스, 프론트엔드 UI를 통합적으로 기술합니다.

---

## 핵심 철학

1. **Unified Source**: 데이터(DB)부터 화면(UI)까지 하나의 파일에 정의하여 정합성을 보장합니다.
2. **Loose Syntax**: 엄격한 문법보다 가독성과 의미 전달을 우선합니다. (LLM 해석용)
3. **Explicit Wiring**: UI가 어떤 API를 호출하고, API가 어떤 Service를 쓰는지 명시합니다.

---

## 4계층 구조

ClariSpec은 다음 4개의 계층으로 구성됩니다:

| 계층 | 섹션 | 역할 | 상세 문서 |
|------|------|------|----------|
| DATA | `models` | 데이터베이스 스키마 및 모델 정의 | [02-A-DataLayer.md](02-A-DataLayer.md) |
| LOGIC | `service` | 비즈니스 로직 및 규칙 | [02-B-LogicLayer.md](02-B-LogicLayer.md) |
| INTERFACE | `api` | HTTP API 계약 (Controller/Router) | [02-C-InterfaceLayer.md](02-C-InterfaceLayer.md) |
| PRESENTATION | `ui` | UI 컴포넌트 및 상태 관리 | [02-D-PresentationLayer.md](02-D-PresentationLayer.md) |

---

## 문서 목차

| 파일 | 내용 |
|------|------|
| [01-Overview.md](01-Overview.md) | 개요 및 핵심 철학 |
| [02-Schema.md](02-Schema.md) | DSL 구조 개요 및 계층간 연결 |
| [02-A-DataLayer.md](02-A-DataLayer.md) | DATA LAYER 상세 (models) |
| [02-B-LogicLayer.md](02-B-LogicLayer.md) | LOGIC LAYER 상세 (service) |
| [02-C-InterfaceLayer.md](02-C-InterfaceLayer.md) | INTERFACE LAYER 상세 (api) |
| [02-D-PresentationLayer.md](02-D-PresentationLayer.md) | PRESENTATION LAYER 상세 (ui) |
| [03-Examples.md](03-Examples.md) | 작성 예시 (댓글 시스템) |
| [04-Guidelines.md](04-Guidelines.md) | 작성 가이드라인 및 네이밍 컨벤션 |

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [CLI/09-FDL.md](../CLI/09-FDL.md) | FDL CLI 명령어 |
| [TTY/04-Phase2.md](../TTY/04-Phase2.md) | Task 실행 시 FDL 활용 |

---

*ClariSpec FDL Specification v0.0.4*
