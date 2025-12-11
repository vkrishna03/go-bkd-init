# go-bkd

Production-ready Go backend starter.

## Quick Start

**Use as starter template:**
```bash
git clone https://github.com/vkrishna03/go-bkd-init myproject
cd myproject
rm -rf .git && git init
make init name=github.com/yourusername/myproject
cp .env.example .env
make dev
```

**Or just run:**
```bash
cp .env.example .env
make docker-up      # or: make dev (DB in docker, app locally)
```

## Tech Stack

Go 1.24 | Gin | PostgreSQL | sqlc | pgx | godotenv | slog

---

## Developer Guide

### Project Structure

```
go-bkd/
├── cmd/app/main.go             # Entry point
├── internal/
│   ├── modules/                # Domain modules
│   │   ├── user/               # Example module
│   │   │   ├── dto.go          # Request/response types
│   │   │   ├── repository.go   # Database operations
│   │   │   ├── service.go      # Business logic
│   │   │   └── handler.go      # HTTP handlers + Setup()
│   │   └── z-to-copy/          # Template module
│   ├── middleware/             # Logging, CORS, request ID
│   ├── errors/                 # Error types + response helper
│   ├── database/               # DB connection
│   ├── config/                 # Config loading
│   └── server/                 # HTTP server, graceful shutdown
├── db/
│   ├── migrations/             # SQL migrations
│   ├── queries/                # sqlc query files
│   └── sqlc/                   # Generated (DO NOT EDIT)
└── test/
```

### Module Pattern

Each module has 4 files:

**dto.go** - Request/response types
```go
type CreateRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

type Response struct {
    ID    int32  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

**repository.go** - Database operations
```go
type Repository struct {
    db *sql.DB
    q  *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db, q: sqlc.New(db)}
}

func (r *Repository) GetByID(ctx context.Context, id int32) (sqlc.User, error) {
    return r.q.GetUser(ctx, id)
}
```

**service.go** - Business logic
```go
type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id int32) (*Response, error) {
    u, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, apperr.Wrap(apperr.ErrNotFound, "user with id %d not found", id)
        }
        return nil, apperr.Wrap(apperr.ErrInternal, "failed to retrieve user")
    }
    return toResponse(u), nil
}
```

**handler.go** - HTTP handlers + Setup()
```go
type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}

func (h *Handler) Get(c *gin.Context) {
    // parse → service → respond
}

// Setup at bottom - wires dependencies + registers routes
func Setup(api *gin.RouterGroup, db *sql.DB) {
    repo := NewRepository(db)
    svc := NewService(repo)
    h := NewHandler(svc)

    r := api.Group("/users")
    r.GET("", h.List)
    r.GET("/:id", h.Get)
    r.POST("", h.Create)
    r.DELETE("/:id", h.Delete)
}
```

### Error Handling

```go
import apperr "github.com/vkrishna03/go-bkd-init/internal/errors"

// Wrap with full message
apperr.Wrap(apperr.ErrNotFound, "user with id %d not found", id)
apperr.Wrap(apperr.ErrValidation, "email is required")
apperr.Wrap(apperr.ErrConflict, "email %s already taken", email)

// In handler
if err != nil {
    apperr.Response(c, err)
    return
}
```

| Sentinel | HTTP | Code |
|----------|------|------|
| `ErrNotFound` | 404 | NOT_FOUND |
| `ErrValidation` | 400 | VALIDATION_ERROR |
| `ErrUnauthorized` | 401 | UNAUTHORIZED |
| `ErrForbidden` | 403 | FORBIDDEN |
| `ErrConflict` | 409 | CONFLICT |
| `ErrInternal` | 500 | INTERNAL_ERROR |

**Response format:**
```json
{"code": "NOT_FOUND", "message": "user with id 123 not found", "request_id": "abc-123"}
```

### Conventions

- **Files**: `snake_case.go`
- **Packages**: singular (`user`, not `users`)
- **Routes**: `/api/v1/users`, `/api/v1/users/:id`
- **Logging**: `slog.Info("msg", "key", value)`

### Adding a New Module

```bash
cp -r internal/modules/z-to-copy internal/modules/<name>
# Edit files, rename package
# Add db/queries/<name>.sql
make sqlc
# Add to main.go: <name>.Setup(api, db)
```

---

## Commands

```bash
# Development
make run          # Run app (expects DB running - local or docker)
make dev          # Start docker DB + run app
make db           # Start only docker DB

# Build
make build        # Dev build
make build-prod   # Production build (optimized)
make build-linux  # Cross-compile for Linux
make docker-build # Docker image (app only)

# Test
make test         # Run tests
make test-cover   # Run tests with coverage

# Docker
make docker-up    # Full docker (app + db)
make docker-down  # Stop all

# Database
make sqlc         # Regenerate sqlc
make migrate-up   # Run migrations
make migrate-down # Rollback one migration
make migrate-create # Create new migration

# Project setup
make init name=github.com/user/project  # Rename module + imports
make tidy         # go mod tidy
```

## Deployment

**Binary:**
```bash
make build-prod   # or: make build-linux (from Mac/Windows)
scp bin/app server:/path/
# On server:
DB_HOST=your-db.com DB_PASSWORD=xxx ./app
```

**Docker:**
```bash
docker build -t go-bkd .
docker run -e DB_HOST=your-db.com -e DB_PASSWORD=xxx go-bkd
```

## API

```bash
curl http://localhost:3000/health
curl http://localhost:3000/api/v1/users
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'
```
