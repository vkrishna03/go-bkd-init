package ws

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Message types
const (
	// Device events
	TypeDeviceOnline  = "device:online"
	TypeDeviceOffline = "device:offline"
	TypeDeviceList    = "device:list"

	// Stream events
	TypeStreamStart = "stream:start"
	TypeStreamEnd   = "stream:end"

	// WebRTC signaling
	TypeOffer     = "webrtc:offer"
	TypeAnswer    = "webrtc:answer"
	TypeCandidate = "webrtc:candidate"

	// Control
	TypeError = "error"
	TypePing  = "ping"
	TypePong  = "pong"
)

// Message is the base WebSocket message envelope
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// DeviceInfo represents device information in events
type DeviceInfo struct {
	ID            uuid.UUID `json:"id"`
	DeviceID      string    `json:"device_id"`
	DeviceName    string    `json:"device_name"`
	DeviceType    string    `json:"device_type"`
	HasCamera     bool      `json:"has_camera"`
	HasMicrophone bool      `json:"has_microphone"`
	IsOnline      bool      `json:"is_online"`
}

// DeviceOnlinePayload is sent when a device comes online
type DeviceOnlinePayload struct {
	Device DeviceInfo `json:"device"`
}

// DeviceOfflinePayload is sent when a device goes offline
type DeviceOfflinePayload struct {
	DeviceID uuid.UUID `json:"device_id"`
}

// DeviceListPayload is sent when client first connects
type DeviceListPayload struct {
	Devices []DeviceInfo `json:"devices"`
}

// StreamStartPayload is sent when a device starts streaming
type StreamStartPayload struct {
	StreamID       uuid.UUID `json:"stream_id"`
	SourceDeviceID uuid.UUID `json:"source_device_id"`
	StreamType     string    `json:"stream_type"`
	Quality        string    `json:"quality"`
}

// StreamEndPayload is sent when a stream ends
type StreamEndPayload struct {
	StreamID uuid.UUID `json:"stream_id"`
}

// WebRTC signaling payloads

type OfferPayload struct {
	FromDeviceID uuid.UUID `json:"from_device_id"`
	ToDeviceID   uuid.UUID `json:"to_device_id"`
	SDP          string    `json:"sdp"`
}

type AnswerPayload struct {
	FromDeviceID uuid.UUID `json:"from_device_id"`
	ToDeviceID   uuid.UUID `json:"to_device_id"`
	SDP          string    `json:"sdp"`
}

type CandidatePayload struct {
	FromDeviceID  uuid.UUID `json:"from_device_id"`
	ToDeviceID    uuid.UUID `json:"to_device_id"`
	Candidate     string    `json:"candidate"`
	SDPMLineIndex *uint16   `json:"sdp_mline_index,omitempty"`
	SDPMid        *string   `json:"sdp_mid,omitempty"`
}

// ErrorPayload is sent when an error occurs
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Helper functions to create messages

func NewMessage(msgType string, payload interface{}) ([]byte, error) {
	var payloadBytes json.RawMessage
	if payload != nil {
		var err error
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	msg := Message{
		Type:    msgType,
		Payload: payloadBytes,
	}

	return json.Marshal(msg)
}

func NewErrorMessage(code, message string) ([]byte, error) {
	return NewMessage(TypeError, ErrorPayload{
		Code:    code,
		Message: message,
	})
}

func NewPongMessage() ([]byte, error) {
	return NewMessage(TypePong, nil)
}
