.PHONY: all build clean backend x86 arm

# Binary names
MAIN_NAME   := main

# Directories
BUILD_DIR   := ./build
ASSETS      := i18n config.json

# Go compiler path
GO          := /snap/bin/go

# Build flags
GO_BUILD_FLAGS := -ldflags="-s -w" -trimpath

all: x86 arm

# x86-64 build
x86:
	@echo ">>> Building x86-64 version..."
	@rm -rf $(BUILD_DIR)/x86
	@mkdir -p $(BUILD_DIR)/x86/bin

	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/x86/bin/$(MAIN_NAME) main.go
	chmod +x $(BUILD_DIR)/x86/bin/$(MAIN_NAME)

	@echo ">>> Copying assets to x86..."
	@for f in $(ASSETS); do \
		if [ -e $$f ]; then cp -r $$f $(BUILD_DIR)/x86/; fi; \
	done
	
	@mkdir -p $(BUILD_DIR)/x86/tmp $(BUILD_DIR)/x86/data $(BUILD_DIR)/x86/logs
	
	@echo ">>> x86-64 build complete: $(BUILD_DIR)/x86/"

# ARM64 build
arm:
	@echo ">>> Building ARM64 version..."
	@rm -rf $(BUILD_DIR)/arm
	@mkdir -p $(BUILD_DIR)/arm/bin

	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/arm/bin/$(MAIN_NAME) main.go
	chmod +x $(BUILD_DIR)/arm/bin/$(MAIN_NAME)

	@echo ">>> Copying assets to ARM64..."
	@for f in $(ASSETS); do \
		if [ -e $$f ]; then cp -r $$f $(BUILD_DIR)/arm/; fi; \
	done
	
	@mkdir -p $(BUILD_DIR)/arm/tmp $(BUILD_DIR)/arm/data $(BUILD_DIR)/arm/logs
	
	@echo ">>> ARM64 build complete: $(BUILD_DIR)/arm/"

# Package builds into tar.gz
backend: x86 arm
	@echo ">>> Packaging builds..."
	
	@cd $(BUILD_DIR)/x86 && COPYFILE_DISABLE=1 tar -czf ../dbmanager_x86-64.tar.gz *
	@echo ">>> Created: $(BUILD_DIR)/dbmanager_x86-64.tar.gz"
	
	@cd $(BUILD_DIR)/arm && COPYFILE_DISABLE=1 tar -czf ../dbmanager_arm64.tar.gz *
	@echo ">>> Created: $(BUILD_DIR)/dbmanager_arm64.tar.gz"
	
	@echo ">>> Packaging complete."

# Quick local build (current architecture)
build:
	@echo ">>> Building for current platform..."
	@mkdir -p $(BUILD_DIR)/local/bin
	$(GO) build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/local/bin/$(MAIN_NAME) main.go
	chmod +x $(BUILD_DIR)/local/bin/$(MAIN_NAME)
	@echo ">>> Local build: $(BUILD_DIR)/local/bin/$(MAIN_NAME)"

clean:
	@echo ">>> Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo ">>> Clean complete."

# Show build info
info:
	@echo "Go version: $$($(GO) version)"
	@echo "Build targets: x86-64, ARM64"
	@echo "Output: $(BUILD_DIR)/{x86,arm}/"
	@echo "UPX available: $$(command -v upx >/dev/null 2>&1 && echo 'Yes' || echo 'No')"
