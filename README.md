# Bookmark Service

A production-ready REST API service for URL shortening and link management built with Go, Gin framework, and Redis, following clean architecture principles.

## Overview

Bookmark Service is a modern, scalable microservice designed to handle URL shortening operations with a focus on reliability, performance, and maintainability. It provides health-check monitoring, structured logging with Zerolog, comprehensive testing, and complete API documentation using Swagger/OpenAPI.

## 🎯 Features

- **URL Shortening**: Create shortened URLs with optional expiration time
- **Link Redirect**: Redirect from shortened codes to original URLs
- **Health Check Endpoint**: Monitor service status and Redis connectivity
- **Redis Backend**: Fast, reliable data storage for shortened links
- **Swagger/OpenAPI Documentation**: Interactive API docs at `/swagger/`
- **Environment Configuration**: Flexible setup via environment variables
- **Structured Logging**: Zerolog integration for comprehensive logging
- **Comprehensive Testing**: Unit and integration tests with 90%+ coverage
- **Docker Ready**: Optimized Dockerfile for containerization
- **Cross-Platform Build**: Support for Linux, macOS, and Windows

## 📋 Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26 |
| Web Framework | Gin | v1.12.0 |
| Database | Redis | v9.19.0 |
| Logger | Zerolog | v1.35.1 |
| Input Validation | Validator | v10.30.2 |
| API Documentation | Swagger/OpenAPI | v1.16.6 |
| Testing | Testify | v1.11.1 |
| UUID Generation | google/uuid | v1.6.0 |
| Config Management | envconfig | v1.4.0 |

## 🚀 Quick Start

### Prerequisites

- **Go 1.26** or higher
- **Git**
- **Redis** (required - can use Docker)
- **Make** or **PowerShell** (for Windows)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/huypham67/bookmark-service.git
   cd bookmark-service
   ```

2. **Install dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Install development tools** (Optional)
   ```bash
   make install-tools    # Linux/macOS/WSL
   ```

### Environment Setup

Create a `.env` file in the project root:

```env
# Application Configuration
APP_PORT=8080
SERVICE_NAME=bookmark-service
INSTANCE_ID=instance-1

# Redis Configuration (optional - defaults to localhost:6379)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

**Configuration Reference:**

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `APP_PORT` | No | 8080 | Port on which the API server runs |
| `SERVICE_NAME` | Yes | - | Name of the service for health checks |
| `INSTANCE_ID` | No | Auto-generated UUID | Unique identifier for this service instance |
| `REDIS_ADDR` | No | localhost:6379 | Redis connection address |
| `REDIS_PASSWORD` | No | Empty | Redis password |
| `REDIS_DB` | No | 0 | Redis database number |

### Running the Application

**Using Make (Linux/macOS/WSL):**
```bash
make run           # Run the application
make dev           # Full development workflow (fmt + vet + test + swagger + run)
make test          # Run all tests with coverage
```

**Direct Go command:**
```bash
go run ./cmd/api/main.go
```

The API will be available at `http://localhost:8080` and Swagger docs at `http://localhost:8080/swagger/`

## 📁 Project Structure

```
bookmark-service/
├── cmd/
│   └── api/
│       └── main.go                      # Application entry point
├── internal/
│   ├── api/
│   │   └── router.go                    # Route definitions and setup
│   ├── bootstrap/
│   │   └── app.go                       # Application initialization and DI
│   ├── config/
│   │   └── config.go                    # Configuration management
│   ├── handler/
│   │   ├── health_handler.go            # Health check HTTP handler
│   │   ├── health_handler_test.go
│   │   ├── link_handler.go              # Link management HTTP handlers
│   │   └── link_handler_test.go
│   ├── service/
│   │   ├── health_service.go            # Health check business logic
│   │   ├── health_service_test.go
│   │   ├── link_service.go              # Link shortening business logic
│   │   ├── link_service_test.go
│   │   └── mocks/                       # Mock services for testing
│   ├── repository/
│   │   ├── link_repository.go           # Redis data access layer
│   │   ├── link_repository_test.go
│   │   └── mocks/                       # Mock repositories for testing
│   ├── dto/
│   │   ├── request/
│   │   │   └── shorten_url_request.go
│   │   └── response/
│   │       ├── health_check_response.go
│   │       └── shorten_url_response.go
│   ├── model/
│   ├── integration/                     # Integration tests
│   │   ├── health_check_test.go
│   │   ├── link_redirect_test.go
│   │   ├── link_shorten_test.go
│   │   └── test_helper.go
├── pkg/
│   ├── logger/                          # Zerolog configuration
│   │   ├── logger.go
│   │   └── config.go
│   ├── redis/                           # Redis client wrapper
│   │   ├── client.go
│   │   ├── config.go
│   │   ├── pinger.go                    # Redis connectivity check
│   │   └── mocks/
│   └── utils/                           # Utilities
│       ├── code_generator.go            # Short code generation
│       └── mocks/
├── docs/                                # Generated Swagger documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── coverage/                            # Test coverage reports
├── Dockerfile                           # Docker configuration
├── Makefile                             # Build automation
├── go.mod                               # Go module definition
├── go.sum                               # Module checksums
├── .env                                 # Environment variables (local)
├── .gitignore
└── README.md                            # This file
```

## 🔌 API Endpoints

### Health Check
Check application health status and Redis connectivity.

```
GET /api/health-check
```

**Response:**
```json
{
  "message": "OK",
  "service_name": "bookmark-service",
  "instance_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Status Codes:**
- `200 OK` - Service is healthy and Redis is connected
- `500 Internal Server Error` - Service or Redis connection failed

---

### Shorten URL
Create a shortened URL code and save it to Redis.

```
POST /api/v1/links/shorten
```

**Request Body:**
```json
{
  "url": "https://example.com/very/long/url",
  "exp": 86400
}
```

**Parameters:**
- `url` (string, required): The original URL to shorten (must be valid URL)
- `exp` (integer, optional): Expiration time in seconds (default: no expiration)

**Response (200 OK):**
```json
{
  "code": "ab12cd34",
  "message": "Shorten URL generated successfully"
}
```

**Status Codes:**
- `200 OK` - Successfully created shortened URL
- `400 Bad Request` - Invalid request body or validation failed
- `500 Internal Server Error` - Internal error (Redis unavailable, etc.)

---

### Redirect to Original URL
Redirect from shortened code to original URL.

```
GET /api/v1/links/redirect/:code
```

**Parameters:**
- `code` (path parameter, required): The shortened code

**Response:**
- `302 Found` - Redirects to the original URL
- `404 Not Found` - Shortened code doesn't exist or expired
- `500 Internal Server Error` - Internal error

---

## Interactive API Documentation

Access Swagger UI at:
```
http://localhost:8080/swagger/
```

This provides an interactive interface to test all API endpoints.

## 🏗️ Architecture

### Clean Architecture Pattern
The project follows clean architecture principles with clear separation of concerns:

- **Handler Layer** (`internal/handler/`): HTTP request/response handling and validation
- **Service Layer** (`internal/service/`): Business logic and domain operations
- **Repository Layer** (`internal/repository/`): Data persistence abstraction
- **DTO Layer** (`internal/dto/`): Data transfer objects for API requests/responses
- **Config Layer** (`internal/config/`): Configuration management
- **API Layer** (`internal/api/`): Route definitions and middleware setup

### Dependency Injection
The application uses constructor-based dependency injection in the bootstrap layer for loose coupling and testability.

### Design Patterns Used
- **Repository Pattern**: Abstracts Redis data access
- **Service Pattern**: Encapsulates business logic
- **Dependency Injection**: Loose coupling and testability
- **Handler Pattern**: Clean HTTP request handling

## 🧪 Testing

### Run Tests

**Using Make:**
```bash
make test              # Run all tests with coverage
make test-verbose      # Run tests with verbose output
make test-coverage     # Generate and view HTML coverage report
```

**Direct Go:**
```bash
go test -v -race ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Coverage

The project aims for 90%+ coverage on main business logic:

- **Handler Tests**: HTTP request/response handling
- **Service Tests**: Business logic and error cases
- **Repository Tests**: Redis operations with miniredis
- **Integration Tests**: End-to-end API flows

### Test Types

1. **Unit Tests**: Individual layer tests with mocks
2. **Integration Tests**: Full API flow tests with miniredis
3. **Mock Generation**: Uses mockery for interface mocking

## 📦 Building

### Development Build

```bash
make build        # Build binary for current OS
```

### Cross-Platform Builds

```bash
make build-linux    # Build for Linux (amd64)
make build-macos    # Build for macOS (arm64)
make build-windows  # Build for Windows (amd64)
make release        # Build for all platforms
```

### Code Quality

**Format Code:**
```bash
make fmt          # Format all Go files
```

**Run Linting:**
```bash
make lint         # Run golangci-lint (requires installation)
```

**Run Go Vet:**
```bash
make vet          # Run go vet analysis
```

## 🐳 Docker

### Build Docker Image

```bash
# Build production binary
make build-prod

# Build Docker image
docker build -t bookmark-service:latest .
```

### Run Docker Container

```bash
docker run -d \
  -e APP_PORT=8080 \
  -e SERVICE_NAME=bookmark-service \
  -e REDIS_ADDR=host.docker.internal:6379 \
  -p 8080:8080 \
  --name bookmark-service \
  bookmark-service:latest
```

### Docker Compose (Optional)

```bash
docker-compose up -d
```

## 🔄 Development Workflow

### Quick Development

```bash
make fmt           # Format code
make vet           # Check code issues
go run ./cmd/api/main.go  # Run application
```

### Full Development Workflow

```bash
make dev           # Format → Vet → Test → Build Swagger → Run
```

## 🧹 Cleanup

**Remove build artifacts:**
```bash
make clean         # Remove binaries and coverage
```

**Remove documentation:**
```bash
make clean-docs    # Remove Swagger docs
```

**Full cleanup:**
```bash
make clean-all     # Remove everything including vendor
```

## 📊 Project Information

View project details:

```bash
make info          # Show version, commit, build time
```

## 🛠️ Available Make Targets

```bash
make help          # Display all available targets
```

### Common Targets

| Target | Description |
|--------|-------------|
| `make run` | Run the application |
| `make test` | Run tests with coverage |
| `make build` | Build binary for current OS |
| `make fmt` | Format code with gofmt |
| `make vet` | Run go vet analysis |
| `make swagger` | Generate Swagger documentation |
| `make clean` | Remove build artifacts |
| `make help` | Show all available targets |

## 🚧 Development Notes

### Adding New Endpoints

1. Create DTOs in `internal/dto/request` and `internal/dto/response`
2. Create handler methods in `internal/handler/`
3. Create service logic in `internal/service/`
4. Create repository methods in `internal/repository/` if needed
5. Register routes in `internal/api/router.go`
6. Add Swagger comments to handler methods
7. Write unit and integration tests
8. Regenerate Swagger docs: `make swagger`

### Adding Tests

```bash
# Unit tests should be in same package
# Integration tests go in internal/integration/

# Run specific test
go test -v -run TestName ./path/to/package

# Run with coverage
go test -coverprofile=coverage.out ./...
```

### Regenerating Swagger Docs

After adding/modifying Swagger comments:

```bash
make swagger       # Regenerates docs.go, swagger.json, swagger.yaml
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow code style: `make fmt`
4. Write tests for new features
5. Ensure tests pass: `make test`
6. Update Swagger docs if API changed
7. Commit with clear messages
8. Push to your branch
9. Open a Pull Request

### Code Standards
- Always run `make fmt` before committing
- Write unit tests for new features (aim for 90%+ coverage)
- Update API documentation for new endpoints
- Follow Go conventions and idioms
- Keep functions small and focused

## 📝 Environment Variables Reference

### Application

- `APP_PORT`: Server port (default: 8080)
- `SERVICE_NAME`: Service identifier for health checks (required)
- `INSTANCE_ID`: Unique instance identifier (auto-generated if not set)

### Redis

- `REDIS_ADDR`: Redis address (default: localhost:6379)
- `REDIS_PASSWORD`: Redis password (optional)
- `REDIS_DB`: Redis database number (default: 0)

## 🔐 Security Considerations

- **Environment Variables**: Store sensitive data in `.env` (not in version control)
- **Input Validation**: All inputs are validated using validator v10
- **SQL Injection**: N/A (using Redis, not SQL)
- **CORS**: Configure as needed for production
- **HTTPS**: Use reverse proxy (nginx, etc.) in production
- **Rate Limiting**: Consider adding middleware for production

## 📄 License

This project is licensed under the MIT License.

## 📞 Support

For issues, questions, or suggestions, please create an issue on GitHub.

## 🔗 Useful Links

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/)
- [Swagger/OpenAPI](https://swagger.io/)
- [Redis Documentation](https://redis.io/docs/)
- [Zerolog Logger](https://github.com/rs/zerolog)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

