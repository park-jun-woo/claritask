.PHONY: all build build-cli build-bot build-gui build-bridge clean install uninstall install-cli install-bot uninstall-cli uninstall-bot install-watchdog uninstall-watchdog

# 변수
BIN_DIR = /usr/local/bin
SYSTEMD_DIR = /etc/systemd/system
SERVICE_NAME = claribot.service
WATCHDOG_NAME = claribot-watchdog.service
PROJECT_DIR = $(shell pwd)
USER = $(shell whoami)
HOME_DIR = $(shell echo ~)

# 기본 타겟
all: build

# 전체 빌드
build: build-gui build-bridge build-cli build-bot

# CLI 빌드
build-cli:
	@echo "Building clari CLI..."
	@mkdir -p bin
	cd cli && go build -o ../bin/clari ./cmd/clari

# Bridge 빌드 (TypeScript → dist)
build-bridge:
	@echo "Building Agent Bridge..."
	cd bot/bridge && npm install && npm run build

# GUI 빌드 (React → dist → Go embed)
build-gui:
	@echo "Building Web UI..."
	cd gui && npm install && npm run build
	@echo "Copying dist to Go embed..."
	rm -rf bot/internal/webui/dist
	cp -r gui/dist bot/internal/webui/dist

# Bot 빌드 (GUI embed 포함)
build-bot:
	@echo "Building claribot..."
	@mkdir -p bin
	cd bot && go build -o ../bin/claribot ./cmd/claribot

# 정리
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf gui/dist
	rm -rf bot/internal/webui/dist
	rm -rf bot/bridge/dist
	rm -rf bot/bridge/node_modules

# 전체 설치
install: build install-cli install-bot
	@echo "Installation complete!"

# CLI 설치 (전역 명령어)
install-cli: build-cli
	@echo "Installing clari to $(BIN_DIR)..."
	sudo cp bin/clari $(BIN_DIR)/clari
	sudo chmod +x $(BIN_DIR)/clari
	@echo "clari installed successfully!"

# Bot 설치 (systemd 서비스)
install-bot: build-bot
	@echo "Installing claribot service..."
	@mkdir -p $(HOME_DIR)/.claribot
	sudo cp bin/claribot $(BIN_DIR)/claribot
	sudo chmod +x $(BIN_DIR)/claribot
	@echo "Creating systemd service file..."
	@sed -e 's|__USER__|$(USER)|g' -e 's|__HOME__|$(HOME_DIR)|g' deploy/claribot.service.template > /tmp/$(SERVICE_NAME)
	sudo mv /tmp/$(SERVICE_NAME) $(SYSTEMD_DIR)/$(SERVICE_NAME)
	sudo systemctl daemon-reload
	sudo systemctl enable $(SERVICE_NAME)
	sudo systemctl start $(SERVICE_NAME)
	@echo "claribot service installed and started!"

# 전체 제거
uninstall: uninstall-bot uninstall-cli
	@echo "Uninstallation complete!"

# CLI 제거
uninstall-cli:
	@echo "Uninstalling clari..."
	-sudo rm -f $(BIN_DIR)/clari
	@echo "clari uninstalled!"

# Bot 제거
uninstall-bot:
	@echo "Stopping and removing claribot service..."
	-sudo systemctl stop $(SERVICE_NAME)
	-sudo systemctl disable $(SERVICE_NAME)
	-sudo rm -f $(SYSTEMD_DIR)/$(SERVICE_NAME)
	-sudo rm -f $(BIN_DIR)/claribot
	sudo systemctl daemon-reload
	@echo "claribot service removed!"

# Watchdog 설치
install-watchdog:
	@echo "Installing claribot-watchdog service..."
	chmod +x deploy/claribot-watchdog.sh
	@sed -e 's|__USER__|$(USER)|g' -e 's|__HOME__|$(HOME_DIR)|g' -e 's|__PROJECT__|$(PROJECT_DIR)|g' deploy/claribot-watchdog.service.template > /tmp/$(WATCHDOG_NAME)
	sudo mv /tmp/$(WATCHDOG_NAME) $(SYSTEMD_DIR)/$(WATCHDOG_NAME)
	sudo systemctl daemon-reload
	sudo systemctl enable $(WATCHDOG_NAME)
	sudo systemctl start $(WATCHDOG_NAME)
	@echo "claribot-watchdog installed and started!"

# Watchdog 제거
uninstall-watchdog:
	@echo "Stopping and removing claribot-watchdog service..."
	-sudo systemctl stop $(WATCHDOG_NAME)
	-sudo systemctl disable $(WATCHDOG_NAME)
	-sudo rm -f $(SYSTEMD_DIR)/$(WATCHDOG_NAME)
	sudo systemctl daemon-reload
	@echo "claribot-watchdog removed!"

# 서비스 상태 확인
status:
	sudo systemctl status $(SERVICE_NAME)

# 서비스 재시작
restart:
	sudo systemctl restart $(SERVICE_NAME)

# 로그 확인
logs:
	sudo journalctl -u $(SERVICE_NAME) -f

# 개발용 로컬 실행
run-bot:
	cd bot && go run ./cmd/claribot

run-cli:
	cd cli && go run ./cmd/clari

# GUI 개발 서버 (Vite HMR + API proxy → 127.0.0.1:9847)
dev-gui:
	cd gui && npm run dev

# 테스트
test:
	cd cli && go test ./...
	cd bot && go test ./...

# 도움말
help:
	@echo "Claribot Makefile"
	@echo ""
	@echo "빌드:"
	@echo "  make build        - GUI, Bridge, CLI, Bot 전체 빌드"
	@echo "  make build-cli    - CLI만 빌드"
	@echo "  make build-bot    - Bot 빌드 (GUI embed 포함)"
	@echo "  make build-gui    - GUI만 빌드 (npm)"
	@echo "  make build-bridge - Agent Bridge 빌드 (TypeScript)"
	@echo "  make clean        - 빌드 결과물 삭제"
	@echo ""
	@echo "설치:"
	@echo "  make install      - 전체 설치 (CLI + Bot 서비스)"
	@echo "  make install-cli  - CLI만 설치 (/usr/local/bin/clari)"
	@echo "  make install-bot      - Bot 서비스 설치 (systemd)"
	@echo "  make install-watchdog - Watchdog 설치 (서비스 죽으면 자동 재빌드+배포)"
	@echo ""
	@echo "제거:"
	@echo "  make uninstall          - 전체 제거"
	@echo "  make uninstall-cli      - CLI만 제거"
	@echo "  make uninstall-bot      - Bot 서비스 제거"
	@echo "  make uninstall-watchdog - Watchdog 제거"
	@echo ""
	@echo "서비스 관리:"
	@echo "  make status       - 서비스 상태 확인"
	@echo "  make restart      - 서비스 재시작"
	@echo "  make logs         - 서비스 로그 확인"
	@echo ""
	@echo "개발:"
	@echo "  make run-bot      - Bot 로컬 실행"
	@echo "  make run-cli      - CLI 로컬 실행"
	@echo "  make dev-gui      - GUI 개발 서버 (Vite HMR)"
	@echo "  make test         - 테스트 실행"
