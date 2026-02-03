.PHONY: build install uninstall test

BINARY_NAME := clari
INSTALL_PATH := /usr/local/bin

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/claritask
	@echo "Built $(BINARY_NAME)"

install: build
	@echo "Installing to $(INSTALL_PATH)..."
	@sudo mv $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_PATH)/$(BINARY_NAME)"

uninstall:
	@rm -f $(BINARY_NAME)
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Uninstalled $(BINARY_NAME)"

test:
	@go test ./test/... -v
