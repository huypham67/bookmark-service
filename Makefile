# Makefile for Bookmark Service API

# =============================================================================
# APPLICATION METADATA
# =============================================================================

APP_NAME    := bookmark-service
CMD_PATH    := ./cmd/api/main.go
MAIN_PKG    := github.com/huypham67/bookmark-service

BIN_DIR     := ./bin
DOCS_DIR    := ./docs

# =============================================================================
# COVERAGE & QUALITY GATES
# =============================================================================

COVERAGE_DIR       ?= coverage_report
COVERAGE_THRESHOLD ?= 80

# ═══════════════════════════════════════════════════════════════════════════
# SINGLE SOURCE OF TRUTH: Coverage & Quality Gate Exclusions
#
# Strategy:
#   1. SYSTEM_DIRS/FILES: Completely excluded (no scan, no coverage)
#      → Auto-generated, vendored, test infrastructure
#      → Used for sonar.exclusions + local coverage filter
#
#   2. INFRA_DIRS / INFRA_FILES: Exclude from coverage % but INCLUDE in scan
#      → Infrastructure/setup code (DI, config, models, adapters)
#      → INFRA_DIRS: whole packages excluded from coverage threshold.
#        Pure adapters/wiring with no testable logic, e.g. cmd, bootstrap
#        (env load, key file I/O, DI), api router, dto, model.
#      → INFRA_FILES: surgical per-file exclusion for packages that mix
#        tested logic with wiring/setup. Currently empty — packages are
#        split so each is wholly one category.
#      → Both are still scanned for security vulnerabilities (SonarQube)
#
#   3. Everything else = business logic → MUST be covered:
#      handler/*, service/{bookmark,health,link} (incl. service/link/resolver),
#      repository/{bookmark,cache,link}
#
# Usage:
#   - make test        → filters coverage.out to exclude infrastructure + system
#   - make docker-test → passes COVERAGE_EXCLUDE to Docker build
#   - make docker-sonar → sonar.exclusions (system only)
#                        sonar.coverage.exclusions (system + infra)
# ═══════════════════════════════════════════════════════════════════════════

# Infrastructure dirs: exclude from coverage % but SCAN for security
INFRA_DIRS := \
	cmd \
	internal/api \
	internal/bootstrap \
	internal/dto \
	internal/model \
	internal/repository/ping

# Infrastructure files: surgical per-file coverage exclusion (still SCANNED).
# For packages that mix tested logic with wiring/setup.
INFRA_FILES :=

# System artifacts: auto-generated, vendored, test infrastructure (NO SCAN)
SYSTEM_DIRS := vendor docs bin internal/test mocks
SYSTEM_FILES := _test.go .pb.go test_helper.go mock.go

# Format conversion for Makefile
comma := ,
space := $(subst ,, )

# Pattern builders for Sonar (Ant-style glob)
SONAR_INFRA_DIRS := $(foreach d,$(INFRA_DIRS),**/$(d)**)
SONAR_INFRA_FILES := $(foreach f,$(INFRA_FILES),**/$(f))
SONAR_SYSTEM_DIRS := $(foreach d,$(SYSTEM_DIRS),**/$(d)**)
SONAR_SYSTEM_FILES := $(foreach f,$(SYSTEM_FILES),**/*$(f))

# Sonar: exclude system artifacts completely (INFRA_FILES intentionally absent → still scanned)
SONAR_EXCLUDE_PATTERNS := $(subst $(space),$(comma),$(strip $(SONAR_SYSTEM_FILES) $(SONAR_SYSTEM_DIRS) $(COVERAGE_DIR)/**))

# Sonar: exclude infrastructure (dirs + files) from coverage % but allow security scan
SONAR_COVERAGE_EXCLUSIONS := $(subst $(space),$(comma),$(strip $(SONAR_INFRA_DIRS) $(SONAR_INFRA_FILES) $(SONAR_SYSTEM_DIRS)))

# Local/Docker: Regex format (coverage.out filtering)
ALL_EXCLUDES := $(INFRA_DIRS) $(INFRA_FILES) $(SYSTEM_DIRS) $(SYSTEM_FILES)
COVERAGE_EXCLUDE := $(subst $(space),|,$(strip $(ALL_EXCLUDES)))

# Go test: Scan all, let grep filter
COVERPKG := ./...

# =============================================================================
# BUILD CONTEXT
# =============================================================================

VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GIT_SHA      ?= $(COMMIT)
GIT_EVENT    ?= local
GIT_REF_TYPE ?= branch
GIT_REF_NAME ?= $(VERSION)

# =============================================================================
# GO COMPILER
# =============================================================================

GO      := go
GOLINT  := golangci-lint
CGO     := 0

LDFLAGS := -ldflags "-s -w \
	-X main.Version=$(VERSION) \
	-X main.Commit=$(COMMIT) \
	-X main.BuildTime=$(BUILD_TIME)"

# =============================================================================
# DOCKER
# =============================================================================

DOCKER_REGISTRY ?= docker.io
DOCKER_NAMESPACE ?= huypham053
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(APP_NAME)
DOCKER_CONTAINER := $(APP_NAME)

CACHE_FROM ?= type=local,src=/tmp/.buildx-cache
CACHE_TO ?= type=local,dest=/tmp/.buildx-cache-new,mode=max

# =============================================================================
# KEYS
# =============================================================================

LOCAL_KEYS_DIR ?= ./keys

# =============================================================================
# MACROS
# =============================================================================

.DEFAULT_GOAL := help

define go-build
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=$(CGO) GOOS=$(1) GOARCH=$(2) $(GO) build $(4) $(LDFLAGS) \
		-o $(BIN_DIR)/$(APP_NAME)-$(1)-$(2)$(3) $(CMD_PATH)
endef

# =============================================================================
# DEVELOPMENT
# =============================================================================

.PHONY: help run dev fmt vet lint tidy vendor

help:
	@echo "Development:"
	@echo "  make run             Run locally"
	@echo "  make dev             Full cycle (fmt → vet → test → swagger → run)"
	@echo "  make fmt             Format code"
	@echo "  make vet             Static analysis"
	@echo "  make lint            Linter"
	@echo "  make tidy            Dependencies"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up [STEPS=n]    Apply all pending migrations (or n)"
	@echo "  make migrate-down [STEPS=n]  Roll back last migration (or n)"
	@echo "  make migrate-version         Show current version + dirty state"
	@echo "  make migrate-force           Recover dirty DB (force version, clears dirty)"
	@echo ""
	@echo "Testing:"
	@echo "  make test            Local tests + coverage report"
	@echo "  make test-coverage   Open coverage HTML"
	@echo ""
	@echo "Build:"
	@echo "  make build           Linux binary"
	@echo "  make build-linux     Cross-compile Linux"
	@echo "  make build-macos     Cross-compile macOS"
	@echo "  make build-windows   Cross-compile Windows"
	@echo "  make release         All platforms"
	@echo ""
	@echo "Mocks:"
	@echo "  make generate-mocks  Generate all mocks"
	@echo "  make clean-mocks     Clean all mocks"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-test     Test in container"
	@echo "  make docker-sonar    SonarCloud scan"
	@echo "  make docker-build    Build image"
	@echo "  make docker-run      Run container"
	@echo "  make docker-stop     Stop container"

run:
	@echo "Starting $(APP_NAME)..."
	SERVICE_NAME=$(APP_NAME) $(GO) run $(CMD_PATH)

dev: fmt vet test swagger run

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint:
	@which $(GOLINT) > /dev/null || (echo "Error: golangci-lint not found. Run: make install-tools"; exit 1)
	$(GOLINT) run ./...

tidy:
	$(GO) mod tidy

vendor:
	$(GO) mod download
	$(GO) mod vendor

# =============================================================================
# TESTING
# =============================================================================

.PHONY: test test-coverage

test:
	@$(GO) clean -testcache
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test ./... -coverprofile=$(COVERAGE_DIR)/coverage.tmp -covermode=atomic -coverpkg=$(COVERPKG) -p 1
	@head -1 $(COVERAGE_DIR)/coverage.tmp > $(COVERAGE_DIR)/coverage.out
	@grep -vE "$(COVERAGE_EXCLUDE)" $(COVERAGE_DIR)/coverage.tmp | tail -n +2 >> $(COVERAGE_DIR)/coverage.out || true
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@total=$$($(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Coverage: $$total%"; \
	if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "FAIL: Below $(COVERAGE_THRESHOLD)% threshold"; exit 1; \
	fi

test-coverage: test
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out

# =============================================================================
# BUILD
# =============================================================================

.PHONY: build build-linux build-macos build-windows build-prod release clean

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)

build-linux:
	$(call go-build,linux,amd64,,)

build-macos:
	$(call go-build,darwin,arm64,,)

build-windows:
	$(call go-build,windows,amd64,.exe,)

build-prod:
	$(call go-build,linux,amd64,-prod,-trimpath)
	@ls -lh $(BIN_DIR)/$(APP_NAME)-linux-amd64-prod

release: clean build-linux build-macos build-windows
	@cd $(BIN_DIR) && sha256sum * > checksums.txt 2>/dev/null || true
	@ls -lh $(BIN_DIR)

# =============================================================================
# CI / CD
# =============================================================================

.PHONY: docker-test docker-sonar docker-login docker-build-push

docker-test:
	@mkdir -p $(COVERAGE_DIR)
	docker buildx build \
		--build-arg COVERAGE_EXCLUDE="$(COVERAGE_EXCLUDE)" \
		--build-arg COVERPKG="$(COVERPKG)" \
		--target test \
		--cache-from=$(CACHE_FROM) \
		--cache-to=$(CACHE_TO) \
		--output type=local,dest=$(COVERAGE_DIR) .
	@if [ -f $(COVERAGE_DIR)/coverage.out ]; then \
		total=$$($(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
		echo "Coverage: $$total%"; \
		if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
			echo "FAIL: Below $(COVERAGE_THRESHOLD)% threshold"; exit 1; \
		fi; \
	else \
		echo "FAIL: coverage.out not found"; exit 1; \
	fi

docker-sonar:
	@[ -n "$(SONAR_TOKEN)" ] || (echo "Error: SONAR_TOKEN not set"; exit 1)
	@docker pull --quiet sonarsource/sonar-scanner-cli:11 || true
	docker run --rm \
		-e SONAR_TOKEN=$(SONAR_TOKEN) \
		-e SONAR_HOST_URL=https://sonarcloud.io \
		-v "$(PWD):/usr/src" \
		sonarsource/sonar-scanner-cli:11 \
		-Dsonar.organization="huypham67" \
		-Dsonar.projectKey="huypham67_bookmark-service" \
		-Dsonar.projectName="$(APP_NAME)" \
		-Dsonar.projectVersion="1.0" \
		-Dsonar.sources="." \
		-Dsonar.tests="." \
		-Dsonar.test.inclusions="**/*_test.go" \
		-Dsonar.test.exclusions="**/vendor/**,**/mocks/**" \
		-Dsonar.exclusions="$(SONAR_EXCLUDE_PATTERNS)" \
		-Dsonar.coverage.exclusions="$(SONAR_COVERAGE_EXCLUSIONS)" \
		-Dsonar.go.coverage.reportPaths="$(COVERAGE_DIR)/coverage.out" \
		-Dsonar.qualitygate.wait=true

docker-login:
	@[ -n "$(DOCKERHUB_USERNAME)" ] && [ -n "$(DOCKERHUB_TOKEN)" ] || (echo "Error: Docker credentials not set"; exit 1)
	echo "$(DOCKERHUB_TOKEN)" | docker login -u "$(DOCKERHUB_USERNAME)" --password-stdin

docker-build-push:
	@if [ "$(GIT_REF_TYPE)" = "tag" ]; then \
		docker buildx build --target final --cache-from=$(CACHE_FROM) --push=true \
			-t $(DOCKER_IMAGE):$(GIT_REF_NAME) -t $(DOCKER_IMAGE):latest .; \
	elif [ "$(GIT_EVENT)" = "pull_request" ]; then \
		docker buildx build --target final --cache-from=$(CACHE_FROM) --push=false \
			-t $(DOCKER_IMAGE):test .; \
	else \
		short_sha=$$(echo "$(GIT_SHA)" | cut -c1-7); \
		docker buildx build --target final --cache-from=$(CACHE_FROM) --push=true \
			-t $(DOCKER_IMAGE):main -t $(DOCKER_IMAGE):$$short_sha -t $(DOCKER_IMAGE):latest .; \
	fi

# =============================================================================
# DOCKER LOCAL
# =============================================================================

.PHONY: docker-run docker-stop docker-logs docker-shell docker-clean

docker-run:
	docker run -d --name $(DOCKER_CONTAINER) -p 8080:8080 --env-file .env $(DOCKER_IMAGE):latest

docker-stop:
	-docker stop $(DOCKER_CONTAINER)
	-docker rm $(DOCKER_CONTAINER)

docker-logs:
	docker logs -f $(DOCKER_CONTAINER)

docker-shell:
	docker exec -it $(DOCKER_CONTAINER) sh

docker-clean:
	-docker rm -f $(DOCKER_CONTAINER)
	-docker rmi -f $$(docker images -q $(DOCKER_IMAGE) 2>/dev/null) 2>/dev/null || true
	docker builder prune --filter type=exec.cachemount --force

# =============================================================================
# DATABASE MIGRATIONS
#
# Thin wrapper over ./cmd/migrate (golang-migrate under the hood). The DB
# connection is read from the service env/config, same as `make run`.
#
#   migrate-up [STEPS=n]   Apply all pending migrations (or only n with STEPS=)
#   migrate-down [STEPS=n] Roll back the last migration (or n with STEPS=)
#   migrate-version        Print the current version + dirty state
#   migrate-force          Recover a DIRTY database. golang-migrate marks the DB
#                          dirty when a migration dies partway, and refuses to
#                          run anything until it is cleared. force sets the
#                          version WITHOUT running SQL and clears the flag — use
#                          it only after manually reconciling the schema.
# =============================================================================

.PHONY: migrate-up migrate-down migrate-force migrate-version

MIGRATE_CMD := $(GO) run ./cmd/migrate
STEPS ?=

migrate-up:
ifeq ($(strip $(STEPS)),)
	@echo "Applying all pending migrations..."
	@$(MIGRATE_CMD)
else
	@echo "Applying $(STEPS) migration(s) up..."
	@$(MIGRATE_CMD) -direction up -steps $(STEPS)
endif

migrate-down:
	@echo "Rolling back $(or $(STEPS),1) migration(s)..."
	@$(MIGRATE_CMD) -direction down -steps $(or $(STEPS),1)

migrate-version:
	@echo "Current migration status:"
	@$(MIGRATE_CMD) -version

migrate-force:
	@echo "⚠  force sets the schema_migrations version WITHOUT running any SQL."
	@echo "   Use it only to clear a 'dirty' state after you have manually"
	@echo "   reconciled the database schema. Run 'make migrate-version' first."
	@read -p "Force schema to which version? " version; \
	if [ -z "$$version" ]; then echo "Aborted: no version given"; exit 1; fi; \
	echo "Forcing migration version to $$version..."; \
	$(MIGRATE_CMD) -force $$version

# =============================================================================
# UTILITIES
# =============================================================================

.PHONY: swagger install-tools info clean clean-docs clean-all generate-mocks clean-mocks

swagger:
	@which swag > /dev/null || (echo "Error: swag not found. Run: make install-tools"; exit 1)
	swag init --parseDependency --parseInternal --generalInfo $(CMD_PATH) --output $(DOCS_DIR)

install-tools:
	$(GO) install github.com/swaggo/swag/cmd/swag@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

info:
	@echo "App:       $(APP_NAME)"
	@echo "Version:   $(VERSION)"
	@echo "Commit:    $(COMMIT)"
	@echo "Built:     $(BUILD_TIME)"
	@echo "Go:        $$($(GO) version)"

generate-mocks:
	@echo "Generating mocks for bookmark repository..."
	cd internal/repository/bookmark && $(GO) generate
	@echo "Generating mocks for cache repository..."
	cd internal/repository/cache && $(GO) generate
	@echo "Generating mocks for link repository..."
	cd internal/repository/link && $(GO) generate
	@echo "Generating mocks for bookmark service..."
	cd internal/service/bookmark && $(GO) generate
	@echo "Generating mocks for health service..."
	cd internal/service/health && $(GO) generate
	@echo "Generating mocks for link service..."
	cd internal/service/link && $(GO) generate
	@echo "Generating mocks for link resolver..."
	cd internal/service/link/resolver && $(GO) generate
	@echo "✓ Mocks generated successfully"

clean-mocks:
	@echo "Cleaning mocks..."
	rm -rf internal/repository/bookmark/mocks
	rm -rf internal/repository/cache/mocks
	rm -rf internal/repository/link/mocks
	rm -rf internal/service/bookmark/mocks
	rm -rf internal/service/health/mocks
	rm -rf internal/service/link/mocks
	rm -rf internal/service/link/resolver/mocks
	@echo "✓ Mocks cleaned"

clean:
	rm -rf $(BIN_DIR) $(COVERAGE_DIR)

clean-docs:
	rm -rf $(DOCS_DIR)

clean-all: clean clean-docs docker-clean
	rm -rf $(LOCAL_KEYS_DIR)
