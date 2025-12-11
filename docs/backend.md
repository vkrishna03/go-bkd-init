# Backend Documentation

## Technology Stack

- **Runtime:** Go 1.21+
- **Framework:** Gin Gonic (HTTP router)
- **Authentication:** JWT (github.com/golang-jwt/jwt) + bcrypt (golang.org/x/crypto/bcrypt)
- **Database:** PostgreSQL 14+ with pgx driver
- **Real-time Communication:** Gorilla WebSocket (github.com/gorilla/websocket)
- **WebRTC Signaling:** Pion WebRTC library (github.com/pion/webrtc)
- **TURN Server:** Coturn (for P2P fallback)
- **Utilities:**
  - UUID generation (github.com/google/uuid)
  - Environment variables (github.com/joho/godotenv)
  - Logging (slog)
  - Database migrations (golang-migrate)

---

## Why Go?

### Concurrency Model
- Goroutines handle WebSocket connections efficiently (1000s of concurrent connections with minimal memory)
- Perfect for WebRTC signaling server (broadcast device updates to many clients)
- Lightweight compared to Node.js event loop

### Performance
- Compiled binary = faster execution
- Garbage collection is predictable (important for real-time streaming)
- Memory footprint: ~5-20MB vs Node.js ~100MB+

### Deployment
- Single binary (no runtime dependencies)
- Docker image size: ~30MB vs Node ~300MB+
- Startup time: ~10-50ms vs Node ~500ms

### Scalability
- Thread-safe with sync.RWMutex for device registry
- Channel-based communication for broadcasting
- Built-in support for graceful shutdowns

---

## Project Structure

```
streamz/
├── cmd/app/main.go             # Entry point
├── internal/
│   ├── modules/                # Domain modules
│   │   ├── auth/               # Authentication (JWT, register, login)
│   │   │   ├── dto.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── handler.go
│   │   ├── device/             # Device management
│   │   │   ├── dto.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── handler.go
│   │   ├── stream/             # Stream session management
│   │   │   ├── dto.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── handler.go
│   │   └── ws/                 # WebSocket hub & signaling
│   │       ├── hub.go
│   │       ├── client.go
│   │       ├── messages.go
│   │       └── handler.go
│   ├── middleware/             # Auth, CORS, request ID
│   ├── errors/                 # Error types + response helper
│   ├── database/               # DB connection
│   ├── config/                 # Config loading
│   └── server/                 # HTTP server, graceful shutdown
├── db/
│   ├── migrations/             # SQL migrations
│   ├── queries/                # sqlc query files
│   └── sqlc/                   # Generated (DO NOT EDIT)
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── sqlc.yaml
└── go.mod
```

---

## Database Schema

### Users Table
```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

### Devices Table
```sql
CREATE TABLE devices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  device_id VARCHAR(255) NOT NULL,
  device_name VARCHAR(255) NOT NULL,
  device_type VARCHAR(50),
  is_online BOOLEAN DEFAULT FALSE,
  last_seen TIMESTAMP DEFAULT NOW(),
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(user_id, device_id)
);

CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_is_online ON devices(is_online);
```

### Streams Table
```sql
CREATE TABLE streams (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  source_device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
  target_device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
  stream_type VARCHAR(50),
  started_at TIMESTAMP DEFAULT NOW(),
  ended_at TIMESTAMP,
  connection_type VARCHAR(50),
  quality VARCHAR(50),
  latency_ms INT
);

CREATE INDEX idx_streams_user_id ON streams(user_id);
CREATE INDEX idx_streams_created_at ON streams(started_at);
```

---

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user (returns JWT)
- `POST /api/auth/refresh` - Refresh JWT token
- `POST /api/auth/logout` - Logout user

### Devices
- `GET /api/devices` - List all user devices
- `POST /api/devices` - Register new device
- `PUT /api/devices/:deviceId` - Update device name
- `DELETE /api/devices/:deviceId` - Remove device

### Health
- `GET /health` - Server health check
- `GET /ping` - Simple ping endpoint

### WebSocket
- `WS /ws` - WebSocket connection for real-time events

---

## WebSocket Message Types

```go
// WebSocket message wrapper
type WSMessage struct {
  Type    string          `json:"type"`
  Payload json.RawMessage `json:"payload"`
}

// Device discovery events
type DeviceOnlineEvent struct {
  DeviceID   string `json:"device_id"`
  DeviceName string `json:"device_name"`
  DeviceType string `json:"device_type"`
}

type DeviceOfflineEvent struct {
  DeviceID string `json:"device_id"`
}

// WebRTC signaling
type StreamStartEvent struct {
  SourceDeviceID string `json:"source_device_id"`
  StreamType     string `json:"stream_type"` // video, audio, both
}

type SDPOfferEvent struct {
  FromDeviceID string `json:"from_device_id"`
  SDP          string `json:"sdp"`
}

type SDPAnswerEvent struct {
  FromDeviceID string `json:"from_device_id"`
  SDP          string `json:"sdp"`
}

type ICECandidateEvent struct {
  FromDeviceID  string `json:"from_device_id"`
  Candidate     string `json:"candidate"`
  SDPMLineIndex uint16 `json:"sdp_mline_index"`
  SDPMid        string `json:"sdp_mid"`
}
```

---

## WebSocket Hub Architecture

```go
// Hub manages all connected clients for a user
type Hub struct {
  clients    map[uuid.UUID]*Client  // device_id -> Client
  register   chan *Client
  unregister chan *Client
  broadcast  chan interface{}
  mu         sync.RWMutex
}

// Client represents a connected device
type Client struct {
  userID   uuid.UUID
  deviceID string
  hub      *Hub
  send     chan interface{} // Message queue
  conn     *websocket.Conn
}

// Run manages goroutines for each connected device
func (h *Hub) Run(ctx context.Context) {
  for {
    select {
    case client := <-h.register:
      h.mu.Lock()
      h.clients[client.userID] = client
      h.mu.Unlock()
      h.broadcast <- DeviceOnlineEvent{...}

    case client := <-h.unregister:
      h.mu.Lock()
      delete(h.clients, client.userID)
      h.mu.Unlock()
      close(client.send)
      h.broadcast <- DeviceOfflineEvent{...}

    case message := <-h.broadcast:
      h.mu.RLock()
      for _, client := range h.clients {
        client.send <- message
      }
      h.mu.RUnlock()

    case <-ctx.Done():
      return
    }
  }
}
```

---

## Environment Variables

```env
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/streamz?sslmode=disable

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# CORS
ALLOWED_ORIGINS=http://localhost:5173
```

---

## Getting Started

```bash
# Run with hot reload
make dev

# Run migrations
make migrate-up

# Generate sqlc
make sqlc

# Build binary
make build

# Run tests
make test
```

---

## Known Challenges & Mitigation

| Challenge | Impact | Mitigation |
|-----------|--------|-----------|
| NAT traversal (P2P behind firewalls) | High | TURN relay server fallback with coturn |
| Goroutine memory leaks | Medium | Proper context cancellation and cleanup |
| Connection state sync | High | Use atomic operations and channels for state |
| Database connection pool | Medium | pgx connection pool tuning (max conns = 25) |
| CORS and WebSocket issues | Medium | Proper headers, upgrade handling |
