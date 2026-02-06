# Claribot

LLM 기반 프로젝트 자동화 시스템. Telegram 봇과 CLI로 Claude Code를 제어하여 프로젝트 작업을 자동화한다.

## Features

- **Telegram 봇 인터페이스**: 모바일에서 프로젝트 관리 및 Claude 실행
- **CLI 클라이언트**: 터미널에서 동일한 기능 사용
- **다중 프로젝트 관리**: 프로젝트별 독립 DB로 작업/태스크 관리
- **Claude Code 연동**: PTY 기반 Claude Code 실행 및 결과 반환
- **Task 기반 워크플로우**: 메시지 → Task 변환 → 실행 → 결과 보고
- **Edge 그래프**: Task 간 의존성 관리
- **Cron 스케줄러**: 지정 시간에 자동 Claude 실행 및 결과 알림

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    claribot (daemon)                    │
│                                                         │
│  ┌───────────┐  ┌───────────┐  ┌───────────────────┐   │
│  │ Telegram  │  │    CLI    │  │    Scheduler      │   │
│  │  Handler  │  │  Handler  │  │     (cron)        │   │
│  └─────┬─────┘  └─────┬─────┘  └─────────┬─────────┘   │
│        │              │                   │             │
│        └──────────────┼───────────────────┘             │
│                       ▼                                 │
│               ┌──────────────┐                          │
│               │ Router/Claude│                          │
│               └──────────────┘                          │
└─────────────────────────────────────────────────────────┘
         ▲                              ▲
         │ Bot API                      │ HTTP POST
    [Telegram]                     [clari CLI]
```

| 컴포넌트 | 역할 |
|----------|------|
| claribot | Telegram + CLI 핸들러, Claude 세션 관리 (systemd 서비스) |
| clari | HTTP 클라이언트 CLI |
| Claude Code | 프로젝트 폴더에서 실제 작업 수행 |

## Requirements

- Go 1.21+
- Claude Code CLI (`~/.local/bin/claude`)
- SQLite3

## Installation

```bash
# Clone
git clone https://github.com/your-username/claribot.git
cd claribot

# Build and install
make install
```

`make install`은 다음을 수행한다:
- `clari` CLI를 `/usr/local/bin/`에 설치
- `claribot` 서비스를 systemd에 등록 및 시작

### Manual Installation

```bash
# Build only
make build

# Install CLI only
make install-cli

# Install service only
make install-bot
```

## Configuration

설정 파일: `~/.claribot/config.yaml`

```yaml
# HTTP Service
service:
  host: 127.0.0.1
  port: 9847

# Telegram Bot
telegram:
  token: "YOUR_BOT_TOKEN"    # @BotFather에서 발급
  allowed_users: []          # 빈 배열 = 모두 허용, [123456789] = 특정 유저만
  admin_chat_id: 0           # 스케줄 실행 결과 알림 대상 (0 = 비활성화)

# Claude Code
claude:
  timeout: 1200              # 유휴 타임아웃 (초)
  max: 3                     # 최대 동시 실행 수

# Project
project:
  path: ~/projects           # 프로젝트 생성 기본 경로

# Pagination
pagination:
  page_size: 10

# Logging
log:
  level: info                # debug, info, warn, error
  file: ~/.claribot/claribot.log
```

예제 파일: `deploy/config.example.yaml`

## Usage

### Service Management

```bash
make status     # 서비스 상태 확인
make restart    # 서비스 재시작
make logs       # 로그 확인 (journalctl)
```

### CLI Commands

```bash
# 프로젝트 관리
clari project list              # 프로젝트 목록
clari project create <id>       # 새 프로젝트 생성
clari project add <path>        # 기존 폴더를 프로젝트로 등록
clari project switch <id>       # 프로젝트 선택
clari project delete <id>       # 프로젝트 삭제

# 태스크 관리
clari task list                 # 태스크 목록
clari task add <title>          # 태스크 추가
clari task get <id>             # 태스크 상세
clari task run [id]             # 태스크 실행

# 메시지 (Claude 실행)
clari send "코드 리뷰해줘"       # 메시지 전송 → Claude 실행
clari message list              # 메시지 기록
clari message status            # 메시지 상태 요약

# 스케줄 관리
clari schedule list             # 스케줄 목록
clari schedule add "0 7 * * *" "아침 인사"  # 스케줄 추가
clari schedule add --once "30 14 * * *" "1회 알림"  # 1회 실행
clari schedule get <id>         # 스케줄 상세
clari schedule enable <id>      # 활성화
clari schedule disable <id>     # 비활성화
clari schedule delete <id>      # 삭제
clari schedule runs <id>        # 실행 기록
clari schedule set project <id> <project>  # 프로젝트 변경

# 상태
clari status                    # 현재 프로젝트 상태
```

### Telegram Commands

| 명령어 | 설명 |
|--------|------|
| `/start` | 봇 시작, 메뉴 키보드 표시 |
| `/project` | 프로젝트 목록 (선택 버튼) |
| `/task` | 태스크 목록 |
| `/status` | 현재 상태 |
| 일반 메시지 | 선택된 프로젝트에서 Claude 실행 |

## Project Structure

```
claribot/
├── bot/                    # claribot 서비스
│   ├── cmd/claribot/       # 진입점
│   ├── internal/           # 내부 패키지
│   │   ├── config/         # 설정 로드
│   │   ├── db/             # SQLite 래퍼
│   │   ├── handler/        # 명령어 라우터
│   │   ├── project/        # 프로젝트 관리
│   │   ├── task/           # 태스크 관리
│   │   ├── message/        # 메시지 처리
│   │   ├── schedule/       # 스케줄 관리
│   │   ├── edge/           # 태스크 의존성
│   │   ├── prompts/        # 시스템 프롬프트 템플릿
│   │   └── tghandler/      # Telegram 핸들러
│   └── pkg/                # 공개 패키지
│       ├── claude/         # Claude Code 실행
│       ├── telegram/       # Telegram Bot API
│       ├── render/         # Markdown → HTML
│       ├── logger/         # 로깅
│       └── errors/         # 에러 타입
├── cli/                    # clari CLI
│   └── cmd/clari/
├── deploy/                 # 배포 파일
│   ├── claribot.service.template
│   └── config.example.yaml
└── Makefile
```

## Database

### Global DB (`~/.claribot/db.clt`)

프로젝트, 스케줄, 메시지 관리

```sql
projects (
    id TEXT PRIMARY KEY,
    name TEXT,
    path TEXT UNIQUE,
    type TEXT,
    description TEXT,
    status TEXT,
    created_at, updated_at
)

schedules (
    id INTEGER PRIMARY KEY,
    project_id TEXT,          -- NULL이면 전역 실행
    cron_expr TEXT,
    message TEXT,
    enabled INTEGER,
    run_once INTEGER,         -- 1회 실행 후 자동 비활성화
    last_run TEXT,
    next_run TEXT,
    created_at, updated_at
)

schedule_runs (
    id INTEGER PRIMARY KEY,
    schedule_id INTEGER,
    status TEXT,              -- 'running', 'done', 'failed'
    result TEXT,
    error TEXT,
    started_at, completed_at
)

messages (
    id INTEGER PRIMARY KEY,
    project_id TEXT,          -- NULL이면 전역 실행
    content TEXT,
    source TEXT,              -- 'telegram', 'cli', 'schedule'
    status TEXT,
    result TEXT,
    error TEXT,
    created_at, completed_at
)
```

### Local DB (`<project>/.claribot/db.clt`)

프로젝트별 태스크 (git으로 관리 가능)

```sql
tasks (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER,
    title TEXT,
    spec TEXT,                -- 요구사항 명세서
    plan TEXT,                -- 실행 계획서
    report TEXT,              -- 완료 보고서
    status TEXT,              -- 'todo', 'planned', 'split', 'done', 'failed'
    error TEXT,
    created_at, updated_at
)

task_edges (
    from_task_id INTEGER,
    to_task_id INTEGER,
    created_at
)
```

## Development

```bash
# 로컬 실행
make run-bot    # claribot 실행
make run-cli    # CLI 실행

# 테스트
make test

# 클린 빌드
make clean && make build
```

## Uninstall

```bash
make uninstall
```

## Disclaimer

이 프로젝트는 Anthropic의 [Claude Code](https://claude.ai/claude-code) CLI를 필요로 한다.

Claribot은 Claude Code를 subprocess로 호출하는 래퍼 프로그램이다. Claude Code 자체를 포함하거나 재배포하지 않으며, 사용자는 별도로 Claude Code를 설치하고 Anthropic 계정을 보유해야 한다.

**사용자 책임:**
- 사용자는 [Anthropic 이용약관](https://www.anthropic.com/legal)을 준수할 책임이 있다
- Consumer 플랜(Free/Pro/Max)의 자동화 사용은 약관에 따라 제한될 수 있다
- 상업적 사용 시 [Commercial Terms](https://www.anthropic.com/legal/commercial-terms) 확인을 권장한다

이 프로젝트의 개발자는 사용자의 Anthropic 약관 위반에 대해 책임지지 않는다.

## License

MIT License - see [LICENSE](LICENSE)
