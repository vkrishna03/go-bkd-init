# Frontend Documentation

## Technology Stack

- **Framework:** React 18 with TypeScript
- **Build Tool:** Vite
- **Package Manager:** npm
- **UI Components:** shadcn/ui
- **Styling:** Tailwind CSS
- **State Management:** Zustand
- **HTTP Client:** Axios
- **Real-time Communication:**
  - WebRTC (peer-to-peer video/audio)
  - WebSocket (signaling + device updates)

---

## Project Structure

```
web/
├── src/
│   ├── components/
│   │   ├── ui/              # shadcn components
│   │   ├── auth/            # Login, Register forms
│   │   ├── device/          # Device list, Device card
│   │   └── stream/          # Stream controls, Video player
│   ├── hooks/
│   │   ├── useAuth.ts
│   │   ├── useWebSocket.ts
│   │   ├── useWebRTC.ts
│   │   └── useDevices.ts
│   ├── stores/
│   │   ├── authStore.ts
│   │   ├── deviceStore.ts
│   │   └── streamStore.ts
│   ├── services/
│   │   ├── api.ts           # Axios instance
│   │   ├── auth.ts          # Auth API calls
│   │   └── device.ts        # Device API calls
│   ├── pages/
│   │   ├── Login.tsx
│   │   ├── Register.tsx
│   │   ├── Dashboard.tsx
│   │   └── Stream.tsx
│   ├── lib/
│   │   ├── utils.ts         # shadcn utils
│   │   └── webrtc.ts        # WebRTC helpers
│   ├── types/
│   │   └── index.ts
│   ├── App.tsx
│   └── main.tsx
├── public/
├── index.html
├── tailwind.config.js
├── tsconfig.json
├── vite.config.ts
└── package.json
```

---

## UI/UX Components

### Pages

1. **Login Page** - Clean, minimal (email + password)
2. **Register Page** - Email, password, confirm password
3. **Dashboard** - Shows "My Devices" and "Available Streams"
4. **Stream Page** - Full-screen stream monitor with controls

### Components

- **DeviceList** - Shows all devices currently logged into account
- **DeviceCard** - Individual device with status indicator
- **StreamControls** - Play/pause, quality selector (high/medium/low)
- **VideoPlayer** - WebRTC video display with full-screen support
- **ConnectionIndicator** - Signal strength display
- **DeviceSelector** - Quick switch between streams

---

## State Management (Zustand)

### Auth Store
```typescript
interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<void>;
}
```

### Device Store
```typescript
interface DeviceState {
  devices: Device[];
  currentDevice: Device | null;
  isLoading: boolean;
  fetchDevices: () => Promise<void>;
  updateDevice: (id: string, name: string) => Promise<void>;
  removeDevice: (id: string) => Promise<void>;
}
```

### Stream Store
```typescript
interface StreamState {
  activeStreams: Stream[];
  currentStream: Stream | null;
  isStreaming: boolean;
  startStream: (type: 'video' | 'audio' | 'both') => Promise<void>;
  stopStream: () => void;
  connectToStream: (deviceId: string) => Promise<void>;
}
```

---

## WebRTC Integration

### Peer Connection Setup
```typescript
const config: RTCConfiguration = {
  iceServers: [
    { urls: 'stun:stun.l.google.com:19302' },
    {
      urls: 'turn:your-turn-server.com:3478',
      username: 'user',
      credential: 'pass'
    }
  ]
};
```

### Stream Types
- **Video only** - Camera stream without audio
- **Audio only** - Microphone stream without video
- **Video + Audio** - Full media stream

---

## WebSocket Events

### Client → Server
- `device:register` - Register device on login
- `device:heartbeat` - Keep-alive ping every 30 seconds
- `stream:start` - Initiate stream from device
- `stream:stop` - End stream
- `sdp:offer` - WebRTC SDP offer
- `sdp:answer` - WebRTC SDP answer
- `ice:candidate` - ICE candidate for P2P connection

### Server → Client
- `device:online` - Another device came online
- `device:offline` - Device went offline
- `stream:available` - Device is now streaming
- `stream:ended` - Stream ended
- `sdp:offer` - Forward SDP offer from source
- `sdp:answer` - Forward SDP answer from target
- `ice:candidate` - Forward ICE candidate

---

## Getting Started

```bash
# Navigate to web directory
cd web

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
```

---

## Environment Variables

```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
```
