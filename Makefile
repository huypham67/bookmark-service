# ============================================================================
# APPLICATION VARIABLES
# ============================================================================
APP_NAME        = bookmark-service
CMD_PATH        = ./cmd/api/main.go
MAIN_PACKAGE    = github.com/huypham67/bookmark-service
BIN_DIR         = ./bin
DOCS_DIR        = ./docs

# ============================================================================
# VERSIONING & GIT CONTEXT
# ============================================================================
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME  ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Nâng cao: Đồng bộ hóa biến GitHub Actions truyền sang và biến Local chạy bằng tay
GIT_SHA        ?= $(COMMIT)
GIT_EVENT_NAME ?= local
GIT_REF_TYPE   ?= branch
GIT_REF_NAME   ?= $(VERSION)

# ============================================================================
# GO CONFIGURATION
# ============================================================================
GO              = go
GOTEST          = go test
GOLINT          = golangci-lint
CGO_ENABLED     = 0

LDFLAGS = -ldflags "\
    -s -w \
    -X main.Version=$(VERSION) \
    -X main.Commit=$(COMMIT) \
    -X main.BuildTime=$(BUILD_TIME)"

# ============================================================================
# DOCKER & COVERAGE CONFIGURATION (Đồng bộ một mối sạch sẽ)
# ============================================================================
DOCKER_REGISTRY   ?= docker.io
DOCKER_NAMESPACE  ?= huypham053
DOCKER_IMAGE      = $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(APP_NAME)
DOCKER_CONTAINER  = $(APP_NAME)

COVERAGE_FOLDER   ?= coverage_report
COVERAGE_EXCLUDE  ?= mocks|main.go|_test.go|docs|bootstrap|config
COVERAGE_THRESHOLD ?= 80

# ============================================================================
# DEFAULT TARGET
# ============================================================================
.DEFAULT_GOAL := help

# ============================================================================
# HELPER FUNCTIONS (Build đa nền tảng)
# ============================================================================
define go-build
    @echo "🚀 Building $(APP_NAME) for $(1)/$(2)..."
    @mkdir -p $(BIN_DIR)
    CGO_ENABLED=$(CGO_ENABLED) GOOS=$(1) GOARCH=$(2) $(GO) build $(4) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-$(1)-$(2)$(3) $(CMD_PATH)
    @echo "✅ Binary created: $(BIN_DIR)/$(APP_NAME)-$(1)-$(2)$(3)"
endef

# ============================================================================
# HELP SYSTEM
# ============================================================================
.PHONY: help
help:
	@echo ""
	@echo "================================================================="
	@echo " 📑 Bookmark Service API - Elite Production Makefile"
	@echo "================================================================="
	@echo "Development:"
	@echo "  make run                 Run application locally"
	@echo "  make dev                 Full development workflow (fmt, vet, test, swagger, run)"
	@echo "Testing & Quality:"
	@echo "  make test                Run tests locally with coverage check"
	@echo "  make fmt | make vet      Format and vet source code"
	@echo "  make lint | make tidy    Run linter and tidy go.mod"
	@echo "Build & Release:"
	@echo "  make build               Build binary for current OS"
	@echo "  make release             Build release binaries for Linux, macOS, Windows"
	@echo "Docker Pipeline (Local & CI/CD):"
	@echo "  make docker-test         Run tests inside Docker Sandbox (Outputs coverage)"
	@echo "  make docker-login        Log in to Docker registry securely"
	@echo "  make docker-build-push   Smart Build & Push using Buildx (Auto-detects PR/Release)"
	@echo "Docker Local Utilities:"
	@echo "  make docker-run | stop   Run/Stop application in local Docker container"
	@echo "  make docker-logs | clean Show container logs / Full cleanup Docker resources"
	@echo "Docker Compose:"
	@echo "  make compose-up | down   Start/Stop full development stack"
	@echo "================================================================="

# ============================================================================
# DEVELOPMENT & QUALITY TARGETS
# ============================================================================
.PHONY: run dev fmt vet lint tidy vendor
run:
	@echo "🚀 Starting application..."
	SERVICE_NAME=bookmark-service $(GO) run $(CMD_PATH)

dev: fmt vet test swagger run

fmt:
	@echo "🎨 Formatting source code..."
	$(GO) fmt ./...

vet:
	@echo "🔍 Running go vet..."
	$(GO) vet ./...

lint:
	@echo "🔎 Running golangci-lint..."
	@which $(GOLINT) > /dev/null || (echo "❌ golangci-lint not installed. Run: make install-tools"; exit 1)
	$(GOLINT) run ./...

tidy:
	@echo "📦 Tidying dependencies..."
	$(GO) mod tidy

vendor:
	@echo "📦 Downloading dependencies..."
	$(GO) mod download
	$(GO) mod vendor

# ============================================================================
# LOCAL TESTING TARGETS
# ============================================================================
.PHONY: test test-coverage
test:
	@echo "🧪 Running local tests..."
	@$(GO) clean -testcache
	@mkdir -p $(COVERAGE_FOLDER)
	@$(GO) test ./... -coverprofile=$(COVERAGE_FOLDER)/coverage.tmp -covermode=atomic -coverpkg=./internal/... -p 1
	@grep -vE "$(COVERAGE_EXCLUDE)" $(COVERAGE_FOLDER)/coverage.tmp > $(COVERAGE_FOLDER)/coverage.out || touch $(COVERAGE_FOLDER)/coverage.out
	@$(GO) tool cover -html=$(COVERAGE_FOLDER)/coverage.out -o $(COVERAGE_FOLDER)/coverage.html
	@echo ""
	@echo "📊 Checking coverage threshold..."
	@total=$$($(GO) tool cover -func=$(COVERAGE_FOLDER)/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total Coverage: $$total%"; \
	if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
	   echo "❌ FAIL: Coverage ($$total%) is below threshold ($(COVERAGE_THRESHOLD)%)"; \
	   exit 1; \
	else \
	   echo "✅ PASS: Coverage ($$total%) meets threshold ($(COVERAGE_THRESHOLD)%)"; \
	fi

test-coverage: test
	@echo "📂 Opening local coverage report..."
	$(GO) tool cover -html=$(COVERAGE_FOLDER)/coverage.out

# ============================================================================
# LOCAL & CROSS-COMPILATION BUILD TARGETS
# ============================================================================
.PHONY: build build-linux build-macos build-windows build-prod release
build:
	@echo "🏗️ Building application..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "✅ Build completed: $(BIN_DIR)/$(APP_NAME)"

build-linux:
	$(call go-build,linux,amd64,,)

build-macos:
	$(call go-build,darwin,arm64,,)

build-windows:
	$(call go-build,windows,amd64,.exe,)

build-prod:
	$(call go-build,linux,amd64,-prod,-trimpath)
	@echo ""
	@echo "📦 Production binary size:"
	@ls -lh $(BIN_DIR)/$(APP_NAME)-linux-amd64-prod

release: clean build-linux build-macos build-windows
	@echo "📦 Creating release checksums..."
	@cd $(BIN_DIR) && sha256sum * > checksums.txt 2>/dev/null || echo "checksums created"
	@echo ""
	@echo "✅ Release artifacts:"
	@ls -lh $(BIN_DIR)

# ============================================================================
# ELITE DOCKER PIPELINE TARGETS (Sạch 100% cho cả Local lẫn GitHub Actions)
# ============================================================================
.PHONY: docker-test docker-login docker-build-push

docker-test:
	@echo "🧪 [SANDBOX] Running tests inside clean Docker container..."
	@mkdir -p $(COVERAGE_FOLDER)
	docker buildx build \
		--build-arg COVERAGE_EXCLUDE="$(COVERAGE_EXCLUDE)" \
		--target test \
		--output type=local,dest=$(COVERAGE_FOLDER) .
	@echo ""
	@echo "📊 [SANDBOX] Analyzing coverage report from Docker..."
	@if [ -f $(COVERAGE_FOLDER)/coverage.out ]; then \
		total=$$(go tool cover -func=$(COVERAGE_FOLDER)/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
		echo "Sandbox Coverage: $$total%"; \
		if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
			echo "❌ FAIL: Sandbox Coverage ($$total%) is below threshold ($(COVERAGE_THRESHOLD)%)"; \
			exit 1; \
		else \
			echo "✅ PASS: Sandbox Coverage ($$total%) meets threshold ($(COVERAGE_THRESHOLD)%)"; \
		fi \
	else \
		echo "❌ Error: Coverage output data not found!"; \
		exit 1; \
	fi

docker-login:
	@echo "🔐 Securely logging in to Docker Registry..."
	@if [ -z "$(DOCKERHUB_USERNAME)" ] || [ -z "$(DOCKERHUB_TOKEN)" ]; then \
		echo "❌ Error: DOCKERHUB_USERNAME or DOCKERHUB_TOKEN environment variables are missing!"; \
		exit 1; \
	fi
	@echo "$(DOCKERHUB_TOKEN)" | docker login -u "$(DOCKERHUB_USERNAME)" --password-stdin

docker-build-push:
	@echo "📦 [PIPELINE] Analyzing Git Context for Containerization..."
	@if [ "$(GIT_REF_TYPE)" = "tag" ]; then \
		IMG_TAG="$(GIT_REF_NAME)"; \
	else \
		IMG_TAG=$$(echo "$(GIT_SHA)" | cut -c1-7); \
	fi; \
	\
	if [ "$(GIT_EVENT_NAME)" = "pull_request" ]; then \
		DOCKER_PUSH="false"; \
		echo "▶️ [PR Mode] Verification Build triggered. IMAGE WILL NOT BE PUSHED."; \
	else \
		DOCKER_PUSH="true"; \
		echo "🚀 [Release Mode] Production Build triggered. TARGET TAG: $$IMG_TAG"; \
	fi; \
	\
	docker buildx build \
		--target final \
		--push=$$DOCKER_PUSH \
		-t $(DOCKER_IMAGE):$$IMG_TAG \
		-t $(DOCKER_IMAGE):$(COMMIT) \
		-t $(DOCKER_IMAGE):latest .

# ============================================================================
# DOCKER LOCAL UTILITIES (Dành riêng cho Dev chạy kiểm thử ở máy cá nhân)
# ============================================================================
.PHONY: docker-run docker-stop docker-logs docker-shell docker-clean

docker-run:
	@echo "🚀 Running local container instance..."
	docker run -d --name $(DOCKER_CONTAINER) -p 8080:8080 --env-file .env $(DOCKER_IMAGE):latest
	@echo "✅ Container successfully spawned. Port mapped to 8080."

docker-stop:
	@echo "🛑 Stopping container instance..."
	-docker stop $(DOCKER_CONTAINER)
	-docker rm $(DOCKER_CONTAINER)

docker-logs:
	docker logs -f $(DOCKER_CONTAINER)

docker-shell:
	docker exec -it $(DOCKER_CONTAINER) sh

docker-clean:
	@echo "🧹 Wiping docker local build images..."
	-docker rm -f $(DOCKER_CONTAINER)
	-docker rmi -f $$(docker images -q $(DOCKER_IMAGE) 2>/dev/null) 2>/dev/null || true

# ============================================================================
# DOCKER COMPOSE TARGETS
# ============================================================================
.PHONY: compose-up compose-down compose-logs compose-restart
compose-up:
	docker compose up --build -d
compose-down:
	docker compose down
compose-logs:
	docker compose logs -f
compose-restart:
	docker compose down && docker compose up --build -d

# ============================================================================
# UTILITIES & SWAGGER
# ============================================================================
.PHONY: swagger install-tools info clean clean-docs clean-all
swagger:
	@echo "📘 Generating Swagger documentation..."
	@which swag > /dev/null || (echo "❌ swag not installed. Run: make install-tools"; exit 1)
	swag init --parseDependency --parseInternal --generalInfo $(CMD_PATH) --output $(DOCS_DIR)

install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

info:
	@echo "App Name:      $(APP_NAME)"
	@echo "Version:       $(VERSION)"
	@echo "Commit:        $(COMMIT)"
	@echo "Build Time:    $(BUILD_TIME)"
	@echo "Go Version:    $$($(GO) version)"

clean:
	rm -rf $(BIN_DIR) $(COVERAGE_FOLDER)
clean-docs:
	rm -rf $(DOCS_DIR)
clean-all: clean clean-docs docker-clean