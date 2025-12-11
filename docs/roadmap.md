# Streamz Roadmap & Progress

## MVP Status: Backend + WebSocket Complete

---

## MVP Feature Checklist

### Authentication
- [x] Email/password registration
- [ ] Email verification (optional for MVP)
- [x] Login with JWT
- [x] JWT refresh tokens
- [x] Logout
- [x] Session persistence
- [x] Password reset flow

### Device Management
- [x] Register device via API
- [x] Device list with real-time status (via WebSocket)
- [x] Device naming
- [x] Device removal
- [x] Online/offline status tracking
- [x] Heartbeat mechanism
- [x] Device capability tracking (camera/mic)

### Streaming
- [x] Stream session management (start/end)
- [x] Stream type selection (video/audio/both)
- [x] Stream quality selection (low/medium/high/auto)
- [ ] P2P WebRTC connection (Pion library)
- [ ] Relay fallback mechanism (TURN server)
- [ ] Real-time stream preview
- [x] Stop streaming
- [x] Connection quality tracking (latency, connection type)

### WebRTC Signaling
- [x] SDP offer/answer exchange
- [x] ICE candidate forwarding
- [x] Connection state management
- [x] Error handling for failed connections
- [x] Graceful disconnect handling

### Security
- [x] Password hashing with bcrypt
- [ ] HTTPS/WSS support
- [x] JWT token validation
- [x] CORS configuration
- [x] Input validation and sanitization
- [x] SQL injection prevention (parameterized queries via sqlc)

---

## Success Metrics (MVP)

### Technical
- P2P connection success rate > 85% (target >90%)
- Average latency < 500ms
- Relay fallback activates in <2 seconds
- 99.9% uptime on signaling server
- Memory usage < 50MB per 1000 concurrent devices
- Single Go binary deployment successful

### User Experience
- Setup time < 2 minutes (login + first stream)
- Device discovery < 3 seconds
- Stream startup time < 5 seconds
- Clear error messages for connection issues
- Responsive UI on both phone and desktop

---

## Known Challenges & Mitigation

| Challenge | Impact | Mitigation |
|-----------|--------|-----------|
| NAT traversal (P2P behind firewalls) | High | TURN relay server fallback with coturn |
| WebRTC browser compatibility | Medium | Feature detection + graceful degradation |
| Goroutine memory leaks | Medium | Proper context cancellation and cleanup |
| Connection state sync | High | Use atomic operations and channels for state |
| Database connection pool | Medium | pgx connection pool tuning (max conns = 25) |
| CORS and WebSocket issues | Medium | Proper headers, upgrade handling |

---

## Future Phases (Post-MVP)

### Phase 2: Cloud Recording
- [ ] Record stream to cloud (AWS S3)
- [ ] Auto-upload after stream ends
- [ ] Video list with duration/date
- [ ] Download/delete recordings

### Phase 3: Advanced Monitoring
- [ ] Multi-stream grid view (monitor 4 devices simultaneously)
- [ ] Picture-in-picture mode
- [ ] Stream switching without interruption
- [ ] Connection quality metrics (latency, jitter, packet loss)

### Phase 4: Creator Tools
- [ ] AI-powered auto-transcription (Whisper API)
- [ ] Auto-generated captions
- [ ] Analytics (stream duration, device types used, peak times)

### Phase 5: Monetization
- [ ] Free tier (3 devices, no recording)
- [ ] Pro tier ($5/month: unlimited devices, cloud recording)
- [ ] Stripe integration

---

## Success Definition

**MVP is successful when:**
- Phone can stream video to Mac with <500ms latency
- Users don't need technical knowledge to use it
- Setup takes <2 minutes (login + first stream)
- No installation required (browser-only)
- P2P works 80%+ of the time, relay fallback works rest
- Single Go binary deploys successfully
- Can handle 1000+ concurrent device connections

---

## Target Users

- Solo YouTubers recording on phone while monitoring framing on Mac
- Podcasters streaming while monitoring audio levels
- Musicians recording audio/video simultaneously
- Content creators (TikTok, Instagram Reels, YouTube Shorts)
- Freelance video editors doing remote client reviews
- Teachers/trainers monitoring student camera feeds
