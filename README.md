# bookmark-service

REST API service for bookmark management, URL shortening, and CSV bulk import. Part of the bookmark microservices architecture.

## Overview

**bookmark-service** provides bookmark CRUD with cache-aside, URL shortening/redirect, bulk import via CSV, and health checks. It uses PostgreSQL for bookmark persistence and Redis for short-link storage, bookmark cache, rate limiting, and the import job queue. JWT validation (no issuance) and shared infrastructure come from `bookmark-common`. User identity is provided by `user-service` and stored here as a plain `user_id` column with no cross-service foreign key.

Migrations run **automatically on startup** via `sqldb.RunMigration` in `NewContainer`. The manual CLI (`cmd/migrate`) is available for rollbacks and recovery.

## Tech Stack

| Component | Technology | Version |
|---|---|---|
| Language | Go | 1.26 |
| Web framework | Gin | v1.12.0 |
| Database | PostgreSQL (GORM) | v1.31.1 / v1.6.0 |
| Cache / Links / Queue | Redis (go-redis) | v9.19.0 |
| Auth | JWT RS256 (validate only) | v5.3.1 |
| Logger | Zerolog | v1.35.1 |
| Shared library | bookmark-common | v0.3.0 |
| API docs | Swagger (swaggo/gin-swagger) | v1.6.1 |
| Testing | Testify + miniredis + SQLite | v1.11.1 |

## Quick Start

```bash
cd bookmark-service
cp .env.example .env          # edit as needed
cp ../user-service/keys/public.pem keys/public.pem   # bookmark-service validates only
make run                      # starts on :8080; migrations run automatically
```

Swagger UI: `http://localhost:8080/swagger/index.html`  
API base: `http://localhost:8080/api/bookmark_service/v1`

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `APP_PORT` | `8080` | HTTP listen port |
| `SERVICE_NAME` | _(required)_ | Service identifier |
| `APP_HOST_NAME` | `/api/bookmark_service` | API base path |
| `APP_ENV` | `development` | Environment (`development` / `production`) |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | — | PostgreSQL user |
| `DB_PASSWORD` | — | PostgreSQL password |
| `DB_NAME` | `bookmark_db` | PostgreSQL database |
| `DB_SSLMODE` | `disable` | PostgreSQL SSL mode |
| `DB_TIMEZONE` | `UTC` | PostgreSQL timezone |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASSWORD` | — | Redis password |
| `REDIS_DATABASE` | `1` | Redis logical DB index |
| `RATELIMIT_LIMIT` | `20` | Requests per window |
| `RATELIMIT_WINDOW` | `10s` | Rate-limit window |
| `JWT_PUBLIC_KEY_PATH` | `keys/public.pem` | RSA public key path (validate only) |
| `JWT_ISSUER` | `user-service` | Expected JWT `iss` claim |
| `JWT_AUDIENCE` | `bookmark-app` | Expected JWT `aud` claim |

## API Endpoints

All routes under `/api/bookmark_service`.

### Health
| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/health-check` | — | Pings PostgreSQL **and** Redis |

### Links (under `/v1`)
| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/v1/links/shorten` | — | Create a short code (Redis-backed) |
| GET | `/v1/links/redirect/:code` | — | Redirect to original URL (302) |

### Bookmarks (under `/v1`, require `Authorization: Bearer <token>`)
| Method | Path | Description |
|---|---|---|
| POST | `/v1/bookmarks` | Create a bookmark |
| POST | `/v1/bookmarks/import` | Bulk import via CSV file upload |
| GET | `/v1/bookmarks` | List own bookmarks (paginated) |
| PUT | `/v1/bookmarks/:id` | Update a bookmark |
| DELETE | `/v1/bookmarks/:id` | Delete a bookmark |

## Short-link Routing

Short codes carry a 1-byte prefix that determines the backing store:

| Prefix | Store | Created by |
|---|---|---|
| `a–h` | Redis | `POST /v1/links/shorten` |
| `i–z` | PostgreSQL | Bookmark creation (encoded from `code_int` serial) |

`Classify(code)` in `bookmark-common/pkg/shortcode` reads the first byte to route redirects.

## Bookmark Import Flow

```
POST /v1/bookmarks/import  (multipart CSV)
  │
  ├── validate CSV (csvutil.Decode)
  ├── build job JSON  {job_id, user_id, records: [{url, description}, ...]}
  └── LPUSH → Redis list  "bookmark:import:jobs"
                              │
                              ▼
                      bookmark-worker (RPOP)
                        └── insert bookmarks → PostgreSQL
```

bookmark-worker handles the actual DB writes asynchronously. bookmark-service is the queue producer only.

## Health Check

```
GET /health-check
  → ping.NewMulti(
        ping.NewSQLDB(db),    // PostgreSQL
        ping.NewRedis(rdb),   // Redis
    )
```

Returns `500` if **either** dependency is unreachable.

## Database Migrations

Migrations run automatically on startup. The `cmd/migrate` binary provides manual control:

```bash
make migrate-up              # apply all pending
make migrate-up STEPS=1      # apply 1 step
make migrate-down STEPS=1    # roll back 1 step
make migrate-version         # show current version + dirty flag
make migrate-force           # clear dirty flag (interactive prompt)
```

### Migration history

| Version | File | Description |
|---|---|---|
| 0001 | `create_bookmarks_table` | `bookmarks` table; `code` column `NOT NULL UNIQUE` |
| 0002 | `add_code_int_to_bookmarks` | `code_int SERIAL UNIQUE` column |
| 0003 | `relax_code_column` | Drop `idx_bookmarks_code`; make `code` nullable |
| 0004 | `drop_code_unique_constraint` | Drop `bookmarks_code_key` unique constraint (enables multi-row insert with empty code before shortcode assignment) |

## Testing

```bash
make test              # unit + integration tests, 80% coverage gate
make test-coverage     # open HTML coverage report
make docker-test       # test inside Docker (CI parity)
```

Integration tests use SQLite and miniredis — no external services required.

**Coverage threshold: 80%** on business logic. Infrastructure packages (`cmd`, `bootstrap`, `api`, `dto`, `model`, `repository/ping`) are excluded from the threshold but still scanned by SonarCloud.

## Make Targets

```
Development:
  make run             Run locally (auto-migrates on start)
  make dev             fmt → vet → test → swagger → run
  make fmt / vet / lint / tidy / vendor

Database:
  make migrate-up [STEPS=n]
  make migrate-down [STEPS=n]
  make migrate-version
  make migrate-force

Testing:
  make test
  make test-coverage

Build:
  make build / build-linux / build-macos / build-windows / build-prod / release

Mocks:
  make generate-mocks
  make clean-mocks

Docker / CI:
  make docker-test / docker-sonar / docker-build-push
  make docker-run / docker-stop / docker-logs / docker-shell / docker-clean

Utilities:
  make swagger / install-tools / info / clean / clean-docs / clean-all
```

## CI/CD

| Trigger | CI | CD |
|---|---|---|
| PR to `main` | test + SonarCloud (no push) | — |
| Push to `main` | test + SonarCloud + build + push `main`/`<sha7>` tags | deploy via self-hosted runner |
| Git tag `v*.*.*` | test + SonarCloud + build + push `<tag>` + `latest` | deploy via self-hosted runner |

CD runner working directory: `/opt/bookmark-system`. Updates `BOOKMARK_SERVICE_TAG` in `.env` and runs `docker compose up -d --force-recreate bookmark-service`.

## Project Structure

```
bookmark-service/
├── cmd/
│   ├── api/main.go           # HTTP server entrypoint
│   └── migrate/main.go       # Migration CLI (up/down/version/force)
├── internal/
│   ├── api/                  # router.go, swagger.go
│   ├── bootstrap/            # app.go, container.go (DI), config.go, routes.go
│   ├── dto/                  # bookmark/, link/, health/ request+response structs
│   ├── handler/              # bookmark/, link/, health/ HTTP handlers
│   ├── model/                # base.go (UUID PK), bookmark.go
│   ├── repository/
│   │   ├── bookmark/         # PostgreSQL CRUD + mocks
│   │   ├── cache/            # Redis cache store + mocks
│   │   ├── link/             # Redis short-link store + mocks
│   │   ├── ping/             # multi.go, sqldb.go, redis.go + mocks
│   │   └── queue/            # Redis LPUSH publisher + mocks
│   ├── service/
│   │   ├── bookmark/         # create, list, update, delete, import + mocks
│   │   ├── bookmark/cache/   # cache-aside decorator for bookmark service
│   │   ├── link/             # shorten, get (redirect), resolver/ + mocks
│   │   └── health/           # check + mocks
│   └── test/
│       ├── integration/      # end-to-end handler tests
│       └── fixtures/         # SQLite testdb helpers, CSV test data
├── migrations/               # 000001–000004 up/down SQL files
├── keys/                     # public.pem (git-ignored)
├── docs/                     # swagger generated output
├── Makefile
├── Dockerfile
├── sonar-project.properties
└── .github/workflows/ci.yaml, cd.yaml
```
