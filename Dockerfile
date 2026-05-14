# Stage 1: Build
FROM golang:tip-20260510-alpine3.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bookmark-service ./cmd/api/main.go

# Stage 2: Run
FROM alpine:3.23

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/bookmark-service .
COPY --from=builder /app/docs ./docs

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./bookmark-service"]