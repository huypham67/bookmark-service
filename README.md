# bookmark-service

Production-ready REST API service for bookmark and short-link management. Part of the bookmark microservices architecture, extracted from `bookmark-service-monolithic` to run independently with its own database.

## Overview

**bookmark-service** provides bookmark CRUD (with list pagination + cache-aside), URL shortening / redirect, and health checks. It uses PostgreSQL for bookmark persistence and Redis for short-link storage, caching, and rate limiting. JWT auth (validation only) and shared infrastructure come from the `bookmark-common` library — user identity is provided by `user-service` and referenced here as a plain `user_id` column with **no cross-service foreign key**.

## 🎯 Features

- **Bookmark Management**: Create, list (paginated), update, delete — scoped to the authenticated user
- **Cache-aside**: Redis-backed caching of bookmark lists with invalidation on writes
- **URL Shortening**: Create short codes and redirect to the original URL
- **JWT Middleware**: Token validation + user context extraction (validation only; issuance lives in user-service)
- **Rate Limiting**: Redis-backed request throttling
- **Dual Health Check**: Pings **both** PostgreSQL and Redis
- **Database-per-service**: Own PostgreSQL DB, own migrations, no foreign keys to other services
- **Docker + CI/CD**: 5-stage Dockerfile, GitHub Actions with coverage gate + SonarCloud

## 📋 Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26 |
| Web Framework | Gin | v1.12.0 |
| Database | PostgreSQL | (via GORM) |
| Cache / Links | Redis | v9.19.0 |
| Auth | JWT (RSA, validation) | v5.3.1 |
| Logger | Zerolog | v1.35.1 |
| Shared Library | bookmark-common | v0.1.0 |
| API Docs | Swagger (swaggo/gin-swagger) | v1.6.1 |
| Testing | Testify | v1.11.1 |

## 🚀 Quick Start

```bash
cd bookmark-service
go mod download
make gen-keys-local    # JWT RSA keys (public key must match user-service's signer in prod)
createdb bookmark_db
make migrate-up        # apply migrations
make run               # build + run
make test              # local tests + coverage (80% gate)
```

API base: `http://localhost:8080/api/bookmark_service/v1` · Swagger UI: `http://localhost:8080/swagger/index.html` (`make swagger` to regenerate docs)

### Environment (`.env`)
```env
APP_PORT=8080
SERVICE_NAME=bookmark-service
APP_HOST_NAME=/api/bookmark_service
JWT_PUBLIC_KEY_PATH=keys/public.pem
DB_HOST=localhost
DB_PORT=5432
DB_USER=admin
DB_PASSWORD=admin
DB_NAME=bookmark_db
REDIS_ADDR=localhost:6379
```

## 🔌 API Endpoints

Base URL: `/api/bookmark_service/v1`

### Bookmarks *(require `Authorization: Bearer <token>`)*
| Method | Path | Description |
|--------|------|-------------|
| POST | `/bookmarks` | Create a bookmark |
| GET | `/bookmarks?page=&limit=&sort=` | List bookmarks (paginated) |
| PUT | `/bookmarks/:id` | Update a bookmark |
| DELETE | `/bookmarks/:id` | Delete a bookmark |

### Links
| Method | Path | Description |
|--------|------|-------------|
| POST | `/links/shorten` | Create a short code |
| GET | `/links/redirect/:code` | Redirect to original URL (302) |

### Health
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/bookmark_service/health-check` | Pings DB **and** Redis |

## 🩺 Health Check (dual)

```
HealthHandler → HealthService → ping.NewMulti(
                                    ping.NewSQLDB(db),    // PostgreSQL
                                    ping.NewRedis(rdb),   // Redis
                                )
```
Returns `500 FAILED` if **either** dependency is unreachable, `200 OK` otherwise.

## 🗄️ Database (own DB, no cross-service FK)

```
migrations/
  000001_create_bookmarks_table.{up,down}.sql   # user_id is a plain UUID column (no FK to users)
  000002_add_code_int_to_bookmarks.{up,down}.sql
```

## 📁 Structure

```
bookmark-service/
├── cmd/{api,migrate}/main.go
├── internal/
│   ├── api/                 # routing
│   ├── bootstrap/           # DI container, config, routes
│   ├── dto/{bookmark,link,health}/
│   ├── handler/{bookmark,link,health}/
│   ├── model/               # base.go, bookmark.go
│   ├── repository/
│   │   ├── bookmark/         # PostgreSQL CRUD
│   │   ├── cache/            # Redis cache store
│   │   ├── link/             # Redis short-link store
│   │   └── ping/             # pinger.go, sqldb.go, redis.go, multi.go
│   ├── service/{bookmark(+cache),link(+resolver),health}/
│   └── test/{integration,fixtures}/
├── migrations/
├── keys/                    # JWT keys (git-ignored)
├── Makefile / Dockerfile / sonar-project.properties
└── .github/workflows/{ci,cd}.yaml
```

## 🛠️ Make Targets

Run `make help` for the full list. Grouped:

- **Dev**: `run` · `dev` (fmt→vet→test→swagger→run) · `fmt` · `vet` · `lint` · `tidy` · `vendor`
- **Database**: `migrate-up` · `migrate-down` · `migrate-force` · `migrate-version`
- **Testing**: `test` · `test-coverage`
- **Build**: `build` · `build-linux` · `build-macos` · `build-windows` · `build-prod` · `release`
- **Mocks**: `generate-mocks` · `clean-mocks`
- **Docker / CI**: `docker-test` · `docker-sonar` · `docker-build-push` · `docker-run` · `docker-stop` · `docker-logs` · `docker-shell` · `docker-clean`
- **Utilities**: `swagger` · `gen-keys-local` · `install-tools` · `info` · `clean` · `clean-all`

The **Makefile is the single source of truth** for coverage/quality-gate exclusions (`INFRA_DIRS` / `SYSTEM_DIRS`); `sonar-project.properties` carries identity/scope only.

## 🔗 Integration

Consumes `github.com/huypham67/bookmark-common` v0.1.0 for JWT middleware, Redis/SQL clients, rate limiting, logging, short-code, and response utilities. `user_id` values originate from `user-service`; there is intentionally **no foreign key** between the two databases.
