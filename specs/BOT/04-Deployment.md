# Claribot Deployment

> **현재 버전**: v0.0.1

---

## 배포 방식

| 방식 | 적합한 환경 | 복잡도 |
|------|------------|--------|
| systemctl | Linux 서버 (권장) | 낮음 |
| Docker | 컨테이너 환경 | 중간 |
| Docker Compose | 다중 서비스 | 중간 |

---

## 사전 요구사항

### 텔레그램 봇 생성

1. [@BotFather](https://t.me/botfather) 대화 시작
2. `/newbot` 명령어 실행
3. 봇 이름 입력: `Claribot`
4. 봇 username 입력: `your_claribot`
5. **API 토큰** 저장 (예: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`)

### 사용자 ID 확인

```bash
# @userinfobot에게 메시지 전송하여 ID 확인
# 또는 curl로 확인
curl "https://api.telegram.org/bot<TOKEN>/getUpdates"
```

---

## 방법 1: systemctl (권장)

### 1.1 빌드

```bash
cd claribot
make build
sudo cp bin/claribot /usr/local/bin/
```

### 1.2 사용자 및 디렉토리 생성

```bash
# 시스템 사용자 생성
sudo useradd -r -s /bin/false claribot

# 설정 디렉토리 생성
sudo mkdir -p /etc/claribot
sudo chown claribot:claribot /etc/claribot
```

### 1.3 환경 설정 파일

```bash
sudo vim /etc/claribot/claribot.env
```

```bash
# /etc/claribot/claribot.env

# 필수: 텔레그램 봇 토큰
TELEGRAM_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz

# 필수: 허용된 사용자 ID (쉼표 구분)
ALLOWED_USERS=123456789,987654321

# 관리자 ID (쉼표 구분)
ADMIN_USERS=123456789

# Claritask DB 경로
CLARITASK_DB=/home/youruser/.claritask/db.clt

# 알림 설정
NOTIFY_ON_COMPLETE=true
NOTIFY_ON_FAIL=true

# 로그 레벨 (debug, info, warn, error)
LOG_LEVEL=info
```

```bash
# 권한 설정 (토큰 보호)
sudo chmod 600 /etc/claribot/claribot.env
sudo chown claribot:claribot /etc/claribot/claribot.env
```

### 1.4 systemd 서비스 파일

```bash
sudo vim /etc/systemd/system/claribot.service
```

```ini
# /etc/systemd/system/claribot.service
[Unit]
Description=Claribot - Claritask Telegram Bot
Documentation=https://github.com/yourrepo/claritask
After=network.target

[Service]
Type=simple
User=claribot
Group=claribot

# 환경 변수 파일
EnvironmentFile=/etc/claribot/claribot.env

# 실행
ExecStart=/usr/local/bin/claribot
ExecReload=/bin/kill -HUP $MAINPID

# 재시작 정책
Restart=always
RestartSec=5

# 리소스 제한
MemoryMax=256M
CPUQuota=50%

# 보안 설정
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/home/youruser/.claritask
PrivateTmp=true

# 로깅
StandardOutput=journal
StandardError=journal
SyslogIdentifier=claribot

[Install]
WantedBy=multi-user.target
```

### 1.5 서비스 시작

```bash
# 서비스 파일 리로드
sudo systemctl daemon-reload

# 서비스 시작
sudo systemctl start claribot

# 부팅 시 자동 시작
sudo systemctl enable claribot

# 상태 확인
sudo systemctl status claribot
```

### 1.6 로그 확인

```bash
# 실시간 로그
sudo journalctl -u claribot -f

# 최근 100줄
sudo journalctl -u claribot -n 100

# 오늘 로그
sudo journalctl -u claribot --since today
```

### 1.7 관리 명령어

```bash
# 재시작
sudo systemctl restart claribot

# 중지
sudo systemctl stop claribot

# 설정 리로드 (HUP 시그널)
sudo systemctl reload claribot
```

---

## 방법 2: Docker

### 2.1 Dockerfile

```dockerfile
# claribot/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o claribot ./cmd/claribot

# 실행 이미지
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g '' claribot

WORKDIR /app
COPY --from=builder /app/claribot .

USER claribot
ENTRYPOINT ["./claribot"]
```

### 2.2 빌드 및 실행

```bash
# 빌드
docker build -t claribot:latest .

# 실행
docker run -d \
  --name claribot \
  --restart always \
  -e TELEGRAM_TOKEN=your-token \
  -e ALLOWED_USERS=123456789 \
  -e CLARITASK_DB=/data/db.clt \
  -v /home/youruser/.claritask:/data:ro \
  claribot:latest
```

### 2.3 Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  claribot:
    build: ./claribot
    container_name: claribot
    restart: always
    env_file:
      - ./claribot.env
    volumes:
      - claritask-data:/data:ro
    networks:
      - claritask-net
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.5'

volumes:
  claritask-data:
    external: true

networks:
  claritask-net:
```

```bash
# 시작
docker-compose up -d

# 로그
docker-compose logs -f claribot

# 재시작
docker-compose restart claribot
```

---

## 업데이트 절차

### systemctl 환경

```bash
#!/bin/bash
# update-claribot.sh

set -e

echo "Building new version..."
cd /path/to/claritask/claribot
git pull
make build

echo "Stopping service..."
sudo systemctl stop claribot

echo "Installing new binary..."
sudo cp bin/claribot /usr/local/bin/

echo "Starting service..."
sudo systemctl start claribot

echo "Checking status..."
sudo systemctl status claribot
```

### Docker 환경

```bash
#!/bin/bash
# update-claribot-docker.sh

set -e

cd /path/to/claritask

echo "Pulling latest..."
git pull

echo "Rebuilding..."
docker-compose build claribot

echo "Restarting..."
docker-compose up -d claribot

echo "Checking logs..."
docker-compose logs --tail=20 claribot
```

---

## 모니터링

### Healthcheck 엔드포인트 (선택적)

```go
// internal/bot/health.go
func (b *Bot) StartHealthServer(port int) {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
```

### systemd 서비스에 healthcheck 추가

```ini
[Service]
# ... 기존 설정 ...
ExecStartPost=/bin/sleep 5
ExecStartPost=/usr/bin/curl -sf http://localhost:8080/health || exit 1
```

### 알림 설정 (선택적)

```bash
# /etc/systemd/system/claribot-notify.service
[Unit]
Description=Claribot Failure Notification
After=claribot.service

[Service]
Type=oneshot
ExecStart=/usr/local/bin/notify-failure.sh claribot
```

---

## 백업

### DB 백업 스크립트

```bash
#!/bin/bash
# /etc/cron.daily/backup-claritask

BACKUP_DIR=/var/backups/claritask
DB_PATH=/home/youruser/.claritask/db.clt
DATE=$(date +%Y%m%d)

mkdir -p $BACKUP_DIR
cp $DB_PATH $BACKUP_DIR/db-$DATE.clt

# 7일 이상 된 백업 삭제
find $BACKUP_DIR -name "db-*.clt" -mtime +7 -delete
```

---

## 트러블슈팅

| 증상 | 원인 | 해결 |
|------|------|------|
| 봇이 응답 안함 | 토큰 오류 | 토큰 확인, 서비스 재시작 |
| 권한 오류 | DB 접근 불가 | 파일 권한 확인 |
| 메모리 초과 | 리소스 제한 | MemoryMax 증가 |
| 재시작 반복 | 설정 오류 | journalctl로 로그 확인 |

---

## 관련 문서

| 문서 | 내용 |
|------|------|
| [01-Overview.md](01-Overview.md) | 전체 개요 |
| [05-Security.md](05-Security.md) | 보안 설정 |

---

*Claribot Deployment v0.0.1*
