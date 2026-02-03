.PHONY: build install clean test run dev tidy lint

BINARY=claribot
BUILD_DIR=bin
GO=go

# Version info
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

# Development: run with .env file
dev: build
	@if [ -f .env ]; then \
		echo "Loading .env and running $(BINARY)..."; \
		set -a && . ./.env && set +a && ./$(BUILD_DIR)/$(BINARY); \
	else \
		echo "Error: .env file not found"; \
		echo "Copy deploy/claribot.env.example to .env and configure it"; \
		exit 1; \
	fi

tidy:
	$(GO) mod tidy

lint:
	golangci-lint run ./...

# Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
