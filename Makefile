.PHONY: help
help:
	@echo "=================================="
	@echo "Bookmark Management API - Makefile"
	@echo "=================================="
	@echo ""
	@echo "Development Targets:"
	@echo "  make dev              - Run development workflow (fmt, vet, test, swagger, run)"
	@echo "  make run              - Run the application"
	@echo "  make test             - Run tests with coverage"
	@echo "  make test-verbose     - Run tests with verbose output"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt              - Format code"
	@echo "  make vet              - Run go vet"
	@echo "  make lint             - Run golangci-lint (requires golangci-lint)"
	@echo "  make tidy             - Tidy Go modules"
	@echo ""
	@echo "Build Targets:"
	@echo "  make build            - Build binary for current OS"
	@echo "  make build-linux      - Build binary for Linux (amd64)"
	@echo "  make build-macos      - Build binary for macOS (arm64)"
	@echo "  make build-windows    - Build binary for Windows (amd64)"
	@echo "  make build-prod       - Build optimized production binary"
	@echo "  make release          - Create release binaries for all platforms"
	@echo ""
	@echo "Documentation:"
	@echo "  make swagger          - Generate Swagger documentation"
	@echo ""
	@echo "Maintenance:"
	@echo "  make install-tools    - Install required tools (swag, golangci-lint)"
	@echo "  make clean            - Remove generated files and binaries"
	@echo "  make clean-all        - Remove all generated files, binaries, and vendor"
	@echo ""

# Variables
APP_NAME = bookmark-management
CMD_PATH = ./cmd/api/main.go
MAIN_PACKAGE = github.com/huypham67/bookmark-management
BIN_DIR = ./bin
DOCS_DIR = ./docs

# Coverage variables
COVERAGE_EXCLUDE = mocks|main.go|_test.go|docs|bootstrap|config
COVERAGE_THRESHOLD = 80

# Version variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME = $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Compiler flags for version embedding
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Default target
.DEFAULT_GOAL := help

# ============================================================================
# DEVELOPMENT TARGETS
# ============================================================================

.PHONY: run
run:
	@echo "Running application..."
	go run $(CMD_PATH)

.PHONY: dev
dev: fmt vet test swagger run
	@echo "Development workflow completed!"

.PHONY: dev-quick
dev-quick: fmt vet run
	@echo "Quick development workflow completed!"

# ============================================================================
# BUILD TARGETS
# ============================================================================

.PHONY: build
build: clean
	@echo "Building $(APP_NAME) for current OS..."
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "✓ Binary created: $(BIN_DIR)/$(APP_NAME)"

.PHONY: build-linux
build-linux: clean
	@echo "Building $(APP_NAME) for Linux..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 $(CMD_PATH)
	@echo "✓ Binary created: $(BIN_DIR)/$(APP_NAME)-linux-amd64"

.PHONY: build-macos
build-macos: clean
	@echo "Building $(APP_NAME) for macOS..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_PATH)
	@echo "✓ Binary created: $(BIN_DIR)/$(APP_NAME)-darwin-arm64"

.PHONY: build-windows
build-windows: clean
	@echo "Building $(APP_NAME) for Windows..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_PATH)
	@echo "✓ Binary created: $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe"

.PHONY: build-prod
build-prod: clean
	@echo "Building optimized production binary..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		$(LDFLAGS) \
		-a -installsuffix cgo \
		-trimpath \
		-o $(BIN_DIR)/$(APP_NAME)-prod $(CMD_PATH)
	@echo "✓ Production binary created: $(BIN_DIR)/$(APP_NAME)-prod"
	@ls -lh $(BIN_DIR)/$(APP_NAME)-prod

.PHONY: release
release: build-linux build-macos build-windows
	@echo "Creating release checksums..."
	@cd $(BIN_DIR) && sha256sum * > checksums.txt 2>/dev/null || echo "checksums created (Windows)"
	@echo "✓ All release binaries created in $(BIN_DIR)"
	@echo ""
	@echo "Release contents:"
	@ls -lh $(BIN_DIR)/

# ============================================================================
# CODE QUALITY TARGETS
# ============================================================================

.PHONY: test
test:
	@echo "Running tests with coverage..."
	go clean -testcache
	go test ./... -coverprofile=coverage.tmp -covermode=atomic -coverpkg=./internal/... -p 1; \
	grep -vE "$(COVERAGE_EXCLUDE)" coverage.tmp > coverage.out || touch coverage.out; \
	go tool cover -html=coverage.out -o coverage.html; \
	echo "✓ Tests completed"; \
	echo "📊 Coverage report: coverage.html"; \
	go tool cover -func=coverage.out | grep total

.PHONY: test-verbose
test-verbose:
	@echo "Running tests with verbose output..."
	go clean -testcache
	go test -v ./... -coverprofile=coverage.tmp -covermode=atomic -coverpkg=./internal/... -p 1; \
	grep -vE "$(COVERAGE_EXCLUDE)" coverage.tmp > coverage.out || touch coverage.out; \
	go tool cover -html=coverage.out -o coverage.html; \
	echo "✓ Tests completed"; \
	echo "📊 HTML Coverage report: coverage.html"; \
	go tool cover -func=coverage.out | grep total

.PHONY: test-coverage
test-coverage: test
	@echo "Opening coverage report..."
	go tool cover -html=coverage.out

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Code formatted"

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✓ No issues found"

.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Run: make install-tools"; exit 1)
	golangci-lint run ./... --deadline=5m
	@echo "✓ Lint check passed"

.PHONY: tidy
tidy:
	@echo "Tidying Go modules..."
	go mod tidy
	@echo "✓ Modules tidied"

.PHONY: vendor
vendor:
	@echo "Downloading dependencies..."
	go mod download
	go mod vendor
	@echo "✓ Dependencies vendored"

# ============================================================================
# DOCUMENTATION TARGETS
# ============================================================================

.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	@which swag > /dev/null || (echo "swag not found. Run: make install-tools"; exit 1)
	swag init \
		--parseDependency \
		--parseInternal \
		--generalInfo $(CMD_PATH) \
		--output $(DOCS_DIR)
	@echo "✓ Swagger docs generated:"
	@echo "  - $(DOCS_DIR)/swagger.json"
	@echo "  - $(DOCS_DIR)/swagger.yaml"

# ============================================================================
# TOOL INSTALLATION
# ============================================================================

.PHONY: install-tools
install-tools:
	@echo "Installing required tools..."
	@echo "Installing swag..."
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✓ All tools installed"

# ============================================================================
# CLEANUP TARGETS
# ============================================================================

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@echo "✓ Cleanup complete"

.PHONY: clean-docs
clean-docs:
	@echo "Removing generated documentation..."
	@rm -rf $(DOCS_DIR)
	@echo "✓ Documentation removed"

.PHONY: clean-all
clean-all: clean clean-docs
	@echo "Removing vendor directory..."
	@rm -rf vendor
	@echo "✓ Complete cleanup done"

# ============================================================================
# DOCKER TARGETS (Optional)
# ============================================================================

.PHONY: docker-build
docker-build: build-prod
	@echo "Building Docker image..."
	@if [ -f Dockerfile ]; then \
		docker build -t $(APP_NAME):$(VERSION) .; \
		echo "✓ Docker image built: $(APP_NAME):$(VERSION)"; \
	else \
		echo "Dockerfile not found"; \
	fi

# ============================================================================
# INFO TARGETS
# ============================================================================

.PHONY: info
info:
	@echo "Project Information:"
	@echo "  App Name: $(APP_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Commit: $(COMMIT)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $$(go version)"
	@echo ""
	@echo "Paths:"
	@echo "  Command: $(CMD_PATH)"
	@echo "  Binary Dir: $(BIN_DIR)"
	@echo "  Docs Dir: $(DOCS_DIR)"

