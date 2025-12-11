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
│   ├── modules/          # auth, device, stream, webrtc, ws
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
# Auth (public)
POST /api/v1/auth/register|login|refresh|logout
POST /api/v1/auth/forgot-password|reset-password

# Devices (protected)
GET|POST /api/v1/devices
GET|PUT|DELETE /api/v1/devices/:id
PUT /api/v1/devices/:id/status
POST /api/v1/devices/:id/heartbeat

# Streams (protected)
GET|POST /api/v1/streams
GET /api/v1/streams/active
GET|DELETE /api/v1/streams/:id
PUT /api/v1/streams/:id/status|latency|connection-type

# WebRTC (protected)
GET /api/v1/webrtc/ice-servers

# WebSocket (protected)
WS /ws?device_id=<uuid>

# Health
GET /health, /ping
```

## WebSocket Events

```
# Device events
device:online   - Device came online
device:offline  - Device went offline
device:list     - List of online devices (sent on connect)

# Stream events
stream:start    - Stream started
stream:end      - Stream ended

# WebRTC signaling
webrtc:offer    - SDP offer
webrtc:answer   - SDP answer
webrtc:candidate - ICE candidate

# Control
ping/pong       - Heartbeat
error           - Error message
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
