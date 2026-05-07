# Bookmark Management API

A production-ready REST API service for bookmark management built with Go, Gin framework, and following clean architecture principles.

## Overview

The Bookmark Management API is a modern, scalable microservice designed to handle bookmark operations with a focus on reliability, performance, and maintainability. It provides comprehensive health-check monitoring, structured logging, and complete API documentation using Swagger/OpenAPI.

## 🎯 Features

- **Health Check Endpoint**: Monitor service status with detailed instance information
- **RESTful API**: Clean and intuitive API design following REST principles
- **Swagger/OpenAPI Documentation**: Interactive API documentation accessible via `/swagger/`
- **Environment Configuration**: Flexible configuration via environment variables
- **Structured Logging**: Integration with Gin framework for logging
- **Cross-Platform Support**: Build binaries for Windows, Linux, and macOS
- **Comprehensive Testing**: Unit tests with code coverage reporting
- **Docker Ready**: Optimized for containerization
- **Production Build Optimization**: Minimal binary sizes with version/commit tracking

## 📋 Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26.2 |
| Web Framework | Gin | v1.12.0 |
| API Documentation | Swagger/OpenAPI | v1.16.6 |
| Testing | Testify | v1.11.1 |
| UUID Generation | google/uuid | v1.6.0 |
| Config Management | envconfig | v1.4.0 |

## 🚀 Quick Start

### Prerequisites

- **Go 1.26.2** or higher
- **Git**
- **Make** or **PowerShell** (for Windows)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/huypham67/bookmark-management.git
   cd bookmark-management
   ```

2. **Install dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Install development tools** (Optional but recommended)
   ```bash
   make install-tools    # Linux/macOS/WSL
   .\build.ps1 -Task install-tools  # Windows PowerShell
   ```

### Environment Setup

Create a `.env` file in the project root (or set environment variables):

```env
# Application Configuration
APP_PORT=8080
SERVICE_NAME=bookmark-service
INSTANCE_ID=your-unique-instance-id
```

**Configuration Reference:**

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `APP_PORT` | No | 8080 | Port on which the API server runs |
| `SERVICE_NAME` | Yes | - | Name of the service for health checks |
| `INSTANCE_ID` | No | Auto-generated UUID | Unique identifier for this service instance |

### Running the Application

**Using Make (Linux/macOS/WSL):**
```bash
make run           # Run the application
make dev           # Full development workflow
make build         # Build binary for current OS
```

**Using PowerShell (Windows):**
```powershell
.\build.ps1 -Task run
.\build.ps1 -Task dev
.\build.ps1 -Task build
```

**Direct Go command:**
```bash
go run ./cmd/api/main.go
```

## 📁 Project Structure

```
bookmark-management/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   └── router.go            # Route definitions and setup
│   ├── bootstrap/
│   │   └── app.go               # Application initialization
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── handler/
│   │   ├── health_handler.go    # HTTP request handlers
│   │   └── health_handler_test.go
│   ├── service/
│   │   ├── health_service.go    # Business logic
│   │   └── health_service_test.go
│   ├── model/
│   │   └── HealthCheckResponse.go # Data models
│   ├── repository/               # Data access layer (extendable)
│   └── integration/              # Integration tests
├── docs/
│   ├── docs.go                  # Generated Swagger documentation
│   ├── swagger.json
│   └── swagger.yaml
├── mocks/                        # Mock implementations for testing
├── Makefile                      # Build automation (Unix-like systems)
├── build.ps1                     # Build automation (Windows PowerShell)
├── go.mod                        # Go module definition
├── go.sum                        # Module checksums
├── .env                          # Environment configuration
└── README.md                     # This file
```

## 🔌 API Endpoints

### Health Check
Returns the current health status of the service.

```
GET /api/v1/health-check
```

**Response Example:**
```json
{
  "message": "OK",
  "service_name": "bookmark-service",
  "instance_id": "7ab8c9d0-e1f2-4a3b-8c9d-0e1f2a3b4c5d"
}
```

**Status Codes:**
- `200 OK` - Service is healthy

### Swagger/OpenAPI Documentation

Interactive API documentation is available at:
```
http://localhost:8080/swagger/
```

## 🏗️ Architecture

### Clean Architecture Pattern
The project follows clean architecture principles with clear separation of concerns:

- **Handler Layer** (`internal/handler/`): HTTP request/response handling
- **Service Layer** (`internal/service/`): Business logic implementation
- **Repository Layer** (`internal/repository/`): Data persistence (extendable)
- **Model Layer** (`internal/model/`): Data structures and DTOs
- **Config Layer** (`internal/config/`): Configuration management
- **API Layer** (`internal/api/`): Route definitions and middleware

### Dependency Injection
The application uses constructor-based dependency injection for loose coupling and testability.

## 🧪 Development & Testing

### Run Tests

**With Make:**
```bash
make test              # Run all tests with coverage
make test-verbose      # Run tests with verbose output
make test-coverage     # Generate and view HTML coverage report
```

**With PowerShell:**
```powershell
.\build.ps1 -Task test
.\build.ps1 -Task test-verbose
.\build.ps1 -Task test-coverage
```

**With Go directly:**
```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Code Quality

**Format Code:**
```bash
make fmt          # Unix-like
.\build.ps1 -Task fmt  # Windows
go fmt ./...      # Direct
```

**Run Linting:**
```bash
make lint         # Unix-like (requires golangci-lint)
.\build.ps1 -Task lint  # Windows
```

**Run Go Vet:**
```bash
make vet          # Unix-like
.\build.ps1 -Task vet  # Windows
go vet ./...      # Direct
```

## 📦 Building

### Development Build

```bash
make build        # Unix-like
.\build.ps1 -Task build  # Windows
```

### Cross-Platform Builds

```bash
# Build for specific platforms
make build-linux
make build-macos
make build-windows

# Windows PowerShell
.\build.ps1 -Task build-linux
.\build.ps1 -Task build-macos
.\build.ps1 -Task build-windows
```

### Production Build

Optimized production binary with minimal size:

```bash
make build-prod   # Unix-like
.\build.ps1 -Task build-prod  # Windows
```

**Build flags applied:**
- `-trimpath`: Remove file system path information
- `-a -installsuffix cgo`: Force rebuild of dependent packages
- Version embedding: `main.Version`, `main.Commit`, `main.BuildTime`

### Release Package

Create release binaries for all platforms:

```bash
make release      # Unix-like
.\build.ps1 -Task release  # Windows
```

## 🐳 Docker Deployment

### Create Dockerfile

```dockerfile
FROM golang:1.26.2-alpine AS builder
WORKDIR /build
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-X main.Version=$(git describe --tags --always) \
    -X main.Commit=$(git rev-parse --short HEAD)" \
    -o bookmark-management ./cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/bookmark-management .
EXPOSE 8080
CMD ["./bookmark-management"]
```

### Build and Run Docker Image

```bash
# Build image
docker build -t bookmark-management:latest .

# Run container
docker run -e APP_PORT=8080 \
           -e SERVICE_NAME=bookmark-service \
           -p 8080:8080 \
           bookmark-management:latest
```

## 🔄 Development Workflow

### Quick Development Cycle
```bash
make dev-quick    # Format → Vet → Run (Unix-like)
.\build.ps1 -Task dev-quick  # Windows
```

### Full Development Workflow
```bash
make dev          # Format → Vet → Test → Swagger → Run (Unix-like)
.\build.ps1 -Task dev  # Windows
```

## 🧹 Cleanup

**Remove build artifacts:**
```bash
make clean        # Unix-like
.\build.ps1 -Task clean  # Windows
```

**Remove documentation:**
```bash
make clean-docs   # Unix-like
.\build.ps1 -Task clean-docs  # Windows
```

**Full cleanup:**
```bash
make clean-all    # Unix-like
.\build.ps1 -Task clean-all  # Windows
```

## 📊 Project Information

Check project details with version and commit information:

```bash
make info         # Unix-like
.\build.ps1 -Task info  # Windows
```

## 🛠️ Available Build Targets

### Unix/Linux/macOS (Make)
```bash
make help         # Display all available targets
```

### Windows (PowerShell)
```powershell
.\build.ps1 -Task help
```

## 📝 Configuration Examples

### Development Environment
```env
APP_PORT=8080
SERVICE_NAME=bookmark-service-dev
INSTANCE_ID=dev-instance-1
```

### Production Environment
```env
APP_PORT=8080
SERVICE_NAME=bookmark-service
INSTANCE_ID=prod-instance-1
```

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Follow** the code style and run tests
4. **Commit** with clear messages (`git commit -m 'Add amazing feature'`)
5. **Push** to your branch (`git push origin feature/amazing-feature`)
6. **Open** a Pull Request

### Code Standards
- Run `make fmt` before committing
- Ensure all tests pass: `make test`
- Write unit tests for new features
- Update API documentation for new endpoints
- Follow Go best practices and idioms

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 📞 Support & Issues

For issues, bug reports, or feature requests, please open an issue on GitHub.

## 🔐 Security

- All dependencies are regularly updated
- Environment-sensitive data must be provided via environment variables
- API endpoints include request validation
- Use HTTPS in production environments
- Implement rate limiting for production deployments

## 📚 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/)
- [Swagger OpenAPI](https://swagger.io/)
- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
