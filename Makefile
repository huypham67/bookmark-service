# Makefile for bookmark-service
# Structure: Y CHANG MONO - Same approach as bookmark-service-monolithic

APP_NAME := bookmark-service
BIN_DIR := ./bin
COVERAGE_DIR ?= coverage_report
COVERAGE_THRESHOLD ?= 80

# ═══════════════════════════════════════════════════════════════════════════
# SINGLE SOURCE OF TRUTH: Coverage & Quality Gate Exclusions
#
# INFRA_DIRS: exclude from coverage % but SCAN for security
INFRA_DIRS := cmd internal/api internal/bootstrap internal/dto internal/model \
              internal/repository/ping

# SYSTEM_DIRS: no scan, no coverage
SYSTEM_DIRS := vendor docs bin mocks internal/test
SYSTEM_FILES := _test.go .pb.go test_helper.go mock.go

# Format conversion for Makefile
comma := ,
space := $(subst ,, )

# Pattern builders for Sonar (Ant-style glob)
SONAR_INFRA_DIRS := $(foreach d,$(INFRA_DIRS),**/$(d)**)
SONAR_SYSTEM_DIRS := $(foreach d,$(SYSTEM_DIRS),**/$(d)**)
SONAR_SYSTEM_FILES := $(foreach f,$(SYSTEM_FILES),**/*$(f))

# Sonar: exclude system artifacts completely
SONAR_EXCLUDE_PATTERNS := $(subst $(space),$(comma),$(strip $(SONAR_SYSTEM_FILES) $(SONAR_SYSTEM_DIRS) $(COVERAGE_DIR)/**))

# Sonar: exclude infrastructure from coverage % but allow security scan
SONAR_INFRA_DIRS_FLAT := $(foreach d,$(INFRA_DIRS),**/$(d)**)
SONAR_COVERAGE_EXCLUSIONS := $(subst $(space),$(comma),$(strip $(SONAR_INFRA_DIRS_FLAT) $(SONAR_SYSTEM_DIRS)))

# Local/Docker: Regex format (coverage.out filtering)
ALL_EXCLUDES := $(INFRA_DIRS) $(INFRA_FILES) $(SYSTEM_DIRS) $(SYSTEM_FILES)
COVERAGE_EXCLUDE := $(subst $(space),|,$(strip $(ALL_EXCLUDES)))

GO := go
COVERPKG := ./...

# Build cache
CACHE_FROM ?= type=local,src=/tmp/.buildx-cache
CACHE_TO ?= type=local,dest=/tmp/.buildx-cache-new,mode=max

# Docker
DOCKER_REGISTRY ?= docker.io
DOCKER_NAMESPACE ?= huypham053
DOCKER_IMAGE := $(DOCKER_REGISTRY)/$(DOCKER_NAMESPACE)/$(APP_NAME)

.PHONY: help test test-coverage fmt vet lint tidy clean \
        docker-test docker-sonar build run swagger gen-keys

help:
	@echo "bookmark-service - Bookmark & Link Management Service"
	@echo ""
	@echo "Local Development:"
	@echo "  make test          Run tests + coverage report (80% threshold)"
	@echo "  make test-coverage Open coverage HTML"
	@echo "  make fmt           Format code (go fmt)"
	@echo "  make vet           Static analysis (go vet)"
	@echo "  make lint          Run linter"
	@echo "  make tidy          Tidy dependencies"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build         Build binary"
	@echo "  make run           Run the service"
	@echo "  make swagger       Generate Swagger/OpenAPI docs"
	@echo "  make gen-keys      Generate JWT RSA key pair"
	@echo ""
	@echo "Docker / CI:"
	@echo "  make docker-test   Test in Docker with coverage extraction"
	@echo "  make docker-sonar  SonarCloud code quality scan"
	@echo "  make clean         Remove artifacts"

# ═══════════════════════════════════════════════════════════════════════════
# LOCAL TESTING
# ═══════════════════════════════════════════════════════════════════════════

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

fmt:
	$(GO) fmt ./...
	@echo "✓ Formatted"

vet:
	$(GO) vet ./...
	@echo "✓ Vet passed"

lint:
	@which golangci-lint > /dev/null || (echo "Error: golangci-lint not found"; exit 1)
	golangci-lint run ./...
	@echo "✓ Lint passed"

tidy:
	$(GO) mod tidy
	@echo "✓ Tidied"

# ═══════════════════════════════════════════════════════════════════════════
# BUILD & RUN
# ═══════════════════════════════════════════════════════════════════════════

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(APP_NAME) ./cmd/api
	@echo "✓ Built $(BIN_DIR)/$(APP_NAME)"

run: build
	./$(BIN_DIR)/$(APP_NAME)

# ═══════════════════════════════════════════════════════════════════════════
# SWAGGER & KEYS
# ═══════════════════════════════════════════════════════════════════════════

swagger:
	@which swag > /dev/null || (echo "Error: swag not found. Install: go install github.com/swaggo/swag/cmd/swag@latest"; exit 1)
	swag init --parseDependency --parseInternal --generalInfo ./cmd/api/main.go --output ./docs
	@echo "✓ Swagger docs generated"

gen-keys:
	@mkdir -p keys
	@if [ -f keys/private.pem ] && [ -f keys/public.pem ]; then \
		echo "✓ Keys already exist"; \
	else \
		echo "Generating RSA key pair..."; \
		openssl genrsa -out keys/private.pem 2048 2>/dev/null; \
		openssl rsa -in keys/private.pem -pubout -out keys/public.pem 2>/dev/null; \
		echo "✓ Keys generated: keys/private.pem, keys/public.pem"; \
	fi

# ═══════════════════════════════════════════════════════════════════════════
# DOCKER TESTING
# ═══════════════════════════════════════════════════════════════════════════

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

# ═══════════════════════════════════════════════════════════════════════════
# SONARCLOUD CODE QUALITY
# ═══════════════════════════════════════════════════════════════════════════

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

# ═══════════════════════════════════════════════════════════════════════════
# CLEANUP
# ═══════════════════════════════════════════════════════════════════════════

clean:
	rm -rf $(BIN_DIR) $(COVERAGE_DIR) docs/
	@echo "✓ Cleaned"
