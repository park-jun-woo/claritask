# Claribot Project

## Overview (v0.2)

LLM 기반 프로젝트 자동 실행 시스템. 텔레그램 봇과 CLI로 Claude Code를 제어하여 프로젝트 작업을 자동화한다.

**아키텍처**: `claribot(daemon)` ← HTTP → `clari(CLI)` / 텔레그램
- **claribot**: 항상 실행되는 서비스. 텔레그램 핸들러 + CLI 핸들러 + TTY 매니저
- **clari CLI**: 서비스에 HTTP 요청을 보내는 클라이언트 (127.0.0.1:9847)

**DB 분리**:
- 전역 DB (`~/.claribot/db.clt`): 프로젝트 목록, 경로 매핑
- 로컬 DB (`프로젝트/.claribot/db.clt`): tasks, memos, features 등 (git 관리)

**Claude Code 실행**: 2-Depth 제한
- 전역 클로드 (~/.claribot/): 메시지 분석, 프로젝트 라우팅
- 프로젝트 클로드 (프로젝트 경로): 실제 작업 수행, 더 이상 네스트 안 함

## Tech Stack

- **Language**: Go 1.21+
- **CLI Framework**: Cobra
- **Database**: SQLite (mattn/go-sqlite3)

## Project Structure

```
claribot/
├── bot/          # Claribot Go 소스코드
├── cli/          # Go CLI 소스코드
├── vsx/          # Claribot VSCode Extension 소스코드
├── docs/         # 설계 문서
└── deploy/       # 배포 설정 파일
```

## 참고 문서

- [docs/Claribot.md](docs/Claribot.md) - 아키텍처, DB 스키마, 구현 현황
- [docs/Task.md](docs/Task.md) - Task 시스템 설계 (분할 정복, 순회 로직)

## Coding Conventions

### Go Style
- `gofmt` 스타일 준수
- 에러는 즉시 처리 (early return)
- 인터페이스는 사용처에서 정의
- 패키지명은 단수형, 소문자

### Go Layout
  - `cmd/앱명/main.go` - 진입점
  - `pkg/` - 공개 라이브러리
  - `internal/` - 내부 전용

### Naming
- 파일명: snake_case (`task_service.go`)
- 변수/함수: camelCase (`taskService`)
- 타입/상수: PascalCase (`TaskService`)
- 약어는 대문자 유지 (`ID`, `URL`, `JSON`)

### CLI 명령어
- Makefile 만들때 CLI 명령어 경로는 `/usr/local/bin/`

### ID 규칙
- **Project**: 영문 소문자, 숫자, 하이픈(-), 언더스코어(_) - 예: `blog`, `api-server`
- **Feature/Task**: 정수 (auto increment)

## 버전 표기 규칙
- vX.X.N 형식이며 테스트하며 수정할때 N 숫자만 올려라. 10이 넘어도 vX.X.11로 표기하라.
