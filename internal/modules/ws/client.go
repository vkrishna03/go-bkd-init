package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a connected WebSocket client
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   uuid.UUID
	deviceID uuid.UUID
	device   *DeviceInfo
	mu       sync.RWMutex
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID, deviceID uuid.UUID, device *DeviceInfo) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		deviceID: deviceID,
		device:   device,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("websocket read error", "error", err, "user_id", c.userID, "device_id", c.deviceID)
			}
			break
		}

		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				slog.Error("websocket write error", "error", err, "user_id", c.userID)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		slog.Error("failed to unmarshal message", "error", err)
		c.sendError("INVALID_MESSAGE", "failed to parse message")
		return
	}

	switch msg.Type {
	case TypePing:
		c.handlePing()

	case TypeOffer:
		c.handleOffer(msg.Payload)

	case TypeAnswer:
		c.handleAnswer(msg.Payload)

	case TypeCandidate:
		c.handleCandidate(msg.Payload)

	default:
		slog.Warn("unknown message type", "type", msg.Type)
		c.sendError("UNKNOWN_TYPE", "unknown message type: "+msg.Type)
	}
}

func (c *Client) handlePing() {
	msg, _ := NewPongMessage()
	c.send <- msg
}

func (c *Client) handleOffer(payload json.RawMessage) {
	var offer OfferPayload
	if err := json.Unmarshal(payload, &offer); err != nil {
		c.sendError("INVALID_PAYLOAD", "invalid offer payload")
		return
	}

	// Set from device to sender's device
	offer.FromDeviceID = c.deviceID

	// Forward to target device
	c.hub.ForwardToDevice(c.userID, offer.ToDeviceID, TypeOffer, offer)
}

func (c *Client) handleAnswer(payload json.RawMessage) {
	var answer AnswerPayload
	if err := json.Unmarshal(payload, &answer); err != nil {
		c.sendError("INVALID_PAYLOAD", "invalid answer payload")
		return
	}

	answer.FromDeviceID = c.deviceID

	// Forward to target device
	c.hub.ForwardToDevice(c.userID, answer.ToDeviceID, TypeAnswer, answer)
}

func (c *Client) handleCandidate(payload json.RawMessage) {
	var candidate CandidatePayload
	if err := json.Unmarshal(payload, &candidate); err != nil {
		c.sendError("INVALID_PAYLOAD", "invalid candidate payload")
		return
	}

	candidate.FromDeviceID = c.deviceID

	// Forward to target device
	c.hub.ForwardToDevice(c.userID, candidate.ToDeviceID, TypeCandidate, candidate)
}

func (c *Client) sendError(code, message string) {
	msg, _ := NewErrorMessage(code, message)
	c.send <- msg
}

// Send sends a message to the client
func (c *Client) Send(data []byte) {
	select {
	case c.send <- data:
	default:
		// Channel full, client is slow
		slog.Warn("client send buffer full", "user_id", c.userID, "device_id", c.deviceID)
	}
}
