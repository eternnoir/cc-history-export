# Makefile for Claude Code History Export

# Binary name
BINARY_NAME=cc-export
MAIN_PACKAGE=./cmd/cc-export

# Build variables
GO=go
BUILD_FLAGS=-ldflags="-s -w"
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
.DEFAULT_GOAL := build

# Platform detection
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Linux)
    GOOS := linux
endif
ifeq ($(UNAME_S),Darwin)
    GOOS := darwin
endif
ifeq ($(OS),Windows_NT)
    GOOS := windows
    BINARY_NAME := $(BINARY_NAME).exe
endif

ifeq ($(UNAME_M),x86_64)
    GOARCH := amd64
endif
ifeq ($(UNAME_M),arm64)
    GOARCH := arm64
endif
ifeq ($(UNAME_M),aarch64)
    GOARCH := arm64
endif

# Build the binary for the current platform
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Run the application
.PHONY: run
run: build
	./$(BINARY_NAME)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run linter (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/"; \
		exit 1; \
	fi

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	$(GO) vet ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f cc-export cc-export.exe
	rm -rf dist/
	rm -f coverage.out coverage.html

# Install the binary to GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(LDFLAGS) $(MAIN_PACKAGE)

# Cross-platform builds
.PHONY: build-all
build-all: build-linux build-darwin build-windows

# Linux builds
.PHONY: build-linux
build-linux: build-linux-amd64 build-linux-arm64 build-linux-386

.PHONY: build-linux-amd64
build-linux-amd64:
	@echo "Building for Linux AMD64..."
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)

.PHONY: build-linux-arm64
build-linux-arm64:
	@echo "Building for Linux ARM64..."
	GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)

.PHONY: build-linux-386
build-linux-386:
	@echo "Building for Linux 386..."
	GOOS=linux GOARCH=386 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-386 $(MAIN_PACKAGE)

# macOS builds
.PHONY: build-darwin
build-darwin: build-darwin-amd64 build-darwin-arm64

.PHONY: build-darwin-amd64
build-darwin-amd64:
	@echo "Building for macOS AMD64..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)

.PHONY: build-darwin-arm64
build-darwin-arm64:
	@echo "Building for macOS ARM64 (Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)

# Windows builds
.PHONY: build-windows
build-windows: build-windows-amd64 build-windows-arm64 build-windows-386

.PHONY: build-windows-amd64
build-windows-amd64:
	@echo "Building for Windows AMD64..."
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

.PHONY: build-windows-arm64
build-windows-arm64:
	@echo "Building for Windows ARM64..."
	GOOS=windows GOARCH=arm64 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe $(MAIN_PACKAGE)

.PHONY: build-windows-386
build-windows-386:
	@echo "Building for Windows 386..."
	GOOS=windows GOARCH=386 $(GO) build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-386.exe $(MAIN_PACKAGE)

# Create release archives
.PHONY: release
release: clean build-all
	@echo "Creating release archives..."
	@mkdir -p dist/archives
	@cd dist && \
	for file in cc-export-*; do \
		if [ -f "$$file" ]; then \
			base=$$(basename $$file .exe); \
			if [[ "$$file" == *.exe ]]; then \
				zip "archives/$$base.zip" "$$file"; \
			else \
				tar -czf "archives/$$base.tar.gz" "$$file"; \
			fi; \
		fi; \
	done
	@echo "Release archives created in dist/archives/"

# Development helpers
.PHONY: dev
dev: deps fmt vet test build

# Quick check before committing
.PHONY: check
check: fmt vet test

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary for the current platform"
	@echo "  run            - Build and run the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  deps           - Install and tidy dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter (requires golangci-lint)"
	@echo "  vet            - Run go vet"
	@echo "  clean          - Remove build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo ""
	@echo "Cross-platform builds:"
	@echo "  build-all      - Build for all supported platforms"
	@echo "  build-linux    - Build for Linux (amd64, arm64, 386)"
	@echo "  build-darwin   - Build for macOS (amd64, arm64)"
	@echo "  build-windows  - Build for Windows (amd64, arm64, 386)"
	@echo ""
	@echo "Other targets:"
	@echo "  release        - Create release archives for all platforms"
	@echo "  dev            - Run full development cycle"
	@echo "  check          - Quick check before committing"
	@echo "  help           - Show this help message"