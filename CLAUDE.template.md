# {PROJECT_NAME}

## Overview

{프로젝트에 대한 1-2문장 설명. 무엇을 하는 프로젝트인지 명확하게.}

## Tech Stack

- **Language**: {Go 1.24+ / Python 3.12+ / TypeScript 등}
- **Framework**: {Gin / React / Cobra 등}
- **Database**: {PostgreSQL / SQLite / 없음}
- **Infra**: {Docker / systemd / 없음}
- **Build**: {Makefile / npm / 없음}

## Project Structure

```
{project-id}/
├── cmd/              # 진입점
├── internal/         # 내부 로직
├── ...
└── go.mod
```

## Architecture

{레이어 구조, 데이터 흐름 등 핵심 아키텍처 설명}

```
{다이어그램 또는 흐름도}
```

## Core Logic

{프로젝트의 핵심 동작 원리 3-5줄 설명}

## Coding Conventions

### Style
- `gofmt` 스타일 준수 (Go 프로젝트의 경우)
- 에러는 즉시 처리 (early return)
- 인터페이스는 사용처에서 정의

### Naming
- 파일명: snake_case (`task_service.go`)
- 변수/함수: camelCase (`taskService`)
- 타입/상수: PascalCase (`TaskService`)
- 약어는 대문자 유지 (`ID`, `URL`, `JSON`)

### Layout (Go 프로젝트)
- `cmd/앱명/main.go` - 진입점
- `pkg/` - 공개 라이브러리
- `internal/` - 내부 전용

## Build & Run

```bash
# 빌드
{빌드 명령어}

# 실행
{실행 명령어}

# 테스트
{테스트 명령어}
```

## Configuration

{설정 파일 위치 및 주요 설정 항목 설명}

## Claribot Integration

이 프로젝트는 [Claribot](https://github.com/{user}/{repo})으로 관리됩니다.

### Task 시스템
- Task는 `clari task` 명령어로 관리 (add, list, get, set, delete)
- Task 상태: `todo` → `planned` → `running` → `done` (또는 `split`)
- Task 분할: 복잡한 작업은 sub task로 분할 (MECE 원칙)

### Spec 관리
- 요구사항 명세서는 **파일이 아닌 DB**에 등록: `clari spec add <title> --content-file <path>`
- 기존 스펙 수정: `clari spec set <id> content --content-file <path>`
- 폴더에 md 파일 직접 생성 금지

### 보고서 규칙
- 작업 완료 시 `.claribot/` 경로에 report 파일 자동 저장
- 보고서는 간결하게 (텔레그램 전송용)

### 금지 사항
- `systemctl stop claribot` 실행 금지
- 배포/재시작은 사용자가 직접 수행 (코드 수정만 수행)

## Deploy

```bash
# 배포 명령어 (사용자가 직접 실행)
{배포 명령어}
```

## Version

- 현재 버전: v0.1.0
- 버전 규칙: vX.X.N (수정 시 N만 증가, 10 초과 허용)

---

<!--
## 템플릿 사용 가이드

### 필수 섹션 (반드시 채워야 함)
1. **Overview**: 프로젝트가 무엇인지 1-2문장
2. **Tech Stack**: 사용 기술 나열
3. **Project Structure**: 디렉토리 구조
4. **Coding Conventions**: 코드 스타일 규칙
5. **Build & Run**: 빌드/실행 방법

### 선택 섹션 (프로젝트 규모에 따라)
- **Architecture**: 중규모 이상 프로젝트
- **Core Logic**: 핵심 알고리즘이 있는 경우
- **Configuration**: 설정 파일이 있는 경우
- **Deploy**: 배포 절차가 있는 경우

### 작성 팁
- Claude Code가 읽는 파일이므로 AI가 이해하기 쉽게 작성
- 구체적인 파일 경로, 함수명, 설정값을 포함
- 프로젝트 고유의 규칙/제약사항을 명시
- 외부 서비스 연동 정보 (API 키 경로, 엔드포인트 등)
- DB 스키마 요약 또는 ERD 링크
- 참고 문서 링크 (docs/ 하위)

### Claribot 프로젝트에서의 CLAUDE.md 역할
- Claude Code가 프로젝트 진입 시 가장 먼저 읽는 파일
- Task Planning/Running 시 프로젝트 컨텍스트로 활용
- 프로젝트별 코딩 규칙과 구조를 AI에게 전달하는 유일한 수단
-->
