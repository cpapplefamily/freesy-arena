# Makefile with semantic versioning, build metadata, and GitHub releases

# The normal order to release a new version:
# 1.  Build, which will create a new VERSION and embed it in the binaries
# 2.  Commit and push the build 
# 3.  Create a Tag and Release from the VERSION

BINARY_NAME := cheesy-arena
VERSION_FILE := VERSION
BUILD_DIR := build
DIST_DIR := dist
BUILD_META := team4859

# Read base version (MAJOR.MINOR.PATCH)
BASE_VERSION := $(shell cut -d+ -f1 $(VERSION_FILE) 2>/dev/null || echo "0.0.0")
DATETIME := $(shell date -u +"%Y%m%d%H%M%S")
FULL_VERSION := $(BASE_VERSION)+$(BUILD_META).$(DATETIME)
RELEASE_TAG := v$(FULL_VERSION)

PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

LDFLAGS := -X main.version=$(FULL_VERSION)

.PHONY: all build tgz package release clean help bump-version

all: help ## Default target: show help

build: bump-version ## Build Go binaries for all platforms and bump patch version
	@echo "🔨 Building $(BINARY_NAME) $(FULL_VERSION) for platforms: $(PLATFORMS)"
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform##*/}; \
		EXT=""; \
		if [ "$$OS" = "windows" ]; then EXT=".exe"; fi; \
		OUT=$(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH$$EXT; \
		echo "-> $$OUT"; \
		GOOS=$$OS GOARCH=$$ARCH go build -ldflags "$(LDFLAGS)" -o $$OUT .; \
	done

archive: ## Create .tar.gz or .zip archives depending on platform
	@mkdir -p $(DIST_DIR)
	@echo "📦 Creating platform-specific archives..."
	@for bin in $(BUILD_DIR)/$(BINARY_NAME)-*; do \
		filename=$$(basename $$bin); \
		platform=$${filename#$(BINARY_NAME)-}; \
		base=$${filename%.*}; \
		ext=$${filename##*.}; \
		TMPDIR=$$(mktemp -d); \
		cp $$bin $$TMPDIR/$(BINARY_NAME); \
		if echo $$platform | grep -q '^windows'; then \
			zip -j $(DIST_DIR)/$$base.zip $$TMPDIR/$(BINARY_NAME); \
			echo "-> $(DIST_DIR)/$$base.zip"; \
		else \
			tar -czf $(DIST_DIR)/$$base.tar.gz -C $$TMPDIR $(BINARY_NAME); \
			echo "-> $(DIST_DIR)/$$base.tar.gz"; \
		fi; \
		rm -rf $$TMPDIR; \
	done

package: archive ## Package and release the current build

release: package ## Create GitHub release and upload .tar.gz assets
	@if ! command -v gh > /dev/null; then \
		echo "❌ GitHub CLI 'gh' not installed."; exit 1; \
	fi
	@echo "🚀 Creating GitHub release for tag $(RELEASE_TAG)"
	@gh release create $(RELEASE_TAG) $(DIST_DIR)/* \
		--title "$(BINARY_NAME) $(RELEASE_TAG)" \
		--notes "Automated release of version $(RELEASE_TAG)"

clean: ## Remove build and dist directories
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@echo "🧹 Cleaned build and dist directories"

bump-version: ## Auto-increment patch version and add build metadata
	@echo "🔁 Current version: $(BASE_VERSION)"
	@MAJOR=$$(echo $(BASE_VERSION) | cut -d. -f1); \
	 MINOR=$$(echo $(BASE_VERSION) | cut -d. -f2); \
	 PATCH=$$(echo $(BASE_VERSION) | cut -d. -f3); \
	 NEW_PATCH=$$((PATCH + 1)); \
	 NEW_VERSION="$$MAJOR.$$MINOR.$$NEW_PATCH+$(BUILD_META).$(DATETIME)"; \
	 echo "$$NEW_VERSION" > $(VERSION_FILE); \
	 echo "🔼 Bumped version to $$NEW_VERSION"

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'
