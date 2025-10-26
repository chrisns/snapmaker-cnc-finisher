.PHONY: build test clean install fmt vet lint coverage help

# Build variables
BINARY_NAME=gcode-optimizer
VERSION?=1.0.0
BUILD_DIR=dist
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"
BUILD_FLAGS=-trimpath $(LDFLAGS)

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary for current platform
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME) ./cmd/gcode-optimizer/

build-all: ## Build binaries for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)

	@echo "Building macOS Intel..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/gcode-optimizer/

	@echo "Building macOS ARM..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/gcode-optimizer/

	@echo "Building Windows..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/gcode-optimizer/

	@echo "Building Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/gcode-optimizer/

	@echo "Done! Binaries are in $(BUILD_DIR)/"

test: ## Run tests
	$(GOTEST) -v -race ./...

coverage: ## Run tests with coverage
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

fmt: ## Format code
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

lint: fmt vet ## Run linters (fmt + vet)
	@echo "Linting complete"

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.txt coverage.html

install: build ## Install binary to /usr/local/bin
	sudo mv $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

.DEFAULT_GOAL := help
