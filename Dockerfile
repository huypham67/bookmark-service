# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bookmark-service ./cmd/api/main.go

# Stage 2: Runtime
FROM alpine:3.23

ENV SERVICE_USER=svc-bookmark \
    SERVICE_GROUP=svc-bookmark

RUN addgroup -S ${SERVICE_GROUP} && \
    adduser -S ${SERVICE_USER} -G ${SERVICE_GROUP}

WORKDIR /app

COPY --from=builder /src/bookmark-service .
COPY --from=builder /src/docs ./docs

RUN chown -R ${SERVICE_USER}:${SERVICE_GROUP} /app

USER ${SERVICE_USER}

EXPOSE 8080

CMD ["./bookmark-service"]