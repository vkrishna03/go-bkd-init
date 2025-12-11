# Streamz: Project Overview

**Project Type:** Portfolio Project / Indie Product
**Status:** Pre-development

---

## Executive Summary

Streamz is a **browser-based, account-driven device streaming platform** designed for solo content creators who need to record video on one device while monitoring on another—without complexity.

### The Problem
Solo creators face a gap:
- **Too simple:** VDO.Ninja requires URL sharing and has no persistent device discovery
- **Too complex:** OBS Studio requires software installation and advanced configuration
- **Hybrid solutions:** Google Meet can screenshare but can't stream camera feeds
- **Real-time lag:** Most solutions require WiFi/wired connections which limit mobility

### The Solution
Streamz allows users to:
1. Log in on multiple devices (phone, Mac, laptop)
2. Automatically discover logged-in devices in their account
3. Select which device streams and which monitors
4. Choose what to stream (video-only, audio-only, or both)
5. Monitor in real-time with minimal lag
6. (Future) Record to cloud storage

---

## Core Features (MVP)

### User Authentication
- **Sign Up:** Email/Password with email verification
- **Log In:** Standard email/password authentication
- **Session Management:** JWT-based tokens with refresh mechanism
- **Device Association:** Each logged-in device gets a unique device ID stored in browser

### Device Discovery & Management
- **Automatic Registration:** When user logs in, device is auto-registered with:
  - Device ID (unique browser fingerprint)
  - Device name (user-assigned or auto-generated)
  - Device type (mobile, desktop, tablet)
  - Last seen timestamp
  - Current status (online/offline)

- **Device List UI:** Shows all devices currently logged into account
  - Device name
  - Type (phone, Mac, etc.)
  - Status indicator (online/offline)
  - Last activity

### Stream Configuration
- **Stream Initiator (Source Device - e.g., Samsung Phone):**
  - "Start Streaming" button
  - Select stream type: Video only, Audio only, Video + Audio
  - Camera/microphone permission prompts
  - Real-time indicator showing stream is active
  - Stop streaming button

- **Stream Receiver (Monitor Device - e.g., MacBook):**
  - "Available Streams" shows all devices currently streaming
  - Click a device to start monitoring
  - Live preview of selected stream
  - Switch between streams without reloading
  - Stop monitoring button
  - Stream info (device name, duration, connection quality)

### WebRTC P2P Streaming
- **Peer-to-Peer Connection:** Direct video/audio stream between devices when possible
- **Fallback Relay Server:** When P2P fails, route through backend relay
- **Low Latency:** Target <500ms latency for monitoring
- **Connection Quality Indicator:** Show signal strength to user

### UI/UX Components
- **Login Page:** Clean, minimal (email + password)
- **Dashboard:** Shows "My Devices" and "Available Streams"
- **Stream Controls:** Play/pause, quality selector (high/medium/low bandwidth)
- **Full-Screen Monitor:** Expand stream to full screen for better monitoring
- **Device Selector:** Quick switch between streams
- **Settings Panel:** Device name, log out, change password

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Frontend (Web App)                          │
├─────────────────────────────────────────────────────────────────┤
│  React App (Mobile Browser, Desktop Browser)                    │
│  ├─ Login/Auth UI                                               │
│  ├─ Device Discovery                                            │
│  ├─ Stream Selector & Controls                                  │
│  └─ WebRTC Client                                               │
└───────────┬──────────────────────────────────────┬──────────────┘
            │                                      │
            │ HTTP (Auth, Device Mgmt)             │ WebSocket
            │                                      │
┌───────────▼──────────────────────────────────────▼──────────────┐
│                    Backend (Go + Gin)                           │
├─────────────────────────────────────────────────────────────────┤
│  ├─ Auth Service (JWT, Registration, Login)                     │
│  ├─ Device Registry Service (Device CRUD)                       │
│  ├─ WebSocket Hub (Device discovery, broadcasts)                │
│  ├─ WebRTC Signaling Service (SDP/ICE exchange)                 │
│  ├─ Stream Session Manager (Connection state)                   │
│  └─ Real-time Events (Goroutine-based broadcasting)             │
└───────────┬──────────────────────────────────────┬──────────────┘
            │                                      │
            │ SQL (pgx)                            │ TCP
            │                                      │
┌───────────▼──────────────┐        ┌─────────────▼──────────────┐
│    PostgreSQL Database   │        │  TURN Relay Server         │
├──────────────────────────┤        ├────────────────────────────┤
│ ├─ users                 │        │ (coturn or equivalent)     │
│ ├─ devices               │        │ Fallback P2P routing       │
│ ├─ sessions              │        │                            │
│ └─ stream_logs           │        │                            │
└──────────────────────────┘        └────────────────────────────┘
```

---

## Documentation Index

| Document | Description |
|----------|-------------|
| [Frontend](frontend.md) | React app, shadcn/ui, state management, WebRTC client |
| [Backend](backend.md) | Go API, database schema, WebSocket hub, signaling |
| [Deployment](deployment.md) | Docker, VPS, cloud platforms, CI/CD |
| [Roadmap](roadmap.md) | MVP checklist, success metrics, future phases |
