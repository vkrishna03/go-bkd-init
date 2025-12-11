# Streamz

Browser-based device streaming platform for solo content creators. Stream from phone, monitor on Mac.

## Tech Stack

| Layer | Stack |
|-------|-------|
| Frontend | React 18, TypeScript, Vite, **shadcn/ui**, Tailwind, Zustand, **npm** |
| Backend | Go 1.21+, Gin, PostgreSQL (pgx), Gorilla WebSocket, Pion WebRTC |
| Infra | Docker, Coturn (TURN), golang-migrate, sqlc |

## Project Structure

```
streamz/
├── cmd/app/main.go
├── internal/
│   ├── modules/          # auth, device, stream, ws
│   ├── middleware/
│   ├── database/
│   ├── config/
│   └── server/
├── db/migrations/
├── docs/                 # Detailed documentation
│   ├── overview.md       # Full project overview
│   ├── frontend.md       # Frontend architecture
│   ├── backend.md        # Backend architecture
│   ├── deployment.md     # Deploy options
│   └── roadmap.md        # MVP checklist & progress
└── web/                  # Frontend (React)
```

## API

```
POST /api/auth/register|login|refresh
GET|POST /api/devices
PUT|DELETE /api/devices/:id
WS /ws
GET /health, /ping
```

## Key Files

- `internal/server/server.go` - HTTP server setup, routes
- `internal/database/database.go` - DB connection
- `internal/modules/*/handler.go` - Route handlers
- `internal/modules/ws/hub.go` - WebSocket hub (device discovery, signaling)

## Commands

```bash
make dev          # Run with hot reload
make build        # Build binary
make migrate-up   # Run migrations
make sqlc         # Generate sqlc
```

## Docs Reference

- **[docs/overview.md](docs/overview.md)** - Full project description, features, architecture
- **[docs/frontend.md](docs/frontend.md)** - React components, state, WebRTC integration
- **[docs/backend.md](docs/backend.md)** - Go modules, DB schema, WebSocket events
- **[docs/deployment.md](docs/deployment.md)** - Docker, VPS, AWS, Railway, CI/CD
- **[docs/roadmap.md](docs/roadmap.md)** - MVP checklist, success metrics, future phases
