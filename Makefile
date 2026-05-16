# ============================================================================
# APPLICATION VARIABLES
# ============================================================================

APP_NAME        = bookmark-service
CMD_PATH        = ./cmd/api/main.go
MAIN_PACKAGE    = github.com/huypham67/bookmark-service

BIN_DIR         = ./bin
DOCS_DIR        = ./docs

# ============================================================================
# COVERAGE CONFIGURATION
# ============================================================================

COVERAGE_FILE       = coverage.out
COVERAGE_HTML       = coverage.html
COVERAGE_TMP        = coverage.tmp

COVERAGE_EXCLUDE    = mocks|main.go|_test.go|docs|bootstrap|config
COVERAGE_THRESHOLD  = 80

# ============================================================================
# VERSIONING
# ============================================================================

VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME  ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# ============================================================================
# GO CONFIGURATION
# ============================================================================

GO              = go
GOFMT           = gofmt
GOVET           = go vet
GOTEST          = go test
GOLINT          = golangci-lint

CGO_ENABLED     = 0

# ============================================================================
# BUILD FLAGS
# ============================================================================

LDFLAGS = -ldflags "\
	-s -w \
	-X main.Version=$(VERSION) \
	-X main.Commit=$(COMMIT) \
	-X main.BuildTime=$(BUILD_TIME)"

# ============================================================================
# DOCKER CONFIGURATION
# ============================================================================

DOCKER_REGISTRY     ?= docker.io
DOCKER_NAMESPACE    ?= huypham67

DOCKER_IMAGE        = $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(APP_NAME)

DOCKER_TAG          ?= $(VERSION)
DOCKER_LATEST_TAG   = latest
DOCKER_COMMIT_TAG   = $(COMMIT)

DOCKER_CONTAINER    = $(APP_NAME)

# ============================================================================
# DEFAULT TARGET
# ============================================================================

.DEFAULT_GOAL := help

# ============================================================================
# HELPER FUNCTIONS
# ============================================================================

define go-build
	@echo "🚀 Building $(APP_NAME) for $(1)/$(2)..."

	@mkdir -p $(BIN_DIR)

	CGO_ENABLED=$(CGO_ENABLED) \
	GOOS=$(1) \
	GOARCH=$(2) \
	$(GO) build \
	$(4) \
	$(LDFLAGS) \
	-o $(BIN_DIR)/$(APP_NAME)-$(1)-$(2)$(3) \
	$(CMD_PATH)

	@echo "✅ Binary created:"
	@echo "   $(BIN_DIR)/$(APP_NAME)-$(1)-$(2)$(3)"
endef

# ============================================================================
# HELP
# ============================================================================

.PHONY: help
help:
	@echo ""
	@echo "==============================================="
	@echo " Bookmark Service API - Makefile Commands"
	@echo "==============================================="
	@echo ""
	@echo "Development:"
	@echo "  make run                 Run application"
	@echo "  make dev                 Full development workflow"
	@echo "  make dev-quick           Fast development workflow"
	@echo ""
	@echo "Testing:"
	@echo "  make test                Run tests with coverage"
	@echo "  make test-verbose        Run verbose tests"
	@echo "  make test-coverage       Open coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt                 Format source code"
	@echo "  make vet                 Run go vet"
	@echo "  make lint                Run golangci-lint"
	@echo "  make tidy                Run go mod tidy"
	@echo "  make vendor              Download dependencies"
	@echo ""
	@echo "Build:"
	@echo "  make build               Build for current OS"
	@echo "  make build-linux         Build Linux binary"
	@echo "  make build-macos         Build macOS binary"
	@echo "  make build-windows       Build Windows binary"
	@echo "  make build-prod          Build optimized production binary"
	@echo "  make release             Build release binaries"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build        Build Docker image with semantic versioning"
	@echo "  make docker-run          Run Docker container"
	@echo "  make docker-push         Push Docker images to registry"
	@echo "  make docker-stop         Stop Docker container"
	@echo "  make docker-logs         Show container logs"
	@echo "  make docker-shell        Open shell in container"
	@echo "  make docker-clean        Remove Docker resources"
	@echo ""
	@echo "Docker Configuration:"
	@echo "  DOCKER_REGISTRY=<url>    Set Docker registry (default: docker.io)"
	@echo "  DOCKER_NAMESPACE=<name>  Set Docker namespace (default: huypham67)"
	@echo "  VERSION=<tag>            Set version tag (default: git describe)"
	@echo ""
	@echo "Docker Compose:"
	@echo "  make compose-up          Start full stack"
	@echo "  make compose-down        Stop full stack"
	@echo "  make compose-logs        Show compose logs"
	@echo "  make compose-restart     Restart full stack"
	@echo ""
	@echo "Utilities:"
	@echo "  make swagger             Generate Swagger docs"
	@echo "  make install-tools       Install development tools"
	@echo "  make info                Show build information"
	@echo "  make clean               Remove generated files"
	@echo "  make clean-all           Full cleanup"
	@echo ""

# ============================================================================
# DEVELOPMENT TARGETS
# ============================================================================

.PHONY: run
run:
	@echo "🚀 Starting application..."
	SERVICE_NAME=bookmark-service $(GO) run $(CMD_PATH)

.PHONY: dev
dev: fmt vet test swagger run

.PHONY: dev-quick
dev-quick: fmt vet run

# ============================================================================
# TEST TARGETS
# ============================================================================

.PHONY: test
test:
	@echo "🧪 Running tests..."

	@$(GO) clean -testcache

	@$(GOTEST) ./... \
		-coverprofile=$(COVERAGE_TMP) \
		-covermode=atomic \
		-coverpkg=./internal/... \
		-p 1

	@grep -vE "$(COVERAGE_EXCLUDE)" $(COVERAGE_TMP) > $(COVERAGE_FILE) || touch $(COVERAGE_FILE)

	@$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

	@echo ""
	@echo "📊 Coverage Summary:"
	@$(GO) tool cover -func=$(COVERAGE_FILE) | grep total

.PHONY: test-verbose
test-verbose:
	@echo "🧪 Running verbose tests..."

	@$(GO) clean -testcache

	@$(GOTEST) -v ./... \
		-coverprofile=$(COVERAGE_TMP) \
		-covermode=atomic \
		-coverpkg=./internal/... \
		-p 1

.PHONY: test-coverage
test-coverage: test
	@echo "📂 Opening coverage report..."
	$(GO) tool cover -html=$(COVERAGE_FILE)

# ============================================================================
# CODE QUALITY TARGETS
# ============================================================================

.PHONY: fmt
fmt:
	@echo "🎨 Formatting source code..."
	$(GO) fmt ./...

.PHONY: vet
vet:
	@echo "🔍 Running go vet..."
	$(GOVET) ./...

.PHONY: lint
lint:
	@echo "🔎 Running golangci-lint..."

	@which $(GOLINT) > /dev/null || \
		(echo "❌ golangci-lint not installed. Run: make install-tools"; exit 1)

	$(GOLINT) run ./...

.PHONY: tidy
tidy:
	@echo "📦 Tidying dependencies..."
	$(GO) mod tidy

.PHONY: vendor
vendor:
	@echo "📦 Downloading dependencies..."
	$(GO) mod download
	$(GO) mod vendor

# ============================================================================
# BUILD TARGETS
# ============================================================================

.PHONY: build
build:
	@echo "🏗️ Building application..."

	@mkdir -p $(BIN_DIR)

	$(GO) build \
	$(LDFLAGS) \
	-o $(BIN_DIR)/$(APP_NAME) \
	$(CMD_PATH)

	@echo "✅ Build completed:"
	@echo "   $(BIN_DIR)/$(APP_NAME)"

.PHONY: build-linux
build-linux:
	$(call go-build,linux,amd64,,)

.PHONY: build-macos
build-macos:
	$(call go-build,darwin,arm64,,)

.PHONY: build-windows
build-windows:
	$(call go-build,windows,amd64,.exe,)

.PHONY: build-prod
build-prod:
	$(call go-build,linux,amd64,-prod,-trimpath)

	@echo ""
	@echo "📦 Production binary size:"
	@ls -lh $(BIN_DIR)/$(APP_NAME)-linux-amd64-prod

.PHONY: release
release: clean build-linux build-macos build-windows
	@echo "📦 Creating release checksums..."

	@cd $(BIN_DIR) && \
	sha256sum * > checksums.txt 2>/dev/null || \
	echo "checksums created"

	@echo ""
	@echo "✅ Release artifacts:"
	@ls -lh $(BIN_DIR)

# ============================================================================
# SWAGGER TARGETS
# ============================================================================

.PHONY: swagger
swagger:
	@echo "📘 Generating Swagger documentation..."

	@which swag > /dev/null || \
		(echo "❌ swag not installed. Run: make install-tools"; exit 1)

	swag init \
		--parseDependency \
		--parseInternal \
		--generalInfo $(CMD_PATH) \
		--output $(DOCS_DIR)

# ============================================================================
# TOOL INSTALLATION
# ============================================================================

.PHONY: install-tools
install-tools:
	@echo "🛠️ Installing development tools..."

	go install github.com/swaggo/swag/cmd/swag@latest

	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

	@echo "✅ Tools installed successfully"

# ============================================================================
# DOCKER TARGETS
# ============================================================================

.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker image..."

	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(DOCKER_IMAGE):$(DOCKER_TAG) \
		-t $(DOCKER_IMAGE):$(DOCKER_COMMIT_TAG) \
		-t $(DOCKER_IMAGE):$(DOCKER_LATEST_TAG) \
		.

	@echo ""
	@echo "✅ Docker images built:"
	@echo "   $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "   $(DOCKER_IMAGE):$(DOCKER_COMMIT_TAG)"
	@echo "   $(DOCKER_IMAGE):$(DOCKER_LATEST_TAG)"

.PHONY: docker-run
docker-run:
	@echo "🚀 Running Docker container..."

	docker run -d \
		--name $(DOCKER_CONTAINER) \
		-p 8080:8080 \
		--env-file .env \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

	@echo "✅ Container started:"
	@echo "   $(DOCKER_CONTAINER) ($(DOCKER_IMAGE):$(DOCKER_TAG))"

.PHONY: docker-stop
docker-stop:
	@echo "🛑 Stopping Docker container..."

	-docker stop $(DOCKER_CONTAINER)
	-docker rm $(DOCKER_CONTAINER)

	@echo "✅ Container stopped"

.PHONY: docker-logs
docker-logs:
	docker logs -f $(DOCKER_CONTAINER)

.PHONY: docker-shell
docker-shell:
	docker exec -it $(DOCKER_CONTAINER) sh

.PHONY: docker-clean
docker-clean:
	@echo "🧹 Cleaning Docker resources..."

	-docker rm -f $(DOCKER_CONTAINER)
	-docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	-docker rmi $(DOCKER_IMAGE):$(DOCKER_COMMIT_TAG) 2>/dev/null || true
	-docker rmi $(DOCKER_IMAGE):$(DOCKER_LATEST_TAG) 2>/dev/null || true

	@echo "✅ Docker cleanup completed"

.PHONY: docker-push
docker-push:
	@echo "📤 Pushing Docker images to $(DOCKER_REGISTRY)..."

	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

	docker push $(DOCKER_IMAGE):$(DOCKER_COMMIT_TAG)

	docker push $(DOCKER_IMAGE):$(DOCKER_LATEST_TAG)

	@echo ""
	@echo "✅ Docker images pushed:"
	@echo "   $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo "   $(DOCKER_IMAGE):$(DOCKER_COMMIT_TAG)"
	@echo "   $(DOCKER_IMAGE):$(DOCKER_LATEST_TAG)"

# ============================================================================
# DOCKER COMPOSE TARGETS
# ============================================================================

.PHONY: compose-up
compose-up:
	@echo "🚀 Starting full stack..."
	docker compose up --build -d

.PHONY: compose-down
compose-down:
	@echo "🛑 Stopping full stack..."
	docker compose down

.PHONY: compose-logs
compose-logs:
	docker compose logs -f

.PHONY: compose-restart
compose-restart:
	@echo "🔄 Restarting full stack..."

	docker compose down
	docker compose up --build -d

# ============================================================================
# INFORMATION TARGETS
# ============================================================================

.PHONY: info
info:
	@echo ""
	@echo "==================================="
	@echo " Build Information"
	@echo "==================================="
	@echo ""
	@echo "App Name:      $(APP_NAME)"
	@echo "Version:       $(VERSION)"
	@echo "Commit:        $(COMMIT)"
	@echo "Build Time:    $(BUILD_TIME)"
	@echo "Go Version:    $$($(GO) version)"
	@echo ""

# ============================================================================
# CLEANUP TARGETS
# ============================================================================

.PHONY: clean
clean:
	@echo "🧹 Cleaning generated files..."

	rm -rf $(BIN_DIR)

	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)
	rm -f $(COVERAGE_TMP)

	@echo "✅ Cleanup completed"

.PHONY: clean-docs
clean-docs:
	@echo "🧹 Cleaning Swagger docs..."
	rm -rf $(DOCS_DIR)

.PHONY: clean-all
clean-all: clean clean-docs docker-clean
	@echo "🧹 Full cleanup completed"