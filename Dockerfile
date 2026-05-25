# Stage 1: BASE - Install Dependencies and Prepare Source
FROM golang:1.26-alpine AS base

RUN apk add --no-cache build-base git

WORKDIR /opt/app

# Tách riêng copy go.mod/go.sum để tận dụng layer cache
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# Stage 2: BUILD - Compile Binary with Optimizations
FROM base AS build

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w" \
    -o bookmark-service \
    ./cmd/api/main.go

# Stage 3: TEST - Run Tests and Generate Coverage Reports
FROM base AS test-exec

ARG COVERAGE_EXCLUDE
ENV _OUTPUTDIR=/tmp/coverage

RUN mkdir -p ${_OUTPUTDIR}

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 go test ./... \
      -coverprofile=coverage.tmp \
      -covermode=atomic \
      -coverpkg=./internal/... \
      -p 1 && \
    grep -v -E "${COVERAGE_EXCLUDE}" coverage.tmp > ${_OUTPUTDIR}/coverage.out && \
    go tool cover -html=${_OUTPUTDIR}/coverage.out -o ${_OUTPUTDIR}/coverage.html

# Stage 4: TEST-REPORT - Extract Coverage Reports for CI/CD
FROM scratch AS test

COPY --from=test-exec /tmp/coverage/coverage.out /coverage.out
COPY --from=test-exec /tmp/coverage/coverage.html /coverage.html

# Stage 5: FINAL - Produce Minimal Runtime Image with Binary and Documentation
FROM alpine:3.19 AS final

ENV SERVICE_USER=svc-bookmark \
    SERVICE_GROUP=svc-bookmark \
    TZ=Asia/Ho_Chi_Minh

RUN addgroup -S ${SERVICE_GROUP} && \
    adduser -S ${SERVICE_USER} -G ${SERVICE_GROUP}

WORKDIR /app

COPY --from=build /opt/app/bookmark-service .
COPY --from=build /opt/app/docs ./docs

RUN chown -R ${SERVICE_USER}:${SERVICE_GROUP} /app && \
    apk add --no-cache tzdata && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone

USER ${SERVICE_USER}

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD test -f /app/bookmark-service || exit 1

CMD ["./bookmark-service"]