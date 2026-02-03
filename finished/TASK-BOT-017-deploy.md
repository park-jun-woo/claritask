# TASK-BOT-017: 배포 파일

## 목표
systemd 서비스 파일 및 환경변수 예시 작성

## 파일
- `bot/deploy/claribot.service`
- `bot/deploy/claribot.env.example`

## 작업 내용

### claribot.service
```ini
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
PrivateTmp=true

# 로깅
StandardOutput=journal
StandardError=journal
SyslogIdentifier=claribot

[Install]
WantedBy=multi-user.target
```

### claribot.env.example
```bash
# Telegram Bot Token (required)
# Get from @BotFather
TELEGRAM_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz

# Allowed Users (comma-separated Telegram user IDs)
# Get your ID from @userinfobot
ALLOWED_USERS=123456789,987654321

# Admin Users (comma-separated)
ADMIN_USERS=123456789

# Claritask Database Path
CLARITASK_DB=/home/youruser/.claritask/db.clt

# Notifications
NOTIFY_ON_COMPLETE=true
NOTIFY_ON_FAIL=true

# Rate Limiting
RATE_LIMIT=1
RATE_BURST=5

# Log Level (debug, info, warn, error)
LOG_LEVEL=info
```

### 설치 스크립트 (선택적)
```bash
#!/bin/bash
# bot/deploy/install.sh

set -e

echo "Creating claribot user..."
sudo useradd -r -s /bin/false claribot || true

echo "Creating config directory..."
sudo mkdir -p /etc/claribot
sudo chown claribot:claribot /etc/claribot

echo "Copying service file..."
sudo cp claribot.service /etc/systemd/system/

echo "Reloading systemd..."
sudo systemctl daemon-reload

echo "Done! Next steps:"
echo "1. Copy claribot.env.example to /etc/claribot/claribot.env"
echo "2. Edit /etc/claribot/claribot.env with your settings"
echo "3. sudo systemctl enable claribot"
echo "4. sudo systemctl start claribot"
```

## 완료 조건
- [ ] claribot.service 작성
- [ ] claribot.env.example 작성
- [ ] 설치 스크립트 작성 (선택적)
