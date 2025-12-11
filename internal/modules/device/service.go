package device

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/vkrishna03/streamz/db/sqlc"
	apperr "github.com/vkrishna03/streamz/internal/errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Register creates a new device for the user
func (s *Service) Register(ctx context.Context, userID uuid.UUID, req RegisterRequest) (*Response, error) {
	// Check if device already exists
	existing, err := s.repo.GetByUserAndDeviceID(ctx, userID, req.DeviceID)
	if err == nil {
		// Device exists, update it and return
		resp := toResponse(existing)
		return &resp, nil
	}
	if err != sql.ErrNoRows {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to check existing device")
	}

	// Check device limit
	settings, err := s.repo.GetUserSettings(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to get user settings")
	}

	count, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to count devices")
	}

	if count >= int64(settings.MaxDevices.Int32) {
		return nil, apperr.Wrap(apperr.ErrValidation, "device limit reached (max %d)", settings.MaxDevices.Int32)
	}

	// Parse device type
	deviceType, err := parseDeviceType(req.DeviceType)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrValidation, "invalid device type")
	}

	// Create device
	device, err := s.repo.Create(ctx, userID, req.DeviceID, req.DeviceName, deviceType, req.HasCamera, req.HasMicrophone)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to create device")
	}

	resp := toResponse(device)
	return &resp, nil
}

// List returns all devices for the user
func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Response, error) {
	devices, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to list devices")
	}

	return toResponseList(devices), nil
}

// ListOnline returns all online devices for the user
func (s *Service) ListOnline(ctx context.Context, userID uuid.UUID) ([]Response, error) {
	devices, err := s.repo.ListOnlineByUser(ctx, userID)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to list online devices")
	}

	return toResponseList(devices), nil
}

// GetByID returns a device by ID
func (s *Service) GetByID(ctx context.Context, userID, deviceID uuid.UUID) (*Response, error) {
	device, err := s.repo.GetByID(ctx, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.Wrap(apperr.ErrNotFound, "device not found")
		}
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to get device")
	}

	// Verify ownership
	if device.UserID != userID {
		return nil, apperr.Wrap(apperr.ErrForbidden, "device not owned by user")
	}

	resp := toResponse(device)
	return &resp, nil
}

// Update updates a device
func (s *Service) Update(ctx context.Context, userID, deviceID uuid.UUID, req UpdateRequest) (*Response, error) {
	// Verify ownership
	device, err := s.repo.GetByID(ctx, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.Wrap(apperr.ErrNotFound, "device not found")
		}
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to get device")
	}

	if device.UserID != userID {
		return nil, apperr.Wrap(apperr.ErrForbidden, "device not owned by user")
	}

	// Update device
	updated, err := s.repo.Update(ctx, deviceID, req.DeviceName, req.HasCamera, req.HasMicrophone)
	if err != nil {
		return nil, apperr.Wrap(apperr.ErrInternal, "failed to update device")
	}

	resp := toResponse(updated)
	return &resp, nil
}

// UpdateStatus updates the online status of a device
func (s *Service) UpdateStatus(ctx context.Context, userID, deviceID uuid.UUID, isOnline bool) error {
	// Verify ownership
	device, err := s.repo.GetByID(ctx, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperr.Wrap(apperr.ErrNotFound, "device not found")
		}
		return apperr.Wrap(apperr.ErrInternal, "failed to get device")
	}

	if device.UserID != userID {
		return apperr.Wrap(apperr.ErrForbidden, "device not owned by user")
	}

	return s.repo.UpdateOnlineStatus(ctx, deviceID, isOnline)
}

// Heartbeat updates the last seen timestamp
func (s *Service) Heartbeat(ctx context.Context, userID, deviceID uuid.UUID) error {
	// Verify ownership
	device, err := s.repo.GetByID(ctx, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperr.Wrap(apperr.ErrNotFound, "device not found")
		}
		return apperr.Wrap(apperr.ErrInternal, "failed to get device")
	}

	if device.UserID != userID {
		return apperr.Wrap(apperr.ErrForbidden, "device not owned by user")
	}

	return s.repo.UpdateLastSeen(ctx, deviceID)
}

// Delete removes a device
func (s *Service) Delete(ctx context.Context, userID, deviceID uuid.UUID) error {
	// Verify ownership
	device, err := s.repo.GetByID(ctx, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperr.Wrap(apperr.ErrNotFound, "device not found")
		}
		return apperr.Wrap(apperr.ErrInternal, "failed to get device")
	}

	if device.UserID != userID {
		return apperr.Wrap(apperr.ErrForbidden, "device not owned by user")
	}

	return s.repo.Delete(ctx, deviceID)
}

// Helpers

func toResponse(d sqlc.Device) Response {
	resp := Response{
		ID:         d.ID,
		DeviceID:   d.DeviceID,
		DeviceName: d.DeviceName,
		DeviceType: string(d.DeviceType),
		CreatedAt:  d.CreatedAt.Time,
	}
	if d.HasCamera.Valid {
		resp.HasCamera = d.HasCamera.Bool
	}
	if d.HasMicrophone.Valid {
		resp.HasMicrophone = d.HasMicrophone.Bool
	}
	if d.IsOnline.Valid {
		resp.IsOnline = d.IsOnline.Bool
	}
	if d.LastSeen.Valid {
		resp.LastSeen = d.LastSeen.Time
	}
	return resp
}

func toResponseList(devices []sqlc.Device) []Response {
	resp := make([]Response, len(devices))
	for i, d := range devices {
		resp[i] = toResponse(d)
	}
	return resp
}

func parseDeviceType(s string) (sqlc.DeviceType, error) {
	switch s {
	case "phone":
		return sqlc.DeviceTypePhone, nil
	case "tablet":
		return sqlc.DeviceTypeTablet, nil
	case "desktop":
		return sqlc.DeviceTypeDesktop, nil
	default:
		return "", apperr.ErrValidation
	}
}
