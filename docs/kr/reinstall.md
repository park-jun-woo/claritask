# Claribot 비밀번호 없는 재설치 가이드

## 문제

현재 `make install`은 매 단계마다 `sudo`를 사용한다:
- 바이너리 복사 (`sudo cp` → `/usr/local/bin/`)
- 서비스 파일 이동 (`sudo mv` → `/etc/systemd/system/`)
- systemctl 명령 (`sudo systemctl daemon-reload/enable/start/restart`)

Claude Code 에이전트는 `sudo` 비밀번호를 입력할 수 없어서, 이 구성에서는 에이전트가 재설치나 서비스 재시작을 할 수 없다.

## 사전 요구사항

- **Go**: 1.24+ (bot), 1.22+ (cli)
- **Node.js**: 18+ (Vite를 사용한 GUI 빌드)
- **npm**: 8+
- **systemd**: 서비스 관리에 필요
- **CGO**: 활성화 필요 (SQLite - `mattn/go-sqlite3`에 필요)

## 해결책: sudoers NOPASSWD 설정

특정 명령에 대해서만 비밀번호 없이 sudo를 실행할 수 있는 sudoers 규칙을 등록한다. 시스템 보안을 최소한으로 열면서 에이전트가 필요한 작업을 수행할 수 있게 한다.

## 초기 설정 (일회성, 사용자가 수동으로 수행)

### 1단계: sudoers 파일 생성

```bash
sudo visudo -f /etc/sudoers.d/claribot
```

아래 내용을 입력한다. `<username>`을 실제 사용자 이름으로 바꿔라:

```
# Claribot service management - no password required
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl daemon-reload
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl start claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl stop claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl restart claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl enable claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl disable claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/systemctl status claribot.service
<username> ALL=(root) NOPASSWD: /bin/cp * /usr/local/bin/claribot
<username> ALL=(root) NOPASSWD: /bin/cp * /usr/local/bin/clari
<username> ALL=(root) NOPASSWD: /bin/chmod +x /usr/local/bin/claribot
<username> ALL=(root) NOPASSWD: /bin/chmod +x /usr/local/bin/clari
<username> ALL=(root) NOPASSWD: /bin/mv /tmp/claribot.service /etc/systemd/system/claribot.service
<username> ALL=(root) NOPASSWD: /bin/rm -f /usr/local/bin/claribot
<username> ALL=(root) NOPASSWD: /bin/rm -f /usr/local/bin/clari
<username> ALL=(root) NOPASSWD: /bin/rm -f /etc/systemd/system/claribot.service
<username> ALL=(root) NOPASSWD: /usr/bin/journalctl -u claribot.service *
```

### 2단계: 권한 확인

```bash
# 파일 권한이 올바르게 설정되었는지 확인
ls -la /etc/sudoers.d/claribot
# 예상: -r--r----- 1 root root ...

# visudo로 문법 검증 (visudo -f 사용 시 이미 검증되지만 재확인)
sudo visudo -c
```

### 3단계: 테스트

```bash
# 비밀번호 없이 명령이 실행되는지 테스트
sudo systemctl status claribot.service
sudo cp /dev/null /dev/null  # 간단한 cp 테스트 (실제 복사 아님)
```

비밀번호 프롬프트 없이 명령이 실행되면 설정 완료.

## 설정 후 에이전트가 사용 가능한 작업

sudoers 설정이 완료되면, Claude Code 에이전트가 비밀번호 없이 다음을 실행할 수 있다:

| 작업 | 명령어 |
|------|--------|
| 빌드 | `make build` (sudo 불필요) |
| CLI 설치 | `make install-cli` |
| Bot 설치 | `make install-bot` |
| 전체 설치 | `make install` |
| 서비스 재시작 | `make restart` |
| 서비스 상태 | `make status` |
| 서비스 로그 | `make logs` |
| 전체 제거 | `make uninstall` |

기존 Makefile 수정 불필요. sudoers 규칙 덕분에 Makefile의 `sudo` 명령이 비밀번호 없이 통과한다.

### 개발 타겟

| 작업 | 명령어 |
|------|--------|
| GUI 개발 서버 | `make dev-gui` (Vite HMR + API 프록시 → 127.0.0.1:9847) |
| Bot 로컬 실행 | `make run-bot` |
| CLI 로컬 실행 | `make run-cli` |
| 테스트 실행 | `make test` |
| 빌드 결과물 삭제 | `make clean` |
| 도움말 | `make help` |

## 빌드 프로세스

`make build` 명령은 세 단계를 순서대로 실행한다:

```
1. build-gui   → cd gui && npm install && npm run build (tsc -b && vite build → dist/)
                → rm -rf bot/internal/webui/dist
                → cp -r gui/dist bot/internal/webui/dist (Go embed 디렉토리에 복사)
2. build-cli   → cd cli && go build -o ../bin/clari ./cmd/clari
3. build-bot   → cd bot && go build -o ../bin/claribot ./cmd/claribot (GUI dist 임베드)
```

GUI는 Go의 `embed` 패키지를 통해 bot 바이너리에 포함된다 (`bot/internal/webui/webui.go`의 `//go:embed dist/*`). 최종 `claribot` 바이너리가 외부 파일 없이 Web UI를 제공한다.

## 배포 스크립트 (셀프 배포)

에이전트가 직접 배포를 트리거할 때, 배포 스크립트(`deploy/claribot-deploy.sh`)가 `nohup`으로 stop→copy→start 사이클을 처리한다. 부모 프로세스(claribot)가 중지되어도 계속 진행된다:

```bash
make build && nohup deploy/claribot-deploy.sh > /tmp/deploy.log 2>&1 &
```

스크립트 동작:
1. 2초 대기 (claribot이 응답을 보낼 시간)
2. claribot 서비스 중지 (`systemctl stop`)
3. 두 바이너리 (`claribot`, `clari`) 모두 `/usr/local/bin/`에 복사
4. 서비스 시작 (`systemctl start`)
5. 2초 대기 후 서비스 활성 상태 확인

배포 로그: `/tmp/claribot-deploy.log`

## 설정

예제 설정 파일을 `~/.claribot/`에 복사:

```bash
cp deploy/config.example.yaml ~/.claribot/config.yaml
```

`config.yaml` 주요 설정:

| 섹션 | 키 | 기본값 | 설명 |
|------|-----|--------|------|
| service | host | 127.0.0.1 | HTTP 수신 주소 |
| service | port | 9847 | HTTP 수신 포트 |
| telegram | token | - | @BotFather에서 받은 봇 토큰 |
| telegram | allowed_users | [] | 허용할 텔레그램 사용자 ID |
| claude | timeout | 1200 | 유휴 타임아웃 (초) |
| claude | max_timeout | 1800 | 절대 타임아웃 (초, 범위: 60-7200) |
| claude | max | 10 | 최대 동시 Claude 인스턴스 수 |
| project | path | ~/projects | 새 프로젝트 기본 경로 |
| pagination | page_size | 10 | 페이지당 항목 수 (최대: 100) |
| log | level | info | 로그 레벨 (debug, info, warn, error) |
| log | file | ~/.claribot/claribot.log | 로그 파일 경로 (빈 값 = stdout만) |

## 서비스 템플릿

systemd 서비스 파일은 `deploy/claribot.service.template`에서 생성된다:

```ini
[Unit]
Description=Claribot - LLM Project Automation Service
After=network.target

[Service]
Type=simple
User=__USER__
WorkingDirectory=__HOME__/.claribot
ExecStart=/usr/local/bin/claribot
Restart=on-failure
RestartSec=5
Environment=HOME=__HOME__
Environment=PATH=__HOME__/.local/bin:/usr/local/bin:/usr/bin:/bin

[Install]
WantedBy=multi-user.target
```

`make install-bot` 실행 시 `__USER__`와 `__HOME__`이 `sed`를 통해 실제 값으로 치환된다.

## 보안 고려사항

### 허용 범위

- claribot 서비스 관련 systemctl 명령만 허용 (다른 서비스 제어 불가)
- 바이너리 복사 경로는 `/usr/local/bin/claribot`과 `/usr/local/bin/clari`로 제한
- 서비스 파일 이동 경로는 `/etc/systemd/system/claribot.service`로 제한

### 허용하지 않는 것

- 다른 시스템 서비스 제어
- 임의 경로로 파일 복사
- 패키지 설치/제거 (apt, yum 등)
- 사용자 관리
- 기타 시스템 관리 명령

### 제거

더 이상 필요하지 않을 때:

```bash
sudo rm /etc/sudoers.d/claribot
```

## 참고: systemctl 경로 확인

배포판에 따라 systemctl, cp, chmod, mv, rm 경로가 다를 수 있다:

```bash
which systemctl  # /usr/bin/systemctl 또는 /bin/systemctl
which cp         # /usr/bin/cp 또는 /bin/cp
which chmod      # /usr/bin/chmod 또는 /bin/chmod
which mv         # /usr/bin/mv 또는 /bin/mv
which rm         # /usr/bin/rm 또는 /bin/rm
which journalctl # /usr/bin/journalctl 또는 /bin/journalctl
```

sudoers 파일의 경로는 실제 시스템에 맞게 조정해야 한다. 이 문서는 Ubuntu/Debian 기준으로 작성되었다.

## 자동 설정 스크립트

위 단계를 한 번에 수행하는 스크립트가 제공된다. **사용자는 이 스크립트를 한 번만 수동으로 실행하면 된다**:

```bash
# 사용법: sudo bash deploy/setup-sudoers.sh
# 제거법: sudo rm /etc/sudoers.d/claribot
```

스크립트는 시스템의 명령어 경로를 자동 감지하여 sudoers 파일을 생성한다. 자세한 내용은 `deploy/setup-sudoers.sh` 파일을 직접 참조하라.

## deploy 디렉토리 내용

| 파일 | 설명 |
|------|------|
| `claribot-deploy.sh` | 셀프 배포 스크립트 (nohup, stop→copy→start) |
| `claribot.service.template` | systemd 서비스 템플릿 |
| `setup-sudoers.sh` | 비밀번호 없는 sudo 설정 스크립트 |
| `config.example.yaml` | 예제 설정 파일 |
| `logo.png` | 텔레그램 봇 프로필 이미지 |

## 운영 흐름

```
[일회성] 사용자 실행: sudo bash deploy/setup-sudoers.sh
                     cp deploy/config.example.yaml ~/.claribot/config.yaml
                     (config.yaml에 텔레그램 토큰 등 설정)
  |
[최초 설치] make install
  |
[이후] Claude Code 에이전트가 자유롭게 실행:
  make build && nohup deploy/claribot-deploy.sh > /tmp/deploy.log 2>&1 &
  또는
  make build -> make install -> make restart
  또는
  make uninstall -> make install
```
