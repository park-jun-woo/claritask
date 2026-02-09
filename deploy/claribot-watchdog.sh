#!/bin/bash
# Claribot Watchdog
# claribot 서비스가 죽어있으면 재빌드 + 배포
# 60초 grace period: 수동 배포 시간 확보

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOG="/tmp/claribot-watchdog.log"
LOCK="/tmp/claribot-deploy.lock"
CHECK_INTERVAL=15
GRACE_PERIOD=60

echo "[$(date)] Watchdog started (interval=${CHECK_INTERVAL}s, grace=${GRACE_PERIOD}s)" >> "$LOG"

DEAD_SINCE=0

while true; do
    sleep "$CHECK_INTERVAL"

    # 배포 lock이 있으면 skip + 타이머 리셋
    if [ -f "$LOCK" ]; then
        DEAD_SINCE=0
        continue
    fi

    if systemctl is-active --quiet claribot.service; then
        DEAD_SINCE=0
        continue
    fi

    # 서비스가 죽어있음
    NOW=$(date +%s)
    if [ "$DEAD_SINCE" -eq 0 ]; then
        DEAD_SINCE=$NOW
        echo "[$(date)] claribot is dead, waiting ${GRACE_PERIOD}s..." >> "$LOG"
        continue
    fi

    ELAPSED=$((NOW - DEAD_SINCE))
    if [ "$ELAPSED" -lt "$GRACE_PERIOD" ]; then
        continue
    fi

    # grace period 경과 — 재빌드 + 배포
    echo "[$(date)] Grace period passed (${ELAPSED}s), rebuilding..." >> "$LOG"
    touch "$LOCK"
    DEAD_SINCE=0

    cd "$PROJECT_ROOT"
    if make build >> "$LOG" 2>&1; then
        echo "[$(date)] Build success, deploying..." >> "$LOG"
        bash "$SCRIPT_DIR/claribot-deploy.sh" >> "$LOG" 2>&1
    else
        echo "[$(date)] Build failed, starting with existing binary..." >> "$LOG"
        sudo /usr/bin/systemctl start claribot.service 2>> "$LOG"
    fi

    rm -f "$LOCK"
done
