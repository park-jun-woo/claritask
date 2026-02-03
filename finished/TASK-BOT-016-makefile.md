# TASK-BOT-016: Makefile

## 목표
빌드 스크립트 작성

## 파일
`bot/Makefile`

## 작업 내용

### Makefile 내용
```makefile
.PHONY: build install clean test run

BINARY=claribot
BUILD_DIR=bin
GO=go

# 버전 정보
VERSION?=0.0.1
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/claribot

install: build
	@echo "Installing $(BINARY) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	$(GO) test -v ./...

run: build
	@echo "Running $(BINARY)..."
	./$(BUILD_DIR)/$(BINARY)

# 개발용: 환경변수 파일로 실행
dev:
	@if [ -f .env ]; then \
		export $$(cat .env | xargs) && ./$(BUILD_DIR)/$(BINARY); \
	else \
		echo "Error: .env file not found"; \
		exit 1; \
	fi

# 의존성 정리
tidy:
	$(GO) mod tidy

# 린트
lint:
	golangci-lint run ./...
```

## 완료 조건
- [ ] Makefile 작성
- [ ] build, install, clean, test, run 타겟
- [ ] 버전 정보 주입
