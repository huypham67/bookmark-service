# ============================================================================
# 1. CORE VARIABLES (Cấu hình hệ thống hệ mặt trời)
# ============================================================================
APP_NAME           = bookmark-service
CMD_PATH           = ./cmd/api/main.go
MAIN_PACKAGE       = github.com/huypham67/bookmark-service
BIN_DIR            = ./bin
DOCS_DIR           = ./docs

# Git Context Execution (Local fallback vs CI injections)
VERSION           ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT            ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME        ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GIT_SHA           ?= $(COMMIT)
GIT_EVENT_NAME    ?= local
GIT_REF_TYPE      ?= branch
GIT_REF_NAME      ?= $(VERSION)

# Go Toolchain Parameters
GO                 = go
GOTEST             = go test
GOLINT             = golangci-lint
CGO_ENABLED        = 0
LDFLAGS            = -ldflags "-s -w \
                     -X main.Version=$(VERSION) \
                     -X main.Commit=$(COMMIT) \
                     -X main.BuildTime=$(BUILD_TIME)"

# Docker Registry & Infrastructure Targets
DOCKER_REGISTRY   ?= docker.io
DOCKER_NAMESPACE  ?= huypham053
DOCKER_IMAGE       = $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(APP_NAME)
DOCKER_CONTAINER   = $(APP_NAME)

# Quality Gates & Verification Layers
COVERAGE_FOLDER   ?= coverage_report
COVERAGE_EXCLUDE  ?= "mocks|main.go|_test.go|docs|bootstrap|config|logger|redis|sqldb"
COVERAGE_THRESHOLD ?= 80

# SonarCloud SAST Exclusion Patterns
SONAR_EXCLUDE     ?= "**/*_test.go,**/vendor/**,**/mocks/**,docs/**,bin/**,$(COVERAGE_FOLDER)/**,**/logger/**,**/redis/**,**/sqldb/**"

# System Default Gate
.DEFAULT_GOAL     := help

# ============================================================================
# 2. INTERNAL MACROS / HELPER FUNCTIONS
# ============================================================================
define go-build
	@echo "🚀 Building $(APP_NAME) for $(1)/$(2)..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(1) GOARCH=$(2) $(GO) build $(4) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-$(1)-$(2)$(3) $(CMD_PATH)
	@echo "✅ Binary created: $(BIN_DIR)/$(APP_NAME)-$(1)-$(2)$(3)"
endef

# ============================================================================
# 3. INTERACTIVE DOCUMENTATION (Help UI)
# ============================================================================
.PHONY: help
help:
	@echo ""
	@echo "================================================================="
	@echo " 📑 Bookmark Service API - Elite Production Makefile"
	@echo "================================================================="
	@echo "Development Workflow:"
	@echo "  make run                 Run application locally with dynamic reload"
	@echo "  make dev                 Trigger complete local cycle (fmt -> vet -> test -> run)"
	@echo "Testing & Linting Core:"
	@echo "  make test                Execute local tests + HTML report visualization"
	@echo "  make fmt | make vet      Execute code style formatting and analysis"
	@echo "  make lint | make tidy    Execute strict golangci-lint and mod verification"
	@echo "Compilation Layers:"
	@echo "  make build               Compile binary optimization for Current Host OS"
	@echo "  make release             Compile cross-platform artifacts (Linux, Mac, Win)"
	@echo "Universal Ops Pipeline (CI/CD):"
	@echo "  make docker-test         Isolate execution test loop inside Buildx Container"
	@echo "  make docker-sonar        Execute SonarCloud SAST Security validation"
	@echo "  make docker-build-push   Automated contextual Buildx engine (Detects PR/Release)"
	@echo "Local Virtualization Infrastructure:"
	@echo "  make docker-run | stop   Spin up / Kill localized single-container target"
	@echo "  make compose-up | down   Orchestrate multi-dependency stack (Redis, Nginx, App)"
	@echo "================================================================="

# ============================================================================
# 4. STANDARD APPLICATION ROAD (Local Dev)
# ============================================================================
.PHONY: run dev fmt vet lint tidy vendor
run:
	@echo "🚀 Starting application..."
	SERVICE_NAME=$(APP_NAME) $(GO) run $(CMD_PATH)

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
# 5. VERIFICATION & QUALITY LAYER (Local Machine Validation)
# ============================================================================
.PHONY: test test-coverage
test:
	@echo "🧪 Running local tests..."
	@$(GO) clean -testcache
	@mkdir -p $(COVERAGE_FOLDER)
	@$(GO) test ./... -coverprofile=$(COVERAGE_FOLDER)/coverage.tmp -covermode=atomic -coverpkg=./internal/... -p 1
	@grep -vE $(COVERAGE_EXCLUDE) $(COVERAGE_FOLDER)/coverage.tmp > $(COVERAGE_FOLDER)/coverage.out || touch $(COVERAGE_FOLDER)/coverage.out
	@$(GO) tool cover -html=$(COVERAGE_FOLDER)/coverage.out -o $(COVERAGE_FOLDER)/coverage.html
	@echo ""
	@echo "📊 Analyzing coverage data criteria..."
	@total=$$(go tool cover -func=$(COVERAGE_FOLDER)/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Total Coverage: $$total%"; \
	if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
	   echo "❌ FAIL: Coverage ($$total%) is below required threshold ($(COVERAGE_THRESHOLD)%)"; \
	   exit 1; \
	else \
	   echo "✅ PASS: Coverage ($$total%) meets threshold ($(COVERAGE_THRESHOLD)%)"; \
	fi

test-coverage: test
	@echo "📂 Opening local coverage report..."
	$(GO) tool cover -html=$(COVERAGE_FOLDER)/coverage.out

# ============================================================================
# 6. COMPILATION ARCHITECTURE (Local Binaries)
# ============================================================================
.PHONY: build build-linux build-macos build-windows build-prod release
build:
	@echo "🏗️ Building application binary..."
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
	@echo "📦 Production binary size optimization profile:"
	@ls -lh $(BIN_DIR)/$(APP_NAME)-linux-amd64-prod

release: clean build-linux build-macos build-windows
	@echo "📦 Generating SHA256 checksum signatures..."
	@cd $(BIN_DIR) && sha256sum * > checksums.txt 2>/dev/null || echo "Checksum database initialized."
	@echo ""
	@echo "✅ Complete release package ready:"
	@ls -lh $(BIN_DIR)

# ============================================================================
# 7. HIGH-TIER CLOUD PIPELINE TARGETS (Docker Sandbox, Sonar & Buildx)
# ============================================================================
.PHONY: docker-test docker-login docker-build-push docker-sonar

docker-test:
	@echo "🧪 Executing isolated test suite within Docker Buildx environment..."
	@mkdir -p $(COVERAGE_FOLDER)
	docker buildx build \
		--build-arg COVERAGE_EXCLUDE=$(COVERAGE_EXCLUDE) \
		--target test \
		--output type=local,dest=$(COVERAGE_FOLDER) .
	@echo ""
	@echo "📊 Evaluating sandbox coverage results..."
	@if [ -f $(COVERAGE_FOLDER)/coverage.out ]; then \
		total=$$(go tool cover -func=$(COVERAGE_FOLDER)/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
		echo "Sandbox Target Coverage: $$total%"; \
		if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
			echo "❌ FAIL: Sandbox Coverage ($$total%) drops below quality gate ($(COVERAGE_THRESHOLD)%)"; \
			exit 1; \
		else \
			echo "✅ PASS: Sandbox Coverage ($$total%) verified successfully!"; \
		fi \
	else \
		echo "❌ Error: Coverage synchronization lost. Output asset missing."; \
		exit 1; \
	fi

docker-login:
	@echo "🔐 Securely initializing Docker Hub Authentication..."
	@if [ -z "$(DOCKERHUB_USERNAME)" ] || [ -z "$(DOCKERHUB_TOKEN)" ]; then \
		echo "❌ Error: Active credentials missing from environment context!"; \
		exit 1; \
	fi
	@echo "$(DOCKERHUB_TOKEN)" | docker login -u "$(DOCKERHUB_USERNAME)" --password-stdin

docker-build-push:
	@echo "📦 [PIPELINE] Evaluating context for secure container deployment..."
	@if [ "$(GIT_REF_TYPE)" = "tag" ]; then \
		echo "::notice title=Docker Buildx::🏷️ [Release] Tagging as: $(GIT_REF_NAME) + latest"; \
		docker buildx build \
			--target final \
			--push=true \
			-t $(DOCKER_IMAGE):$(GIT_REF_NAME) \
			-t $(DOCKER_IMAGE):latest .; \
	elif [ "$(GIT_EVENT_NAME)" = "pull_request" ]; then \
		echo "::notice title=Docker Buildx::▶️ [PR Mode] Build only. NO PUSH"; \
		docker buildx build \
			--target final \
			--push=false \
			-t $(DOCKER_IMAGE):test .; \
	else \
		SHORT_SHA=$$(echo "$(GIT_SHA)" | cut -c1-7); \
		echo "::notice title=Docker Buildx::🚀 [Main] Tagging as: main + $$SHORT_SHA + latest"; \
		docker buildx build \
			--target final \
			--push=true \
			-t $(DOCKER_IMAGE):main \
			-t $(DOCKER_IMAGE):$$SHORT_SHA \
			-t $(DOCKER_IMAGE):latest .; \
	fi

docker-sonar:
	@echo "🔍 [SONAR] Initiating Cloud Vulnerability & Code Smell Scan..."
	@if [ -z "$(SONAR_TOKEN)" ]; then \
		echo "❌ Error: Scan blocked. SONAR_TOKEN context variable is missing!"; \
		exit 1; \
	fi
	@docker pull --quiet sonarsource/sonar-scanner-cli:11 || true
	docker run --rm \
		-e SONAR_TOKEN=$(SONAR_TOKEN) \
		-e SONAR_HOST_URL=https://sonarcloud.io \
		-v "$(PWD):/usr/src" \
		sonarsource/sonar-scanner-cli:11 \
		-Dsonar.exclusions=$(SONAR_EXCLUDE) \

# ============================================================================
# 8. LOCAL DESKTOP DOCKER UTILITIES (Chạy nhanh một container)
# ============================================================================
.PHONY: docker-run docker-stop docker-logs docker-shell docker-clean

docker-run:
	@echo "🚀 Launching local detached container instance..."
	docker run -d --name $(DOCKER_CONTAINER) -p 8080:8080 --env-file .env $(DOCKER_IMAGE):latest
	@echo "✅ Instance deployed successfully. Traffic mapping enabled on port 8080."

docker-stop:
	@echo "🛑 Destroying localized runtime stack..."
	-docker stop $(DOCKER_CONTAINER)
	-docker rm $(DOCKER_CONTAINER)

docker-logs:
	docker logs -f $(DOCKER_CONTAINER)

docker-shell:
	docker exec -it $(DOCKER_CONTAINER) sh

docker-clean:
	@echo "🧹 Executing full structural cache purge..."
	-docker rm -f $(DOCKER_CONTAINER)
	-docker rmi -f $$(docker images -q $(DOCKER_IMAGE) 2>/dev/null) 2>/dev/null || true

# ============================================================================
# 9. INTEGRATED ORCHESTRATION (Docker Compose)
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
# 10. SYSTEM UTILITIES & METADATA INFRASTRUCTURE
# ============================================================================
.PHONY: swagger install-tools info clean clean-docs clean-all
swagger:
	@echo "📘 Compiling API Swagger reference system..."
	@which swag > /dev/null || (echo "❌ System Error: Executable 'swag' dependency missing. Run: make install-tools"; exit 1)
	swag init --parseDependency --parseInternal --generalInfo $(CMD_PATH) --output $(DOCS_DIR)

install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

info:
	@echo "App Tracking Identity: $(APP_NAME)"
	@echo "SemVer Version:        $(VERSION)"
	@echo "Commit Hash Signature: $(COMMIT)"
	@echo "Compilation Time:      $(BUILD_TIME)"
	@echo "Host Runtime Engine:   $$($(GO) version)"

clean:
	rm -rf $(BIN_DIR) $(COVERAGE_FOLDER)
clean-docs:
	rm -rf $(DOCS_DIR)
clean-all: clean clean-docs docker-clean