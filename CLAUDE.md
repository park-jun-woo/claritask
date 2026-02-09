# Claribot Project

## Overview (v0.2)

LLM 기반 프로젝트 자동 실행 시스템. 텔레그램 봇과 CLI로 Claude Code를 제어하여 프로젝트 작업을 자동화한다.

**아키텍처**: `claribot(daemon)` ← HTTP → `clari(CLI)` / 텔레그램 / Web UI
- **claribot**: 항상 실행되는 서비스. 텔레그램 핸들러 + CLI 핸들러 + Web UI + TTY 매니저
- **clari CLI**: 서비스에 HTTP 요청을 보내는 클라이언트 (127.0.0.1:9847)
- **Web UI**: 브라우저 기반 대시보드 (Go embed, 같은 포트)

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
- **Web UI**: React + TypeScript + Vite + shadcn/ui + TanStack Query

## Project Structure

```
claribot/
├── bot/          # Claribot Go 소스코드
├── cli/          # Go CLI 소스코드
├── gui/          # Web UI (React + TypeScript + Vite + shadcn/ui)
├── vsx/          # Claribot VSCode Extension 소스코드
├── docs/         # 설계 문서
└── deploy/       # 배포 설정 파일
```

## 참고 문서

- [docs/Claribot.md](docs/Claribot.md) - 아키텍처, DB 스키마, 구현 현황
- [docs/Task.md](docs/Task.md) - Task 시스템 설계 (분할 정복, 순회 로직)
- [docs/Schedule.md](docs/Schedule.md) - Schedule 시스템 설계 (cron 기반 자동 실행)
- [docs/webui.md](docs/webui.md) - Web UI 기획 및 구현 현황
- [docs/MobileApp.md](docs/MobileApp.md) - 모바일 앱 기획 문서

> 한글 버전은 [docs/kr/](docs/kr/) 디렉토리를 참고하세요.

### 외부 프로젝트
- [../bastion/CLAUDE.md](../bastion/CLAUDE.md) - Bastion 서버 (인프라, SSH, 보안 관련 작업 시 참고)

### Bastion 서버 접속

Bastion(AWS EC2)은 Security Group이 현재 IP만 허용하므로 `bastion-secugate` 서비스가 실행 중이어야 접속 가능.

```bash
# 1. bastion-secugate 서비스 상태 확인 (SG에 현재 IP 등록)
systemctl is-active bastion-secugate

# 서비스가 inactive면 bastion 프로젝트에서 설치
cd /mnt/c/Users/mail/git/bastion && make install

# 2. SSH 접속 (~/.ssh/config에 Host bastion 정의됨)
ssh bastion

# 3. 파일 전송
scp <파일> bastion:~/
```

> **주의**: bastion IP(현재 43.201.60.23)는 인스턴스 재시작 시 변경됨.
> 변경 시 `~/.ssh/config`의 bastion HostName과 `/etc/default/bastion-tunnel`의 BASTION_IP 수정 필요.

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

## 배포

### 배포 Trigger
- 명확하게 Claribot을 배포하라고 말하면 배포한다.

```bash
# 빌드 + 배포
make build && bash deploy/claribot-deploy.sh
```

### 서비스 중지 규칙
- **서비스 중지가 필요하면 반드시 `deploy/claribot-deploy.sh`를 사용**한다. `systemctl stop`을 직접 호출하지 않는다.
- deploy.sh는 lock 파일(`/tmp/claribot-deploy.lock`)을 생성하여 watchdog과 충돌을 방지한다.
- Watchdog(`claribot-watchdog.service`)가 서비스 상태를 감시하며, 60초 이상 죽어있으면 자동 재빌드+배포한다.

## 버전 표기 규칙
- vX.X.N 형식이며 테스트하며 수정할때 N 숫자만 올려라. 10이 넘어도 vX.X.11로 표기하라.
