# Makefile for datakeg

# Binary name
BINARY_NAME=datakeg

# Build directory
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Main package path
MAIN_PATH=./cmd/datakeg

# Version information
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Check if working directory is dirty (only if not already marked dirty)
GIT_STATUS := $(shell git status --porcelain 2>/dev/null)
ifneq ($(GIT_STATUS),)
    ifeq (,$(findstring dirty,$(VERSION)))
        VERSION := $(VERSION)-dirty
    endif
endif

# Linker flags to set version information
LDFLAGS := -X 'main.version=$(VERSION)' \
           -X 'main.commit=$(COMMIT)' \
           -X 'main.date=$(DATE)' \
           -X 'main.goVersion=$(GO_VERSION)'

# Find all Go source files
GO_FILES := $(shell find . -name '*.go' -type f)

.PHONY: all clean test coverage install help deps tidy lint test-nocache

# Default target
all: $(BINARY_NAME)

## build: Build the binary (same as default target)
build: $(BINARY_NAME)

# Actual build target - rebuilds only when source files change
$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	@echo "  Version: $(VERSION)"
	@echo "  Commit:  $(COMMIT)"
	@echo "  Date:    $(DATE)"
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## clean: Remove built binaries and test cache
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Clean complete"

## test: Run all tests (uses Go's built-in test cache)
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-nocache: Run all tests without cache
test-nocache:
	@echo "Running tests without cache..."
	$(GOTEST) -v -count=1 ./...

## coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@PATH="$(HOME)/go/bin:$(PATH)" golangci-lint run ./...

## install: Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) -v ./...

## tidy: Tidy up go.mod and go.sum
tidy:
	@echo "Tidying go modules..."
	$(GOMOD) tidy

## version: Print version information
version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Date: $(DATE)"
	@echo "Go Version: $(GO_VERSION)"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
