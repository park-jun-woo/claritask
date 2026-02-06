#!/bin/bash
# Claribot self-deploy script
# claribot이 자기 자신을 배포할 때 사용
# nohup으로 실행하면 부모 프로세스가 죽어도 계속 진행
# 반드시 프로젝트 루트 디렉토리에서 실행할 것

set -e

# 스크립트 위치 기준으로 프로젝트 루트 찾기
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

LOG="/tmp/claribot-deploy.log"
BIN_SRC="$PROJECT_ROOT/bin"
BIN_DST="/usr/local/bin"

echo "[$(date)] Deploy started" > "$LOG"

# 잠시 대기 (claribot이 응답 보낼 시간)
sleep 2

echo "[$(date)] Stopping claribot..." >> "$LOG"
sudo /usr/bin/systemctl stop claribot.service

echo "[$(date)] Copying binaries..." >> "$LOG"
sudo /usr/bin/cp "$BIN_SRC/claribot" "$BIN_DST/claribot"
sudo /usr/bin/cp "$BIN_SRC/clari" "$BIN_DST/clari"

echo "[$(date)] Starting claribot..." >> "$LOG"
sudo /usr/bin/systemctl start claribot.service

echo "[$(date)] Deploy completed" >> "$LOG"

# 결과 확인
sleep 2
if systemctl is-active --quiet claribot.service; then
    echo "[$(date)] Service is active" >> "$LOG"
else
    echo "[$(date)] WARNING: Service is not active!" >> "$LOG"
fi
