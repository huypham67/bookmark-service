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

# Single source of truth: list infrastructure packages that don't require tests
INFRA_DIRS  := cmd internal/bootstrap pkg/logger pkg/redis pkg/sqldb
INFRA_FILES := config.go loader.go claims.go router.go test_helper.go

# System artifacts: auto-generated, vendored, or test infrastructure
SYSTEM_FILES := _test.go mocks vendor docs bin testutil .pb.go

# Build all exclude patterns from raw lists (automatic format conversion)
comma := ,
space := $(subst ,, )

# SONAR: Ant-style glob format
SONAR_DIRS := $(foreach d,$(INFRA_DIRS),**/$(d)**)
SONAR_FILES := $(foreach f,$(INFRA_FILES),**/$(f))
SONAR_TRASH_FILES := $(foreach f,$(filter %.go,$(SYSTEM_FILES)),**/*$(f))
SONAR_TRASH_DIRS := $(foreach d,$(filter-out %.go,$(SYSTEM_FILES)),**/$(d)**)

SONAR_EXCLUDE_PATTERNS := $(subst $(space),$(comma),$(strip $(SONAR_TRASH_FILES) $(SONAR_TRASH_DIRS) $(COVERAGE_DIR)/**))
SONAR_COVERAGE_EXCLUSIONS := $(subst $(space),$(comma),$(strip $(SONAR_DIRS) $(SONAR_FILES)))

# Local/Docker: Regex format
ALL_EXCLUDES := $(INFRA_DIRS) $(INFRA_FILES) $(SYSTEM_FILES)
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

VM_KEYS_DIR ?= /opt/bookmark-service/keys
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
	@echo "Docker:"
	@echo "  make docker-test     Test in container"
	@echo "  make docker-sonar    SonarCloud scan"
	@echo "  make docker-build    Build image"
	@echo "  make docker-run      Run container"
	@echo "  make docker-stop     Stop container"
	@echo ""
	@echo "Compose:"
	@echo "  make compose-up      Start stack (Redis, Postgres, App)"
	@echo "  make compose-down    Stop stack"
	@echo ""
	@echo "Keys:"
	@echo "  make gen-keys        Generate keys on VM (sudo)"
	@echo "  make gen-keys-local  Generate keys locally"

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
	@grep -vE "$(COVERAGE_EXCLUDE)" $(COVERAGE_DIR)/coverage.tmp > $(COVERAGE_DIR)/coverage.out || touch $(COVERAGE_DIR)/coverage.out
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
# COMPOSE
# =============================================================================

.PHONY: compose-up compose-down compose-logs compose-restart

compose-up:
	docker compose up --build -d

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f

compose-restart:
	docker compose down && docker compose up --build -d

# =============================================================================
# UTILITIES
# =============================================================================

.PHONY: swagger install-tools info clean clean-docs clean-all gen-keys gen-keys-local

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

gen-keys:
	sudo mkdir -p $(VM_KEYS_DIR)
	sudo openssl genpkey -algorithm RSA -out $(VM_KEYS_DIR)/private.pem -pkeyopt rsa_keygen_bits:2048
	sudo openssl rsa -pubout -in $(VM_KEYS_DIR)/private.pem -out $(VM_KEYS_DIR)/public.pem
	sudo chmod 600 $(VM_KEYS_DIR)/private.pem
	sudo chmod 644 $(VM_KEYS_DIR)/public.pem

gen-keys-local:
	mkdir -p $(LOCAL_KEYS_DIR)
	openssl genpkey -algorithm RSA -out $(LOCAL_KEYS_DIR)/private.pem -pkeyopt rsa_keygen_bits:2048
	openssl rsa -pubout -in $(LOCAL_KEYS_DIR)/private.pem -out $(LOCAL_KEYS_DIR)/public.pem

clean:
	rm -rf $(BIN_DIR) $(COVERAGE_DIR)

clean-docs:
	rm -rf $(DOCS_DIR)

clean-all: clean clean-docs docker-clean
	rm -rf $(LOCAL_KEYS_DIR)