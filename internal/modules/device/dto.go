package device

import (
	"time"

	"github.com/google/uuid"
)

// Request DTOs

type RegisterRequest struct {
	DeviceID      string `json:"device_id" binding:"required"`
	DeviceName    string `json:"device_name" binding:"required"`
	DeviceType    string `json:"device_type" binding:"required,oneof=phone tablet desktop"`
	HasCamera     bool   `json:"has_camera"`
	HasMicrophone bool   `json:"has_microphone"`
}

type UpdateRequest struct {
	DeviceName    *string `json:"device_name"`
	HasCamera     *bool   `json:"has_camera"`
	HasMicrophone *bool   `json:"has_microphone"`
}

type UpdateStatusRequest struct {
	IsOnline bool `json:"is_online"`
}

// Response DTOs

type Response struct {
	ID            uuid.UUID `json:"id"`
	DeviceID      string    `json:"device_id"`
	DeviceName    string    `json:"device_name"`
	DeviceType    string    `json:"device_type"`
	HasCamera     bool      `json:"has_camera"`
	HasMicrophone bool      `json:"has_microphone"`
	IsOnline      bool      `json:"is_online"`
	LastSeen      time.Time `json:"last_seen"`
	CreatedAt     time.Time `json:"created_at"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
