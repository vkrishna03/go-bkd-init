package stream

import (
	"time"

	"github.com/google/uuid"
)

// Request DTOs

type StartRequest struct {
	SourceDeviceID uuid.UUID `json:"source_device_id" binding:"required"`
	TargetDeviceID uuid.UUID `json:"target_device_id"`
	StreamType     string    `json:"stream_type" binding:"required,oneof=video audio both"`
	Quality        string    `json:"quality" binding:"omitempty,oneof=low medium high auto"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=connecting active paused ended failed"`
}

type UpdateLatencyRequest struct {
	LatencyMs int `json:"latency_ms" binding:"required,min=0"`
}

type UpdateConnectionTypeRequest struct {
	ConnectionType string `json:"connection_type" binding:"required,oneof=p2p relay"`
}

// Response DTOs

type Response struct {
	ID             uuid.UUID      `json:"id"`
	UserID         uuid.UUID      `json:"user_id"`
	SourceDeviceID *uuid.UUID     `json:"source_device_id,omitempty"`
	TargetDeviceID *uuid.UUID     `json:"target_device_id,omitempty"`
	StreamType     string         `json:"stream_type"`
	Status         string         `json:"status"`
	ConnectionType *string        `json:"connection_type,omitempty"`
	Quality        string         `json:"quality"`
	LatencyMs      *int           `json:"latency_ms,omitempty"`
	StartedAt      time.Time      `json:"started_at"`
	EndedAt        *time.Time     `json:"ended_at,omitempty"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
