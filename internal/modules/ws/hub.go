package ws

import (
	"context"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients grouped by user ID
	// map[userID]map[deviceID]*Client
	clients map[uuid.UUID]map[uuid.UUID]*Client

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for clients map
	mu sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's event loop
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case <-ctx.Done():
			slog.Info("hub shutting down")
			return
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Create user's device map if not exists
	if h.clients[client.userID] == nil {
		h.clients[client.userID] = make(map[uuid.UUID]*Client)
	}

	// Check if device already connected (close old connection)
	if existing, ok := h.clients[client.userID][client.deviceID]; ok {
		close(existing.send)
	}

	h.clients[client.userID][client.deviceID] = client

	slog.Info("client registered",
		"user_id", client.userID,
		"device_id", client.deviceID,
		"total_user_devices", len(h.clients[client.userID]),
	)

	// Send device list to newly connected client
	h.sendDeviceList(client)

	// Broadcast device online to other user's devices
	h.broadcastToUser(client.userID, client.deviceID, TypeDeviceOnline, DeviceOnlinePayload{
		Device: *client.device,
	})
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if userClients, ok := h.clients[client.userID]; ok {
		if _, ok := userClients[client.deviceID]; ok {
			delete(userClients, client.deviceID)
			close(client.send)

			slog.Info("client unregistered",
				"user_id", client.userID,
				"device_id", client.deviceID,
			)

			// Clean up empty user map
			if len(userClients) == 0 {
				delete(h.clients, client.userID)
			} else {
				// Broadcast device offline to other user's devices
				h.broadcastToUserLocked(client.userID, client.deviceID, TypeDeviceOffline, DeviceOfflinePayload{
					DeviceID: client.deviceID,
				})
			}
		}
	}
}

// sendDeviceList sends the list of online devices to a client
func (h *Hub) sendDeviceList(client *Client) {
	devices := make([]DeviceInfo, 0)

	if userClients, ok := h.clients[client.userID]; ok {
		for _, c := range userClients {
			if c.device != nil {
				devices = append(devices, *c.device)
			}
		}
	}

	msg, err := NewMessage(TypeDeviceList, DeviceListPayload{Devices: devices})
	if err != nil {
		slog.Error("failed to create device list message", "error", err)
		return
	}

	client.Send(msg)
}

// broadcastToUser sends a message to all of a user's devices except the sender
func (h *Hub) broadcastToUser(userID, excludeDeviceID uuid.UUID, msgType string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	h.broadcastToUserLocked(userID, excludeDeviceID, msgType, payload)
}

// broadcastToUserLocked broadcasts without acquiring lock (caller must hold lock)
func (h *Hub) broadcastToUserLocked(userID, excludeDeviceID uuid.UUID, msgType string, payload interface{}) {
	userClients, ok := h.clients[userID]
	if !ok {
		return
	}

	msg, err := NewMessage(msgType, payload)
	if err != nil {
		slog.Error("failed to create broadcast message", "error", err)
		return
	}

	for deviceID, client := range userClients {
		if deviceID != excludeDeviceID {
			client.Send(msg)
		}
	}
}

// ForwardToDevice forwards a message to a specific device
func (h *Hub) ForwardToDevice(userID, targetDeviceID uuid.UUID, msgType string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userClients, ok := h.clients[userID]
	if !ok {
		slog.Warn("user not found for forwarding", "user_id", userID)
		return
	}

	client, ok := userClients[targetDeviceID]
	if !ok {
		slog.Warn("target device not found", "user_id", userID, "device_id", targetDeviceID)
		return
	}

	msg, err := NewMessage(msgType, payload)
	if err != nil {
		slog.Error("failed to create forward message", "error", err)
		return
	}

	client.Send(msg)
}

// BroadcastStreamStart notifies all user's devices about a new stream
func (h *Hub) BroadcastStreamStart(userID uuid.UUID, streamID, sourceDeviceID uuid.UUID, streamType, quality string) {
	h.broadcastToUser(userID, uuid.Nil, TypeStreamStart, StreamStartPayload{
		StreamID:       streamID,
		SourceDeviceID: sourceDeviceID,
		StreamType:     streamType,
		Quality:        quality,
	})
}

// BroadcastStreamEnd notifies all user's devices about a stream ending
func (h *Hub) BroadcastStreamEnd(userID uuid.UUID, streamID uuid.UUID) {
	h.broadcastToUser(userID, uuid.Nil, TypeStreamEnd, StreamEndPayload{
		StreamID: streamID,
	})
}

// GetOnlineDevices returns the list of online devices for a user
func (h *Hub) GetOnlineDevices(userID uuid.UUID) []DeviceInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	devices := make([]DeviceInfo, 0)
	if userClients, ok := h.clients[userID]; ok {
		for _, c := range userClients {
			if c.device != nil {
				devices = append(devices, *c.device)
			}
		}
	}
	return devices
}

// IsDeviceOnline checks if a specific device is online
func (h *Hub) IsDeviceOnline(userID, deviceID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if userClients, ok := h.clients[userID]; ok {
		_, online := userClients[deviceID]
		return online
	}
	return false
}
