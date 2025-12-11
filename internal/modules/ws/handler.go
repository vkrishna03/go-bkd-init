package ws

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vkrishna03/streamz/db/sqlc"
	"github.com/vkrishna03/streamz/internal/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking in production
		return true
	},
}

// Handler handles WebSocket connections
type Handler struct {
	hub *Hub
	db  *sql.DB
	q   *sqlc.Queries
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, db *sql.DB) *Handler {
	return &Handler{
		hub: hub,
		db:  db,
		q:   sqlc.New(db),
	}
}

// HandleWebSocket handles the WebSocket upgrade and connection
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Get user ID from auth middleware
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get device ID from query param
	deviceIDStr := c.Query("device_id")
	if deviceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id query parameter required"})
		return
	}

	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device_id"})
		return
	}

	// Verify device belongs to user
	device, err := h.q.GetDeviceByID(c.Request.Context(), deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
			return
		}
		slog.Error("failed to get device", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if device.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "device not owned by user"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	// Create device info
	deviceInfo := &DeviceInfo{
		ID:         device.ID,
		DeviceID:   device.DeviceID,
		DeviceName: device.DeviceName,
		DeviceType: string(device.DeviceType),
		IsOnline:   true,
	}
	if device.HasCamera.Valid {
		deviceInfo.HasCamera = device.HasCamera.Bool
	}
	if device.HasMicrophone.Valid {
		deviceInfo.HasMicrophone = device.HasMicrophone.Bool
	}

	// Create client
	client := NewClient(h.hub, conn, userID, deviceID, deviceInfo)

	// Register client
	h.hub.register <- client

	// Update device online status in DB
	_ = h.q.UpdateDeviceOnlineStatus(c.Request.Context(), sqlc.UpdateDeviceOnlineStatusParams{
		ID:       deviceID,
		IsOnline: sql.NullBool{Bool: true, Valid: true},
	})

	// Start pumps
	go client.WritePump()
	go func() {
		client.ReadPump()
		// When ReadPump exits, mark device offline
		_ = h.q.UpdateDeviceOnlineStatus(context.Background(), sqlc.UpdateDeviceOnlineStatusParams{
			ID:       deviceID,
			IsOnline: sql.NullBool{Bool: false, Valid: true},
		})
	}()
}

// Setup registers WebSocket routes and returns the hub
func Setup(router *gin.Engine, db *sql.DB, jwtSecret string) *Hub {
	hub := NewHub()
	handler := NewHandler(hub, db)

	// WebSocket endpoint (requires auth via query param token or header)
	ws := router.Group("/ws")
	ws.Use(middleware.Auth(jwtSecret))
	ws.GET("", handler.HandleWebSocket)

	return hub
}
