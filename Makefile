# Binary name
BINARY_NAME=curlgen

# Go command
GO=go

# Build directory
BUILD_DIR=build

# Platforms
PLATFORMS=linux darwin windows

# Version
VERSION=$(shell git rev-parse --short HEAD)

# Build flags
BUILD_FLAGS=-ldflags "-X main.Version=$(VERSION)"

# Debug build flags
DEBUG_BUILD_FLAGS=$(BUILD_FLAGS) -tags debug

.PHONY: all clean build build-debug build-linux build-darwin build-windows zip-linux zip-darwin zip-windows zip-all

all: clean build

clean:
	rm -rf $(BUILD_DIR)

build: clean build-linux build-darwin build-windows

build-debug: clean build-debug-linux build-debug-darwin build-debug-windows

build-linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux

build-darwin:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin

build-windows:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe

build-debug-linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(DEBUG_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux

build-debug-darwin:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO) build $(DEBUG_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin

build-debug-windows:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(DEBUG_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe

zip-linux:
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-linux.tar.gz $(BINARY_NAME)-linux

zip-darwin:
	cd $(BUILD_DIR) && tar -czf $(BINARY_NAME)-darwin.tar.gz $(BINARY_NAME)-darwin

zip-windows:
	cd $(BUILD_DIR) && zip $(BINARY_NAME)-windows.zip $(BINARY_NAME)-windows.exe

zip-all: zip-linux zip-darwin zip-windows

.PHONY: run
run:
	$(GO) run .